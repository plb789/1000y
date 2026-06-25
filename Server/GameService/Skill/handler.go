package skill

import (
	"log"
	"net/http"
	"strconv"

	attribute "game-server/GameService/Attribute"
	cache "game-server/GameService/Cache"

	"github.com/gin-gonic/gin"
)

// Handler HTTP请求处理
type Handler struct {
	service     *Service
	playerCache *cache.PlayerCache // ★ 新增：玩家状态缓存引用
}

// NewHandler 创建Handler实例
func NewHandler() *Handler {
	return &Handler{
		service: NewService(),
	}
}

// SetPlayerCache 注入PlayerCache（在main.go中初始化后调用）
// ★ 使用依赖注入模式，避免循环依赖
func (h *Handler) SetPlayerCache(pc *cache.PlayerCache) {
	h.playerCache = pc
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	skillGroup := r.Group("/api/skill")
	{
		// 武学基础信息
		skillGroup.GET("/base/list", h.GetAllSkillBase)          // 获取所有武学列表
		skillGroup.GET("/base/type/:type", h.GetSkillBaseByType) // 获取指定类型武学
		skillGroup.GET("/base/:id", h.GetSkillBase)              // 获取武学详情
		skillGroup.GET("/type/list", h.SkillTypeList)            // 获取武学类型列表

		// 角色武学
		skillGroup.GET("/role/:roleId/list", h.GetRoleSkills)                        // 获取角色所有武学
		skillGroup.GET("/role/:roleId/type/:type", h.GetRoleSkillsByType)            // 获取角色指定类型武学
		skillGroup.GET("/role/:roleId/equipped", h.GetEquippedSkills)                // 获取已装备武学
		skillGroup.GET("/role/:roleId/bonus", h.GetSkillBonus)                       // 获取武学加成
		skillGroup.GET("/role/:roleId/exp_progress/:skillId", h.GetSkillExpProgress) // 获取熟练度进度

		// 武学操作
		skillGroup.POST("/role/:roleId/learn", h.LearnSkill)     // 学习武学
		skillGroup.POST("/role/:roleId/add_exp", h.AddExp)       // 增加熟练度
		skillGroup.POST("/role/:roleId/upgrade", h.UpgradeSkill) // 升级武学
		skillGroup.POST("/role/:roleId/equip", h.EquipSkill)     // 装备武学
		skillGroup.POST("/role/:roleId/unequip", h.UnequipSkill) // 卸下武学
		skillGroup.POST("/role/:roleId/forget", h.ForgetSkill)   // 遗忘武学
	}
}

// GetAllSkillBase 获取所有武学列表
func (h *Handler) GetAllSkillBase(c *gin.Context) {
	skills, err := h.service.GetAllSkillBase()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "获取武学列表失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": skills})
}

// GetSkillBaseByType 获取指定类型武学
func (h *Handler) GetSkillBaseByType(c *gin.Context) {
	typeStr := c.Param("type")
	skillType, err := strconv.ParseUint(typeStr, 10, 8)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的武学类型"})
		return
	}

	skills, err := h.service.GetSkillBaseByType(uint8(skillType))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "获取武学列表失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": skills})
}

// GetSkillBase 获取武学详情
func (h *Handler) GetSkillBase(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的武学ID"})
		return
	}

	skill, err := h.service.GetSkillBase(uint32(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "武学不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": skill})
}

// GetRoleSkills 获取角色所有武学
func (h *Handler) GetRoleSkills(c *gin.Context) {
	roleIdStr := c.Param("roleId")
	roleId, err := strconv.ParseUint(roleIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	skills, err := h.service.GetRoleSkills(roleId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "获取角色武学失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": skills})
}

// GetRoleSkillsByType 获取角色指定类型武学
func (h *Handler) GetRoleSkillsByType(c *gin.Context) {
	roleIdStr := c.Param("roleId")
	roleId, err := strconv.ParseUint(roleIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	typeStr := c.Param("type")
	skillType, err := strconv.ParseUint(typeStr, 10, 8)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的武学类型"})
		return
	}

	skills, err := h.service.GetRoleSkillsByType(roleId, uint8(skillType))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "获取角色武学失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": skills})
}

// GetEquippedSkills 获取已装备武学
func (h *Handler) GetEquippedSkills(c *gin.Context) {
	roleIdStr := c.Param("roleId")
	roleId, err := strconv.ParseUint(roleIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	skills, err := h.service.GetEquippedSkills(roleId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "获取已装备武学失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": skills})
}

// GetSkillBonus 获取武学加成
func (h *Handler) GetSkillBonus(c *gin.Context) {
	roleIdStr := c.Param("roleId")
	roleId, err := strconv.ParseUint(roleIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	bonus, err := h.service.CalculateSkillBonus(roleId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "计算武学加成失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": bonus})
}

// GetSkillExpProgress 获取熟练度进度
func (h *Handler) GetSkillExpProgress(c *gin.Context) {
	roleIdStr := c.Param("roleId")
	roleId, err := strconv.ParseUint(roleIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	skillIdStr := c.Param("skillId")
	skillId, err := strconv.ParseUint(skillIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的武学ID"})
		return
	}

	currentExp, expNeeded, level, maxLevel, err := h.service.GetSkillExpProgress(roleId, uint32(skillId))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": gin.H{
			"current_exp": currentExp,
			"exp_needed":  expNeeded,
			"level":       level,
			"max_level":   maxLevel,
		},
	})
}

// LearnSkillRequest 学习武学请求
type LearnSkillRequest struct {
	SkillID   uint32 `json:"skill_id" binding:"required"`
	RoleLevel uint32 `json:"role_level"` // 角色等级用于校验
}

// LearnSkill 学习武学
func (h *Handler) LearnSkill(c *gin.Context) {
	roleIdStr := c.Param("roleId")
	roleId, err := strconv.ParseUint(roleIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	var req LearnSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	// 检查角色等级是否满足武学学习条件
	canLearn, err := h.service.CanLearnSkillByLevel(req.RoleLevel, req.SkillID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}
	if !canLearn {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "角色等级不足"})
		return
	}

	if err := h.service.LearnSkill(roleId, req.SkillID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "学习成功"})
}

// AddExpRequest 增加熟练度请求
type AddExpRequest struct {
	SkillID uint32 `json:"skill_id" binding:"required"`
	Exp     int64  `json:"exp" binding:"required"`
}

// AddExp 增加熟练度
func (h *Handler) AddExp(c *gin.Context) {
	roleIdStr := c.Param("roleId")
	roleId, err := strconv.ParseUint(roleIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	var req AddExpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	leveledUp, newLevel, err := h.service.AddExp(roleId, req.SkillID, req.Exp)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":       200,
		"msg":        "success",
		"leveled_up": leveledUp,
		"new_level":  newLevel,
	})
}

// UpgradeSkillRequest 升级武学请求
type UpgradeSkillRequest struct {
	SkillID uint32 `json:"skill_id" binding:"required"`
}

// UpgradeSkill 升级武学
func (h *Handler) UpgradeSkill(c *gin.Context) {
	roleIdStr := c.Param("roleId")
	roleId, err := strconv.ParseUint(roleIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	var req UpgradeSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	newLevel, err := h.service.UpgradeSkill(roleId, req.SkillID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "升级成功", "new_level": newLevel})
}

// EquipSkillRequest 装备武学请求
type EquipSkillRequest struct {
	SkillID uint32 `json:"skill_id" binding:"required"`
}

// EquipSkill 装备武学
// ★ 完整流程：DB操作 → 缓存失效 → 重新计算属性 → 返回完整数据
func (h *Handler) EquipSkill(c *gin.Context) {
	roleIdStr := c.Param("roleId")
	roleId, err := strconv.ParseUint(roleIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	var req EquipSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	// 1. 执行装备操作（写DB）
	if err := h.service.EquipSkill(roleId, req.SkillID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	// 2. ★ 使旧缓存失效（强制下次重新计算）
	if h.playerCache != nil {
		if invalidateErr := h.playerCache.Invalidate(c.Request.Context(), roleId); invalidateErr != nil {
			log.Printf("⚠️ 使缓存失败失败: %v", invalidateErr)
			// 失败不阻塞主流程
		}
	}

	// 3. ★ 重新计算并获取最新完整属性（含装备+技能+BUFF加成）
	var finalAttrs *attribute.Attribute
	var bonusDetail *attribute.AttributeBonus

	if h.playerCache != nil {
		playerState, calcErr := h.playerCache.GetOrLoad(c.Request.Context(), roleId)
		if calcErr != nil {
			log.Printf("⚠️ 获取最新属性失败: %v", calcErr)
			// 降级：仅返回技能加成（不包含完整的最终属性）
			bonus, _ := h.service.CalculateSkillBonus(roleId)
			c.JSON(http.StatusOK, gin.H{
				"code": 200,
				"msg":  "装备成功",
				"data": gin.H{
					"skill_id":     req.SkillID,
					"bonus":        bonus,
					"final_attrs":  nil,
					"bonus_detail": nil,
				},
			})
			return
		}
		finalAttrs = playerState.FinalAttributes
		bonusDetail = playerState.BonusDetail
	} else {
		// 未配置缓存时，使用旧的简单逻辑
		bonus, _ := h.service.CalculateSkillBonus(roleId)
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"msg":  "装备成功",
			"data": gin.H{
				"skill_id": req.SkillID,
				"bonus":    bonus,
			},
		})
		return
	}

	// 4. ★ 返回完整的最新属性（供前端立即更新所有UI面板）
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "装备成功",
		"data": gin.H{
			"skill_id":     req.SkillID,
			"final_attrs":  finalAttrs,  // ★ 最终属性（HP/MP/攻击等）
			"bonus_detail": bonusDetail, // ★ 加成明细（[武+5][装+15]等）
			"equipped_skills": func() []map[string]interface{} {
				if finalAttrs != nil && h.playerCache != nil {
					state, _ := h.playerCache.GetOrLoad(c.Request.Context(), roleId)
					if state != nil {
						return state.EquippedSkills
					}
				}
				return nil
			}(),
		},
	})
}

// UnequipSkill 卸下武学
// ★ 完整流程：DB操作 → 缓存失效 → 重新计算属性 → 返回完整数据
func (h *Handler) UnequipSkill(c *gin.Context) {
	roleIdStr := c.Param("roleId")
	roleId, err := strconv.ParseUint(roleIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	var req EquipSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	// 1. 执行卸载操作（写DB）
	if err := h.service.UnequipSkill(roleId, req.SkillID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	// 2. ★ 使旧缓存失效（强制下次重新计算）
	if h.playerCache != nil {
		if invalidateErr := h.playerCache.Invalidate(c.Request.Context(), roleId); invalidateErr != nil {
			log.Printf("⚠️ 使缓存失败失败: %v", invalidateErr)
			// 失败不阻塞主流程
		}
	}

	// 3. ★ 重新计算并获取最新完整属性
	var finalAttrs *attribute.Attribute
	var bonusDetail *attribute.AttributeBonus

	if h.playerCache != nil {
		playerState, calcErr := h.playerCache.GetOrLoad(c.Request.Context(), roleId)
		if calcErr != nil {
			log.Printf("⚠️ 获取最新属性失败: %v", calcErr)
			bonus, _ := h.service.CalculateSkillBonus(roleId)
			c.JSON(http.StatusOK, gin.H{
				"code": 200,
				"msg":  "卸下成功",
				"data": gin.H{
					"skill_id":     req.SkillID,
					"bonus":        bonus,
					"final_attrs":  nil,
					"bonus_detail": nil,
				},
			})
			return
		}
		finalAttrs = playerState.FinalAttributes
		bonusDetail = playerState.BonusDetail
	} else {
		bonus, _ := h.service.CalculateSkillBonus(roleId)
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"msg":  "卸下成功",
			"data": gin.H{
				"skill_id": req.SkillID,
				"bonus":    bonus,
			},
		})
		return
	}

	// 4. ★ 返回完整的最新属性（供前端立即更新所有UI面板）
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "卸下成功",
		"data": gin.H{
			"skill_id":     req.SkillID,
			"final_attrs":  finalAttrs,
			"bonus_detail": bonusDetail,
			"equipped_skills": func() []map[string]interface{} {
				if finalAttrs != nil && h.playerCache != nil {
					state, _ := h.playerCache.GetOrLoad(c.Request.Context(), roleId)
					if state != nil {
						return state.EquippedSkills
					}
				}
				return nil
			}(),
		},
	})
}

// ForgetSkill 遗忘武学
func (h *Handler) ForgetSkill(c *gin.Context) {
	roleIdStr := c.Param("roleId")
	roleId, err := strconv.ParseUint(roleIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	var req EquipSkillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	if err := h.service.ForgetSkill(roleId, req.SkillID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "遗忘成功"})
}

// SkillTypeList 武学类型列表(用于前端展示)
func (h *Handler) SkillTypeList(c *gin.Context) {
	typeList := []gin.H{
		{"type": 1, "name": "内功", "description": "提升生命、内力上限和内力恢复"},
		{"type": 2, "name": "外功", "description": "提升攻击力和暴击"},
		{"type": 3, "name": "身法", "description": "提升速度和闪避"},
		{"type": 4, "name": "护体", "description": "提升防御和生命"},
		{"type": 5, "name": "拳法", "description": "徒手武学,拳拳到肉"},
		{"type": 6, "name": "剑法", "description": "剑意无双,剑气伤人"},
		{"type": 7, "name": "刀法", "description": "刀势威猛,刀刀见血"},
		{"type": 8, "name": "枪法", "description": "枪如游龙,气势如虹"},
		{"type": 9, "name": "斧法", "description": "力沉势猛,开山裂石"},
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": typeList})
}

// SkillInfoWithTypeName 武学信息(带类型名称)
type SkillInfoWithTypeName struct {
	SkillBase
	TypeName    string `json:"type_name"`
	SubTypeName string `json:"sub_type_name"`
}

// GetAllSkillBaseWithTypeName 获取所有武学(带类型名称)
func (h *Handler) GetAllSkillBaseWithTypeName(c *gin.Context) {
	skills, err := h.service.GetAllSkillBase()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "获取武学列表失败"})
		return
	}

	result := make([]map[string]interface{}, len(skills))
	for i, skill := range skills {
		result[i] = skill
		if t, ok := skill["type"].(uint8); ok {
			result[i]["type_name"] = SkillTypeName[t]
		}
		if st, ok := skill["sub_type"].(uint8); ok {
			result[i]["sub_type_name"] = SkillSubTypeName[st]
		}
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": result})
}
