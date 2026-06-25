package common

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/streadway/amqp"
)

// CrossGatewayBroadcast 跨网关广播管理器
type CrossGatewayBroadcast struct {
	redisClient     *redis.Client
	rabbitMQConn    *amqp.Connection
	rabbitMQChannel *amqp.Channel
	gatewayID       string
	messageDedup    *MessageDeduplicator
	broadcastStats  *BroadcastMetrics
	mutex           sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
	useHTTPFallback bool // ★ 标记是否使用HTTP降级模式
}

// MessageDeduplicator 消息去重器
type MessageDeduplicator struct {
	seenMessages map[string]time.Time // messageID -> timestamp
	ttl          time.Duration        // 消息TTL
	maxSize      int                  // 最大缓存大小
	mutex        sync.RWMutex
}

// BroadcastMetrics 广播指标
type BroadcastMetrics struct {
	TotalBroadcasts        int64
	LocalBroadcasts        int64
	CrossGatewayBroadcasts int64
	DuplicatedMessages     int64
	FailedBroadcasts       int64
	AvgLatency             time.Duration
	mutex                  sync.RWMutex
}

// BroadcastMessage 广播消息结构
type BroadcastMessage struct {
	MessageID   string      `json:"message_id"`   // 唯一消息ID（用于去重）
	GatewayID   string      `json:"gateway_id"`   // 发送方网关ID
	MessageType uint16      `json:"message_type"` // 消息类型
	MapID       uint32      `json:"map_id"`       // 地图ID
	SenderID    uint64      `json:"sender_id"`    // 发送者ID
	Data        interface{} `json:"data"`         // 消息数据
	Timestamp   int64       `json:"timestamp"`    // 时间戳
	ViewRange   int         `json:"view_range"`   // 视野范围
	Priority    uint8       `json:"priority"`     // 优先级 (0-9, 0最高)
}

var crossGatewayBC *CrossGatewayBroadcast

// InitCrossGatewayBroadcast 初始化跨网关广播
func InitCrossGatewayBroadcast(gatewayID string, redisAddr, rabbitMQURL string) error {
	ctx, cancel := context.WithCancel(context.Background())

	cbc := &CrossGatewayBroadcast{
		gatewayID:      gatewayID,
		messageDedup:   NewMessageDeduplicator(5*time.Minute, 10000),
		broadcastStats: &BroadcastMetrics{},
		ctx:            ctx,
		cancel:         cancel,
	}

	var transportMode string // 记录实际使用的传输模式

	// 初始化Redis连接（用于Pub/Sub）
	if redisAddr != "" {
		cbc.redisClient = redis.NewClient(&redis.Options{
			Addr:     redisAddr,
			Password: "", // 无密码
			DB:       0,
			PoolSize: 50,
		})

		// 测试Redis连接
		_, err := cbc.redisClient.Ping(ctx).Result()
		if err != nil {
			log.Printf("⚠️ Redis连接失败，跨网关广播将降级: %v", err)
			cbc.redisClient = nil
		} else {
			log.Printf("✅ Redis Pub/Sub就绪: %s", redisAddr)
			go cbc.subscribeToRedisChannel()
			transportMode += "Redis+"
		}
	}

	// 初始化RabbitMQ连接
	if rabbitMQURL != "" {
		var err error
		cbc.rabbitMQConn, err = amqp.Dial(rabbitMQURL)
		if err != nil {
			log.Printf("⚠️ RabbitMQ连接失败，将降级为HTTP轮询: %v", err)
		} else {
			log.Printf("✅ RabbitMQ就绪")
			cbc.setupRabbitMQExchange()
			go cbc.consumeRabbitMQMessages()
			transportMode += "RabbitMQ"
		}
	}

	// ★ 确定最终传输模式
	if transportMode == "" {
		transportMode = "HTTP(降级)"
		cbc.useHTTPFallback = true // 标记使用HTTP降级模式
	} else {
		transportMode = strings.TrimSuffix(transportMode, "+")
	}

	crossGatewayBC = cbc
	log.Printf(`🌐 跨网关广播就绪: gatewayID=%s, 传输=%s`, gatewayID, transportMode)

	return nil
}

// NewMessageDeduplicator 创建消息去重器
func NewMessageDeduplicator(ttl time.Duration, maxSize int) *MessageDeduplicator {
	dedup := &MessageDeduplicator{
		seenMessages: make(map[string]time.Time),
		ttl:          ttl,
		maxSize:      maxSize,
	}

	// 启动清理goroutine
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			dedup.cleanup()
		}
	}()

	return dedup
}

// IsDuplicate 检查是否重复消息
func (d *MessageDeduplicator) IsDuplicate(messageID string) bool {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if _, seen := d.seenMessages[messageID]; seen {
		return true
	}

	d.seenMessages[messageID] = time.Now()
	return false
}

// cleanup 清理过期消息
func (d *MessageDeduplicator) cleanup() {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	now := time.Now()
	for msgID, ts := range d.seenMessages {
		if now.Sub(ts) > d.ttl {
			delete(d.seenMessages, msgID)
		}
	}

	// 如果超过最大大小，删除最旧的一半消息
	if len(d.seenMessages) > d.maxSize {
		count := 0
		for msgID := range d.seenMessages {
			if count >= len(d.seenMessages)/2 {
				break
			}
			delete(d.seenMessages, msgID)
			count++
		}
	}
}

// GenerateMessageID 生成唯一消息ID
func GenerateMessageID(gatewayID string, msgType uint16, senderID uint64, data interface{}) string {
	hashInput := fmt.Sprintf("%s-%d-%d-%v-%d",
		gatewayID, msgType, senderID, data, time.Now().UnixNano())
	hash := md5.Sum([]byte(hashInput))
	return fmt.Sprintf("%x", hash)[:16]
}

// BroadcastToAllGateways 广播到所有网关
func BroadcastToAllGateways(msg *BroadcastMessage) error {
	if crossGatewayBC == nil {
		return fmt.Errorf("跨网关广播未初始化")
	}

	startTime := time.Now()

	// 生成唯一消息ID（如果还没有）
	if msg.MessageID == "" {
		msg.MessageID = GenerateMessageID(
			crossGatewayBC.gatewayID,
			msg.MessageType,
			msg.SenderID,
			msg.Data,
		)
		msg.GatewayID = crossGatewayBC.gatewayID
		msg.Timestamp = time.Now().Unix()
	}

	// 本地广播统计
	crossGatewayBC.broadcastStats.mutex.Lock()
	crossGatewayBC.broadcastStats.TotalBroadcasts++
	crossGatewayBC.broadcastStats.LocalBroadcasts++
	crossGatewayBC.broadcastStats.mutex.Unlock()

	// 通过Redis Pub/Sub发布
	if crossGatewayBC.redisClient != nil {
		if err := crossGatewayBC.publishToRedis(msg); err != nil {
			log.Printf(`❌ Redis发布失败: %v`, err)
			crossGatewayBC.recordFailedBroadcast()
		} else {
			crossGatewayBC.recordCrossGatewayBroadcast()
		}
	}

	// 通过RabbitMQ Fanout发布
	if crossGatewayBC.rabbitMQConn != nil {
		if err := crossGatewayBC.publishToRabbitMQ(msg); err != nil {
			log.Printf(`❌ RabbitMQ发布失败: %v`, err)
			crossGatewayBC.recordFailedBroadcast()
		} else {
			crossGatewayBC.recordCrossGatewayBroadcast()
		}
	}

	duration := time.Since(startTime)
	crossGatewayBC.updateAvgLatency(duration)

	log.Printf(`📢 跨网关广播完成: type=%d, sender=%d, latency=%v`,
		msg.MessageType, msg.SenderID, duration)

	return nil
}

// publishToRedis 发布到Redis Pub/Sub
func (cbc *CrossGatewayBroadcast) publishToRedis(msg *BroadcastMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %v", err)
	}

	channelName := fmt.Sprintf("gateway_broadcast:%d", msg.MapID)

	err = cbc.redisClient.Publish(cbc.ctx, channelName, data).Err()
	if err != nil {
		return fmt.Errorf("Redis Publish失败: %v", err)
	}

	log.Printf(`📤 已发布到Redis频道: %s`, channelName)
	return nil
}

// subscribeToRedisChannel 订阅Redis频道
func (cbc *CrossGatewayBroadcast) subscribeToRedisChannel() {
	pubsub := cbc.redisClient.Subscribe(cbc.ctx, "gateway_broadcast:*")
	defer pubsub.Close()

	ch := pubsub.Channel()

	for msg := range ch {
		go cbc.handleIncomingRedisMessage(msg.Payload)
	}
}

// handleIncomingRedisMessage 处理收到的Redis消息
func (cbc *CrossGatewayBroadcast) handleIncomingRedisMessage(payload string) {
	var broadcastMsg BroadcastMessage
	if err := json.Unmarshal([]byte(payload), &broadcastMsg); err != nil {
		log.Printf(`❌ 解析Redis消息失败: %v`, err)
		return
	}

	// 跳过自己发送的消息
	if broadcastMsg.GatewayID == cbc.gatewayID {
		return
	}

	// 消息去重
	if cbc.messageDedup.IsDuplicate(broadcastMsg.MessageID) {
		cbc.broadcastStats.mutex.Lock()
		cbc.broadcastStats.DuplicatedMessages++
		cbc.broadcastStats.mutex.Unlock()
		return
	}

	log.Printf(`📥 收到跨网关消息(Redis): type=%d, from=%s`,
		broadcastMsg.MessageType, broadcastMsg.GatewayID)

	// 处理消息（调用回调或直接转发给本地玩家）
	cbc.forwardToLocalPlayers(&broadcastMsg)
}

// setupRabbitMQExchange 设置RabbitMQ Exchange
func (cbc *CrossGatewayBroadcast) setupRabbitMQExchange() error {
	var err error
	cbc.rabbitMQChannel, err = cbc.rabbitMQConn.Channel()
	if err != nil {
		return fmt.Errorf("创建RabbitMQ Channel失败: %v", err)
	}

	// 声明Fanout Exchange（广播模式）
	err = cbc.rabbitMQChannel.ExchangeDeclare(
		"gateway_broadcast_fanout", // exchange名称
		"fanout",                   // 类型：扇出（广播给所有队列）
		true,                       // durable
		false,                      // autoDelete
		false,                      // internal
		false,                      // noWait
		nil,                        // arguments
	)
	if err != nil {
		return fmt.Errorf("声明Exchange失败: %v", err)
	}

	// 声明独占队列并绑定到Exchange
	queueName := fmt.Sprintf("gateway_%s_queue", cbc.gatewayID)
	_, err = cbc.rabbitMQChannel.QueueDeclare(
		queueName,
		false, // durable
		true,  // deleteWhenUnused
		true,  // exclusive
		false, // noWait
		nil,
	)
	if err != nil {
		return fmt.Errorf("声明Queue失败: %v", err)
	}

	err = cbc.rabbitMQChannel.QueueBind(
		queueName,
		"",                         // routingKey（fanout不需要）
		"gateway_broadcast_fanout", // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("绑定Queue失败: %v", err)
	}

	log.Printf(`✅ RabbitMQ Exchange设置完成: queue=%s`, queueName)
	return nil
}

// publishToRabbitMQ 发布到RabbitMQ
func (cbc *CrossGatewayBroadcast) publishToRabbitMQ(msg *BroadcastMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %v", err)
	}

	err = cbc.rabbitMQChannel.Publish(
		"gateway_broadcast_fanout", // exchange
		"",                         // routingKey
		false,                      // mandatory
		false,                      // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         data,
			DeliveryMode: amqp.Transient, // 非持久化（实时消息）
			Priority:     msg.Priority,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("RabbitMQ Publish失败: %v", err)
	}

	log.Printf(`📤 已发布到RabbitMQ`)
	return nil
}

// consumeRabbitMQMessages 消费RabbitMQ消息
func (cbc *CrossGatewayBroadcast) consumeRabbitMQMessages() {
	queueName := fmt.Sprintf("gateway_%s_queue", cbc.gatewayID)

	msgs, err := cbc.rabbitMQChannel.Consume(
		queueName,
		"",    // consumer
		true,  // autoAck
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,   // args
	)
	if err != nil {
		log.Printf(`❌ RabbitMQ消费失败: %v`, err)
		return
	}

	for msg := range msgs {
		go cbc.handleIncomingRabbitMQMessage(msg.Body)
	}
}

// handleIncomingRabbitMQMessage 处理收到的RabbitMQ消息
func (cbc *CrossGatewayBroadcast) handleIncomingRabbitMQMessage(body []byte) {
	var broadcastMsg BroadcastMessage
	if err := json.Unmarshal(body, &broadcastMsg); err != nil {
		log.Printf(`❌ 解析RabbitMQ消息失败: %v`, err)
		return
	}

	// 跳过自己发送的消息
	if broadcastMsg.GatewayID == cbc.gatewayID {
		return
	}

	// 消息去重
	if cbc.messageDedup.IsDuplicate(broadcastMsg.MessageID) {
		cbc.broadcastStats.mutex.Lock()
		cbc.broadcastStats.DuplicatedMessages++
		cbc.broadcastStats.mutex.Unlock()
		return
	}

	log.Printf(`📥 收到跨网关消息(RabbitMQ): type=%d, from=%s`,
		broadcastMsg.MessageType, broadcastMsg.GatewayID)

	// 处理消息
	cbc.forwardToLocalPlayers(&broadcastMsg)
}

// forwardToLocalPlayers 转发给本地玩家
func (cbc *CrossGatewayBroadcast) forwardToLocalPlayers(msg *BroadcastMessage) {
	// 这里需要与Gateway的PlayerManager集成
	// 将消息转发给该地图视野范围内的本地玩家

	// TODO: 实现具体的玩家查找和消息发送逻辑
	// 示例伪代码：
	/*
		players := GlobalManager.GetPlayersInMapAndRange(msg.MapID, msg.ViewRange)
		for _, player := range players {
			if player.ID != msg.SenderID { // 不发送给自己
				player.sendPacket(msg.MessageType, mustMarshal(msg.Data))
			}
		}
	*/
}

// recordCrossGatewayBroadcast 记录跨网关广播
func (cbc *CrossGatewayBroadcast) recordCrossGatewayBroadcast() {
	cbc.broadcastStats.mutex.Lock()
	defer cbc.broadcastStats.mutex.Unlock()
	cbc.broadcastStats.CrossGatewayBroadcasts++
}

// recordFailedBroadcast 记录失败的广播
func (cbc *CrossGatewayBroadcast) recordFailedBroadcast() {
	cbc.broadcastStats.mutex.Lock()
	defer cbc.broadcastStats.mutex.Unlock()
	cbc.broadcastStats.FailedBroadcasts++
}

// updateAvgLatency 更新平均延迟
func (cbc *CrossGatewayBroadcast) updateAvgLatency(duration time.Duration) {
	cbc.broadcastStats.mutex.Lock()
	defer cbc.broadcastStats.mutex.Unlock()

	// 简单的移动平均
	if cbc.broadcastStats.AvgLatency == 0 {
		cbc.broadcastStats.AvgLatency = duration
	} else {
		cbc.broadcastStats.AvgLatency = (cbc.broadcastStats.AvgLatency*9 + duration) / 10
	}
}

// GetBroadcastMetrics 获取广播指标
func GetBroadcastMetrics() *BroadcastMetrics {
	if crossGatewayBC == nil {
		return nil
	}
	return crossGatewayBC.broadcastStats
}

// Shutdown 关闭跨网关广播
func ShutdownCrossGatewayBroadcast() {
	if crossGatewayBC == nil {
		return
	}

	crossGatewayBC.cancel()

	if crossGatewayBC.redisClient != nil {
		crossGatewayBC.redisClient.Close()
	}

	if crossGatewayBC.rabbitMQConn != nil {
		crossGatewayBC.rabbitMQConn.Close()
	}

	log.Println("🔒 跨网关广播已关闭")
}
