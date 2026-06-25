package quest

import (
	"bytes"
	"encoding/json"
	"fmt"
	common "game-server/Common"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Handler HTTP请求处理
type Handler struct {
	Service      *Service
	gatewayURL   string // Gateway服务地址（用于推送WebSocket消息）
	gatewayMutex sync.RWMutex
	client       *http.Client // HTTP客户端（复用连接）
}

// NewHandler 创建Handler实例
func NewHandler(gatewayURL string) *Handler {
	return &Handler{
		Service:    NewService(),
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
	questGroup := r.Group("/api/quest")
	{
		questGroup.GET("/list/:roleId", h.GetQuestList)              // 获取任务列表
		questGroup.GET("/detail/:roleId/:questId", h.GetQuestDetail) // 获取任务详情
		questGroup.POST("/accept", h.AcceptQuest)                    // 接取任务
		questGroup.POST("/complete", h.CompleteQuest)                // 完成任务（领取奖励）
		questGroup.POST("/abandon", h.AbandonQuest)                  // 放弃任务
		questGroup.POST("/progress", h.UpdateProgress)               // 更新进度
		questGroup.POST("/auto-accept", h.AutoAcceptQuests)          // 自动接取任务
		questGroup.POST("/monster-killed", h.OnMonsterKilled)        // 怪物击杀事件
		questGroup.POST("/item-gathered", h.OnItemGathered)          // 物品采集事件
		questGroup.POST("/npc-dialog", h.OnNPCDialog)                // NPC对话事件
		questGroup.GET("/guide/:roleId", h.GetQuestGuide)            // 获取任务指引（目标位置）
		questGroup.POST("/batch-accept", h.BatchAcceptQuests)        // 批量接取任务
	}
}

// GetQuestList 获取任务列表
// GET /api/quest/list/:roleId
func (h *Handler) GetQuestList(c *gin.Context) {
	roleIdStr := c.Param("roleId")
	roleId, err := strconv.ParseUint(roleIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	// 从数据库获取玩家等级
	playerLevel := uint32(1)
	if roleInfo, err := common.DBRoleGet(roleId); err == nil && roleInfo != nil {
		playerLevel = uint32(roleInfo.Level)
	}

	list, err := h.Service.GetQuestList(roleId, playerLevel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "获取任务列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": list})
}

// AcceptQuest 接取任务
// POST /api/quest/accept
// Body: { role_id: xxx, quest_id: xxx }
func (h *Handler) AcceptQuest(c *gin.Context) {
	var req AcceptQuestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	// 从数据库获取玩家等级
	playerLevel := uint32(1)
	if roleInfo, err := common.DBRoleGet(req.RoleID); err == nil && roleInfo != nil {
		playerLevel = uint32(roleInfo.Level)
	}

	info, err := h.Service.AcceptQuest(req.RoleID, req.QuestID, playerLevel)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	// WebSocket 推送任务更新
	h.BroadcastQuestAccept(req.RoleID, info)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "接取成功",
		"data": info,
	})
}

// CompleteQuest 完成任务（领取奖励）
// POST /api/quest/complete
// Body: { role_id: xxx, quest_id: xxx }
func (h *Handler) CompleteQuest(c *gin.Context) {
	var req CompleteQuestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	info, err := h.Service.CompleteQuest(req.RoleID, req.QuestID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	// WebSocket 推送任务更新
	h.BroadcastQuestComplete(req.RoleID, info)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "奖励已领取",
		"data": info,
	})
}

// AbandonQuest 放弃任务
// POST /api/quest/abandon
// Body: { role_id: xxx, quest_id: xxx }
func (h *Handler) AbandonQuest(c *gin.Context) {
	var req AbandonQuestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	err := h.Service.AbandonQuest(req.RoleID, req.QuestID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	// WebSocket 推送任务更新
	h.BroadcastQuestAbandon(req.RoleID, req.QuestID)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "放弃成功",
	})
}

// UpdateProgress 更新任务进度
// POST /api/quest/progress
// Body: { role_id: xxx, quest_id: xxx, target_type: x, target_id: x, objective_id?: x, count?: x }
func (h *Handler) UpdateProgress(c *gin.Context) {
	var req UpdateProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	// 默认增加1个进度
	if req.Count <= 0 {
		req.Count = 1
	}

	update, err := h.Service.UpdateProgress(req.RoleID, req.QuestID, req.TargetType, req.TargetID, req.ObjectiveID, req.Count)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	// WebSocket 推送进度更新
	h.BroadcastQuestProgress(req.RoleID, update)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "进度更新成功",
		"data": update,
	})
}

// AutoAcceptQuests 自动接取任务
// POST /api/quest/auto-accept
// Body: { role_id: xxx }
func (h *Handler) AutoAcceptQuests(c *gin.Context) {
	var req struct {
		RoleID uint64 `json:"role_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	// 从数据库获取玩家等级
	playerLevel := uint32(1)
	if roleInfo, err := common.DBRoleGet(req.RoleID); err == nil && roleInfo != nil {
		playerLevel = uint32(roleInfo.Level)
	}

	accepted := h.Service.AutoAcceptQuests(req.RoleID, playerLevel)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "自动接取成功",
		"data": accepted,
	})
}

// BatchAcceptQuestsRequest 批量接取任务请求
type BatchAcceptQuestsRequest struct {
	RoleID   uint64   `json:"role_id" binding:"required"`
	QuestIDs []uint32 `json:"quest_ids"` // 指定任务ID列表，为空则接取所有可接任务
}

// BatchAcceptQuests 批量接取任务
// POST /api/quest/batch-accept
// Body: { role_id: xxx, quest_ids: [1,2,3] } quest_ids为空则接取所有可接任务
func (h *Handler) BatchAcceptQuests(c *gin.Context) {
	var req BatchAcceptQuestsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	// 从数据库获取玩家等级
	playerLevel := uint32(1)
	if roleInfo, err := common.DBRoleGet(req.RoleID); err == nil && roleInfo != nil {
		playerLevel = uint32(roleInfo.Level)
	}

	// 获取可接取的任务列表
	list, err := h.Service.GetQuestList(req.RoleID, playerLevel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "获取任务列表失败"})
		return
	}

	var accepted []*QuestInfo
	if len(req.QuestIDs) > 0 {
		// 指定任务ID，接取指定任务
		for _, quest := range list.AvailableQuests {
			for _, questID := range req.QuestIDs {
				if quest.ID == questID {
					info, err := h.Service.AcceptQuest(req.RoleID, quest.ID, playerLevel)
					if err == nil && info != nil {
						accepted = append(accepted, info)
					}
					break
				}
			}
		}
	} else {
		// 未指定任务ID，接取所有可接任务
		for _, quest := range list.AvailableQuests {
			info, err := h.Service.AcceptQuest(req.RoleID, quest.ID, playerLevel)
			if err == nil && info != nil {
				accepted = append(accepted, info)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  fmt.Sprintf("成功接取%d个任务", len(accepted)),
		"data": accepted,
	})
}

// OnMonsterKilled 怪物击杀事件
// POST /api/quest/monster-killed
// Body: { role_id: xxx, monster_id: xxx }
func (h *Handler) OnMonsterKilled(c *gin.Context) {
	var req struct {
		RoleID    uint64 `json:"role_id" binding:"required"`
		MonsterID uint32 `json:"monster_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	updates := h.Service.OnMonsterKilled(req.RoleID, req.MonsterID)

	// WebSocket 推送所有受影响的进度更新
	for _, update := range updates {
		h.BroadcastQuestProgress(req.RoleID, update)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "任务进度已更新",
		"data": updates,
	})
}

// OnItemGathered 物品采集事件
// POST /api/quest/item-gathered
// Body: { role_id: xxx, item_id: xxx }
func (h *Handler) OnItemGathered(c *gin.Context) {
	var req struct {
		RoleID uint64 `json:"role_id" binding:"required"`
		ItemID uint32 `json:"item_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	updates := h.Service.OnItemGathered(req.RoleID, req.ItemID)

	// WebSocket 推送所有受影响的进度更新
	for _, update := range updates {
		h.BroadcastQuestProgress(req.RoleID, update)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "任务进度已更新",
		"data": updates,
	})
}

// OnNPCDialog NPC对话事件
// POST /api/quest/npc-dialog
// Body: { role_id: xxx, npc_id: xxx }
func (h *Handler) OnNPCDialog(c *gin.Context) {
	var req struct {
		RoleID uint64 `json:"role_id" binding:"required"`
		NPCID  uint32 `json:"npc_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	updates := h.Service.OnNPCDialog(req.RoleID, req.NPCID)

	// WebSocket 推送所有受影响的进度更新
	for _, update := range updates {
		h.BroadcastQuestProgress(req.RoleID, update)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "任务进度已更新",
		"data": updates,
	})
}

// GetQuestDetail 获取任务详情
// GET /api/quest/detail/:roleId/:questId
func (h *Handler) GetQuestDetail(c *gin.Context) {
	roleIdStr := c.Param("roleId")
	questIdStr := c.Param("questId")

	roleId, err := strconv.ParseUint(roleIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	questId, err := strconv.ParseUint(questIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的任务ID"})
		return
	}

	// 从数据库获取玩家等级
	playerLevel := uint32(1)
	if roleInfo, err := common.DBRoleGet(roleId); err == nil && roleInfo != nil {
		playerLevel = uint32(roleInfo.Level)
	}

	list, err := h.Service.GetQuestList(roleId, playerLevel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "获取任务详情失败"})
		return
	}

	// 查找指定任务
	for _, quest := range list.ActiveQuests {
		if quest.ID == uint32(questId) {
			c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": quest})
			return
		}
	}
	for _, quest := range list.CompletedQuests {
		if quest.ID == uint32(questId) {
			c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": quest})
			return
		}
	}
	for _, quest := range list.AvailableQuests {
		if quest.ID == uint32(questId) {
			c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": quest})
			return
		}
	}
	for _, quest := range list.FinishedQuests {
		if quest.ID == uint32(questId) {
			c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": quest})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "任务不存在"})
}

// QuestGuide 任务指引信息
type QuestGuide struct {
	QuestID     uint32 `json:"quest_id"`    // 任务ID
	QuestName   string `json:"quest_name"`  // 任务名称
	TargetType  uint8  `json:"target_type"` // 目标类型: 1=杀怪, 2=采集, 3=对话, 4=到达
	TargetID    uint32 `json:"target_id"`   // 目标ID (怪物ID/物品ID/NPCID/地图ID)
	TargetName  string `json:"target_name"` // 目标名称
	MapID       uint32 `json:"map_id"`      // 所在地图ID
	MapName     string `json:"map_name"`    // 地图名称
	X           int    `json:"x"`           // 目标X坐标
	Y           int    `json:"y"`           // 目标Y坐标
	Description string `json:"description"` // 指引描述
}

// GetQuestGuide 获取任务指引（用于自动寻路）
// GET /api/quest/guide/:roleId
func (h *Handler) GetQuestGuide(c *gin.Context) {
	roleIdStr := c.Param("roleId")
	roleId, err := strconv.ParseUint(roleIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	// 从数据库获取玩家等级
	playerLevel := uint32(1)
	if roleInfo, err := common.DBRoleGet(roleId); err == nil && roleInfo != nil {
		playerLevel = uint32(roleInfo.Level)
	}

	// 获取玩家的活动任务
	list, err := h.Service.GetQuestList(roleId, playerLevel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "获取任务列表失败"})
		return
	}

	// 构建指引信息列表
	var guides []QuestGuide
	addGuide := func(quest QuestInfo) {
		config := &quest.QuestBaseConfig

		// 根据目标类型获取指引
		var targetType uint8 = config.TargetType
		var targetID uint32 = config.TargetID
		var targetName string = config.TargetName
		var mapID uint32 = 0
		var mapName string = ""
		var x, y int = 0, 0

		// 如果任务有多个目标，取第一个未完成的目标
		if len(quest.Objectives) > 0 {
			for _, obj := range quest.Objectives {
				if int(obj.Progress) < int(obj.TargetCount) {
					targetType = obj.TargetType
					targetID = obj.TargetID
					targetName = obj.TargetName
					// 根据目标类型查找位置
					switch targetType {
					case 1: // 杀怪
						if monsterConfig := common.GetMonsterConfig(targetID); monsterConfig != nil {
							mapID = monsterConfig.MapID
							mapName = getMapName(mapID)
							// 怪物坐标需要通过MonsterBaseConfig获取，此处简化处理
						}
					case 2: // 采集
						// 采集物位置需要额外配置支持
					case 3: // NPC对话
						if npcConfig := common.GetNPCConfig(targetID); npcConfig != nil {
							mapID = npcConfig.MapID
							mapName = getMapName(mapID)
							x, y = int(npcConfig.X), int(npcConfig.Y)
						}
					case 4: // 到达
						mapID = targetID
						mapName = getMapName(mapID)
					}
					break // 只取第一个未完成目标
				}
			}
		}

		description := fmt.Sprintf("去%s击杀%s", mapName, targetName)
		if targetType == 2 {
			description = fmt.Sprintf("去%s采集%s", mapName, targetName)
		} else if targetType == 3 {
			description = fmt.Sprintf("去%s与%s对话", mapName, targetName)
		} else if targetType == 4 {
			description = fmt.Sprintf("前往%s", mapName)
		}

		guides = append(guides, QuestGuide{
			QuestID:     quest.ID,
			QuestName:   config.Name,
			TargetType:  targetType,
			TargetID:    targetID,
			TargetName:  targetName,
			MapID:       mapID,
			MapName:     mapName,
			X:           x,
			Y:           y,
			Description: description,
		})
	}

	// 遍历活动任务
	for _, quest := range list.ActiveQuests {
		addGuide(quest)
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": guides,
	})
}

// QuestPushType 任务推送类型
const (
	QuestPushProgress = 1 // 进度更新
	QuestPushAccept   = 2 // 接取任务
	QuestPushComplete = 3 // 完成任务
	QuestPushAbandon  = 4 // 放弃任务
	QuestPushReset    = 5 // 任务重置
)

// QuestPushData 任务推送数据
type QuestPushData struct {
	Type   uint8       `json:"type"`    // 推送类型
	RoleID uint64      `json:"role_id"` // 角色ID
	Data   interface{} `json:"data"`    // 推送数据
}

// BroadcastQuestUpdate 广播任务更新给客户端（通过MessageBus，支持分布式架构）
func (h *Handler) BroadcastQuestUpdate(roleID uint64, pushType uint8, data interface{}) {
	pushData := QuestPushData{
		Type:   pushType,
		RoleID: roleID,
		Data:   data,
	}

	// 使用 MessageBus 发布任务更新消息（支持 RabbitMQ 分布式）
	// 路由键格式: quest.push.{roleId} 用于精确推送
	topic := "quest.push"
	err := common.GlobalMessageBus.Publish(topic, pushData)
	if err != nil {
		log.Printf("[QUEST] 发布任务更新消息失败: %v", err)
		// 如果 MessageBus 不可用，尝试 HTTP 降级
		h.broadcastViaHTTP(roleID, pushData)
	}
}

// broadcastViaHTTP HTTP降级推送（当 RabbitMQ 不可用时使用）
func (h *Handler) broadcastViaHTTP(roleID uint64, pushData QuestPushData) {
	gatewayURL := h.GetGatewayURL()
	if gatewayURL == "" {
		return
	}

	jsonData, err := json.Marshal(pushData)
	if err != nil {
		log.Printf("[QUEST] 序列化推送数据失败: %v", err)
		return
	}

	url := gatewayURL + "/internal/quest-push"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("[QUEST] 创建推送请求失败: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		log.Printf("[QUEST] HTTP降级推送失败: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[QUEST] HTTP降级推送失败，状态码: %d", resp.StatusCode)
	}
}

// BroadcastQuestProgress 广播任务进度更新
func (h *Handler) BroadcastQuestProgress(roleID uint64, update *QuestProgressUpdate) {
	h.BroadcastQuestUpdate(roleID, QuestPushProgress, update)
}

// BroadcastQuestAccept 广播接取任务
func (h *Handler) BroadcastQuestAccept(roleID uint64, quest *QuestInfo) {
	h.BroadcastQuestUpdate(roleID, QuestPushAccept, quest)
}

// BroadcastQuestComplete 广播完成任务
func (h *Handler) BroadcastQuestComplete(roleID uint64, quest *QuestInfo) {
	h.BroadcastQuestUpdate(roleID, QuestPushComplete, quest)
}

// BroadcastQuestAbandon 广播放弃任务
func (h *Handler) BroadcastQuestAbandon(roleID uint64, questID uint32) {
	h.BroadcastQuestUpdate(roleID, QuestPushAbandon, map[string]interface{}{"quest_id": questID})
}

// BroadcastQuestReset 广播任务重置（日常/周常）
func (h *Handler) BroadcastQuestReset(roleID uint64, questType uint8, resetQuests []*QuestInfo) {
	h.BroadcastQuestUpdate(roleID, QuestPushReset, map[string]interface{}{
		"quest_type":   questType,
		"reset_quests": resetQuests,
	})
}

// getMapName 根据地图ID获取地图名称
func getMapName(mapID uint32) string {
	if mapID == 0 {
		return "未知地图"
	}
	if mapConfig := common.GetMapConfig(mapID); mapConfig != nil {
		return mapConfig.Name
	}
	return fmt.Sprintf("地图%d", mapID)
}
