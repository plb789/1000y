package battle

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Handler HTTP请求处理
type Handler struct {
	service      *Service
	gatewayURL   string // Gateway服务地址（用于推送WebSocket消息）
	gatewayMutex sync.RWMutex
	client       *http.Client // HTTP客户端（复用连接）
}

// NewHandler 创建Handler实例
func NewHandler(gatewayURL string) *Handler {
	return &Handler{
		service:    NewService(),
		gatewayURL: gatewayURL,
		client:     &http.Client{Timeout: 5 * time.Second},
	}
}

// SetGatewayURL 设置Gateway服务地址
func (h *Handler) SetGatewayURL(url string) {
	h.gatewayMutex.Lock()
	defer h.gatewayMutex.Unlock()
	h.gatewayURL = url
}

// GetGatewayURL 获取Gateway服务地址
func (h *Handler) GetGatewayURL() string {
	h.gatewayMutex.RLock()
	defer h.gatewayMutex.RUnlock()
	return h.gatewayURL
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	battleGroup := r.Group("/api/battle")
	{
		// 战斗相关
		battleGroup.POST("/damage", h.HandleDamage)   // 处理伤害
		battleGroup.POST("/death", h.HandleDeath)     // 处理死亡
		battleGroup.POST("/respawn", h.HandleRespawn) // 处理复活
		battleGroup.POST("/levelup", h.HandleLevelUp) // 处理升级
		battleGroup.POST("/buff", h.HandleBuff)       // 处理增益
		battleGroup.POST("/debuff", h.HandleDeBuff)   // 处理减益
		battleGroup.POST("/event", h.HandleMapEvent)  // 处理地图事件
	}
}

// HandleDamage 处理伤害
func (h *Handler) HandleDamage(c *gin.Context) {
	var req DamageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	// 处理伤害逻辑
	result := h.service.ProcessDamage(req)

	// 推送给目标玩家
	h.pushToClient(req.TargetID, "damage", result)

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": result})
}

// HandleDeath 处理死亡
func (h *Handler) HandleDeath(c *gin.Context) {
	var req DeathRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	// 处理死亡逻辑
	result := h.service.ProcessDeath(req)

	// 推送给目标玩家
	h.pushToClient(req.TargetID, "death", result)

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": result})
}

// HandleRespawn 处理复活
func (h *Handler) HandleRespawn(c *gin.Context) {
	var req RespawnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	// 处理复活逻辑
	result := h.service.ProcessRespawn(req)

	// 推送给玩家
	h.pushToClient(req.RoleID, "respawn", result)

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": result})
}

// HandleLevelUp 处理升级
func (h *Handler) HandleLevelUp(c *gin.Context) {
	var req LevelUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	// 处理升级逻辑
	result := h.service.ProcessLevelUp(req)

	// 推送给玩家
	h.pushToClient(req.RoleID, "levelup", result)

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": result})
}

// HandleBuff 处理增益
func (h *Handler) HandleBuff(c *gin.Context) {
	var req BuffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	// 处理增益逻辑
	result := h.service.ProcessBuff(req)

	// 推送给玩家
	h.pushToClient(req.TargetID, "buff", result)

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": result})
}

// HandleDeBuff 处理减益
func (h *Handler) HandleDeBuff(c *gin.Context) {
	var req DeBuffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	// 处理减益逻辑
	result := h.service.ProcessDeBuff(req)

	// 推送给玩家
	h.pushToClient(req.TargetID, "debuff", result)

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": result})
}

// HandleMapEvent 处理地图事件
func (h *Handler) HandleMapEvent(c *gin.Context) {
	var req MapEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	// 处理地图事件逻辑
	result := h.service.ProcessMapEvent(req)

	// 广播给地图所有玩家
	h.broadcastToMap(req.MapID, "mapevent", result)

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": result})
}

// pushToClient 推送消息给客户端（通过Gateway）
func (h *Handler) pushToClient(roleID uint64, msgType string, data interface{}) {
	gatewayURL := h.GetGatewayURL()
	if gatewayURL == "" {
		log.Printf("Gateway URL未配置，跳过推送")
		return
	}

	// 构建推送请求
	pushReq := map[string]interface{}{
		"role_id":  roleID,
		"msg_type": msgType,
		"data":     data,
	}

	jsonData, err := json.Marshal(pushReq)
	if err != nil {
		log.Printf("序列化推送数据失败: %v", err)
		return
	}

	// 异步发送HTTP请求到Gateway（复用client）
	go func() {
		resp, err := h.client.Post(gatewayURL+"/internal/push", "application/json", bytes.NewReader(jsonData))
		if err != nil {
			log.Printf("推送消息到Gateway失败: %v", err)
			return
		}
		defer resp.Body.Close()
	}()
}

// broadcastToMap 广播消息给地图所有玩家（通过Gateway）
func (h *Handler) broadcastToMap(mapID uint32, msgType string, data interface{}) {
	gatewayURL := h.GetGatewayURL()
	if gatewayURL == "" {
		log.Printf("Gateway URL未配置，跳过广播")
		return
	}

	// 构建广播请求
	broadcastReq := map[string]interface{}{
		"map_id":   mapID,
		"msg_type": msgType,
		"data":     data,
	}

	jsonData, err := json.Marshal(broadcastReq)
	if err != nil {
		log.Printf("序列化广播数据失败: %v", err)
		return
	}

	// 异步发送HTTP请求到Gateway（复用client）
	go func() {
		resp, err := h.client.Post(gatewayURL+"/internal/broadcast", "application/json", bytes.NewReader(jsonData))
		if err != nil {
			log.Printf("广播消息到Gateway失败: %v", err)
			return
		}
		defer resp.Body.Close()
	}()
}
