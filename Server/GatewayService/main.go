package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

	// 会话管理
	sessionStore   = make(map[string]*SessionData) // token -> session
	sessionMutex   sync.RWMutex
	sessionTimeout = 5 * time.Minute // 会话超时时间
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
	ID         uint64          // 角色ID
	Name       string          // 角色名
	Conn       *websocket.Conn // WebSocket连接
	Send       chan []byte     // 发送通道
	GameSvc    string          // 所在游戏服务
	MapID      uint32          // 当前地图
	X          int             // X坐标
	Y          int             // Y坐标
	LastPing   time.Time       // 最后心跳时间
	Registered bool            // 是否已注册到管理器
	Token      string          // 重连token
}

// SessionData 会话数据（用于断线重连）
type SessionData struct {
	RoleID uint64    // 角色ID
	Name   string    // 角色名
	MapID  uint32    // 当前地图
	X      int       // X坐标
	Y      int       // Y坐标
	Expire time.Time // 过期时间
}

// generateToken 生成重连token
func generateToken(roleID uint64) string {
	return fmt.Sprintf("%d_%d", roleID, time.Now().UnixNano())
}

// saveSession 保存会话
func saveSession(token string, roleID uint64, name string, mapID uint32, x, y int) {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()
	sessionStore[token] = &SessionData{
		RoleID: roleID,
		Name:   name,
		MapID:  mapID,
		X:      x,
		Y:      y,
		Expire: time.Now().Add(sessionTimeout),
	}
}

// getSession 获取会话
func getSession(token string) (*SessionData, bool) {
	sessionMutex.RLock()
	defer sessionMutex.RUnlock()
	session, ok := sessionStore[token]
	if !ok {
		return nil, false
	}
	// 检查是否过期
	if session.Expire.Before(time.Now()) {
		return nil, false
	}
	return session, true
}

// removeSession 删除会话
func removeSession(token string) {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()
	delete(sessionStore, token)
}

// getSessionByRoleID 通过角色ID获取会话
func getSessionByRoleID(roleID uint64) (*SessionData, bool) {
	sessionMutex.RLock()
	defer sessionMutex.RUnlock()
	for _, session := range sessionStore {
		if session.RoleID == roleID && session.Expire.After(time.Now()) {
			return session, true
		}
	}
	return nil, false
}

// cleanupExpiredSessions 清理过期会话（定时调用）
func cleanupExpiredSessions() {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()
	now := time.Now()
	for token, session := range sessionStore {
		if session.Expire.Before(now) {
			delete(sessionStore, token)
		}
	}
}

// ClientManager 客户端管理器
type ClientManager struct {
	clients     map[uint64]*Client          // 客户端列表 key=roleID
	byConn      map[*websocket.Conn]*Client // by connection
	register    chan *Client                // 注册通道
	unregister  chan *Client                // 注销通道
	broadcast   chan *Message               // 广播通道
	mutex       sync.RWMutex
	onlineCount int64 // 在线人数（原子计数器）
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
	CmdLogin        uint16 = 1001 // 登录
	CmdLogout       uint16 = 1002 // 登出
	CmdHeartbeat    uint16 = 1003 // 心跳
	CmdMove         uint16 = 2001 // 移动
	CmdAttack       uint16 = 2002 // 攻击
	CmdUseSkill     uint16 = 2003 // 使用技能
	CmdChat         uint16 = 2004 // 聊天
	CmdPickup       uint16 = 2005 // 拾取
	CmdUseItem      uint16 = 2006 // 使用物品
	CmdEquip        uint16 = 2007 // 装备
	CmdTrade        uint16 = 2008 // 交易
	CmdEnterMap     uint16 = 3001 // 进入地图
	CmdLeaveMap     uint16 = 3002 // 离开地图
	CmdMapPlayer    uint16 = 3003 // 地图玩家列表
	CmdOnlineCount  uint16 = 3004 // 在线人数广播
	CmdNpcTalk      uint16 = 3005 // NPC对话
	CmdNpcTrade     uint16 = 3006 // NPC交易
	CmdSkillLearn   uint16 = 4001 // 学习武学
	CmdSkillUpgrade uint16 = 4002 // 升级武学
	CmdRoleInfo     uint16 = 5001 // 角色信息
	CmdRoleAttrib   uint16 = 5002 // 角色属性
	CmdSync         uint16 = 5003 // 属性同步
)

// NewClientManager 创建客户端管理器
func NewClientManager() *ClientManager {
	return &ClientManager{
		clients:    make(map[uint64]*Client),
		byConn:     make(map[*websocket.Conn]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Message, broadcastChannelSize),
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
			log.Printf("客户端断开: roleID=%d", client.ID)

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
		close(client.Send)
		atomic.AddInt64(&cm.onlineCount, -1) // 原子减少在线人数
	} else {
		log.Printf("removeClient: 玩家 %d 不存在于客户端列表", client.ID)
	}
}

func (cm *ClientManager) sendMessage(msg *Message) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	if msg.To > 0 {
		// 私聊
		if client, ok := cm.clients[msg.To]; ok {
			select {
			case client.Send <- msg.Data:
			default:
			}
		}
	} else {
		// 广播
		for _, client := range cm.clients {
			if client.ID == msg.From {
				select {
				case client.Send <- msg.Data:
				default:
				}
			}
		}
	}
}

// BroadcastToMap 广播到地图
func (cm *ClientManager) BroadcastToMap(mapID uint32, msg *Message) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	// 封包
	pkg := make([]byte, 2+len(msg.Data))
	binary.LittleEndian.PutUint16(pkg[0:2], msg.Type)
	copy(pkg[2:], msg.Data)

	for _, client := range cm.clients {
		if client.MapID == mapID && client.ID != msg.From {
			select {
			case client.Send <- pkg:
			default:
			}
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

	var players []*Client
	for _, client := range cm.clients {
		if client.MapID == mapID {
			players = append(players, client)
		}
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

	var result []*Client
	for _, client := range cm.clients {
		if client.MapID == mapID {
			result = append(result, client)
		}
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
	Code   int    `json:"code"`
	Msg    string `json:"msg"`
	RoleID uint64 `json:"role_id,omitempty"`
	Token  string `json:"token,omitempty"`
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
	log.Printf("收到消息: cmd=%d, bodyLen=%d, roleID=%d", cmd, len(body), c.ID)

	// 未登录玩家（roleID=0）只能发送登录和心跳消息
	if c.ID == 0 && cmd != CmdLogin && cmd != CmdHeartbeat {
		log.Printf("拒绝未登录玩家的消息: cmd=%d", cmd)
		return
	}

	switch cmd {
	case CmdHeartbeat:
		c.LastPing = time.Now()
		c.sendPacket(CmdHeartbeat, []byte(`{"time":`+fmt.Sprintf("%d", time.Now().Unix())+`}`))

	case CmdMove:
		c.handleMove(body)

	case CmdChat:
		c.handleChat(body)

	case CmdLogin:
		c.handleLogin(body)

	case CmdLogout:
		c.handleLogout()

	case CmdEnterMap:
		c.handleEnterMap(body)

	default:
		// 未知命令
	}
}

func (c *Client) handleLogin(body []byte) {
	log.Printf("处理登录请求: roleID=%d, body=%s", c.ID, string(body))
	var req LoginRequest
	if err := json.Unmarshal(body, &req); err != nil {
		c.sendPacket(CmdLogin, mustMarshal(LoginResponse{Code: 400, Msg: "请求格式错误"}))
		return
	}

	switch req.Type {
	case "account":
		// TODO: 调用登录服务验证
		c.sendPacket(CmdLogin, mustMarshal(LoginResponse{Code: 200, Msg: "登录成功", RoleID: 1}))

	case "guest":
		// 游客登录 - 生成唯一ID
		roleID := uint64(time.Now().UnixNano() % 1000000)
		if roleID == 0 {
			roleID = uint64(time.Now().Unix())
		}
		c.ID = roleID
		c.Name = req.Username
		// 更新管理器中的客户端ID
		GlobalManager.updateClientID(c, roleID)
		// 生成重连token
		token := generateToken(roleID)
		c.Token = token // 保存token到client
		c.sendPacket(CmdLogin, mustMarshal(LoginResponse{Code: 200, Msg: "登录成功", RoleID: roleID, Token: token}))

	case "reconnect":
		// 断线重连
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
		c.ID = session.RoleID
		c.Name = session.Name
		c.MapID = session.MapID
		c.X = session.X
		c.Y = session.Y

		// 更新管理器中的客户端ID
		GlobalManager.updateClientID(c, session.RoleID)

		// 注册到管理器
		GlobalManager.addClient(c)

		// 获取当前在线人数
		currentCount := GlobalManager.GetOnlineCount()
		log.Printf("重连成功: roleID=%d, mapID=%d, 在线人数=%d", c.ID, c.MapID, currentCount)

		// 发送登录成功响应（包含在线人数）
		c.sendPacket(CmdLogin, mustMarshal(LoginResponse{
			Code:   200,
			Msg:    "重连成功",
			RoleID: session.RoleID,
			Token:  req.Token, // 复用token
		}))

		// 发送在线人数
		pkg := mustMarshal(map[string]interface{}{"count": currentCount})
		sendPkg := make([]byte, 2+len(pkg))
		binary.LittleEndian.PutUint16(sendPkg[0:2], CmdOnlineCount)
		copy(sendPkg[2:], pkg)
		c.Send <- sendPkg

		// 广播在线人数更新给所有玩家
		GlobalManager.BroadcastToAll(&Message{
			Type: CmdOnlineCount,
			Data: mustMarshal(map[string]interface{}{
				"count": currentCount,
			}),
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

	log.Printf("玩家进入地图: roleID=%d, mapID=%d, x=%d, y=%d", c.ID, req.MapID, req.X, req.Y)

	// 如果是从其他地图切换，先广播离开旧地图
	if c.MapID != 0 && c.MapID != req.MapID {
		GlobalManager.BroadcastToMap(c.MapID, &Message{
			From: c.ID,
			Type: CmdLeaveMap,
			Data: mustMarshal(map[string]interface{}{
				"role_id": c.ID,
			}),
		})
	}

	// 更新玩家地图和位置
	c.MapID = req.MapID
	c.X = req.X
	c.Y = req.Y

	// 如果还未注册，现在注册（直接调用addClient，确保同步）
	if !c.Registered {
		GlobalManager.addClient(c)
		log.Printf("客户端注册: roleID=%d, mapID=%d", c.ID, c.MapID)

		// 保存会话（用于断线重连）
		if c.Token != "" {
			saveSession(c.Token, c.ID, c.Name, c.MapID, c.X, c.Y)
			log.Printf("保存会话: token=%s, roleID=%d", c.Token[:10]+"...", c.ID)
		}

		// 先获取当前在线人数（注册后的值）
		currentCount := GlobalManager.GetOnlineCount()
		log.Printf("当前在线人数: %d", currentCount)

		// 发送当前在线人数给新玩家（使用阻塞发送，确保消息送达）
		pkg := mustMarshal(map[string]interface{}{"count": currentCount})
		sendPkg := make([]byte, 2+len(pkg))
		binary.LittleEndian.PutUint16(sendPkg[0:2], CmdOnlineCount)
		copy(sendPkg[2:], pkg)
		log.Printf("准备发送在线人数给玩家 %d: count=%d", c.ID, currentCount)
		c.Send <- sendPkg
		log.Printf("成功发送在线人数给玩家 %d", c.ID)

		// 广播在线人数更新给所有其他玩家
		log.Printf("广播在线人数给所有玩家: count=%d", currentCount)
		GlobalManager.BroadcastToAll(&Message{
			Type: CmdOnlineCount,
			Data: mustMarshal(map[string]interface{}{
				"count": currentCount,
			}),
		})
		log.Printf("广播完成")
	}

	// 先向新玩家发送当前地图视野范围内的其他玩家列表
	existingPlayers := GlobalManager.GetMapPlayers(req.MapID)
	log.Printf("当前地图%d有%d个玩家", req.MapID, len(existingPlayers))

	if len(existingPlayers) > 1 { // 有其他玩家
		playerList := make([]map[string]interface{}, 0)
		viewRangeSq := VIEW_RANGE * VIEW_RANGE

		for _, p := range existingPlayers {
			if p.ID != c.ID {
				// 只添加视野范围内的玩家
				dx := p.X - c.X
				dy := p.Y - c.Y
				distanceSq := dx*dx + dy*dy

				if distanceSq <= viewRangeSq {
					playerList = append(playerList, map[string]interface{}{
						"role_id": p.ID,
						"name":    p.Name,
						"map_id":  p.MapID,
						"x":       p.X,
						"y":       p.Y,
					})
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
		}
	}

	// 使用视野范围并发广播新玩家进入（只发送给视野范围内的玩家）
	log.Printf("广播玩家进入: mapID=%d, 玩家ID=%d, 玩家名=%s", c.MapID, c.ID, c.Name)
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
	log.Printf("广播完成: mapID=%d, 当前在线玩家数=%d", c.MapID, len(GlobalManager.GetMapPlayers(c.MapID)))
}

func (c *Client) handleLogout() {
	GlobalManager.unregister <- c
	c.sendPacket(CmdLogout, mustMarshal(map[string]interface{}{"code": 200, "msg": "登出成功"}))
}

func (c *Client) handleMove(body []byte) {
	// 尝试解析JSON格式
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

	c.X = moveData.X
	c.Y = moveData.Y

	log.Printf("玩家移动: roleID=%d, x=%d, y=%d", c.ID, c.X, c.Y)

	// 使用消息合并和并发广播（同一玩家短时间内多次移动会合并成一条消息）
	data := mustMarshal(map[string]interface{}{
		"role_id": c.ID,
		"x":       c.X,
		"y":       c.Y,
	})
	GlobalManager.MergeAndBroadcastMove(c.MapID, c.ID, c.X, c.Y, data)
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

	// 启动定期广播在线人数（每10秒）
	go broadcastOnlineCount()

	// 启动心跳超时检测
	go checkHeartbeatTimeout()

	wsPort := common.AppConfig.GetWSPort()
	log.Printf("WebSocket: :%d", wsPort)

	http.HandleFunc("/ws", HandleWebSocket)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok","online":` + fmt.Sprintf("%d", GlobalManager.GetOnlineCount()) + `}`))
	})

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", wsPort), nil))
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
