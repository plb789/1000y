package role

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
	roleGroup := r.Group("/api/role")
	{
		// 角色创建与查询
		roleGroup.POST("/create", h.CreateRole)                        // 创建角色
		roleGroup.GET("/account/:accountId/list", h.GetRolesByAccount) // 获取账号下角色列表
		roleGroup.GET("/:id", h.GetRoleByID)                           // 获取角色详情
		roleGroup.GET("/name/:name", h.GetRoleByName)                  // 根据名称获取角色

		// 角色更新
		roleGroup.POST("/:id/update", h.UpdateRole)               // 更新角色信息
		roleGroup.POST("/:id/attributes", h.UpdateRoleAttributes) // 更新角色属性
		roleGroup.POST("/:id/delete", h.DeleteRole)               // 删除角色

		// 属性操作
		roleGroup.POST("/:id/add_exp", h.AddExp)               // 增加经验
		roleGroup.POST("/:id/add_gold", h.AddGold)             // 增加金币
		roleGroup.POST("/:id/consume_gold", h.ConsumeGold)     // 消耗金币
		roleGroup.POST("/:id/change_hp", h.ChangeHP)           // 改变生命
		roleGroup.POST("/:id/change_mp", h.ChangeMP)           // 改变内力
		roleGroup.POST("/:id/change_stamina", h.ChangeStamina) // 改变体力
		roleGroup.POST("/:id/recovery", h.FullRecovery)        // 完全恢复
		roleGroup.POST("/:id/save", h.SaveRole)                // 保存角色

		// 位置与状态
		roleGroup.POST("/:id/change_map", h.ChangeMap)    // 切换地图
		roleGroup.POST("/:id/position", h.UpdatePosition) // 更新位置
		roleGroup.POST("/:id/status", h.SetStatus)        // 设置状态
		roleGroup.POST("/:id/pk_mode", h.SetPKMode)       // 设置PK模式

		// 在线管理
		roleGroup.POST("/:id/login", h.Login)            // 登录
		roleGroup.POST("/:id/logout", h.Logout)          // 登出
		roleGroup.GET("/online/count", h.GetOnlineCount) // 获取在线人数
		roleGroup.GET("/online/list", h.GetOnlineList)   // 获取在线玩家列表

		// PK相关
		roleGroup.POST("/:id/record_kill", h.RecordKill)   // 记录击杀
		roleGroup.POST("/:id/record_death", h.RecordDeath) // 记录死亡
		roleGroup.POST("/:id/pk_value", h.UpdatePkValue)   // 更新PK值
	}
}

// CreateRole 创建角色
func (h *Handler) CreateRole(c *gin.Context) {
	var req RoleCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误: " + err.Error()})
		return
	}

	role, err := h.service.CreateRole(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "创建成功", "data": role})
}

// GetRoleByID 获取角色详情
func (h *Handler) GetRoleByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	role, err := h.service.GetRoleByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "角色不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": role})
}

// GetRoleByName 根据名称获取角色
func (h *Handler) GetRoleByName(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "角色名不能为空"})
		return
	}

	role, err := h.service.GetRoleByName(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "角色不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": role})
}

// GetRolesByAccount 获取账号下角色列表
func (h *Handler) GetRolesByAccount(c *gin.Context) {
	accountIdStr := c.Param("accountId")
	accountId, err := strconv.ParseUint(accountIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的账号ID"})
		return
	}

	roles, err := h.service.GetRolesByAccount(accountId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "获取角色列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": roles})
}

// UpdateRole 更新角色信息
func (h *Handler) UpdateRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	var req RoleUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	if err := h.service.UpdateRole(id, req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "更新成功"})
}

// UpdateRoleAttributes 批量更新角色属性
func (h *Handler) UpdateRoleAttributes(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	var req RoleAttributeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	if err := h.service.UpdateRoleAttributes(id, req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "更新成功"})
}

// DeleteRole 删除角色
func (h *Handler) DeleteRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	// 验证Token
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "未提供Token"})
		return
	}

	claims, err := common.ValidateToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "Token验证失败: " + err.Error()})
		return
	}

	// 需要验证账号所有权
	accountIDStr := c.GetHeader("X-Account-ID")
	accountID, err := strconv.ParseUint(accountIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的账号ID"})
		return
	}

	// 验证Token中的UID与请求中的accountID一致
	if claims.UID != accountID {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "msg": "无权删除此角色"})
		return
	}

	if err := h.service.DeleteRole(id, accountID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "删除成功"})
}

// AddExp 增加经验
func (h *Handler) AddExp(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	var req struct {
		Exp int64 `json:"exp" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	leveledUp, newLevel, newExp, err := h.service.AddExp(id, req.Exp)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":       200,
		"msg":        "success",
		"leveled_up": leveledUp,
		"new_level":  newLevel,
		"new_exp":    newExp,
	})
}

// AddGold 增加金币
func (h *Handler) AddGold(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	var req struct {
		Gold int64 `json:"gold" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	if err := h.service.AddGold(id, req.Gold); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "增加成功"})
}

// ConsumeGold 消耗金币
func (h *Handler) ConsumeGold(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	var req struct {
		Gold int64 `json:"gold" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	if err := h.service.ConsumeGold(id, req.Gold); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "消耗成功"})
}

// ChangeHP 改变生命
func (h *Handler) ChangeHP(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	var req struct {
		Change int `json:"change" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	newHP, err := h.service.ChangeHP(id, req.Change)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "new_hp": newHP})
}

// ChangeMP 改变内力
func (h *Handler) ChangeMP(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	var req struct {
		Change int `json:"change" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	newMP, err := h.service.ChangeMP(id, req.Change)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "new_mp": newMP})
}

// ChangeStamina 改变体力
func (h *Handler) ChangeStamina(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	var req struct {
		Change int `json:"change" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	newStamina, err := h.service.ChangeStamina(id, req.Change)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "new_stamina": newStamina})
}

// FullRecovery 完全恢复
func (h *Handler) FullRecovery(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	if err := h.service.FullRecovery(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "恢复成功"})
}

// SaveRole 保存角色
func (h *Handler) SaveRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	if err := h.service.SaveRole(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "保存成功"})
}

// ChangeMap 切换地图
func (h *Handler) ChangeMap(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	var req struct {
		MapID int `json:"map_id" binding:"required"`
		X     int `json:"x" binding:"required"`
		Y     int `json:"y" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	if err := h.service.ChangeMap(id, req.MapID, req.X, req.Y); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "切换成功"})
}

// UpdatePosition 更新位置
func (h *Handler) UpdatePosition(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	var req struct {
		X int `json:"x" binding:"required"`
		Y int `json:"y" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	if err := h.service.UpdatePosition(id, req.X, req.Y); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	// 同时更新在线玩家位置
	h.service.UpdatePlayerPosition(id, req.X, req.Y)

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "更新成功"})
}

// SetStatus 设置状态
func (h *Handler) SetStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	var req struct {
		Status uint8 `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	if err := h.service.SetStatus(id, req.Status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "设置成功"})
}

// SetPKMode 设置PK模式
func (h *Handler) SetPKMode(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	var req struct {
		Mode uint8 `json:"mode" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	if err := h.service.SetPKMode(id, req.Mode); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "设置成功"})
}

// Login 登录
func (h *Handler) Login(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	ip := c.ClientIP()

	// 记录登录
	if err := h.service.LoginRecord(id, ip); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	// 添加到在线列表
	player := h.service.PlayerLogin(id)

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "登录成功", "data": player})
}

// Logout 登出
func (h *Handler) Logout(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	// 记录登出
	if err := h.service.LogoutRecord(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	// 从在线列表移除
	h.service.PlayerLogout(id)

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "登出成功"})
}

// GetOnlineCount 获取在线人数
func (h *Handler) GetOnlineCount(c *gin.Context) {
	count := h.service.GetOnlineCount()
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": gin.H{"count": count}})
}

// GetOnlineList 获取在线玩家列表
func (h *Handler) GetOnlineList(c *gin.Context) {
	players := h.service.GetAllOnlinePlayers()
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": players})
}

// RecordKill 记录击杀
func (h *Handler) RecordKill(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	if err := h.service.RecordKill(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "记录成功"})
}

// RecordDeath 记录死亡
func (h *Handler) RecordDeath(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	if err := h.service.RecordDeath(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "记录成功"})
}

// UpdatePkValue 更新PK值
func (h *Handler) UpdatePkValue(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	var req struct {
		Change int `json:"change" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	if err := h.service.UpdatePkValue(id, req.Change); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "更新成功"})
}
