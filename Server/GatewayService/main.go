package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
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
}

// ClientManager 客户端管理器
type ClientManager struct {
	clients    map[uint64]*Client          // 客户端列表 key=roleID
	byConn     map[*websocket.Conn]*Client // by connection
	register   chan *Client                // 注册通道
	unregister chan *Client                // 注销通道
	broadcast  chan *Message               // 广播通道
	mutex      sync.RWMutex
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
	CmdNpcTalk      uint16 = 3004 // NPC对话
	CmdNpcTrade     uint16 = 3005 // NPC交易
	CmdSkillLearn   uint16 = 4001 // 学习武学
	CmdSkillUpgrade uint16 = 4002 // 升级武学
	CmdRoleInfo     uint16 = 5001 // 角色信息
	CmdRoleAttrib   uint16 = 5002 // 角色属性
	CmdSync         uint16 = 5003 // 属性同步
)

// GlobalManager 全局客户端管理器
var GlobalManager = NewClientManager()

// NewClientManager 创建客户端管理器
func NewClientManager() *ClientManager {
	return &ClientManager{
		clients:    make(map[uint64]*Client),
		byConn:     make(map[*websocket.Conn]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Message, 1000),
	}
}

// Start 启动管理器
func (cm *ClientManager) Start() {
	for {
		select {
		case client := <-cm.register:
			cm.addClient(client)
			log.Printf("客户端连接: roleID=%d, addr=%s", client.ID, client.Conn.RemoteAddr())

			// 1. 向新玩家发送当前地图的玩家列表
			existingPlayers := cm.GetMapPlayers(client.MapID)
			if len(existingPlayers) > 1 { // 有其他玩家
				playerList := make([]map[string]interface{}, 0)
				for _, p := range existingPlayers {
					if p.ID != client.ID {
						playerList = append(playerList, map[string]interface{}{
							"role_id": p.ID,
							"map_id":  p.MapID,
							"x":       p.X,
							"y":       p.Y,
							"name":    p.Name,
						})
					}
				}
				if len(playerList) > 0 {
					cm.SendToClient(client.ID, &Message{
						Type: CmdMapPlayer,
						Data: mustMarshal(map[string]interface{}{
							"players": playerList,
						}),
					})
				}
			}

			// 2. 广播新玩家进入（让其他玩家知道）
			cm.BroadcastToMap(client.MapID, &Message{
				From: client.ID,
				Type: CmdEnterMap,
				Data: mustMarshal(map[string]interface{}{
					"role_id": client.ID,
					"map_id":  client.MapID,
					"x":       client.X,
					"y":       client.Y,
					"name":    client.Name,
				}),
			})

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
	client.Registered = true // 只有成功添加到列表后才标记为已注册
}

func (cm *ClientManager) removeClient(client *Client) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if _, ok := cm.clients[client.ID]; ok {
		log.Printf("removeClient: 移除玩家 %d, 当前地图=%d", client.ID, client.MapID)
		delete(cm.clients, client.ID)
		delete(cm.byConn, client.Conn)
		close(client.Send)
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

// GetOnlineCount 获取在线人数
func (cm *ClientManager) GetOnlineCount() int {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	return len(cm.clients)
}

// ClientAuth 客户端认证
type ClientAuth struct {
	Token  string `json:"token"`
	RoleID uint64 `json:"role_id"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Type     string `json:"type"` // "guest", "account"
	Username string `json:"username"`
	Password string `json:"password"`
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
		Send:       make(chan []byte, 256),
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
	log.Printf("收到消息: cmd=%d, bodyLen=%d", cmd, len(body))
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
		c.sendPacket(CmdLogin, mustMarshal(LoginResponse{Code: 200, Msg: "登录成功", RoleID: roleID}))

	default:
		c.sendPacket(CmdLogin, mustMarshal(LoginResponse{Code: 400, Msg: "无效的登录类型"}))
	}
}

func (c *Client) handleEnterMap(body []byte) {
	var req struct {
		MapID uint32 `json:"map_id"`
		X     int    `json:"x"`
		Y     int    `json:"y"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		log.Printf("handleEnterMap解析失败: %v", err)
		return
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

	// 如果还未注册，现在注册
	if !c.Registered {
		GlobalManager.register <- c
		log.Printf("客户端注册: roleID=%d, mapID=%d", c.ID, c.MapID)
	}

	// 先向新玩家发送当前地图的其他玩家列表
	existingPlayers := GlobalManager.GetMapPlayers(req.MapID)
	log.Printf("当前地图%d有%d个玩家", req.MapID, len(existingPlayers))

	if len(existingPlayers) > 1 { // 有其他玩家
		playerList := make([]map[string]interface{}, 0)
		for _, p := range existingPlayers {
			if p.ID != c.ID {
				playerList = append(playerList, map[string]interface{}{
					"role_id": p.ID,
					"name":    p.Name,
					"map_id":  p.MapID,
					"x":       p.X,
					"y":       p.Y,
				})
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

	// 广播新玩家进入给当前地图的其他玩家
	log.Printf("广播玩家进入: mapID=%d, 玩家ID=%d, 玩家名=%s", c.MapID, c.ID, c.Name)
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

	// 广播移动给同地图玩家
	GlobalManager.BroadcastToMap(c.MapID, &Message{
		From: c.ID,
		Type: CmdMove,
		Data: mustMarshal(map[string]interface{}{
			"role_id": c.ID,
			"x":       c.X,
			"y":       c.Y,
		}),
	})
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
		"from_name": "玩家",
		"channel":   chat.Channel,
		"content":   chat.Content,
		"time":      time.Now().Unix(),
	})

	switch chat.Channel {
	case 0: // 世界
		GlobalManager.broadcast <- &Message{Type: CmdChat, Data: msg}
	case 1: // 当前地图
		GlobalManager.BroadcastToMap(c.MapID, &Message{From: c.ID, Type: CmdChat, Data: msg})
	case 3: // 私聊
		GlobalManager.broadcast <- &Message{To: chat.ToID, Type: CmdChat, Data: msg}
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

func main() {
	log.Println("===== 网关服务启动 =====")
	log.Println("WebSocket: :8080")

	http.HandleFunc("/ws", HandleWebSocket)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok","online":` + fmt.Sprintf("%d", GlobalManager.GetOnlineCount()) + `}`))
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
