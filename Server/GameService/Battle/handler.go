package battle

import (
	"bytes"
	"encoding/json"
	common "game-server/Common"
	monster "game-server/GameService/Monster"
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
		battleGroup.POST("/attack", h.HandleAttack)   // 玩家攻击怪物
		battleGroup.POST("/damage", h.HandleDamage)   // 处理伤害
		battleGroup.POST("/death", h.HandleDeath)     // 处理死亡
		battleGroup.POST("/respawn", h.HandleRespawn) // 处理复活
		battleGroup.POST("/levelup", h.HandleLevelUp) // 处理升级
		battleGroup.POST("/buff", h.HandleBuff)       // 处理增益
		battleGroup.POST("/debuff", h.HandleDeBuff)   // 处理减益
		battleGroup.POST("/event", h.HandleMapEvent)  // 处理地图事件
	}
}

// HandleAttack 处理玩家攻击怪物请求
func (h *Handler) HandleAttack(c *gin.Context) {
	var req AttackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	log.Printf("⚔️ 收到攻击请求: roleID=%d, monsterID=%d, skillID=%d",
		req.AttackerID, req.TargetID, req.SkillID)

	// 调用战斗服务处理攻击
	result := h.service.PlayerAttackMonster(req.AttackerID, req.TargetID, req.SkillID)

	if result == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "战斗计算失败",
			"data": nil,
		})
		return
	}

	if !result.Success && result.ErrorCode > 0 {
		// 攻击失败（距离过远、冷却中等）
		c.JSON(http.StatusOK, gin.H{
			"code": result.ErrorCode,
			"msg":  result.ErrorMsg,
			"data": result,
		})
		return
	}

	// 攻击成功，返回完整结果
	log.Printf("✅ 攻击结果: 怪物%d 受到%d点伤害 (暴击=%v, 闪避=%v, 死亡=%v)",
		req.TargetID, result.Damage, result.IsCrit, result.IsMiss, result.IsDead)

	// 推送战斗结果给客户端（通过Gateway WebSocket）
	h.pushToClient(req.AttackerID, "attack_result", result)

	// 如果怪物死亡，广播给同地图所有玩家
	if result.IsDead {
		deathNotice := map[string]interface{}{
			"monster_id": req.TargetID,
			"killer_id":  req.AttackerID,
			"exp_gain":   result.ExpGain,
			"gold_gain":  result.GoldGain,
			"drops":      result.Drops,
		}
		h.broadcastToMap(getPlayerMapID(req.AttackerID), "monster_death", deathNotice)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": result,
	})
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

// BroadcastMonsterPositions 广播怪物位置更新给地图所有玩家
func (h *Handler) BroadcastMonsterPositions(mapID uint32, data interface{}) {
	log.Printf("📡 广播怪物位置: 地图%d, %d个怪物", mapID, len(data.(gin.H)["monsters"].([]monster.MonsterPositionInfo)))
	h.broadcastToMap(mapID, "monster_position_update", data)
}

// getPlayerMapID 获取玩家当前地图ID
func getPlayerMapID(roleID uint64) uint32 {
	pos, err := common.DBGetRolePosition(roleID)
	if err != nil || pos == nil {
		return 1 // 默认返回新手村
	}
	return pos.MapID
}
