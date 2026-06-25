package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	common "game-server/Common"

	"github.com/gorilla/websocket"
)

// WebSocket配置(从配置文件读取)
var (
	upgrader             websocket.Upgrader
	sendChannelSize      int
	broadcastChannelSize int
	maxConnectionsPerIP  int
	pongTimeout          time.Duration
	pingInterval         time.Duration
)

// 初始化WebSocket配置
func initWebSocketConfig() {
	readBuf, writeBuf, _, pingInt, pongOut := common.AppConfig.GetWebSocketConfig()
	sendChan, broadcastChan, maxConnPerIP, _, _ := common.AppConfig.GetGatewayConfig()

	sendChannelSize = sendChan
	broadcastChannelSize = broadcastChan
	maxConnectionsPerIP = maxConnPerIP
	pingInterval = pingInt
	pongTimeout = pongOut

	upgrader = websocket.Upgrader{
		ReadBufferSize:  readBuf,
		WriteBufferSize: writeBuf,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	log.Printf("WebSocket配置: ReadBuf=%d, WriteBuf=%d, SendChan=%d, BroadcastChan=%d, MaxConnPerIP=%d, PingInterval=%v, PongTimeout=%v",
		readBuf, writeBuf, sendChan, broadcastChan, maxConnPerIP, pingInt, pongOut)
}

// Client 客户端连接
type Client struct {
	AccountID  uint64          // 账号ID（登录后设置）
	ID         uint64          // 角色ID（选择角色后设置）
	Name       string          // 角色名
	Conn       *websocket.Conn // WebSocket连接
	Send       chan []byte     // 发送通道
	GameSvc    string          // 所在游戏服务
	MapID      uint32          // 当前地图
	X          int             // X坐标
	Y          int             // Y坐标
	LastPing   time.Time       // 最后心跳时间
	Registered bool            // 是否已注册到管理器
	Token      string          // 登录token
}

// SessionData 会话数据（用于断线重连）
type SessionData struct {
	AccountID uint64    // 账号ID
	RoleID    uint64    // 角色ID
	Name      string    // 角色名
	MapID     uint32    // 当前地图
	X         int       // X坐标
	Y         int       // Y坐标
	Expire    time.Time // 过期时间
}

// generateToken 生成重连token
func generateToken(accountID, roleID uint64) string {
	return fmt.Sprintf("%d_%d_%d", accountID, roleID, time.Now().UnixNano())
}

// ClientManager 客户端管理器
type ClientManager struct {
	clients      map[uint64]*Client            // 客户端列表 key=roleID
	byConn       map[*websocket.Conn]*Client   // by connection
	clientsByMap map[uint32]map[uint64]*Client // 按 mapID 分桶索引（key=roleID），加速同地图查询
	register     chan *Client                  // 注册通道
	unregister   chan *Client                  // 注销通道
	broadcast    chan *Message                 // 广播通道
	mutex        sync.RWMutex
	onlineCount  int64 // 在线人数（原子计数器）
}

// Message 消息结构
type Message struct {
	From uint64
	To   uint64 // 0=广播
	Type uint16 // 消息类型
	Data []byte // 消息内容
}

// PacketID 消息命令ID
const (
	CmdLogin                 uint16 = 1001 // 登录
	CmdLogout                uint16 = 1002 // 登出
	CmdHeartbeat             uint16 = 1003 // 心跳
	CmdRegister              uint16 = 1004 // 注册
	CmdMove                  uint16 = 2001 // 移动
	CmdMoveBlocked           uint16 = 2002 // 移动被阻挡（新增）
	CmdAttack                uint16 = 2003 // 攻击（原2002，后移）
	CmdUseSkill              uint16 = 2004 // 使用技能（原2003，后移）
	CmdChat                  uint16 = 2005 // 聊天（原2004，后移）
	CmdPickup                uint16 = 2006 // 拾取（原2005，后移）
	CmdUseItem               uint16 = 2007 // 使用物品（原2006，后移）
	CmdEquip                 uint16 = 2008 // 装备（原2007，后移）
	CmdTrade                 uint16 = 2009 // 交易（原2008，后移）
	CmdDamage                uint16 = 2010 // 伤害（原2009，后移）
	CmdDeath                 uint16 = 2011 // 死亡（原2010，后移）
	CmdRespawn               uint16 = 2012 // 复活（原2011，后移）
	CmdLevelUp               uint16 = 2013 // 升级（原2012，后移）
	CmdBuff                  uint16 = 2014 // 增益（原2013，后移）
	CmdDeBuff                uint16 = 2015 // 减益（原2014，后移）
	CmdSetPKMode             uint16 = 2016 // 切换PK模式（原2015，后移）
	CmdPositionSync          uint16 = 2017 // 服务端位置同步确认（新增，原CmdSync改名避免冲突）
	CmdEnterMap              uint16 = 3001 // 进入地图
	CmdLeaveMap              uint16 = 3002 // 离开地图
	CmdMapPlayer             uint16 = 3003 // 地图玩家列表
	CmdOnlineCount           uint16 = 3004 // 在线人数广播
	CmdNpcTalk               uint16 = 3005 // NPC对话
	CmdNpcTrade              uint16 = 3006 // NPC交易
	CmdMapEvent              uint16 = 3007 // 地图事件
	CmdMonsterPositionUpdate uint16 = 3101 // 怪物位置同步
	CmdMonsterSpawn          uint16 = 3102 // 怪物生成
	CmdMonsterDeath          uint16 = 3103 // 怪物死亡
	CmdSkillLearn            uint16 = 4001 // 学习武学
	CmdSkillUpgrade          uint16 = 4002 // 升级武学
	CmdSkillList             uint16 = 4003 // 获取技能列表
	CmdSkillEquip            uint16 = 4004 // 装备技能
	CmdSkillUnequip          uint16 = 4005 // 卸下技能
	CmdSkillExp              uint16 = 4006 // 技能熟练度
	CmdItemList              uint16 = 7001 // 背包列表
	CmdEquipList             uint16 = 7002 // 装备列表
	CmdRoleInfo              uint16 = 5001 // 角色信息
	CmdRoleAttrib            uint16 = 5002 // 角色属性
	CmdSync                  uint16 = 5003 // 属性同步
	CmdRoleList              uint16 = 5004 // 角色列表
	CmdRoleCreate            uint16 = 5005 // 创建角色
	CmdRoleSelect            uint16 = 5006 // 选择角色
	CmdQuestUpdate           uint16 = 6001 // 任务更新推送
	CmdQuestList             uint16 = 6101 // 任务列表
	CmdQuestAccept           uint16 = 6102 // 接取任务
	CmdQuestComplete         uint16 = 6103 // 完成任务
	CmdQuestAbandon          uint16 = 6104 // 放弃任务
	CmdAchievementList       uint16 = 6201 // 成就列表
	CmdAchievementStats      uint16 = 6202 // 成就统计
)

// NewClientManager 创建客户端管理器
func NewClientManager() *ClientManager {
	return &ClientManager{
		clients:      make(map[uint64]*Client),
		byConn:       make(map[*websocket.Conn]*Client),
		clientsByMap: make(map[uint32]map[uint64]*Client),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		broadcast:    make(chan *Message, broadcastChannelSize),
	}
}

// Start 启动管理器
func (cm *ClientManager) Start() {
	for {
		select {
		case client := <-cm.register:
			log.Printf("客户端连接: roleID=%d, addr=%s", client.ID, client.Conn.RemoteAddr())
			// 不立即注册，等待用户登录并进入地图后再注册
			// 这样可以确保在线人数统计只包含已登录的玩家

		case client := <-cm.unregister:
			cm.removeClient(client)
			log.Printf("客户端断开: roleID=%d, 最后位置=(%d,%d)", client.ID, client.X, client.Y)

			// 通知GameService保存玩家最后位置并离开地图
			go func(c *Client) {
				roleID := c.ID
				mapID := c.MapID
				if mapID > 0 {
					// 获取处理该地图的GameService实例
					instance := common.GetInstanceByMapID(mapID)
					if instance == nil {
						log.Printf("未找到处理地图 %d 的GameService实例", mapID)
						return
					}
					gatewayURL := instance.URL

					// ★ 关键修复：发送玩家的最后位置坐标
					body, _ := json.Marshal(map[string]interface{}{
						"role_id": roleID,
						"map_id":  mapID,
						"x":       c.X, // ← 新增：网关保存的最后一次移动坐标
						"y":       c.Y, // ← 新增：网关保存的最后一次移动坐标
					})
					resp, err := http.Post(gatewayURL+"/api/map/leave", "application/json", bytes.NewReader(body))
					if err == nil {
						resp.Body.Close()
						log.Printf("💾 通知GameService保存最后位置: roleID=%d, mapID=%d, pos=(%d,%d)", roleID, mapID, c.X, c.Y)
					} else {
						log.Printf("❌ 保存位置失败: roleID=%d, error=%v", roleID, err)
					}
				}
			}(client)

			// 广播玩家离开
			cm.BroadcastToMap(client.MapID, &Message{
				From: client.ID,
				Type: CmdLeaveMap,
				Data: mustMarshal(map[string]interface{}{
					"role_id": client.ID,
				}),
			})

		case msg := <-cm.broadcast:
			cm.sendMessage(msg)
		}
	}
}

func (cm *ClientManager) addClient(client *Client) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.clients[client.ID] = client
	cm.byConn[client.Conn] = client
	// 维护按 mapID 分桶索引
	if client.MapID > 0 {
		bucket, ok := cm.clientsByMap[client.MapID]
		if !ok {
			bucket = make(map[uint64]*Client)
			cm.clientsByMap[client.MapID] = bucket
		}
		bucket[client.ID] = client
	}
	client.Registered = true            // 只有成功添加到列表后才标记为已注册
	atomic.AddInt64(&cm.onlineCount, 1) // 原子增加在线人数
}

func (cm *ClientManager) removeClient(client *Client) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if _, ok := cm.clients[client.ID]; ok {
		log.Printf("removeClient: 移除玩家 %d, 当前地图=%d", client.ID, client.MapID)
		delete(cm.clients, client.ID)
		delete(cm.byConn, client.Conn)
		// 从 mapID 分桶索引中移除
		if client.MapID > 0 {
			if bucket, ok := cm.clientsByMap[client.MapID]; ok {
				delete(bucket, client.ID)
				// 如果桶为空，删除整个桶避免内存泄漏
				if len(bucket) == 0 {
					delete(cm.clientsByMap, client.MapID)
				}
			}
		}
		close(client.Send)
		atomic.AddInt64(&cm.onlineCount, -1) // 原子减少在线人数
	} else {
		log.Printf("removeClient: 玩家 %d 不存在于客户端列表", client.ID)
	}
}

// updateClientMap 更新客户端所在地图（维护分桶索引）
// 必须在持有 cm.mutex 写锁的情况下调用
func (cm *ClientManager) updateClientMap(client *Client, newMapID uint32) {
	oldMapID := client.MapID
	if oldMapID == newMapID {
		return
	}
	// 从旧桶移除
	if oldMapID > 0 {
		if bucket, ok := cm.clientsByMap[oldMapID]; ok {
			delete(bucket, client.ID)
			if len(bucket) == 0 {
				delete(cm.clientsByMap, oldMapID)
			}
		}
	}
	// 加入新桶
	client.MapID = newMapID
	if newMapID > 0 {
		bucket, ok := cm.clientsByMap[newMapID]
		if !ok {
			bucket = make(map[uint64]*Client)
			cm.clientsByMap[newMapID] = bucket
		}
		bucket[client.ID] = client
	}
}

func (cm *ClientManager) sendMessage(msg *Message) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	// 封包: [cmd(2字节)][body(N字节)]
	pkg := make([]byte, 2+len(msg.Data))
	binary.LittleEndian.PutUint16(pkg[0:2], msg.Type)
	copy(pkg[2:], msg.Data)

	if msg.To > 0 {
		// 私聊
		if client, ok := cm.clients[msg.To]; ok {
			select {
			case client.Send <- pkg:
			default:
				log.Printf("警告: 客户端 %d 发送通道已满，丢弃消息", msg.To)
			}
		}
	} else {
		// 广播 - 发送给所有连接的客户端（包括发送者自己）
		// 注：战斗广播等场景需要发送者也收到消息以更新UI（伤害飘字、MP等）
		for _, client := range cm.clients {
			select {
			case client.Send <- pkg:
			default:
				log.Printf("警告: 客户端 %d 发送通道已满，丢弃消息", client.ID)
			}
		}
	}
}

// BroadcastToMap 广播到地图
// ★ 优化：本地发送 + 发布到 Redis（支持多 Gateway 实例跨实例广播）
// 本地客户端由本函数直接发送，其他 Gateway 实例的客户端通过 Redis Pub/Sub 接收
func (cm *ClientManager) BroadcastToMap(mapID uint32, msg *Message) {
	cm.mutex.RLock()

	// 封包
	pkg := make([]byte, 2+len(msg.Data))
	binary.LittleEndian.PutUint16(pkg[0:2], msg.Type)
	copy(pkg[2:], msg.Data)

	// ★ 使用分桶索引，O(M) 而非 O(N)（M=同地图人数）
	bucket, ok := cm.clientsByMap[mapID]
	if ok {
		for _, client := range bucket {
			if client.ID != msg.From {
				select {
				case client.Send <- pkg:
				default:
				}
			}
		}
	}

	cm.mutex.RUnlock()

	// ★ 发布到 Redis，让其他 Gateway 实例也广播给它们的同地图客户端
	// 只有 Redis 可用时才发布，避免单实例模式下浪费资源
	if redisClient != nil {
		// 解析 msg.Data 为 map（Redis 消息格式要求）
		var dataMap map[string]interface{}
		if err := json.Unmarshal(msg.Data, &dataMap); err == nil {
			broadcastMsg := BroadcastMessage{
				Type:    "map_msg",
				MapID:   mapID,
				FromID:  msg.From,
				MsgType: msg.Type,
				Data:    dataMap,
			}
			go func() {
				if err := PublishToMap(broadcastMsg); err != nil {
					log.Printf("Redis 跨实例广播失败: %v", err)
				}
			}()
		}
	}
}

// BroadcastToAll 广播到所有玩家（使用非阻塞发送，但记录丢弃的消息）
func (cm *ClientManager) BroadcastToAll(msg *Message) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	pkg := make([]byte, 2+len(msg.Data))
	binary.LittleEndian.PutUint16(pkg[0:2], msg.Type)
	copy(pkg[2:], msg.Data)

	droppedCount := 0
	for _, client := range cm.clients {
		select {
		case client.Send <- pkg:
		default:
			droppedCount++
		}
	}

	if droppedCount > 0 {
		log.Printf("BroadcastToAll: 丢弃了 %d 条消息", droppedCount)
	}
}

// SendToClient 向指定客户端发送消息
func (cm *ClientManager) SendToClient(roleID uint64, msg *Message) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	client, ok := cm.clients[roleID]
	if !ok {
		return
	}

	// 封包
	pkg := make([]byte, 2+len(msg.Data))
	binary.LittleEndian.PutUint16(pkg[0:2], msg.Type)
	copy(pkg[2:], msg.Data)

	select {
	case client.Send <- pkg:
	default:
	}
}

// GetMapPlayers 获取指定地图的所有玩家
func (cm *ClientManager) GetMapPlayers(mapID uint32) []*Client {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	// ★ 使用分桶索引，O(M) 而非 O(N)
	bucket, ok := cm.clientsByMap[mapID]
	if !ok {
		return nil
	}
	players := make([]*Client, 0, len(bucket))
	for _, client := range bucket {
		players = append(players, client)
	}
	return players
}

// GetClient 获取客户端
func (cm *ClientManager) GetClient(roleID uint64) (*Client, bool) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	client, ok := cm.clients[roleID]
	return client, ok
}

// updateClientID 更新客户端ID（登录后）
func (cm *ClientManager) updateClientID(client *Client, newID uint64) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// 从旧ID删除
	delete(cm.clients, client.ID)
	// 设置新ID
	client.ID = newID
	// 添加新ID
	cm.clients[newID] = client
}

// GetClientsByMap 获取地图上的客户端
func (cm *ClientManager) GetClientsByMap(mapID uint32) []*Client {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	// ★ 使用分桶索引，O(M) 而非 O(N)
	bucket, ok := cm.clientsByMap[mapID]
	if !ok {
		return nil
	}
	result := make([]*Client, 0, len(bucket))
	for _, client := range bucket {
		result = append(result, client)
	}
	return result
}

// GetOnlineCount 获取在线人数（原子操作，无锁）
func (cm *ClientManager) GetOnlineCount() int64 {
	return atomic.LoadInt64(&cm.onlineCount)
}

// ClientAuth 客户端认证
type ClientAuth struct {
	Token  string `json:"token"`
	RoleID uint64 `json:"role_id"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Type     string `json:"type"` // "guest", "account", "reconnect"
	Username string `json:"username"`
	Password string `json:"password"`
	Token    string `json:"token"` // 重连时使用的token
}

// LoginResponse 登录响应
type LoginResponse struct {
	Code      int    `json:"code"`
	Msg       string `json:"msg"`
	AccountID uint64 `json:"account_id,omitempty"`
	RoleID    uint64 `json:"role_id,omitempty"`
	Token     string `json:"token,omitempty"`
}

// HandleWebSocket 处理WebSocket连接
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket升级失败: %v", err)
		return
	}

	client := &Client{
		Conn:       conn,
		Send:       make(chan []byte, sendChannelSize),
		LastPing:   time.Now(),
		Registered: false, // 延迟注册，等登录并进入地图后再注册
	}

	// 启动写循环
	go client.writeLoop()
	// 启动读循环
	go client.readLoop()
}

// writeLoop 写循环
func (c *Client) writeLoop() {
	defer func() {
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteMessage(websocket.BinaryMessage, message); err != nil {
				return
			}
		}
	}
}

// readLoop 读循环
func (c *Client) readLoop() {
	defer func() {
		// 只有已注册的客户端才注销
		if c.Registered {
			GlobalManager.unregister <- c
		}
		c.Conn.Close()
	}()

	for {
		_, data, err := c.Conn.ReadMessage()
		if err != nil {
			return
		}

		// 简单协议: 前2字节是命令,后面是数据
		if len(data) < 2 {
			continue
		}

		cmd := binary.LittleEndian.Uint16(data[0:2])
		body := data[2:]
		c.handleMessage(cmd, body)
	}
}

func (c *Client) handleMessage(cmd uint16, body []byte) {
	log.Printf("收到消息: cmd=%d, bodyLen=%d, accountID=%d, roleID=%d", cmd, len(body), c.AccountID, c.ID)

	// 未登录玩家（AccountID=0）只能发送登录、注册、心跳消息
	if c.AccountID == 0 && cmd != CmdLogin && cmd != CmdRegister && cmd != CmdHeartbeat {
		log.Printf("拒绝未登录玩家的消息: cmd=%d", cmd)
		return
	}

	// 未选择角色的玩家（ID=0）只能发送角色相关消息
	if c.ID == 0 && c.AccountID > 0 && cmd != CmdRoleList && cmd != CmdRoleCreate && cmd != CmdRoleSelect && cmd != CmdLogout && cmd != CmdHeartbeat {
		log.Printf("拒绝未选择角色的消息: cmd=%d", cmd)
		return
	}

	switch cmd {
	case CmdHeartbeat:
		c.LastPing = time.Now()
		c.sendPacket(CmdHeartbeat, []byte(`{"time":`+fmt.Sprintf("%d", time.Now().Unix())+`}`))
		// 刷新会话过期时间
		if c.Token != "" {
			if err := refreshSession(c.Token); err != nil {
				log.Printf("刷新会话失败: %v", err)
			}
		}

	case CmdRegister:
		c.handleRegister(body)

	case CmdLogin:
		c.handleLogin(body)

	case CmdLogout:
		c.handleLogout()

	case CmdRoleList:
		c.handleRoleList()

	case CmdRoleCreate:
		c.handleRoleCreate(body)

	case CmdRoleSelect:
		c.handleRoleSelect(body)

	case CmdMove:
		c.handleMove(body)

	case CmdAttack:
		c.handleAttack(body)

	case CmdUseSkill:
		// 使用技能：复用攻击流程，前端需在body中带target_id和skill_id
		c.handleAttack(body)

	case CmdSetPKMode:
		c.handleSetPKMode(body)

	case CmdChat:
		c.handleChat(body)

	case CmdEnterMap:
		c.handleEnterMap(body)

	case CmdDamage:
		c.handleDamage(body)

	case CmdDeath:
		c.handleDeath(body)

	case CmdRespawn:
		c.handleRespawn(body)

	case CmdLevelUp:
		c.handleLevelUp(body)

	case CmdBuff:
		c.handleBuff(body)

	case CmdDeBuff:
		c.handleDeBuff(body)

	case CmdMapEvent:
		c.handleMapEvent(body)

	// ========== 任务相关 ==========
	case CmdQuestList:
		c.handleQuestList(body)

	case CmdQuestAccept:
		c.handleQuestAccept(body)

	case CmdQuestComplete:
		c.handleQuestComplete(body)

	case CmdQuestAbandon:
		c.handleQuestAbandon(body)

	// ========== 成就相关 ==========
	case CmdAchievementList:
		c.handleAchievementList(body)

	case CmdAchievementStats:
		c.handleAchievementStats(body)

	// ========== 技能相关 ==========
	case CmdSkillList:
		c.handleSkillList(body)

	case CmdSkillLearn:
		c.handleSkillLearn(body)

	case CmdSkillUpgrade:
		c.handleSkillUpgrade(body)

	case CmdSkillEquip:
		c.handleSkillEquip(body)

	case CmdSkillUnequip:
		c.handleSkillUnequip(body)

	case CmdSkillExp:
		c.handleSkillExp(body)

	// ========== 物品/装备相关 ==========
	case CmdItemList:
		c.handleItemList(body)

	case CmdEquipList:
		c.handleEquipList(body)

	default:
		// 未知命令
	}
}

func (c *Client) handleLogin(body []byte) {
	log.Printf("处理登录请求: accountID=%d, roleID=%d, body=%s", c.AccountID, c.ID, string(body))
	var req LoginRequest
	if err := json.Unmarshal(body, &req); err != nil {
		c.sendPacket(CmdLogin, mustMarshal(LoginResponse{Code: 400, Msg: "请求格式错误"}))
		return
	}

	switch req.Type {
	case "account":
		// 调用LoginService验证账号密码
		loginReq := map[string]string{
			"username": req.Username,
			"password": req.Password,
		}
		resp, err := common.LoginClient.Post("/api/login", loginReq)
		if err != nil {
			log.Printf("调用LoginService失败: %v", err)
			c.sendPacket(CmdLogin, mustMarshal(LoginResponse{Code: 500, Msg: "服务异常，请稍后重试"}))
			return
		}

		var loginResp struct {
			Code  int    `json:"code"`
			UID   uint   `json:"uid"`
			Token string `json:"token"`
			Msg   string `json:"msg"`
		}
		if err := json.Unmarshal(resp, &loginResp); err != nil {
			c.sendPacket(CmdLogin, mustMarshal(LoginResponse{Code: 500, Msg: "响应解析失败"}))
			return
		}

		if loginResp.Code != 0 {
			c.sendPacket(CmdLogin, mustMarshal(LoginResponse{Code: loginResp.Code, Msg: loginResp.Msg}))
			return
		}

		// 登录成功，设置账号ID和token
		c.AccountID = uint64(loginResp.UID)
		c.Token = loginResp.Token

		log.Printf("账号登录成功: accountID=%d, token=%s", c.AccountID, c.Token[:min(10, len(c.Token))]+"...")

		// 返回登录成功响应（包含账号ID，等待前端选择角色）
		c.sendPacket(CmdLogin, mustMarshal(LoginResponse{
			Code:      200,
			Msg:       "登录成功",
			AccountID: c.AccountID,
			Token:     c.Token,
		}))

	case "guest":
		// 游客登录 - 创建临时账号
		guestID := time.Now().UnixNano() % 10000000 // 7位随机数
		guestUsername := fmt.Sprintf("g_%d", guestID)
		guestPassword := fmt.Sprintf("g_%d", time.Now().UnixNano()%10000000)

		// 调用LoginService注册临时账号
		registerReq := map[string]string{
			"username": guestUsername,
			"password": guestPassword,
		}
		resp, err := common.LoginClient.Post("/api/register", registerReq)
		if err != nil {
			log.Printf("游客注册失败: %v", err)
			// 如果注册失败，尝试直接登录（可能账号已存在）
			loginResp, err := common.LoginClient.Post("/api/login", registerReq)
			if err != nil {
				c.sendPacket(CmdLogin, mustMarshal(LoginResponse{Code: 500, Msg: "服务异常"}))
				return
			}
			resp = loginResp
		}

		var result struct {
			Code  int    `json:"code"`
			UID   uint   `json:"uid"`
			Token string `json:"token"`
			Msg   string `json:"msg"`
		}
		if err := json.Unmarshal(resp, &result); err != nil {
			c.sendPacket(CmdLogin, mustMarshal(LoginResponse{Code: 500, Msg: "响应解析失败"}))
			return
		}

		if result.Code != 0 {
			c.sendPacket(CmdLogin, mustMarshal(LoginResponse{Code: result.Code, Msg: result.Msg}))
			return
		}

		// 游客登录成功
		c.AccountID = uint64(result.UID)
		c.Token = result.Token
		c.Name = guestUsername

		log.Printf("游客登录成功: accountID=%d", c.AccountID)

		c.sendPacket(CmdLogin, mustMarshal(LoginResponse{
			Code:      200,
			Msg:       "游客登录成功",
			AccountID: c.AccountID,
			Token:     c.Token,
		}))

	case "reconnect":
		// 断线重连 - 验证token
		if req.Token == "" {
			c.sendPacket(CmdLogin, mustMarshal(LoginResponse{Code: 400, Msg: "token不能为空"}))
			return
		}

		// 获取会话数据
		session, ok := getSession(req.Token)
		if !ok {
			c.sendPacket(CmdLogin, mustMarshal(LoginResponse{Code: 401, Msg: "会话已过期，请重新登录"}))
			return
		}

		// 恢复会话
		c.AccountID = session.AccountID
		c.ID = session.RoleID
		c.Name = session.Name
		c.MapID = session.MapID
		c.X = session.X
		c.Y = session.Y

		// 注册到管理器
		GlobalManager.addClient(c)

		// 获取当前在线人数
		currentCount := GlobalManager.GetOnlineCount()
		log.Printf("重连成功: accountID=%d, roleID=%d, mapID=%d, 在线人数=%d", c.AccountID, c.ID, c.MapID, currentCount)

		// 发送登录成功响应
		c.sendPacket(CmdLogin, mustMarshal(LoginResponse{
			Code:      200,
			Msg:       "重连成功",
			AccountID: c.AccountID,
			RoleID:    session.RoleID,
			Token:     req.Token,
		}))

		// 发送在线人数
		pkg := mustMarshal(map[string]interface{}{"count": currentCount})
		sendPkg := make([]byte, 2+len(pkg))
		binary.LittleEndian.PutUint16(sendPkg[0:2], CmdOnlineCount)
		copy(sendPkg[2:], pkg)
		c.Send <- sendPkg

		// 广播在线人数更新
		GlobalManager.BroadcastToAll(&Message{
			Type: CmdOnlineCount,
			Data: mustMarshal(map[string]interface{}{"count": currentCount}),
		})

		// 广播玩家进入地图
		GlobalManager.BroadcastToMap(c.MapID, &Message{
			From: c.ID,
			Type: CmdEnterMap,
			Data: mustMarshal(map[string]interface{}{
				"role_id": c.ID,
				"name":    c.Name,
				"map_id":  c.MapID,
				"x":       c.X,
				"y":       c.Y,
			}),
		})

	default:
		c.sendPacket(CmdLogin, mustMarshal(LoginResponse{Code: 400, Msg: "无效的登录类型"}))
	}
}

func (c *Client) handleEnterMap(body []byte) {
	var req struct {
		RoleID uint64 `json:"role_id"`
		MapID  uint32 `json:"map_id"`
		X      int    `json:"x"`
		Y      int    `json:"y"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		log.Printf("handleEnterMap解析失败: %v", err)
		return
	}

	// 如果客户端还没有设置角色ID（可能是重连），从消息中获取
	if c.ID == 0 && req.RoleID != 0 {
		c.ID = req.RoleID
		log.Printf("从消息中恢复角色ID: %d", c.ID)

		// 尝试从会话中获取玩家名称
		session, ok := getSessionByRoleID(c.ID)
		if ok {
			c.Name = session.Name
			log.Printf("从会话中恢复玩家名称: %s", c.Name)
		}
	}

	log.Printf("🗺️ 玩家进入地图: roleID=%d, mapID=%d, 客户端坐标(%d,%d), 当前网关坐标(%d,%d)",
		c.ID, req.MapID, req.X, req.Y, c.X, c.Y)

	// ★ 关键修复1: 使用客户端坐标更新网关缓存（如果网关坐标为0或无效）
	if c.X == 0 && c.Y == 0 && (req.X != 0 || req.Y != 0) {
		c.X = req.X
		c.Y = req.Y
		log.Printf("✅ 使用客户端坐标更新网关缓存: (%d,%d)", c.X, c.Y)
	}

	// 如果是从其他地图切换，先广播离开旧地图
	if c.MapID != 0 && c.MapID != req.MapID {
		log.Printf("📍 玩家切换地图: %d -> %d", c.MapID, req.MapID)
		GlobalManager.BroadcastToMap(c.MapID, &Message{
			From: c.ID,
			Type: CmdLeaveMap,
			Data: mustMarshal(map[string]interface{}{
				"role_id": c.ID,
				"name":    c.Name,
			}),
		})
	}

	// 更新玩家地图（使用 updateClientMap 维护分桶索引）
	if c.Registered {
		GlobalManager.mutex.Lock()
		GlobalManager.updateClientMap(c, req.MapID)
		GlobalManager.mutex.Unlock()
	} else {
		c.MapID = req.MapID
	}

	// 如果还未注册，现在注册（确保同步完成）
	if !c.Registered {
		GlobalManager.addClient(c)
		log.Printf("✅ 客户端注册成功: roleID=%d, name=%s, mapID=%d, pos=(%d,%d)",
			c.ID, c.Name, c.MapID, c.X, c.Y)

		// 保存会话（用于断线重连）
		if c.Token != "" {
			if err := saveSession(c.Token, c.AccountID, c.ID, c.Name, c.MapID, c.X, c.Y); err != nil {
				log.Printf("保存会话失败: %v", err)
			}
		}

		// 向客户端发送当前位置同步
		log.Printf("📤 发送位置同步给玩家 %d: (%d,%d)", c.ID, c.X, c.Y)
		c.sendPacket(CmdPositionSync, mustMarshal(map[string]interface{}{
			"x": c.X,
			"y": c.Y,
		}))

		// 获取当前在线人数
		currentCount := GlobalManager.GetOnlineCount()
		log.Printf("👥 当前在线人数: %d", currentCount)

		// 发送在线人数给新玩家
		pkg := mustMarshal(map[string]interface{}{"count": currentCount})
		sendPkg := make([]byte, 2+len(pkg))
		binary.LittleEndian.PutUint16(sendPkg[0:2], CmdOnlineCount)
		copy(sendPkg[2:], pkg)
		c.Send <- sendPkg

		// 广播在线人数给所有玩家
		GlobalManager.BroadcastToAll(&Message{
			Type: CmdOnlineCount,
			Data: mustMarshal(map[string]interface{}{"count": currentCount}),
		})
	}

	// ★ 关键修复2: 先向新玩家发送当前地图的所有其他玩家列表（完整同步）
	existingPlayers := GlobalManager.GetMapPlayers(req.MapID)
	log.Printf("📍 当前地图%d共有%d个玩家", req.MapID, len(existingPlayers))

	if len(existingPlayers) > 1 { // 有其他玩家
		playerList := make([]map[string]interface{}, 0)
		validPlayerCount := 0

		for _, p := range existingPlayers {
			if p.ID != c.ID {
				// ★ 关键修复：直接使用玩家当前坐标（不再过滤(0,0)）
				// 原因：(0,0)可能是有效的出生点或初始位置
				// 如果坐标为(0,0)，优先使用客户端提供的坐标，其次使用会话缓存
				pX, pY := p.X, p.Y

				// 只添加有role_id的玩家（基本验证）
				if p.ID > 0 {
					playerList = append(playerList, map[string]interface{}{
						"role_id": p.ID,
						"name":    p.Name,
						"map_id":  p.MapID,
						"x":       pX,
						"y":       pY,
					})
					validPlayerCount++
					log.Printf("  👤 添加玩家: ID=%d, name=%s, pos=(%d,%d)", p.ID, p.Name, pX, pY)
				} else {
					log.Printf("  ⚠️ 跳过无效玩家（缺少ID）: name=%s", p.Name)
				}
			}
		}

		if len(playerList) > 0 {
			GlobalManager.SendToClient(c.ID, &Message{
				Type: CmdMapPlayer,
				Data: mustMarshal(map[string]interface{}{
					"players": playerList,
				}),
			})
			log.Printf("📤 发送地图玩家列表给玩家%d: 共%d个有效玩家", c.ID, validPlayerCount)
		}
	}

	// ★ 关键修复3: 使用视野范围广播新玩家进入（性能优化，只广播给附近玩家）
	log.Printf("📢 广播玩家进入（视野范围内）: mapID=%d, roleID=%d, name=%s, pos=(%d,%d)",
		c.MapID, c.ID, c.Name, c.X, c.Y)

	// 使用BroadcastToViewRangeConcurrent进行视野范围广播
	// 只向以新玩家为中心、VIEW_RANGE范围内的其他玩家广播进入消息
	GlobalManager.BroadcastToViewRangeConcurrent(c.MapID, c.X, c.Y, &Message{
		From: c.ID,
		Type: CmdEnterMap,
		Data: mustMarshal(map[string]interface{}{
			"role_id": c.ID,
			"name":    c.Name,
			"map_id":  c.MapID,
			"x":       c.X,
			"y":       c.Y,
		}),
	})
	log.Printf("📢 视野广播完成: mapID=%d, pos=(%d,%d), viewRange=%d", req.MapID, c.X, c.Y, VIEW_RANGE)

	// ★ 关键修复4: 延迟补充推送怪物列表和二次同步玩家（容错机制）
	if c.Registered {
		go func() {
			// 等待200ms确保所有异步操作完成
			time.Sleep(200 * time.Millisecond)

			// 补充推送怪物列表
			instance := common.GetInstanceByMapID(c.MapID)
			if instance != nil {
				client := &http.Client{Timeout: 3 * time.Second}
				resp, err := client.Get(instance.URL + fmt.Sprintf("/api/map/%d/monsters", c.MapID))
				if err == nil && resp.StatusCode == http.StatusOK {
					defer resp.Body.Close()
					var result map[string]interface{}
					if json.NewDecoder(resp.Body).Decode(&result) == nil {
						if monsters, ok := result["data"].([]interface{}); ok && len(monsters) > 0 {
							log.Printf("🔄 补充推送: 地图%d 有 %d 个怪物", c.MapID, len(monsters))
							GlobalManager.SendToClient(c.ID, &Message{
								Type: CmdSync,
								Data: mustMarshal(map[string]interface{}{
									"monster_list": monsters,
								}),
							})
						}
					}
				}
			}

			// 二次同步：再次检查是否有新玩家进入（防止时序问题导致的漏发）
			currentPlayers := GlobalManager.GetMapPlayers(c.MapID)
			if len(currentPlayers) > 1 {
				newPlayerList := make([]map[string]interface{}, 0)
				for _, p := range currentPlayers {
					if p.ID != c.ID && p.X != 0 && p.Y != 0 {
						newPlayerList = append(newPlayerList, map[string]interface{}{
							"role_id": p.ID,
							"name":    p.Name,
							"map_id":  p.MapID,
							"x":       p.X,
							"y":       p.Y,
						})
					}
				}
				if len(newPlayerList) > 0 {
					log.Printf("🔄 二次同步: 再次发送%d个玩家给玩家%d", len(newPlayerList), c.ID)
					GlobalManager.SendToClient(c.ID, &Message{
						Type: CmdMapPlayer,
						Data: mustMarshal(map[string]interface{}{
							"players": newPlayerList,
						}),
					})
				}
			}
		}()
	}
}

func (c *Client) handleLogout() {
	GlobalManager.unregister <- c
	c.sendPacket(CmdLogout, mustMarshal(map[string]interface{}{"code": 200, "msg": "登出成功"}))
}

func (c *Client) handleMove(body []byte) {
	// 解析移动数据
	var moveData struct {
		X int `json:"x"`
		Y int `json:"y"`
	}
	if err := json.Unmarshal(body, &moveData); err != nil {
		// 如果JSON解析失败，尝试二进制格式
		if len(body) < 4 {
			log.Printf("handleMove: 数据长度不足 %d", len(body))
			return
		}
		moveData.X = int(binary.LittleEndian.Uint16(body[0:2]))
		moveData.Y = int(binary.LittleEndian.Uint16(body[2:4]))
	}

	log.Printf("🚶 [Gateway] 收到移动请求: roleID=%d, mapID=%d, 目标(%d,%d), 当前(%d,%d)",
		c.ID, c.MapID, moveData.X, moveData.Y, c.X, c.Y)

	// ★ Prometheus：记录移动请求
	common.RecordMoveRequest(c.MapID, "received")

	// ★ 分布式架构：使用分布式路由器获取GameService实例（支持多实例+负载均衡）
	startTime := time.Now()

	var validationResult MoveValidationResult

	// 尝试使用分布式路由器
	instance, err := common.RouteByMapID(c.MapID)
	if err != nil {
		log.Printf("⚠️ [Gateway] 分布式路由失败，回退到默认路由: %v", err)
		common.IncrFallbackRouting()
		// 回退到旧的验证方式
		validationResult = c.validateMoveWithGameServiceLegacy(moveData.X, moveData.Y)
	} else {
		validationResult = c.validateMoveWithDistributedInstance(instance, moveData.X, moveData.Y)
	}

	duration := time.Since(startTime)

	// ★ Prometheus：记录验证耗时
	if instance != nil {
		common.RecordValidationDuration(instance.InstanceID, duration)
	}
	common.RecordRoutingDuration("map_id", duration)

	if !validationResult.Success {
		// ❌ 验证失败：通知客户端移动被阻挡
		log.Printf(`🛡️ [Gateway] 移动被阻挡: roleID=%d, 原因=%s`, c.ID, validationResult.Reason)

		// ★ Prometheus：记录失败的移动请求
		common.RecordMoveRequest(c.MapID, "blocked")

		c.sendPacket(CmdMoveBlocked, mustMarshal(map[string]interface{}{
			"code":    400,
			"msg":     validationResult.Reason,
			"x":       c.X, // 返回当前有效位置
			"y":       c.Y,
			"block_x": validationResult.BlockX,
			"block_y": validationResult.BlockY,
		}))
		return
	}

	// ✅ 验证通过：更新本地缓存位置
	c.X = moveData.X
	c.Y = moveData.Y

	log.Printf("✅ [Gateway] 移动验证通过: roleID=%d → (%d,%d)", c.ID, c.X, c.Y)

	// ★ Prometheus：记录成功的移动请求
	common.RecordMoveRequest(c.MapID, "success")

	// ★ 通知客户端移动成功（可选，用于确认）
	c.sendPacket(CmdSync, mustMarshal(map[string]interface{}{
		"x": c.X,
		"y": c.Y,
	}))

	// ★ 使用视野范围广播移动消息给同地图其他玩家（分布式支持）
	go func() {
		// 本地广播
		GlobalManager.BroadcastToViewRangeConcurrent(c.MapID, c.X, c.Y, &Message{
			From: c.ID,
			Type: CmdMove,
			Data: mustMarshal(map[string]interface{}{
				"role_id": c.ID,
				"map_id":  c.MapID,
				"x":       c.X,
				"y":       c.Y,
			}),
		})

		// ★ Prometheus：记录本地广播
		common.RecordBroadcastMessage("move", "local")

		log.Printf(`📢 [Gateway] 视野广播完成: mapID=%d, roleID=%d, pos=(%d,%d), viewRange=%d`,
			c.MapID, c.ID, c.X, c.Y, VIEW_RANGE)

		// ★ 跨网关广播（通过Redis Pub/Sub + RabbitMQ通知其他网关实例）
		broadcastMsg := &common.BroadcastMessage{
			MessageType: CmdMove,
			MapID:       c.MapID,
			SenderID:    c.ID,
			Data: map[string]interface{}{
				"role_id": c.ID,
				"map_id":  c.MapID,
				"x":       c.X,
				"y":       c.Y,
			},
			ViewRange: VIEW_RANGE,
			Priority:  3, // 高优先级（移动消息需要实时性）
		}

		if err := common.BroadcastToAllGateways(broadcastMsg); err != nil {
			log.Printf(`⚠️ [Gateway] 跨网关广播失败: %v`, err)
			common.IncrFailedBroadcasts()
		} else {
			common.RecordBroadcastMessage("move", "cross_gateway")
		}
	}()
}

// MoveValidationResult 移动验证结果
type MoveValidationResult struct {
	Success bool   // 是否验证通过
	Reason  string // 失败原因
	BlockX  int    // 阻挡位置X
	BlockY  int    // 阻挡位置Y
}

// validateMoveWithGameService 转发到GameService进行碰撞检测验证
func (c *Client) validateMoveWithGameService(x, y int) MoveValidationResult {
	// 获取处理该地图的GameService实例（分布式路由）
	instance := common.GetInstanceByMapID(c.MapID)
	if instance == nil {
		// 如果找不到GameService实例，允许本地移动（容错降级）
		log.Printf("⚠️ [Gateway] 未找到地图%d的GameService实例，跳过服务端验证", c.MapID)
		return MoveValidationResult{Success: true}
	}

	// 构建验证请求
	moveReq := map[string]interface{}{
		"role_id": c.ID,
		"map_id":  c.MapID,
		"x":       x,
		"y":       y,
	}

	jsonData, err := json.Marshal(moveReq)
	if err != nil {
		log.Printf("❌ [Gateway] 序列化移动数据失败: %v", err)
		return MoveValidationResult{Success: true} // 容错：序列化失败时允许移动
	}

	// 调用GameService的移动接口（带5秒超时）
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(instance.URL+"/api/map/move", "application/json", bytes.NewReader(jsonData))
	if err != nil {
		// GameService不可用时，允许本地移动（离线容错）
		log.Printf("⚠️ [Gateway] GameService不可用，允许本地移动: %v", err)
		return MoveValidationResult{Success: true}
	}
	defer resp.Body.Close()

	// 解析响应
	var result struct {
		Code    int    `json:"code"`
		Success bool   `json:"success"`
		Msg     string `json:"msg"`
		X       int    `json:"x"`
		Y       int    `json:"y"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("⚠️ [Gateway] 解析GameService响应失败，允许移动: %v", err)
		return MoveValidationResult{Success: true}
	}

	// 检查验证结果
	if resp.StatusCode != http.StatusOK || (!result.Success && result.Code != 200) {
		reason := result.Msg
		if reason == "" {
			reason = "移动路径被阻挡"
		}
		log.Printf(`🛡️ [Gateway] GameService拒绝移动: roleID=%d, 原因=%s`, c.ID, reason)
		return MoveValidationResult{
			Success: false,
			Reason:  reason,
			BlockX:  result.X,
			BlockY:  result.Y,
		}
	}

	log.Printf(`✅ [Gateway] GameService验证通过: roleID=%d`, c.ID)
	return MoveValidationResult{Success: true}
}

// validateMoveWithDistributedInstance 使用分布式路由器验证移动（新增）
func (c *Client) validateMoveWithDistributedInstance(instance *common.GameServiceInstance, x, y int) MoveValidationResult {
	if instance == nil {
		log.Printf("⚠️ [Gateway] 分布式实例为空，允许本地移动")
		return MoveValidationResult{Success: true}
	}

	log.Printf(`🔀 [Gateway] 使用分布式实例验证: instanceID=%d, url=%s`, instance.InstanceID, instance.URL)

	// 构建验证请求
	moveReq := map[string]interface{}{
		"role_id": c.ID,
		"map_id":  c.MapID,
		"x":       x,
		"y":       y,
	}

	jsonData, err := json.Marshal(moveReq)
	if err != nil {
		log.Printf("❌ [Gateway] 序列化移动数据失败: %v", err)
		return MoveValidationResult{Success: true}
	}

	// 调用GameService的移动接口（带5秒超时）
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(instance.URL+"/api/map/move", "application/json", bytes.NewReader(jsonData))
	if err != nil {
		// 记录失败到分布式路由器
		common.RecordFailure(instance.InstanceID)
		common.RecordCircuitBreakerTrip(instance.InstanceID, "connection_error")

		log.Printf("⚠️ [Gateway] GameService实例不可用，允许本地移动: %v", err)
		return MoveValidationResult{Success: true}
	}
	defer resp.Body.Close()

	// 记录成功
	common.RecordSuccess(instance.InstanceID)

	// 解析响应
	var result struct {
		Code    int    `json:"code"`
		Success bool   `json:"success"`
		Msg     string `json:"msg"`
		X       int    `json:"x"`
		Y       int    `json:"y"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("⚠️ [Gateway] 解析GameService响应失败，允许移动: %v", err)
		return MoveValidationResult{Success: true}
	}

	// 检查验证结果
	if resp.StatusCode != http.StatusOK || (!result.Success && result.Code != 200) {
		reason := result.Msg
		if reason == "" {
			reason = "移动路径被阻挡"
		}
		log.Printf(`🛡️ [Gateway] GameService拒绝移动: roleID=%d, 原因=%s`, c.ID, reason)

		// 记录被阻挡（不是服务端错误）
		common.RecordRoutingRequest("map_id", fmt.Sprintf("%d", instance.InstanceID), "blocked")

		return MoveValidationResult{
			Success: false,
			Reason:  reason,
			BlockX:  result.X,
			BlockY:  result.Y,
		}
	}

	// 记录成功的路由请求
	common.RecordRoutingRequest("map_id", fmt.Sprintf("%d", instance.InstanceID), "success")

	log.Printf(`✅ [Gateway] 分布式实例验证通过: instanceID=%d`, instance.InstanceID)
	return MoveValidationResult{Success: true}
}

// validateMoveWithGameServiceLegacy 旧的验证方式（作为降级方案）
func (c *Client) validateMoveWithGameServiceLegacy(x, y int) MoveValidationResult {
	// 获取处理该地图的GameService实例（使用旧的注册中心方式）
	instance := common.GetInstanceByMapID(c.MapID)
	if instance == nil {
		log.Printf("⚠️ [Gateway] 未找到地图%d的GameService实例，跳过服务端验证", c.MapID)
		return MoveValidationResult{Success: true}
	}

	// 构建验证请求
	moveReq := map[string]interface{}{
		"role_id": c.ID,
		"map_id":  c.MapID,
		"x":       x,
		"y":       y,
	}

	jsonData, err := json.Marshal(moveReq)
	if err != nil {
		log.Printf("❌ [Gateway] 序列化移动数据失败: %v", err)
		return MoveValidationResult{Success: true}
	}

	// 调用GameService的移动接口（带5秒超时）
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(instance.URL+"/api/map/move", "application/json", bytes.NewReader(jsonData))
	if err != nil {
		log.Printf("⚠️ [Gateway] GameService不可用，允许本地移动: %v", err)
		return MoveValidationResult{Success: true}
	}
	defer resp.Body.Close()

	// 解析响应
	var result struct {
		Code    int    `json:"code"`
		Success bool   `json:"success"`
		Msg     string `json:"msg"`
		X       int    `json:"x"`
		Y       int    `json:"y"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("⚠️ [Gateway] 解析GameService响应失败，允许移动: %v", err)
		return MoveValidationResult{Success: true}
	}

	// 检查验证结果
	if resp.StatusCode != http.StatusOK || (!result.Success && result.Code != 200) {
		reason := result.Msg
		if reason == "" {
			reason = "移动路径被阻挡"
		}
		return MoveValidationResult{
			Success: false,
			Reason:  reason,
			BlockX:  result.X,
			BlockY:  result.Y,
		}
	}

	return MoveValidationResult{Success: true}
}

// handleAttack 处理玩家攻击请求（转发到GameService战斗接口）
func (c *Client) handleAttack(body []byte) {
	var req struct {
		TargetID   uint64 `json:"target_id"`
		TargetType uint8  `json:"target_type"` // 2=怪物
		SkillID    uint32 `json:"skill_id"`    // 0=普通攻击
		X          int    `json:"x"`
		Y          int    `json:"y"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		log.Printf("解析攻击请求失败: %v", err)
		return
	}

	// 默认目标类型为怪物
	if req.TargetType == 0 {
		req.TargetType = 2
	}

	log.Printf("玩家攻击: roleID=%d, targetID=%d, targetType=%d, skillID=%d",
		c.ID, req.TargetID, req.TargetType, req.SkillID)

	// 转发到GameService处理（异步，避免阻塞WebSocket读取）
	go c.forwardAttackToGameService(req.TargetID, req.TargetType, req.SkillID, req.X, req.Y)
}

// forwardAttackToGameService 转发攻击请求到GameService
func (c *Client) forwardAttackToGameService(targetID uint64, targetType uint8, skillID uint32, x, y int) {
	instance := common.GetInstanceByMapID(c.MapID)
	if instance == nil {
		log.Printf("未找到处理地图 %d 的GameService实例", c.MapID)
		return
	}

	// 构建攻击请求（与GameService Battle.AttackRequest结构一致）
	attackReq := map[string]interface{}{
		"attacker_id":   c.ID,
		"attacker_type": 1, // 1=玩家
		"target_id":     targetID,
		"target_type":   targetType,
		"skill_id":      skillID,
		"x":             x,
		"y":             y,
	}

	jsonData, err := json.Marshal(attackReq)
	if err != nil {
		log.Printf("序列化攻击数据失败: %v", err)
		return
	}

	client := &http.Client{Timeout: 5 * time.Second}

	// 根据目标类型选择接口：1=玩家→PVP接口，2=怪物→普通攻击接口
	url := instance.URL + "/api/battle/attack"
	if targetType == 1 {
		url = instance.URL + "/api/battle/pvp"
	}
	resp, err := client.Post(url, "application/json", bytes.NewReader(jsonData))
	if err != nil {
		log.Printf("调用GameService攻击接口失败: %v", err)
		return
	}
	defer resp.Body.Close()

	// GameService会通过/internal/push接口异步推送attack_result给客户端
	// 这里不需要同步返回结果给客户端
}

// handleSetPKMode 处理切换PK模式请求
func (c *Client) handleSetPKMode(body []byte) {
	var req struct {
		PKMode uint8 `json:"pk_mode"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		log.Printf("解析PK模式切换数据失败: %v", err)
		return
	}

	if req.PKMode > 3 {
		c.sendPacket(CmdSetPKMode, mustMarshal(map[string]interface{}{
			"code": 1,
			"msg":  "无效的PK模式",
		}))
		return
	}

	// 调用DBService更新PK模式
	if err := common.DBRoleSetPKMode(c.ID, req.PKMode); err != nil {
		log.Printf("设置PK模式失败: roleID=%d, mode=%d, err=%v", c.ID, req.PKMode, err)
		c.sendPacket(CmdSetPKMode, mustMarshal(map[string]interface{}{
			"code": 2,
			"msg":  "设置失败",
		}))
		return
	}

	log.Printf("玩家[%d]切换PK模式为: %d", c.ID, req.PKMode)

	// 推送成功结果给客户端
	c.sendPacket(CmdSetPKMode, mustMarshal(map[string]interface{}{
		"code":    0,
		"msg":     "success",
		"pk_mode": req.PKMode,
	}))
}

func (c *Client) handleChat(body []byte) {
	var chat struct {
		Channel int    `json:"channel"` // 0=世界, 1=当前地图, 2=门派, 3=私聊
		Content string `json:"content"`
		ToID    uint64 `json:"to_id,omitempty"`
	}
	if err := json.Unmarshal(body, &chat); err != nil {
		return
	}

	msg := mustMarshal(map[string]interface{}{
		"from_id":   c.ID,
		"from_name": c.Name,
		"channel":   chat.Channel,
		"content":   chat.Content,
		"time":      time.Now().Unix(),
	})

	switch chat.Channel {
	case 0: // 世界
		// 先回显给自己
		c.sendPacket(CmdChat, msg)
		GlobalManager.broadcast <- &Message{Type: CmdChat, Data: msg}
	case 1: // 当前地图
		// 先回显给自己
		c.sendPacket(CmdChat, msg)
		GlobalManager.BroadcastToMap(c.MapID, &Message{From: c.ID, Type: CmdChat, Data: msg})
	case 3: // 私聊
		// 发送给目标玩家
		GlobalManager.broadcast <- &Message{To: chat.ToID, Type: CmdChat, Data: msg}
		// 回显给自己
		c.sendPacket(CmdChat, msg)
	}
}

// handleDamage 处理伤害消息（转发到GameService）
func (c *Client) handleDamage(body []byte) {
	var dmg struct {
		TargetID   uint64 `json:"target_id"`
		Damage     int    `json:"damage"`
		IsCritical bool   `json:"is_critical"`
		IsBlocked  bool   `json:"is_blocked"`
		IsDodged   bool   `json:"is_dodged"`
	}
	if err := json.Unmarshal(body, &dmg); err != nil {
		return
	}

	// 调用GameService处理伤害逻辑
	go func() {
		// 获取GameService实例
		instances := common.GetAllInstances()
		if len(instances) == 0 {
			log.Printf("没有可用的GameService实例")
			return
		}

		// 使用第一个可用的GameService实例
		gameServiceURL := instances[0].URL

		// 构建请求数据
		reqData := map[string]interface{}{
			"target_id":   dmg.TargetID,
			"damage":      dmg.Damage,
			"is_critical": dmg.IsCritical,
			"is_blocked":  dmg.IsBlocked,
			"is_dodged":   dmg.IsDodged,
		}

		jsonData, err := json.Marshal(reqData)
		if err != nil {
			log.Printf("序列化伤害请求失败: %v", err)
			return
		}

		// 调用GameService的HTTP接口
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Post(gameServiceURL+"/api/battle/damage", "application/json", bytes.NewReader(jsonData))
		if err != nil {
			log.Printf("调用GameService伤害接口失败: %v", err)
			return
		}
		defer resp.Body.Close()

		log.Printf("伤害处理完成，状态码: %d", resp.StatusCode)
	}()
}

// handleDeath 处理死亡消息（转发到GameService）
func (c *Client) handleDeath(body []byte) {
	var death struct {
		TargetID uint64 `json:"target_id"`
	}
	if err := json.Unmarshal(body, &death); err != nil {
		return
	}

	// 调用GameService处理死亡逻辑
	go func() {
		// 获取GameService实例
		instances := common.GetAllInstances()
		if len(instances) == 0 {
			log.Printf("没有可用的GameService实例")
			return
		}

		// 使用第一个可用的GameService实例
		gameServiceURL := instances[0].URL

		// 构建请求数据
		reqData := map[string]interface{}{
			"target_id": death.TargetID,
		}

		jsonData, err := json.Marshal(reqData)
		if err != nil {
			log.Printf("序列化死亡请求失败: %v", err)
			return
		}

		// 调用GameService的HTTP接口
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Post(gameServiceURL+"/api/battle/death", "application/json", bytes.NewReader(jsonData))
		if err != nil {
			log.Printf("调用GameService死亡接口失败: %v", err)
			return
		}
		defer resp.Body.Close()

		log.Printf("死亡处理完成，状态码: %d", resp.StatusCode)
	}()
}

// handleRespawn 处理复活请求（转发到GameService）
func (c *Client) handleRespawn(body []byte) {
	var req struct {
		Type   string `json:"type"` // "here"=原地复活, "town"=回城复活
		RoleID uint64 `json:"role_id"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		return
	}

	// 调用GameService处理复活逻辑
	go func() {
		// 获取GameService实例
		instances := common.GetAllInstances()
		if len(instances) == 0 {
			log.Printf("没有可用的GameService实例")
			return
		}

		// 使用第一个可用的GameService实例
		gameServiceURL := instances[0].URL

		// 构建请求数据
		reqData := map[string]interface{}{
			"type":    req.Type,
			"role_id": req.RoleID,
		}

		jsonData, err := json.Marshal(reqData)
		if err != nil {
			log.Printf("序列化复活请求失败: %v", err)
			return
		}

		// 调用GameService的HTTP接口
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Post(gameServiceURL+"/api/battle/respawn", "application/json", bytes.NewReader(jsonData))
		if err != nil {
			log.Printf("调用GameService复活接口失败: %v", err)
			return
		}
		defer resp.Body.Close()

		log.Printf("复活处理完成，状态码: %d", resp.StatusCode)
	}()
}

// handleLevelUp 处理升级消息（转发到GameService）
func (c *Client) handleLevelUp(body []byte) {
	var levelUp struct {
		TargetID uint64 `json:"target_id"`
		Level    int    `json:"level"`
		MaxHP    int    `json:"max_hp"`
		MaxMP    int    `json:"max_mp"`
		Attack   int    `json:"attack"`
		Defense  int    `json:"defense"`
		Speed    int    `json:"speed"`
	}
	if err := json.Unmarshal(body, &levelUp); err != nil {
		return
	}

	// 调用GameService处理升级逻辑
	go c.forwardLevelUpToGameService(levelUp.TargetID, levelUp.Level)
}

// forwardLevelUpToGameService 将升级请求转发到GameService
func (c *Client) forwardLevelUpToGameService(targetID uint64, level int) {
	instances := common.GetAllInstances()
	if len(instances) == 0 {
		log.Printf("没有可用的GameService实例")
		return
	}

	reqData := map[string]interface{}{
		"target_id": targetID,
		"level":     level,
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		log.Printf("序列化升级数据失败: %v", err)
		return
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(instances[0].URL+"/api/battle/level_up", "application/json", bytes.NewReader(jsonData))
	if err != nil {
		log.Printf("调用GameService升级接口失败: %v", err)
		return
	}
	defer resp.Body.Close()
}

// handleBuff 处理增益效果消息（转发到GameService）
func (c *Client) handleBuff(body []byte) {
	var buff struct {
		TargetID uint64 `json:"target_id"`
		BuffType string `json:"buff_type"` // attack, defense, speed, heal
	}
	if err := json.Unmarshal(body, &buff); err != nil {
		return
	}

	// 调用GameService处理增益逻辑
	go c.forwardBuffToGameService(buff.TargetID, buff.BuffType)
}

// forwardBuffToGameService 将增益请求转发到GameService
func (c *Client) forwardBuffToGameService(targetID uint64, buffType string) {
	instances := common.GetAllInstances()
	if len(instances) == 0 {
		log.Printf("没有可用的GameService实例")
		return
	}

	reqData := map[string]interface{}{
		"target_id": targetID,
		"buff_type": buffType,
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		log.Printf("序列化增益数据失败: %v", err)
		return
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(instances[0].URL+"/api/battle/buff", "application/json", bytes.NewReader(jsonData))
	if err != nil {
		log.Printf("调用GameService增益接口失败: %v", err)
		return
	}
	defer resp.Body.Close()
}

// handleDeBuff 处理减益效果消息（转发到GameService）
func (c *Client) handleDeBuff(body []byte) {
	var debuff struct {
		TargetID   uint64 `json:"target_id"`
		DeBuffType string `json:"debuff_type"` // poison, burn, freeze, stun, bleed, silence, fear
	}
	if err := json.Unmarshal(body, &debuff); err != nil {
		return
	}

	// 调用GameService处理减益逻辑
	go c.forwardDeBuffToGameService(debuff.TargetID, debuff.DeBuffType)
}

// forwardDeBuffToGameService 将减益请求转发到GameService
func (c *Client) forwardDeBuffToGameService(targetID uint64, debuffType string) {
	instances := common.GetAllInstances()
	if len(instances) == 0 {
		log.Printf("没有可用的GameService实例")
		return
	}

	reqData := map[string]interface{}{
		"target_id":   targetID,
		"debuff_type": debuffType,
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		log.Printf("序列化减益数据失败: %v", err)
		return
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(instances[0].URL+"/api/battle/debuff", "application/json", bytes.NewReader(jsonData))
	if err != nil {
		log.Printf("调用GameService减益接口失败: %v", err)
		return
	}
	defer resp.Body.Close()
}

// handleMapEvent 处理地图事件消息（转发到GameService）
func (c *Client) handleMapEvent(body []byte) {
	var event struct {
		EventType string `json:"event_type"` // spawn, end, portal_open, chest_open
		X         int    `json:"x"`
		Y         int    `json:"y"`
	}
	if err := json.Unmarshal(body, &event); err != nil {
		return
	}

	// 调用GameService处理地图事件逻辑
	go c.forwardMapEventToGameService(event.EventType, c.MapID, event.X, event.Y)
}

// forwardMapEventToGameService 将地图事件转发到GameService
func (c *Client) forwardMapEventToGameService(eventType string, mapID uint32, x, y int) {
	instance := common.GetInstanceByMapID(mapID)
	if instance == nil {
		log.Printf("没有找到处理地图 %d 的GameService实例", mapID)
		return
	}

	reqData := map[string]interface{}{
		"event_type": eventType,
		"map_id":     mapID,
		"x":          x,
		"y":          y,
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		log.Printf("序列化地图事件数据失败: %v", err)
		return
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(instance.URL+"/api/battle/map_event", "application/json", bytes.NewReader(jsonData))
	if err != nil {
		log.Printf("调用GameService地图事件接口失败: %v", err)
		return
	}
	defer resp.Body.Close()
}

// ==================== 任务相关 Handler ====================

// handleQuestList 处理获取任务列表请求
func (c *Client) handleQuestList(body []byte) {
	var req struct {
		RoleID uint64 `json:"role_id"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		c.sendPacket(CmdQuestList, mustMarshal(map[string]interface{}{"code": 400, "msg": "请求格式错误"}))
		return
	}

	go c.forwardQuestToGameService(req.RoleID, "/api/quest/list/"+fmt.Sprintf("%d", req.RoleID), "GET", nil, CmdQuestList)
}

// handleQuestAccept 处理接取任务请求
func (c *Client) handleQuestAccept(body []byte) {
	var req struct {
		RoleID  uint64 `json:"role_id"`
		QuestID uint32 `json:"quest_id"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		c.sendPacket(CmdQuestAccept, mustMarshal(map[string]interface{}{"code": 400, "msg": "请求格式错误"}))
		return
	}

	reqData := map[string]interface{}{
		"role_id":  req.RoleID,
		"quest_id": req.QuestID,
	}
	go c.forwardQuestToGameService(req.RoleID, "/api/quest/accept", "POST", reqData, CmdQuestAccept)
}

// handleQuestComplete 处理完成任务请求
func (c *Client) handleQuestComplete(body []byte) {
	var req struct {
		RoleID  uint64 `json:"role_id"`
		QuestID uint32 `json:"quest_id"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		c.sendPacket(CmdQuestComplete, mustMarshal(map[string]interface{}{"code": 400, "msg": "请求格式错误"}))
		return
	}

	reqData := map[string]interface{}{
		"role_id":  req.RoleID,
		"quest_id": req.QuestID,
	}
	go c.forwardQuestToGameService(req.RoleID, "/api/quest/complete", "POST", reqData, CmdQuestComplete)
}

// handleQuestAbandon 处理放弃任务请求
func (c *Client) handleQuestAbandon(body []byte) {
	var req struct {
		RoleID  uint64 `json:"role_id"`
		QuestID uint32 `json:"quest_id"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		c.sendPacket(CmdQuestAbandon, mustMarshal(map[string]interface{}{"code": 400, "msg": "请求格式错误"}))
		return
	}

	reqData := map[string]interface{}{
		"role_id":  req.RoleID,
		"quest_id": req.QuestID,
	}
	go c.forwardQuestToGameService(req.RoleID, "/api/quest/abandon", "POST", reqData, CmdQuestAbandon)
}

// forwardQuestToGameService 将任务请求转发到GameService
func (c *Client) forwardQuestToGameService(roleID uint64, path string, method string, reqData map[string]interface{}, respCmd uint16) {
	// 获取角色所在地图对应的GameService实例
	instance := common.GetInstanceByRoleID(roleID)
	if instance == nil {
		// 如果没有找到实例，尝试获取任意实例（用于处理角色未在地图上的情况）
		instances := common.GetAllInstances()
		if len(instances) == 0 {
			c.sendPacket(respCmd, mustMarshal(map[string]interface{}{"code": 500, "msg": "没有可用的GameService"}))
			return
		}
		instance = instances[0]
	}

	var resp *http.Response
	var err error
	client := &http.Client{Timeout: 10 * time.Second}

	if method == "GET" {
		resp, err = client.Get(instance.URL + path)
	} else {
		jsonData, _ := json.Marshal(reqData)
		resp, err = client.Post(instance.URL+path, "application/json", bytes.NewReader(jsonData))
	}

	if err != nil {
		log.Printf("调用GameService任务接口失败: %v", err)
		c.sendPacket(respCmd, mustMarshal(map[string]interface{}{"code": 500, "msg": "服务调用失败"}))
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	c.sendPacket(respCmd, body)
}

// ==================== 成就相关 Handler ====================

// handleAchievementList 处理获取成就列表请求
func (c *Client) handleAchievementList(body []byte) {
	var req struct {
		RoleID uint64 `json:"role_id"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		c.sendPacket(CmdAchievementList, mustMarshal(map[string]interface{}{"code": 400, "msg": "请求格式错误"}))
		return
	}

	go c.forwardAchievementToGameService(req.RoleID, "/api/quest/achievement/list/"+fmt.Sprintf("%d", req.RoleID), "GET", nil, CmdAchievementList)
}

// handleAchievementStats 处理获取成就统计请求
func (c *Client) handleAchievementStats(body []byte) {
	var req struct {
		RoleID uint64 `json:"role_id"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		c.sendPacket(CmdAchievementStats, mustMarshal(map[string]interface{}{"code": 400, "msg": "请求格式错误"}))
		return
	}

	go c.forwardAchievementToGameService(req.RoleID, "/api/quest/achievement/stats/"+fmt.Sprintf("%d", req.RoleID), "GET", nil, CmdAchievementStats)
}

// forwardAchievementToGameService 将成就请求转发到GameService
func (c *Client) forwardAchievementToGameService(roleID uint64, path string, method string, reqData map[string]interface{}, respCmd uint16) {
	instance := common.GetInstanceByRoleID(roleID)
	if instance == nil {
		instances := common.GetAllInstances()
		if len(instances) == 0 {
			c.sendPacket(respCmd, mustMarshal(map[string]interface{}{"code": 500, "msg": "没有可用的GameService"}))
			return
		}
		instance = instances[0]
	}

	var resp *http.Response
	var err error
	client := &http.Client{Timeout: 10 * time.Second}

	if method == "GET" {
		resp, err = client.Get(instance.URL + path)
	} else {
		jsonData, _ := json.Marshal(reqData)
		resp, err = client.Post(instance.URL+path, "application/json", bytes.NewReader(jsonData))
	}

	if err != nil {
		log.Printf("调用GameService成就接口失败: %v", err)
		c.sendPacket(respCmd, mustMarshal(map[string]interface{}{"code": 500, "msg": "服务调用失败"}))
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	c.sendPacket(respCmd, body)
}

// ==================== 技能相关 Handler ====================

// handleSkillList 处理获取技能列表请求
// ★ BUG修复：正确处理type参数 + 正确传递msg_id
func (c *Client) handleSkillList(body []byte) {
	var req struct {
		RoleID uint64   `json:"role_id"`
		Type   string   `json:"type"`   // 'equipped' | 'learned' | 'base' | ''
		MsgID  *float64 `json:"msg_id"` // ★ 新增：保留消息ID用于RPC匹配
	}
	if err := json.Unmarshal(body, &req); err != nil {
		c.sendPacket(CmdSkillList, mustMarshal(map[string]interface{}{"code": 400, "msg": "请求格式错误"}))
		return
	}

	// ★ 构建请求数据时包含msg_id（确保RPC响应能正确匹配）
	reqData := map[string]interface{}{"role_id": req.RoleID}
	if req.MsgID != nil {
		reqData["msg_id"] = *req.MsgID
	}

	// ★ 根据type参数路由到不同的API
	switch req.Type {
	case "equipped":
		// 获取已装备武学 → 调用GameService的GetEquippedSkills接口
		go c.forwardToGameServiceSync(req.RoleID,
			fmt.Sprintf("/api/skill/role/%d/equipped", req.RoleID),
			"GET", reqData, CmdSkillList)
	case "learned":
		// 获取已学武学 → 调用GameService的GetRoleSkills接口
		go c.forwardToGameServiceSync(req.RoleID,
			fmt.Sprintf("/api/skill/role/%d/list", req.RoleID),
			"GET", reqData, CmdSkillList)
	default:
		// 默认：获取武学基础配置列表（向后兼容）
		go c.forwardToGameService(req.RoleID, "/api/skill/base/list", "GET", reqData, CmdSkillList)
	}
}

// handleSkillLearn 处理学习技能请求
// ★ BUG修复：正确传递msg_id用于RPC匹配
func (c *Client) handleSkillLearn(body []byte) {
	var req struct {
		RoleID    uint64   `json:"role_id"`
		SkillID   uint32   `json:"skill_id"`
		RoleLevel uint32   `json:"role_level"` // 角色等级（用于学习条件检查）
		MsgID     *float64 `json:"msg_id"`     // ★ 新增：保留消息ID
	}
	if err := json.Unmarshal(body, &req); err != nil {
		c.sendPacket(CmdSkillLearn, mustMarshal(map[string]interface{}{"code": 400, "msg": "请求格式错误"}))
		return
	}

	reqData := map[string]interface{}{
		"role_id":    req.RoleID,
		"skill_id":   req.SkillID,
		"role_level": req.RoleLevel,
	}
	// ★ 传递msg_id确保RPC响应能正确匹配
	if req.MsgID != nil {
		reqData["msg_id"] = *req.MsgID
	}

	// 同步转发（学习技能需要立即返回结果给前端）
	c.forwardToGameServiceSync(req.RoleID, "/api/skill/role/"+fmt.Sprintf("%d", req.RoleID)+"/learn", "POST", reqData, CmdSkillLearn)
}

// handleSkillUpgrade 处理升级技能请求
// ★ BUG修复：正确传递msg_id用于RPC匹配
func (c *Client) handleSkillUpgrade(body []byte) {
	var req struct {
		RoleID  uint64   `json:"role_id"`
		SkillID uint32   `json:"skill_id"`
		MsgID   *float64 `json:"msg_id"` // ★ 新增：保留消息ID
	}
	if err := json.Unmarshal(body, &req); err != nil {
		c.sendPacket(CmdSkillUpgrade, mustMarshal(map[string]interface{}{"code": 400, "msg": "请求格式错误"}))
		return
	}

	reqData := map[string]interface{}{
		"role_id":  req.RoleID,
		"skill_id": req.SkillID,
	}
	// ★ 传递msg_id确保RPC响应能正确匹配
	if req.MsgID != nil {
		reqData["msg_id"] = *req.MsgID
	}

	// 同步转发（升级技能需要立即返回结果）
	c.forwardToGameServiceSync(req.RoleID, "/api/skill/role/"+fmt.Sprintf("%d", req.RoleID)+"/upgrade", "POST", reqData, CmdSkillUpgrade)
}

// handleSkillEquip 处理装备技能请求
// ★ BUG修复：正确传递msg_id用于RPC匹配
func (c *Client) handleSkillEquip(body []byte) {
	var req struct {
		RoleID    uint64   `json:"role_id"`
		SkillID   uint32   `json:"skill_id"`
		SlotIndex int      `json:"slot_index"`
		MsgID     *float64 `json:"msg_id"` // ★ 新增：保留消息ID
	}
	if err := json.Unmarshal(body, &req); err != nil {
		c.sendPacket(CmdSkillEquip, mustMarshal(map[string]interface{}{"code": 400, "msg": "请求格式错误"}))
		return
	}

	reqData := map[string]interface{}{
		"role_id":    req.RoleID,
		"skill_id":   req.SkillID,
		"slot_index": req.SlotIndex,
	}
	// ★ 传递msg_id确保RPC响应能正确匹配
	if req.MsgID != nil {
		reqData["msg_id"] = *req.MsgID
	}

	// 同步转发（装备技能需要立即返回结果）
	c.forwardToGameServiceSync(req.RoleID, "/api/skill/role/"+fmt.Sprintf("%d", req.RoleID)+"/equip", "POST", reqData, CmdSkillEquip)
}

// handleSkillUnequip 处理卸下技能请求
// ★ BUG修复：正确传递msg_id用于RPC匹配
func (c *Client) handleSkillUnequip(body []byte) {
	var req struct {
		RoleID  uint64   `json:"role_id"`
		SkillID uint32   `json:"skill_id"`
		MsgID   *float64 `json:"msg_id"` // ★ 新增：保留消息ID
	}
	if err := json.Unmarshal(body, &req); err != nil {
		c.sendPacket(CmdSkillUnequip, mustMarshal(map[string]interface{}{"code": 400, "msg": "请求格式错误"}))
		return
	}

	reqData := map[string]interface{}{
		"role_id":  req.RoleID,
		"skill_id": req.SkillID,
	}
	// ★ 传递msg_id确保RPC响应能正确匹配
	if req.MsgID != nil {
		reqData["msg_id"] = *req.MsgID
	}

	// 同步转发（卸下技能需要立即返回结果）
	c.forwardToGameServiceSync(req.RoleID, "/api/skill/role/"+fmt.Sprintf("%d", req.RoleID)+"/unequip", "POST", reqData, CmdSkillUnequip)
}

// handleSkillExp 处理技能熟练度更新请求
func (c *Client) handleSkillExp(body []byte) {
	var req struct {
		RoleID  uint64 `json:"role_id"`
		SkillID uint32 `json:"skill_id"`
		Exp     int64  `json:"exp"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		c.sendPacket(CmdSkillExp, mustMarshal(map[string]interface{}{"code": 400, "msg": "请求格式错误"}))
		return
	}

	reqData := map[string]interface{}{
		"role_id":  req.RoleID,
		"skill_id": req.SkillID,
		"exp":      req.Exp,
	}
	go c.forwardToGameService(req.RoleID, "/api/skill/role/"+fmt.Sprintf("%d", req.RoleID)+"/add_exp", "POST", reqData, CmdSkillExp)
}

// ==================== 物品/装备相关 Handler ====================

// handleItemList 处理获取背包列表请求
func (c *Client) handleItemList(body []byte) {
	var req struct {
		RoleID uint64 `json:"role_id"`
		MsgID  uint64 `json:"msg_id"` // ★ 新增：保存消息ID
	}
	if err := json.Unmarshal(body, &req); err != nil {
		c.sendPacket(CmdItemList, mustMarshal(map[string]interface{}{"code": 400, "msg": "请求格式错误"}))
		return
	}

	// ★ 改为同步转发，确保返回 msg_id
	reqData := map[string]interface{}{
		"role_id": req.RoleID,
		"msg_id":  req.MsgID,
	}
	c.forwardToGameServiceSync(req.RoleID, "/api/item/bag/"+fmt.Sprintf("%d", req.RoleID)+"/list", "GET", reqData, CmdItemList)
}

// handleEquipList 处理获取装备列表请求
func (c *Client) handleEquipList(body []byte) {
	var req struct {
		RoleID uint64 `json:"role_id"`
		MsgID  uint64 `json:"msg_id"` // ★ 新增：保存消息ID
	}
	if err := json.Unmarshal(body, &req); err != nil {
		c.sendPacket(CmdEquipList, mustMarshal(map[string]interface{}{"code": 400, "msg": "请求格式错误"}))
		return
	}

	// ★ 改为同步转发，确保返回 msg_id
	reqData := map[string]interface{}{
		"role_id": req.RoleID,
		"msg_id":  req.MsgID,
	}
	c.forwardToGameServiceSync(req.RoleID, "/api/item/equip/"+fmt.Sprintf("%d", req.RoleID)+"/list", "GET", reqData, CmdEquipList)
}

// forwardToGameService 通用转发函数（支持RabbitMQ降级）
// forwardToGameServiceSync 同步转发到GameService（用于需要立即返回结果的请求）
// ★ 与forwardToGameService的区别：此函数是同步的，会等待GameService响应后再返回给前端
func (c *Client) forwardToGameServiceSync(roleID uint64, path string, method string, reqData map[string]interface{}, respCmd uint16) {
	instance := common.GetInstanceByRoleID(roleID)
	if instance == nil {
		instances := common.GetAllInstances()
		if len(instances) == 0 {
			c.sendPacket(respCmd, mustMarshal(map[string]interface{}{"code": 500, "msg": "没有可用的GameService"}))
			return
		}
		instance = instances[0]
	}

	var resp *http.Response
	var err error
	client := &http.Client{Timeout: 10 * time.Second}

	if method == "GET" {
		resp, err = client.Get(instance.URL + path)
	} else {
		jsonData, _ := json.Marshal(reqData)
		resp, err = client.Post(instance.URL+path, "application/json", bytes.NewReader(jsonData))
	}

	if err != nil {
		log.Printf("❌ [Sync] 调用GameService接口失败: %v", err)
		c.sendPacket(respCmd, mustMarshal(map[string]interface{}{"code": 500, "msg": "服务调用失败: " + err.Error()}))
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Printf(`✅ [Sync] GameService响应: %s %s → %d`, method, path, resp.StatusCode)

	// 添加msg_id到响应中，以便前端匹配请求
	var respData map[string]interface{}
	if err := json.Unmarshal(body, &respData); err == nil {
		respData["msg_id"] = reqData["msg_id"]
		body, _ = json.Marshal(respData)
	}
	c.sendPacket(respCmd, body)
}

func (c *Client) forwardToGameService(roleID uint64, path string, method string, reqData map[string]interface{}, respCmd uint16) {
	instance := common.GetInstanceByRoleID(roleID)
	if instance == nil {
		instances := common.GetAllInstances()
		if len(instances) == 0 {
			c.sendPacket(respCmd, mustMarshal(map[string]interface{}{"code": 500, "msg": "没有可用的GameService"}))
			return
		}
		instance = instances[0]
	}

	var resp *http.Response
	var err error
	client := &http.Client{Timeout: 10 * time.Second}

	if method == "GET" {
		resp, err = client.Get(instance.URL + path)
	} else {
		jsonData, _ := json.Marshal(reqData)
		resp, err = client.Post(instance.URL+path, "application/json", bytes.NewReader(jsonData))
	}

	if err != nil {
		log.Printf("调用GameService接口失败: %v", err)
		c.sendPacket(respCmd, mustMarshal(map[string]interface{}{"code": 500, "msg": "服务调用失败"}))
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	// 添加msg_id到响应中，以便前端匹配请求
	var respData map[string]interface{}
	if err := json.Unmarshal(body, &respData); err == nil {
		respData["msg_id"] = reqData["msg_id"]
		body, _ = json.Marshal(respData)
	}
	c.sendPacket(respCmd, body)
}

func (c *Client) sendPacket(cmd uint16, data []byte) {
	// 简单封包: | 命令(2字节) | 数据(N字节) |
	pkg := make([]byte, 2+len(data))
	binary.LittleEndian.PutUint16(pkg[0:2], cmd)
	copy(pkg[2:], data)

	select {
	case c.Send <- pkg:
	default:
	}
}

func mustMarshal(v interface{}) []byte {
	data, _ := json.Marshal(v)
	return data
}

// handleRegister 处理注册请求
func (c *Client) handleRegister(body []byte) {
	log.Printf("处理注册请求: body=%s", string(body))
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		c.sendPacket(CmdRegister, mustMarshal(map[string]interface{}{
			"code": 400,
			"msg":  "请求格式错误",
		}))
		return
	}

	if len(req.Username) < 4 || len(req.Username) > 20 {
		c.sendPacket(CmdRegister, mustMarshal(map[string]interface{}{
			"code": 400,
			"msg":  "用户名长度需4-20位",
		}))
		return
	}

	if len(req.Password) < 6 || len(req.Password) > 20 {
		c.sendPacket(CmdRegister, mustMarshal(map[string]interface{}{
			"code": 400,
			"msg":  "密码长度需6-20位",
		}))
		return
	}

	// 调用LoginService注册
	registerReq := map[string]string{
		"username": req.Username,
		"password": req.Password,
	}
	resp, err := common.LoginClient.Post("/api/register", registerReq)
	if err != nil {
		log.Printf("调用LoginService注册失败: %v", err)
		c.sendPacket(CmdRegister, mustMarshal(map[string]interface{}{
			"code": 500,
			"msg":  "服务异常，请稍后重试",
		}))
		return
	}

	var result struct {
		Code  int    `json:"code"`
		UID   uint   `json:"uid"`
		Token string `json:"token"`
		Msg   string `json:"msg"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		c.sendPacket(CmdRegister, mustMarshal(map[string]interface{}{
			"code": 500,
			"msg":  "响应解析失败",
		}))
		return
	}

	if result.Code != 0 {
		c.sendPacket(CmdRegister, mustMarshal(map[string]interface{}{
			"code": result.Code,
			"msg":  result.Msg,
		}))
		return
	}

	// 注册成功，自动登录
	c.AccountID = uint64(result.UID)
	c.Token = result.Token

	log.Printf("注册成功: accountID=%d", c.AccountID)

	c.sendPacket(CmdRegister, mustMarshal(map[string]interface{}{
		"code":       200,
		"msg":        "注册成功",
		"account_id": c.AccountID,
		"token":      c.Token,
	}))
}

// handleRoleList 处理获取角色列表请求
func (c *Client) handleRoleList() {
	log.Printf("获取角色列表: accountID=%d", c.AccountID)

	if c.AccountID == 0 {
		c.sendPacket(CmdRoleList, mustMarshal(map[string]interface{}{
			"code": 400,
			"msg":  "请先登录",
		}))
		return
	}

	// 调用DBService获取角色列表
	resp, err := common.DBPost("/api/role/list", map[string]interface{}{
		"account_id": c.AccountID,
	})
	if err != nil {
		log.Printf("获取角色列表失败: %v", err)
		c.sendPacket(CmdRoleList, mustMarshal(map[string]interface{}{
			"code": 500,
			"msg":  "服务异常",
		}))
		return
	}

	code, _ := resp["code"].(float64)
	if code != 0 {
		msg, _ := resp["msg"].(string)
		c.sendPacket(CmdRoleList, mustMarshal(map[string]interface{}{
			"code": int(code),
			"msg":  msg,
		}))
		return
	}

	// 返回角色列表
	roles, _ := resp["data"].([]interface{})
	log.Printf("角色列表: accountID=%d, count=%d", c.AccountID, len(roles))

	c.sendPacket(CmdRoleList, mustMarshal(map[string]interface{}{
		"code":  200,
		"msg":   "获取成功",
		"roles": roles,
	}))
}

// handleRoleCreate 处理创建角色请求
func (c *Client) handleRoleCreate(body []byte) {
	log.Printf("创建角色: accountID=%d, body=%s", c.AccountID, string(body))

	if c.AccountID == 0 {
		c.sendPacket(CmdRoleCreate, mustMarshal(map[string]interface{}{
			"code": 400,
			"msg":  "请先登录",
		}))
		return
	}

	var req struct {
		Name       string `json:"name"`
		Gender     uint8  `json:"gender"`
		Appearance uint32 `json:"appearance"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		c.sendPacket(CmdRoleCreate, mustMarshal(map[string]interface{}{
			"code": 400,
			"msg":  "请求格式错误",
		}))
		return
	}

	if len(req.Name) < 2 || len(req.Name) > 12 {
		c.sendPacket(CmdRoleCreate, mustMarshal(map[string]interface{}{
			"code": 400,
			"msg":  "角色名长度需2-12位",
		}))
		return
	}

	// 调用DBService创建角色
	resp, err := common.DBPost("/api/role/create", map[string]interface{}{
		"account_id": c.AccountID,
		"name":       req.Name,
		"gender":     req.Gender,
		"appearance": req.Appearance,
	})
	if err != nil {
		log.Printf("创建角色失败: %v", err)
		c.sendPacket(CmdRoleCreate, mustMarshal(map[string]interface{}{
			"code": 500,
			"msg":  "服务异常",
		}))
		return
	}

	code, _ := resp["code"].(float64)
	if code != 0 {
		msg, _ := resp["msg"].(string)
		c.sendPacket(CmdRoleCreate, mustMarshal(map[string]interface{}{
			"code": int(code),
			"msg":  msg,
		}))
		return
	}

	roleID, _ := resp["data"].(float64)
	log.Printf("创建角色成功: accountID=%d, roleID=%d, name=%s", c.AccountID, uint64(roleID), req.Name)

	c.sendPacket(CmdRoleCreate, mustMarshal(map[string]interface{}{
		"code":    200,
		"msg":     "创建成功",
		"role_id": uint64(roleID),
		"name":    req.Name,
	}))
}

// handleRoleSelect 处理选择角色请求
func (c *Client) handleRoleSelect(body []byte) {
	log.Printf("选择角色: accountID=%d, body=%s", c.AccountID, string(body))

	if c.AccountID == 0 {
		c.sendPacket(CmdRoleSelect, mustMarshal(map[string]interface{}{
			"code": 400,
			"msg":  "请先登录",
		}))
		return
	}

	var req struct {
		RoleID uint64 `json:"role_id"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		c.sendPacket(CmdRoleSelect, mustMarshal(map[string]interface{}{
			"code": 400,
			"msg":  "请求格式错误",
		}))
		return
	}

	// 调用DBService获取角色信息
	resp, err := common.DBPost("/api/role/get", map[string]interface{}{
		"id": req.RoleID,
	})
	if err != nil {
		log.Printf("获取角色信息失败: %v", err)
		c.sendPacket(CmdRoleSelect, mustMarshal(map[string]interface{}{
			"code": 500,
			"msg":  "服务异常",
		}))
		return
	}

	code, _ := resp["code"].(float64)
	if code != 0 {
		msg, _ := resp["msg"].(string)
		c.sendPacket(CmdRoleSelect, mustMarshal(map[string]interface{}{
			"code": int(code),
			"msg":  msg,
		}))
		return
	}

	// 解析角色数据
	data, _ := json.Marshal(resp["data"])
	var role struct {
		ID        uint64 `json:"id"`
		AccountID uint64 `json:"account_id"`
		Name      string `json:"name"`
		Level     int    `json:"level"`
		Gender    uint8  `json:"gender"`
		MapID     int    `json:"map_id"`
		MapX      int    `json:"map_x"`
		MapY      int    `json:"map_y"`
		Hp        int    `json:"hp"`
		MaxHp     int    `json:"max_hp"`
		Mp        int    `json:"mp"`
		MaxMp     int    `json:"max_mp"`
		Attack    int    `json:"attack"`
		Defense   int    `json:"defense"`
		Speed     int    `json:"speed"`
		Gold      int64  `json:"gold"`
	}
	if err := json.Unmarshal(data, &role); err != nil {
		c.sendPacket(CmdRoleSelect, mustMarshal(map[string]interface{}{
			"code": 500,
			"msg":  "角色数据解析失败",
		}))
		return
	}

	// 验证角色归属
	if role.AccountID != c.AccountID {
		c.sendPacket(CmdRoleSelect, mustMarshal(map[string]interface{}{
			"code": 403,
			"msg":  "角色不属于当前账号",
		}))
		return
	}

	// 设置角色信息
	c.ID = role.ID
	c.Name = role.Name
	c.MapID = uint32(role.MapID)
	c.X = role.MapX
	c.Y = role.MapY

	log.Printf("选择角色成功: accountID=%d, roleID=%d, name=%s, mapID=%d", c.AccountID, c.ID, c.Name, c.MapID)

	// 返回角色完整信息
	c.sendPacket(CmdRoleSelect, mustMarshal(map[string]interface{}{
		"code":    200,
		"msg":     "选择成功",
		"role_id": role.ID,
		"name":    role.Name,
		"level":   role.Level,
		"gender":  role.Gender,
		"map_id":  role.MapID,
		"x":       role.MapX,
		"y":       role.MapY,
		"hp":      role.Hp,
		"max_hp":  role.MaxHp,
		"mp":      role.Mp,
		"max_mp":  role.MaxMp,
		"attack":  role.Attack,
		"defense": role.Defense,
		"speed":   role.Speed,
		"gold":    role.Gold,
	}))

	// 通知 GameService 进入地图（同步执行，确保怪物列表在 EnterMap 之前准备好）
	c.notifyGameServiceEnterMap()
}

// notifyGameServiceEnterMap 通知 GameService 玩家进入地图
func (c *Client) notifyGameServiceEnterMap() {
	instance := common.GetInstanceByMapID(c.MapID)
	if instance == nil {
		log.Printf("未找到处理地图 %d 的GameService实例，跳过进入地图通知", c.MapID)
		return
	}

	enterReq := map[string]interface{}{
		"role_id": c.ID,
		"map_id":  c.MapID,
		"x":       c.X,
		"y":       c.Y,
	}

	jsonData, err := json.Marshal(enterReq)
	if err != nil {
		log.Printf("序列化进入地图数据失败: %v", err)
		return
	}

	log.Printf("通知GameService进入地图: roleID=%d, mapID=%d, x=%d, y=%d, body=%s",
		c.ID, c.MapID, c.X, c.Y, string(jsonData))

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(instance.URL+"/api/map/enter", "application/json", bytes.NewReader(jsonData))
	if err != nil {
		log.Printf("通知GameService进入地图失败: %v", err)
		return
	}
	defer resp.Body.Close()

	// 解析响应，获取实际的坐标和怪物列表
	if resp.StatusCode == http.StatusOK {
		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err == nil {
			if x, ok := result["x"].(float64); ok {
				c.X = int(x)
			}
			if y, ok := result["y"].(float64); ok {
				c.Y = int(y)
			}

			// 提取怪物列表并推送给客户端
			if monsterList, ok := result["monster_list"].([]interface{}); ok && len(monsterList) > 0 {
				log.Printf("玩家 %d 进入地图 %d，收到 %d 个怪物", c.ID, c.MapID, len(monsterList))

				// 发送怪物列表给客户端
				GlobalManager.SendToClient(c.ID, &Message{
					Type: CmdSync, // 使用同步消息类型
					Data: mustMarshal(map[string]interface{}{
						"monster_list": monsterList,
					}),
				})
			} else {
				log.Printf("玩家 %d 进入地图 %d，没有怪物或怪物列表为空", c.ID, c.MapID)
			}

			log.Printf("玩家 %d 进入地图 %d，实际坐标: x=%d, y=%d", c.ID, c.MapID, c.X, c.Y)
		}
	}

	log.Printf("玩家 %d 进入地图 %d 成功，GameService状态码: %d", c.ID, c.MapID, resp.StatusCode)
}

func init() {
	// 启动客户端管理器
	go GlobalManager.Start()

	// 启动心跳检测
	go heartbeatCheck()
}

func heartbeatCheck() {
	ticker := time.NewTicker(30 * time.Second)
	for range ticker.C {
		GlobalManager.mutex.RLock()
		now := time.Now()
		for _, client := range GlobalManager.clients {
			if now.Sub(client.LastPing) > 60*time.Second {
				// 超时断开
				GlobalManager.unregister <- client
			}
		}
		GlobalManager.mutex.RUnlock()
	}
}

// GlobalManager 全局客户端管理器
var GlobalManager = NewClientManager()

// initMessageBus 初始化消息总线
func initMessageBus() {
	busConfig := common.AppConfig.MessageBus

	if busConfig.Type == "" {
		busConfig.Type = "http" // 默认使用HTTP
	}

	config := common.MessageBusConfig{
		Type:         busConfig.Type,
		RabbitMQURL:  busConfig.RabbitMQURL,
		KafkaBrokers: busConfig.KafkaBrokers,
		HTTPURL:      "", // Gateway不需要HTTPURL
	}

	common.InitMessageBus(config)

	// 如果消息总线可用，订阅消息
	if common.GlobalMessageBus != nil && common.GlobalMessageBus.IsAvailable() {
		go subscribeMessages()
	}
}

// subscribeMessages 订阅消息总线消息
func subscribeMessages() {
	// 订阅移动消息
	common.GlobalMessageBus.Subscribe("map_move", func(data []byte) {
		var moveData struct {
			RoleID uint64 `json:"role_id"`
			MapID  uint32 `json:"map_id"`
			X      int    `json:"x"`
			Y      int    `json:"y"`
		}
		if err := json.Unmarshal(data, &moveData); err != nil {
			log.Printf("解析移动消息失败: %v", err)
			return
		}

		// 广播给同地图玩家
		GlobalManager.BroadcastToMap(moveData.MapID, &Message{
			From: moveData.RoleID,
			Type: CmdMove,
			Data: data,
		})
	})

	// 订阅怪物位置二进制消息（来自GameService，通过RabbitMQ或HTTP降级）
	// 元数据: map_id, cmd；body: 紧凑二进制格式
	common.GlobalMessageBus.SubscribeRaw("monster_position", func(metadata map[string]string, body []byte) {
		mapIDStr, ok1 := metadata["map_id"]
		cmdStr, ok2 := metadata["cmd"]
		if !ok1 || !ok2 {
			log.Printf("怪物位置消息缺少元数据: %v", metadata)
			return
		}

		mapID64, err := strconv.ParseUint(mapIDStr, 10, 32)
		if err != nil {
			log.Printf("解析map_id失败: %v", err)
			return
		}
		cmd64, err := strconv.ParseUint(cmdStr, 10, 16)
		if err != nil {
			log.Printf("解析cmd失败: %v", err)
			return
		}

		// 直接透传二进制body给同地图所有客户端 [cmd][body]
		GlobalManager.BroadcastToMap(uint32(mapID64), &Message{
			Type: uint16(cmd64),
			Data: body,
		})
	})

	// 订阅任务更新消息（来自GameService，通过RabbitMQ或HTTP降级）
	// 推送给指定玩家
	common.GlobalMessageBus.Subscribe("quest.push", func(data []byte) {
		var pushData struct {
			Type   uint8       `json:"type"`
			RoleID uint64      `json:"role_id"`
			Data   interface{} `json:"data"`
		}
		if err := json.Unmarshal(data, &pushData); err != nil {
			log.Printf("解析任务推送消息失败: %v", err)
			return
		}

		// 通过WebSocket推送给指定玩家
		GlobalManager.SendToClient(pushData.RoleID, &Message{
			Type: CmdQuestUpdate,
			Data: data,
		})

		log.Printf("任务推送: roleID=%d, type=%d", pushData.RoleID, pushData.Type)
	})

	// 订阅成就更新消息（来自GameService，通过RabbitMQ或HTTP降级）
	common.GlobalMessageBus.Subscribe("achievement.push", func(data []byte) {
		var pushData struct {
			Type   string      `json:"type"`
			RoleID uint64      `json:"role_id"`
			Data   interface{} `json:"data"`
		}
		if err := json.Unmarshal(data, &pushData); err != nil {
			log.Printf("解析成就推送消息失败: %v", err)
			return
		}

		// 通过WebSocket推送给指定玩家
		GlobalManager.SendToClient(pushData.RoleID, &Message{
			Type: CmdAchievementList, // 使用成就列表命令码推送
			Data: data,
		})

		log.Printf("成就推送: roleID=%d, type=%s", pushData.RoleID, pushData.Type)
	})

	log.Printf("消息总线订阅完成: map_move, monster_position, quest.push, achievement.push")
}

func main() {
	log.Println("===== 网关服务启动 =====")

	// 加载配置文件(必须先加载)
	if err := common.LoadConfig("./Config/Gateway.yaml"); err != nil {
		log.Println("加载配置失败(使用默认配置):", err)
	}

	// 初始化WebSocket配置(从配置文件读取)
	initWebSocketConfig()

	// 初始化广播配置(从配置文件读取)
	initBroadcastConfig()

	// 初始化服务客户端
	common.InitServiceClients()

	// 初始化分片游戏客户端
	InitShardedGameClient()

	// 初始化消息总线
	initMessageBus()

	// 初始化 Redis（从配置读取）
	redisAddr := common.AppConfig.Redis.Addr
	if redisAddr == "" {
		redisAddr = "127.0.0.1:6379"
	}
	redisPassword := common.AppConfig.Redis.Password
	redisDB := common.AppConfig.Redis.DB
	redisPoolSize := common.AppConfig.Redis.PoolSize
	redisMinIdle := common.AppConfig.Redis.MinIdleConns
	InitRedis(redisAddr, redisPassword, redisDB, redisPoolSize, redisMinIdle)

	// 初始化 Redis Session（分布式会话）
	InitRedisSession(redisAddr, redisPassword, redisDB, redisPoolSize, redisMinIdle)

	// ★ 新增：初始化分布式路由器（多GameService实例支持）
	if err := common.InitDistributedRouter("./Config/GameServiceInstances.yaml"); err != nil {
		log.Printf("⚠️ 分布式路由器初始化失败，将使用默认路由: %v", err)
	} else {
		log.Println(`✅ 分布式路由器初始化成功`)
	}

	// ★ 新增：初始化跨网关广播（Redis Pub/Sub + RabbitMQ + 消息去重）
	gatewayID := fmt.Sprintf("gateway-%d", time.Now().UnixNano()%10000)
	// ★ 修复：复用主消息总线的RabbitMQ配置（如果未配置则直接走HTTP降级）
	rabbitMQURL := common.AppConfig.MessageBus.RabbitMQURL // 从配置文件读取
	if err := common.InitCrossGatewayBroadcast(gatewayID, redisAddr, rabbitMQURL); err != nil {
		log.Printf("⚠️ 跨网关广播初始化失败: %v", err)
	} else {
		log.Printf(`✅ 跨网关广播初始化成功: gatewayID=%s`, gatewayID)
	}

	// ★ 新增：初始化怪物同步管理器（通过网关广播怪物状态）
	common.InitMonsterSyncManager(
		func(msg *common.BroadcastMessage) {
			// 回调：将怪物更新消息通过网关广播给视野范围内的玩家
			go func() {
				GlobalManager.BroadcastToViewRangeConcurrent(msg.MapID, 0, 0, &Message{
					Type: msg.MessageType,
					Data: mustMarshal(msg.Data),
				})
			}()
		},
		50*time.Millisecond, // 批量发送间隔50ms
		100,                 // 最大批量大小
	)
	log.Println(`✅ 怪物同步管理器初始化成功`)

	// ★ 新增：初始化Prometheus性能监控指标
	metricsPort := 9090 // Prometheus metrics端口
	if err := common.InitPrometheusMetrics(metricsPort); err != nil {
		log.Printf("⚠️ Prometheus监控初始化失败: %v", err)
	} else {
		log.Printf(`✅ Prometheus监控初始化成功: http://localhost:%d/metrics`, metricsPort)
	}

	// 启动定期广播在线人数（每10秒）
	go broadcastOnlineCount()

	// 启动心跳超时检测
	go checkHeartbeatTimeout()

	// 启动定期会话清理日志（每5分钟）
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			cleanupExpiredSessions()
		}
	}()

	wsPort := common.AppConfig.GetWSPort()
	log.Printf("WebSocket: :%d", wsPort)

	// 添加CORS中间件（允许跨域请求）
	corsHandler := corsMiddleware(http.DefaultServeMux)

	http.HandleFunc("/ws", HandleWebSocket)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok","online":` + fmt.Sprintf("%d", GlobalManager.GetOnlineCount()) + `}`))
	})

	// 登录注册API路由（转发到LoginService）
	http.HandleFunc("/api/auth/login", handleAuthLogin)
	http.HandleFunc("/api/auth/register", handleAuthRegister)

	// 内部API：GameService调用推送消息给客户端
	http.HandleFunc("/internal/push", handleInternalPush)
	// 内部API：GameService调用广播消息给地图玩家
	http.HandleFunc("/internal/broadcast", handleInternalBroadcast)
	// 内部API：GameService调用广播移动消息给同地图玩家
	http.HandleFunc("/internal/broadcast_map", handleInternalBroadcastMap)
	// 内部API：GameService调用二进制广播（透传二进制body给客户端，零拷贝）
	http.HandleFunc("/internal/broadcast_binary", handleInternalBroadcastBinary)
	// 内部API：GameService调用任务推送（通过WebSocket推送给指定玩家）
	http.HandleFunc("/internal/quest-push", handleInternalQuestPush)

	// ★ 新增：静态数据代理API（前端通过网关获取配置数据，支持分布式架构）
	// 技能配置代理
	http.HandleFunc("/api/skill/base/list", handleProxyToGameService)
	http.HandleFunc("/api/skill/base/type/", handleProxyToGameService)
	http.HandleFunc("/api/skill/base/", handleProxyToGameService)
	http.HandleFunc("/api/skill/type/list", handleProxyToGameService)

	// ★ 物品操作代理（优先注册，避免被其他路由覆盖）
	http.HandleFunc("/api/item/bag/", handleProxyToGameService)
	http.HandleFunc("/api/item/equip/", handleProxyToGameService)

	// 道具配置代理（注意：这些路由可能覆盖上面的路由，所以放在后面）
	http.HandleFunc("/api/item/base/list", handleProxyToGameService)
	http.HandleFunc("/api/item/base/", handleProxyToGameService)

	// 怪物配置代理
	http.HandleFunc("/api/monster/base/list", handleProxyToGameService)
	http.HandleFunc("/api/monster/base/", handleProxyToGameService)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", wsPort), corsHandler))
}

// corsMiddleware CORS中间件，允许跨域请求
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 设置CORS头
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Max-Age", "86400")

		// 处理预检请求
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// handleAuthLogin 处理登录请求（转发到LoginService）
func handleAuthLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	forwardToAuthService(w, r, "/api/login")
}

// handleAuthRegister 处理注册请求（转发到LoginService）
func handleAuthRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	forwardToAuthService(w, r, "/api/register")
}

// forwardToAuthService 转发认证请求到LoginService
// handleProxyToGameService 静态数据代理处理器
// ★ 将前端的配置数据请求转发到GameService，支持分布式架构
// 前端只需知道网关地址，无需直接连接GameService
func handleProxyToGameService(w http.ResponseWriter, r *http.Request) {
	// 记录请求日志
	startTime := time.Now()
	log.Printf(`🔄 [Gateway Proxy] 收到代理请求: %s %s`, r.Method, r.URL.Path)

	// ★ Prometheus：记录代理请求
	common.RecordProxyRequest(r.URL.Path, "received")

	// 处理预检请求
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// ★ 获取目标GameService实例（支持分布式路由）
	var targetURL string

	// ★ 尝试从URL中提取角色ID，进行精准路由
	// URL格式: /api/item/bag/{roleId}/... 或 /api/item/equip/{roleId}/...
	var roleID uint64
	pathParts := strings.Split(r.URL.Path, "/")
	for i, part := range pathParts {
		if part == "bag" || part == "equip" {
			if i+1 < len(pathParts) {
				parsedID, err := strconv.ParseUint(pathParts[i+1], 10, 64)
				if err == nil {
					roleID = parsedID
					break
				}
			}
		}
	}

	var instance *common.GameServiceInstance

	if roleID > 0 {
		// ★ 根据角色ID获取对应的GameService实例（分布式路由）
		instance = common.GetInstanceByRoleID(roleID)
		if instance == nil {
			log.Printf("⚠️ [Gateway Proxy] 根据角色ID%d获取实例失败", roleID)
		}
	}

	if instance == nil {
		// ★ 回退到负载均衡
		// 尝试获取任意健康实例
		var err error
		instance, err = common.GetAnyHealthyInstance()
		if err != nil {
			log.Printf("⚠️ [Gateway Proxy] 未找到健康的GameService实例，使用默认地址")
			// 使用配置的默认地址
			targetURL = common.AppConfig.Services.GameService
			if targetURL == "" {
				targetURL = "http://127.0.0.1:8082"
			}
		} else {
			targetURL = instance.URL
			log.Printf("✅ [Gateway Proxy] 使用负载均衡路由到实例%d: %s", instance.InstanceID, targetURL)
		}
	} else {
		targetURL = instance.URL
		log.Printf("✅ [Gateway Proxy] 根据角色ID%d路由到实例%d: %s", roleID, instance.InstanceID, targetURL)
	}

	// 构建完整的转发URL（保留原始路径和查询参数）
	forwardURL := targetURL + r.URL.Path
	if r.URL.RawQuery != "" {
		forwardURL += "?" + r.URL.RawQuery
	}

	// 创建HTTP客户端（带超时控制）
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	var resp *http.Response
	var err error

	// 根据请求方法转发
	switch r.Method {
	case http.MethodGet:
		resp, err = client.Get(forwardURL)
	case http.MethodPost, http.MethodPut, http.MethodDelete:
		// 读取请求体
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, `{"code":400,"msg":"读取请求失败"}`, http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// 创建新请求
		req, err := http.NewRequest(r.Method, forwardURL, bytes.NewReader(body))
		if err != nil {
			http.Error(w, `{"code":500,"msg":"创建请求失败"}`, http.StatusInternalServerError)
			return
		}

		// 复制原始请求头
		req.Header = r.Header.Clone()
		resp, err = client.Do(req)
	default:
		http.Error(w, `{"code":405,"msg":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if err != nil {
		log.Printf("❌ [Gateway Proxy] 转发失败: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(fmt.Sprintf(`{"code":502,"msg":"GameService不可用: %v"}`, err)))
		common.RecordProxyRequest(r.URL.Path, "error")
		return
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("❌ [Gateway Proxy] 读取响应失败: %v", err)
		http.Error(w, `{"code":500,"msg":"读取响应失败"}`, http.StatusInternalServerError)
		return
	}

	// ★ 设置缓存头（静态数据可缓存24小时）
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=86400") // 缓存24小时
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// 转发状态码和响应体
	w.WriteHeader(resp.StatusCode)
	w.Write(respBody)

	// 记录成功日志
	duration := time.Since(startTime)
	log.Printf(`✅ [Gateway Proxy] 代理完成: %s %s → %d (%v)`,
		r.Method, r.URL.Path, resp.StatusCode, duration)

	// ★ Prometheus：记录代理耗时
	common.RecordProxyDuration(r.URL.Path, duration)
}

func forwardToAuthService(w http.ResponseWriter, r *http.Request, path string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "读取请求失败", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// 获取LoginService地址（从配置或默认值）
	loginAddr := common.AppConfig.Services.LoginService
	if loginAddr == "" {
		loginAddr = "http://127.0.0.1:8082"
	}

	// 转发请求到LoginService
	targetURL := loginAddr + path
	resp, err := http.Post(targetURL, "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("转发到LoginService失败: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"code":500,"msg":"登录服务不可用"}`))
		return
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"code":500,"msg":"读取响应失败"}`))
		return
	}

	log.Printf("LoginService响应 (%s): %s", path, string(respBody))

	// 转换响应格式：将 code:0 改为 code:200（前端期望的格式）
	var loginResp struct {
		Code      int    `json:"code"`
		UID       uint   `json:"uid"`
		Token     string `json:"token"`
		Msg       string `json:"msg"`
		AccountID uint   `json:"account_id,omitempty"`
		Name      string `json:"name,omitempty"`
		RoleID    uint64 `json:"role_id,omitempty"`
	}

	if err := json.Unmarshal(respBody, &loginResp); err == nil {
		// 转换为前端期望的格式
		if loginResp.Code == 0 {
			loginResp.Code = 200
		}
		// 确保 account_id 字段存在
		if loginResp.AccountID == 0 && loginResp.UID > 0 {
			loginResp.AccountID = loginResp.UID
		}

		// 登录成功时，将token保存到Redis Session（用于后续WebSocket认证）
		if loginResp.Code == 200 && loginResp.Token != "" {
			log.Printf("保存登录session到Redis: accountID=%d, token=%s...", loginResp.AccountID, loginResp.Token[:min(10, len(loginResp.Token))])
			if err := saveSession(loginResp.Token, uint64(loginResp.AccountID), 0, "", 1, 0, 0); err != nil {
				log.Printf("保存session失败: %v", err)
			}
		}

		// 返回转换后的JSON
		result, _ := json.Marshal(loginResp)
		w.Header().Set("Content-Type", "application/json")
		w.Write(result)
	} else {
		// 如果解析失败，直接返回原始响应（可能是错误信息）
		w.Header().Set("Content-Type", "application/json")
		w.Write(respBody)
	}
}

// handleInternalPush 处理GameService的推送请求（推送给指定玩家）
func handleInternalPush(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		RoleID  uint64      `json:"role_id"`
		MsgType string      `json:"msg_type"`
		Data    interface{} `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("解析推送请求失败: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// 构建消息
	msgData, err := json.Marshal(req.Data)
	if err != nil {
		log.Printf("序列化消息数据失败: %v", err)
		http.Error(w, "Invalid data", http.StatusBadRequest)
		return
	}

	// 根据MsgType获取命令码
	cmd := getMsgTypeCmd(req.MsgType)

	// 发送到广播通道
	GlobalManager.broadcast <- &Message{
		To:   req.RoleID,
		Type: cmd,
		Data: msgData,
	}

	log.Printf("推送消息到玩家: roleID=%d, msgType=%s, cmd=%d", req.RoleID, req.MsgType, cmd)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

// handleInternalBroadcast 处理GameService的广播请求（广播给地图所有玩家）
func handleInternalBroadcast(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		MapID   uint32      `json:"map_id"`
		MsgType string      `json:"msg_type"`
		Data    interface{} `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("解析广播请求失败: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// 构建消息
	msgData, err := json.Marshal(req.Data)
	if err != nil {
		log.Printf("序列化消息数据失败: %v", err)
		http.Error(w, "Invalid data", http.StatusBadRequest)
		return
	}

	// 根据MsgType获取命令码
	cmd := getMsgTypeCmd(req.MsgType)

	// 使用 BroadcastToMap 只广播给同地图玩家（更高效）
	GlobalManager.BroadcastToMap(req.MapID, &Message{
		Type: cmd,
		Data: msgData,
	})

	log.Printf("广播消息到地图: mapID=%d, msgType=%s, cmd=%d", req.MapID, req.MsgType, cmd)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

// handleInternalBroadcastMap 处理GameService的移动广播请求
func handleInternalBroadcastMap(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		RoleID uint64 `json:"role_id"`
		MapID  uint32 `json:"map_id"`
		X      int    `json:"x"`
		Y      int    `json:"y"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("解析移动广播请求失败: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// 构建移动消息
	msgData, err := json.Marshal(map[string]interface{}{
		"role_id": req.RoleID,
		"x":       req.X,
		"y":       req.Y,
	})
	if err != nil {
		log.Printf("序列化移动数据失败: %v", err)
		http.Error(w, "Invalid data", http.StatusBadRequest)
		return
	}

	// 使用 BroadcastToMap 只广播给同地图玩家（排除自己）
	GlobalManager.BroadcastToMap(req.MapID, &Message{
		From: req.RoleID,
		Type: CmdMove,
		Data: msgData,
	})

	log.Printf("广播移动消息: roleID=%d, mapID=%d, x=%d, y=%d", req.RoleID, req.MapID, req.X, req.Y)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

// handleInternalBroadcastBinary 处理GameService的二进制广播请求
// 直接透传二进制body给客户端，避免JSON编解码开销
// Header: X-Map-Id (地图ID), X-Cmd (命令码)
// Body: 纯二进制数据，Gateway直接拼 [cmd][body] 发给客户端
func handleInternalBroadcastBinary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 从Header读取元数据
	mapIDStr := r.Header.Get("X-Map-Id")
	cmdStr := r.Header.Get("X-Cmd")
	if mapIDStr == "" || cmdStr == "" {
		http.Error(w, "Missing X-Map-Id or X-Cmd header", http.StatusBadRequest)
		return
	}

	mapID64, err := strconv.ParseUint(mapIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid X-Map-Id", http.StatusBadRequest)
		return
	}
	mapID := uint32(mapID64)

	cmd64, err := strconv.ParseUint(cmdStr, 10, 16)
	if err != nil {
		http.Error(w, "Invalid X-Cmd", http.StatusBadRequest)
		return
	}
	cmd := uint16(cmd64)

	// 读取二进制body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("读取二进制广播body失败: %v", err)
		http.Error(w, "Read body failed", http.StatusBadRequest)
		return
	}

	if len(body) == 0 {
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	}

	// 直接构建客户端数据包 [cmd(2)][body]，零拷贝透传
	GlobalManager.BroadcastToMap(mapID, &Message{
		Type: cmd,
		Data: body,
	})

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

// getMsgTypeCmd 根据消息类型字符串获取命令码
func getMsgTypeCmd(msgType string) uint16 {
	switch msgType {
	case "damage":
		return CmdDamage
	case "attack_result":
		// 玩家攻击怪物的结果，复用CmdDamage通道（前端按damage处理）
		return CmdDamage
	case "attack_failed":
		// 攻击失败（距离过远、冷却中等），复用CmdDamage通道
		// 前端通过error_code/error_msg字段区分失败类型
		return CmdDamage
	case "death":
		return CmdDeath
	case "respawn":
		return CmdRespawn
	case "level_up":
		return CmdLevelUp
	case "buff":
		return CmdBuff
	case "debuff":
		return CmdDeBuff
	case "map_event":
		return CmdMapEvent
	case "monster_position_update":
		return CmdMonsterPositionUpdate
	case "monster_spawn":
		return CmdMonsterSpawn
	case "monster_death":
		return CmdMonsterDeath
	default:
		return 0
	}
}

// handleInternalQuestPush 处理GameService的任务推送请求
func handleInternalQuestPush(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var pushData struct {
		Type   uint8       `json:"type"`    // 推送类型: 1=进度更新, 2=接取, 3=完成, 4=放弃, 5=重置
		RoleID uint64      `json:"role_id"` // 角色ID
		Data   interface{} `json:"data"`    // 推送数据
	}

	if err := json.NewDecoder(r.Body).Decode(&pushData); err != nil {
		log.Printf("解析任务推送请求失败: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// 构建消息
	msgData, err := json.Marshal(pushData)
	if err != nil {
		log.Printf("序列化任务推送数据失败: %v", err)
		http.Error(w, "Invalid data", http.StatusBadRequest)
		return
	}

	// 发送到广播通道（只推送给指定角色）
	GlobalManager.broadcast <- &Message{
		Type: CmdQuestUpdate, // 使用任务更新命令码
		Data: msgData,
	}

	log.Printf("任务推送: roleID=%d, type=%d", pushData.RoleID, pushData.Type)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

// checkHeartbeatTimeout 心跳超时检测(清理僵尸连接)
func checkHeartbeatTimeout() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		GlobalManager.mutex.RLock()
		now := time.Now()
		for _, client := range GlobalManager.clients {
			if now.Sub(client.LastPing) > pongTimeout {
				log.Printf("心跳超时断开: roleID=%d, 超时时间=%v", client.ID, now.Sub(client.LastPing))
				// 关闭连接
				client.Conn.Close()
				// 触发注销
				go func(c *Client) {
					GlobalManager.unregister <- c
				}(client)
			}
		}
		GlobalManager.mutex.RUnlock()
	}
}

// broadcastOnlineCount 定期广播在线人数
func broadcastOnlineCount() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C
		count := GlobalManager.GetOnlineCount()
		log.Printf("广播在线人数: %d", count)

		// 广播给所有在线玩家
		GlobalManager.BroadcastToAll(&Message{
			Type: CmdOnlineCount,
			Data: mustMarshal(map[string]interface{}{
				"count": count,
			}),
		})
	}
}
