package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	common "game-server/Common"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()
var redisClient *redis.Client

// VIEW_RANGE 视野范围（像素）- 从配置文件读取
var VIEW_RANGE int

// MOVE_MERGE_DELAY 移动消息合并延迟（毫秒）- 从配置文件读取
var MOVE_MERGE_DELAY int

// gatewayInstanceID 当前 Gateway 实例ID（用于跨实例广播时避免循环）
// 启动时生成，保证全局唯一
var gatewayInstanceID string

// initBroadcastConfig 初始化广播配置
func initBroadcastConfig() {
	_, _, _, viewRange, moveMergeDelay := common.AppConfig.GetGatewayConfig()
	VIEW_RANGE = viewRange
	MOVE_MERGE_DELAY = moveMergeDelay
	// 生成实例ID：进程PID + 时间戳，保证多实例唯一
	gatewayInstanceID = fmt.Sprintf("gw-%d-%d", os.Getpid(), time.Now().UnixNano())
	log.Printf("广播配置: ViewRange=%d, MoveMergeDelay=%dms, InstanceID=%s", VIEW_RANGE, MOVE_MERGE_DELAY, gatewayInstanceID)
}

// MoveMergeItem 待合并的移动消息
type MoveMergeItem struct {
	MapID  uint32
	FromID uint64
	X, Y   int
	Data   []byte
}

// MoveMerger 移动消息合并器
type MoveMerger struct {
	mu      sync.Mutex
	pending map[uint64]*MoveMergeItem // key=roleID
	timers  map[uint64]*time.Timer    // key=roleID
}

var moveMerger = &MoveMerger{
	pending: make(map[uint64]*MoveMergeItem),
	timers:  make(map[uint64]*time.Timer),
}

// InitRedis 初始化 Redis 客户端
func InitRedis(addr, password string, db int, poolSize, minIdle int) {
	if poolSize <= 0 {
		poolSize = 50
	}
	if minIdle <= 0 {
		minIdle = 5
	}

	redisClient = redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		PoolSize:     poolSize,
		MinIdleConns: minIdle,
	})

	// 测试连接
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Redis 连接失败: %v", err)
	}
	log.Printf("Redis 连接成功 (地址: %s, 连接池: PoolSize=%d, MinIdle=%d)", addr, poolSize, minIdle)

	// 启动 Redis 订阅监听
	go subscribeToMapChannels()
}

// subscribeToMapChannels 订阅所有地图频道
func subscribeToMapChannels() {
	pubsub := redisClient.PSubscribe(ctx, "map:*")
	defer pubsub.Close()

	log.Println("开始订阅地图频道...")

	for {
		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			log.Printf("Redis 订阅错误: %v", err)
			continue
		}

		// 解析消息
		var broadcastMsg BroadcastMessage
		if err := json.Unmarshal([]byte(msg.Payload), &broadcastMsg); err != nil {
			log.Printf("消息解析失败: %v", err)
			continue
		}

		// ★ 跳过自己实例发布的消息（避免循环广播）
		// 发布方已经处理过本地客户端，订阅方只处理其他实例的消息
		if broadcastMsg.SourceInstance == gatewayInstanceID {
			continue
		}

		// 根据消息类型处理
		switch broadcastMsg.Type {
		case "move":
			handleMoveBroadcast(broadcastMsg)
		case "enter":
			handleEnterBroadcast(broadcastMsg)
		case "leave":
			handleLeaveBroadcast(broadcastMsg)
		case "chat":
			handleChatBroadcast(broadcastMsg)
		case "map_msg":
			// ★ 通用地图消息（替代 BroadcastToMap 的跨实例广播）
			handleMapMsgBroadcast(broadcastMsg)
		}
	}
}

// BroadcastMessage Redis 广播消息结构
type BroadcastMessage struct {
	Type           string                 `json:"type"`
	MapID          uint32                 `json:"map_id"`
	FromID         uint64                 `json:"from_id"`
	FromX          int                    `json:"from_x"`
	FromY          int                    `json:"from_y"`
	SourceInstance string                 `json:"source_instance"` // ★ 来源 Gateway 实例ID
	MsgType        uint16                 `json:"msg_type"`        // ★ 通用消息类型（用于 map_msg）
	Data           map[string]interface{} `json:"data"`
}

// PublishToMap 通过 Redis 发布消息到地图频道
func PublishToMap(msg BroadcastMessage) error {
	// ★ 自动填充来源实例ID（调用方无需设置）
	msg.SourceInstance = gatewayInstanceID
	channel := fmt.Sprintf("map:%d", msg.MapID)
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return redisClient.Publish(ctx, channel, data).Err()
}

// handleMapMsgBroadcast 处理通用地图消息（跨实例 BroadcastToMap）
// 收到其他 Gateway 实例发布的消息，转发给本实例的同地图玩家
func handleMapMsgBroadcast(msg BroadcastMessage) {
	GlobalManager.mutex.RLock()

	// 封包：[cmd(2字节)][body(N字节)]
	data, _ := json.Marshal(msg.Data)
	pkg := make([]byte, 2+len(data))
	binary.LittleEndian.PutUint16(pkg[0:2], msg.MsgType)
	copy(pkg[2:], data)

	// ★ 使用分桶索引，O(M) 而非 O(N)
	bucket, ok := GlobalManager.clientsByMap[msg.MapID]
	if !ok {
		GlobalManager.mutex.RUnlock()
		return
	}

	var targets []*Client
	for _, client := range bucket {
		if client.ID != msg.FromID {
			targets = append(targets, client)
		}
	}

	GlobalManager.mutex.RUnlock()

	// 并发发送
	var wg sync.WaitGroup
	wg.Add(len(targets))
	for _, client := range targets {
		go func(c *Client, p []byte) {
			defer wg.Done()
			select {
			case c.Send <- p:
			default:
			}
		}(client, pkg)
	}
	wg.Wait()
}

// MergeAndBroadcastMove 合并并广播移动消息
func (cm *ClientManager) MergeAndBroadcastMove(mapID uint32, fromID uint64, x, y int, data []byte) {
	// 如果合并延迟为0，直接广播不合并
	if MOVE_MERGE_DELAY <= 0 {
		msg := &Message{
			From: fromID,
			Type: CmdMove,
			Data: data,
		}
		cm.BroadcastToViewRangeConcurrent(mapID, x, y, msg)
		return
	}

	moveMerger.mu.Lock()

	// 如果已有待合并的消息，更新位置
	if item, exists := moveMerger.pending[fromID]; exists {
		item.X = x
		item.Y = y
		item.Data = data
		moveMerger.mu.Unlock()
		return
	}

	// 创建新的待合并项
	item := &MoveMergeItem{
		MapID:  mapID,
		FromID: fromID,
		X:      x,
		Y:      y,
		Data:   data,
	}
	moveMerger.pending[fromID] = item

	// 设置定时器，延迟后发送合并后的消息
	timer := time.AfterFunc(time.Duration(MOVE_MERGE_DELAY)*time.Millisecond, func() {
		moveMerger.mu.Lock()
		defer moveMerger.mu.Unlock()

		if pendingItem, exists := moveMerger.pending[fromID]; exists {
			// 发送合并后的消息
			msg := &Message{
				From: fromID,
				Type: CmdMove,
				Data: pendingItem.Data,
			}
			cm.BroadcastToViewRangeConcurrent(pendingItem.MapID, pendingItem.X, pendingItem.Y, msg)

			// 清理
			delete(moveMerger.pending, fromID)
			delete(moveMerger.timers, fromID)
		}
	})

	moveMerger.timers[fromID] = timer
	moveMerger.mu.Unlock()
}

// BroadcastToViewRangeConcurrent 并发视野范围广播
func (cm *ClientManager) BroadcastToViewRangeConcurrent(mapID uint32, fromX, fromY int, msg *Message) {
	cm.mutex.RLock()

	// 封包
	pkg := make([]byte, 2+len(msg.Data))
	binary.LittleEndian.PutUint16(pkg[0:2], msg.Type)
	copy(pkg[2:], msg.Data)

	// 计算视野范围的平方（避免开方运算）
	viewRangeSq := VIEW_RANGE * VIEW_RANGE

	// ★ 使用分桶索引，O(M) 而非 O(N)（M=同地图人数）
	bucket, ok := cm.clientsByMap[mapID]
	if !ok {
		cm.mutex.RUnlock()
		return
	}

	// 收集视野范围内的目标玩家
	var targets []*Client
	for _, client := range bucket {
		if client.ID != msg.From {
			dx := client.X - fromX
			dy := client.Y - fromY
			distanceSq := dx*dx + dy*dy

			if distanceSq <= viewRangeSq {
				targets = append(targets, client)
			}
		}
	}

	cm.mutex.RUnlock()

	// 并发发送消息
	var wg sync.WaitGroup
	wg.Add(len(targets))

	for _, client := range targets {
		go func(c *Client, p []byte) {
			defer wg.Done()
			select {
			case c.Send <- p:
			default:
				log.Printf("玩家 %d 的发送通道已满", c.ID)
			}
		}(client, pkg)
	}

	wg.Wait()
}

// BroadcastToViewRange 视野范围广播（同步版本，用于快速消息）
func (cm *ClientManager) BroadcastToViewRange(mapID uint32, fromX, fromY int, msg *Message) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	// 封包
	pkg := make([]byte, 2+len(msg.Data))
	binary.LittleEndian.PutUint16(pkg[0:2], msg.Type)
	copy(pkg[2:], msg.Data)

	// 计算视野范围的平方（避免开方运算）
	viewRangeSq := VIEW_RANGE * VIEW_RANGE

	// ★ 使用分桶索引，O(M) 而非 O(N)
	bucket, ok := cm.clientsByMap[mapID]
	if !ok {
		return
	}
	for _, client := range bucket {
		if client.ID != msg.From {
			// 计算距离平方
			dx := client.X - fromX
			dy := client.Y - fromY
			distanceSq := dx*dx + dy*dy

			// 只发送给视野范围内的玩家
			if distanceSq <= viewRangeSq {
				select {
				case client.Send <- pkg:
				default:
					log.Printf("玩家 %d 的发送通道已满", client.ID)
				}
			}
		}
	}
}

// AsyncBroadcastToViewRange 异步视野范围广播（通过 Redis）
func (cm *ClientManager) AsyncBroadcastToViewRange(mapID uint32, fromX, fromY int, fromID uint64, msgType uint16, data []byte) {
	// 创建广播消息
	broadcastMsg := BroadcastMessage{
		Type:   getMessageTypeName(msgType),
		MapID:  mapID,
		FromID: fromID,
		FromX:  fromX,
		FromY:  fromY,
	}

	// 解析数据为 map
	var dataMap map[string]interface{}
	if err := json.Unmarshal(data, &dataMap); err == nil {
		broadcastMsg.Data = dataMap
	}

	// 发布到 Redis
	go func() {
		if err := PublishToMap(broadcastMsg); err != nil {
			log.Printf("Redis 发布失败: %v", err)
		}
	}()
}

// handleMoveBroadcast 处理移动广播（并发版本）
func handleMoveBroadcast(msg BroadcastMessage) {
	GlobalManager.mutex.RLock()

	// 封包
	data, _ := json.Marshal(msg.Data)
	pkg := make([]byte, 2+len(data))
	binary.LittleEndian.PutUint16(pkg[0:2], CmdMove)
	copy(pkg[2:], data)

	viewRangeSq := VIEW_RANGE * VIEW_RANGE

	// ★ 使用分桶索引，O(M) 而非 O(N)
	bucket, ok := GlobalManager.clientsByMap[msg.MapID]
	if !ok {
		GlobalManager.mutex.RUnlock()
		return
	}

	// 收集目标玩家
	var targets []*Client
	for _, client := range bucket {
		if client.ID != msg.FromID {
			dx := client.X - msg.FromX
			dy := client.Y - msg.FromY
			distanceSq := dx*dx + dy*dy

			if distanceSq <= viewRangeSq {
				targets = append(targets, client)
			}
		}
	}

	GlobalManager.mutex.RUnlock()

	// 并发发送
	var wg sync.WaitGroup
	wg.Add(len(targets))

	for _, client := range targets {
		go func(c *Client, p []byte) {
			defer wg.Done()
			select {
			case c.Send <- p:
			default:
			}
		}(client, pkg)
	}

	wg.Wait()
}

// handleEnterBroadcast 处理进入广播（并发版本）
func handleEnterBroadcast(msg BroadcastMessage) {
	GlobalManager.mutex.RLock()

	data, _ := json.Marshal(msg.Data)
	pkg := make([]byte, 2+len(data))
	binary.LittleEndian.PutUint16(pkg[0:2], CmdEnterMap)
	copy(pkg[2:], data)

	viewRangeSq := VIEW_RANGE * VIEW_RANGE

	// ★ 使用分桶索引，O(M) 而非 O(N)
	bucket, ok := GlobalManager.clientsByMap[msg.MapID]
	if !ok {
		GlobalManager.mutex.RUnlock()
		return
	}

	var targets []*Client
	for _, client := range bucket {
		if client.ID != msg.FromID {
			dx := client.X - msg.FromX
			dy := client.Y - msg.FromY
			distanceSq := dx*dx + dy*dy

			if distanceSq <= viewRangeSq {
				targets = append(targets, client)
			}
		}
	}

	GlobalManager.mutex.RUnlock()

	// 并发发送
	var wg sync.WaitGroup
	wg.Add(len(targets))

	for _, client := range targets {
		go func(c *Client, p []byte) {
			defer wg.Done()
			select {
			case c.Send <- p:
			default:
			}
		}(client, pkg)
	}

	wg.Wait()
}

// handleLeaveBroadcast 处理离开广播（并发版本）
func handleLeaveBroadcast(msg BroadcastMessage) {
	GlobalManager.mutex.RLock()

	data, _ := json.Marshal(msg.Data)
	pkg := make([]byte, 2+len(data))
	binary.LittleEndian.PutUint16(pkg[0:2], CmdLeaveMap)
	copy(pkg[2:], data)

	// ★ 使用分桶索引，O(M) 而非 O(N)
	bucket, ok := GlobalManager.clientsByMap[msg.MapID]
	if !ok {
		GlobalManager.mutex.RUnlock()
		return
	}

	var targets []*Client
	for _, client := range bucket {
		if client.ID != msg.FromID {
			targets = append(targets, client)
		}
	}

	GlobalManager.mutex.RUnlock()

	// 并发发送
	var wg sync.WaitGroup
	wg.Add(len(targets))

	for _, client := range targets {
		go func(c *Client, p []byte) {
			defer wg.Done()
			select {
			case c.Send <- p:
			default:
			}
		}(client, pkg)
	}

	wg.Wait()
}

// handleChatBroadcast 处理聊天广播（并发版本）
func handleChatBroadcast(msg BroadcastMessage) {
	GlobalManager.mutex.RLock()

	data, _ := json.Marshal(msg.Data)
	pkg := make([]byte, 2+len(data))
	binary.LittleEndian.PutUint16(pkg[0:2], CmdChat)
	copy(pkg[2:], data)

	// ★ 使用分桶索引，O(M) 而非 O(N)
	bucket, ok := GlobalManager.clientsByMap[msg.MapID]
	if !ok {
		GlobalManager.mutex.RUnlock()
		return
	}

	var targets []*Client
	for _, client := range bucket {
		if client.ID != msg.FromID {
			targets = append(targets, client)
		}
	}

	GlobalManager.mutex.RUnlock()

	// 并发发送
	var wg sync.WaitGroup
	wg.Add(len(targets))

	for _, client := range targets {
		go func(c *Client, p []byte) {
			defer wg.Done()
			select {
			case c.Send <- p:
			default:
			}
		}(client, pkg)
	}

	wg.Wait()
}

// getMessageTypeName 获取消息类型名称
func getMessageTypeName(msgType uint16) string {
	switch msgType {
	case CmdMove:
		return "move"
	case CmdEnterMap:
		return "enter"
	case CmdLeaveMap:
		return "leave"
	case CmdChat:
		return "chat"
	default:
		return "unknown"
	}
}
