package quest

import (
	common "game-server/Common"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handler HTTP请求处理
type Handler struct {
	service *Service
}

// NewHandler 创建Handler实例
func NewHandler() *Handler {
	return &Handler{
		service: NewService(),
	}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	questGroup := r.Group("/api/quest")
	{
		questGroup.GET("/list/:roleId", h.GetQuestList) // 获取任务列表
		questGroup.POST("/accept", h.AcceptQuest)       // 接取任务
		questGroup.POST("/complete", h.CompleteQuest)   // 完成任务（领取奖励）
		questGroup.POST("/abandon", h.AbandonQuest)     // 放弃任务
		questGroup.POST("/progress", h.UpdateProgress)  // 更新进度
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

	list, err := h.service.GetQuestList(roleId, playerLevel)
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

	info, err := h.service.AcceptQuest(req.RoleID, req.QuestID, playerLevel)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

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

	info, err := h.service.CompleteQuest(req.RoleID, req.QuestID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

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

	err := h.service.AbandonQuest(req.RoleID, req.QuestID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "放弃成功",
	})
}

// UpdateProgress 更新任务进度
// POST /api/quest/progress
// Body: { role_id: xxx, quest_id: xxx, target_type: x, target_id: x, count?: x }
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

	update, err := h.service.UpdateProgress(req.RoleID, req.QuestID, req.TargetType, req.TargetID, req.Count)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "进度更新成功",
		"data": update,
	})
}
