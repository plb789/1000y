package achievement

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler 成就处理器
type Handler struct {
	Service *Service
}

// NewHandler 创建成就处理器
func NewHandler() *Handler {
	return &Handler{
		Service: GetAchievementService(),
	}
}

// RegisterRoutes 注册成就相关路由
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	// 在 /api/quest 下注册成就路由
	apiQuestGroup := r.Group("/api/quest")
	{
		apiQuestGroup.GET("/achievement/list/:roleId", h.GetAchievementList)   // 获取成就列表
		apiQuestGroup.GET("/achievement/stats/:roleId", h.GetAchievementStats) // 获取成就统计
	}
}

// GetAchievementList 获取玩家成就列表
// GET /api/quest/achievement/list/:roleId
func (h *Handler) GetAchievementList(c *gin.Context) {
	roleIDStr := c.Param("roleId")
	var roleID uint64
	if _, err := parseUint64(roleIDStr, &roleID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	infos, err := h.Service.GetRoleAchievements(roleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "获取成就列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": infos,
	})
}

// GetAchievementStats 获取玩家成就统计
// GET /api/quest/achievement/stats/:roleId
func (h *Handler) GetAchievementStats(c *gin.Context) {
	roleIDStr := c.Param("roleId")
	var roleID uint64
	if _, err := parseUint64(roleIDStr, &roleID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	stats, err := h.Service.GetAchievementStats(roleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "获取成就统计失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": stats,
	})
}

// parseUint64 解析uint64
func parseUint64(s string, result *uint64) (int, error) {
	var v uint64
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0, nil
		}
		v = v*10 + uint64(s[i]-'0')
	}
	*result = v
	return 1, nil
}
