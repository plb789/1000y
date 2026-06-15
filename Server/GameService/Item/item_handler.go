package item

import (
	"game-server/GameService/Item/model"
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
	itemGroup := r.Group("/api/item")
	{
		// 道具基础信息
		itemGroup.GET("/base/list", h.GetAllItems)                  // 获取所有道具
		itemGroup.GET("/base/type/:type", h.GetItemsByType)        // 获取指定类型道具
		itemGroup.GET("/base/:id", h.GetItemBase)                   // 获取道具详情

		// 背包管理
		itemGroup.GET("/bag/:roleId/list", h.GetBagItems)           // 获取背包物品
		itemGroup.GET("/bag/:roleId/item/:grid", h.GetBagItem)    // 获取指定格子物品
		itemGroup.GET("/bag/:roleId/empty_count", h.GetEmptySlotCount) // 获取空位数

		// 物品操作
		itemGroup.POST("/bag/add", h.AddItem)                      // 添加物品
		itemGroup.POST("/bag/:roleId/move", h.MoveItem)             // 移动物品
		itemGroup.POST("/bag/:roleId/split", h.SplitItem)           // 拆分物品
		itemGroup.POST("/bag/:roleId/use", h.UseItem)               // 使用物品
		itemGroup.POST("/bag/:roleId/discard", h.DiscardItem)       // 丢弃物品
		itemGroup.POST("/bag/:roleId/sell", h.SellItem)             // 出售物品
		itemGroup.POST("/bag/:roleId/destroy", h.DestroyItem)       // 销毁物品
		itemGroup.POST("/bag/:roleId/clear", h.ClearBag)           // 清空背包

		// 装备管理
		itemGroup.GET("/equip/:roleId/list", h.GetEquippedItems)   // 获取已装备列表
		itemGroup.GET("/equip/:roleId/type/:type", h.GetEquippedItemByType) // 获取指定位置装备
		itemGroup.POST("/equip/:roleId/equip", h.EquipItem)         // 穿戴装备
		itemGroup.POST("/equip/:roleId/unequip", h.UnequipItem)    // 卸下装备
	}
}

// GetAllItems 获取所有道具
func (h *Handler) GetAllItems(c *gin.Context) {
	items, err := h.service.GetAllItems()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "获取道具列表失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": items})
}

// GetItemsByType 获取指定类型道具
func (h *Handler) GetItemsByType(c *gin.Context) {
	typeStr := c.Param("type")
	itemType, err := strconv.ParseUint(typeStr, 10, 8)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的道具类型"})
		return
	}

	items, err := h.service.GetItemsByType(uint8(itemType))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "获取道具列表失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": items})
}

// GetItemBase 获取道具详情
func (h *Handler) GetItemBase(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的道具ID"})
		return
	}

	item, err := h.service.GetItemBase(uint32(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "道具不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": item})
}

// GetBagItems 获取背包物品
func (h *Handler) GetBagItems(c *gin.Context) {
	roleIdStr := c.Param("roleId")
	roleId, err := strconv.ParseUint(roleIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	items, err := h.service.GetBagItems(roleId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "获取背包失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": items})
}

// GetBagItem 获取指定格子物品
func (h *Handler) GetBagItem(c *gin.Context) {
	roleIdStr := c.Param("roleId")
	roleId, err := strconv.ParseUint(roleIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	gridStr := c.Param("grid")
	grid, err := strconv.Atoi(gridStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的格子索引"})
		return
	}

	item, err := h.service.GetBagItemByGrid(roleId, grid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "格子为空"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": item})
}

// GetEmptySlotCount 获取空位数
func (h *Handler) GetEmptySlotCount(c *gin.Context) {
	roleIdStr := c.Param("roleId")
	roleId, err := strconv.ParseUint(roleIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	count, err := h.service.GetEmptySlotCount(roleId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "获取空位数失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": gin.H{"empty_count": count}})
}

// AddItem 添加物品
func (h *Handler) AddItem(c *gin.Context) {
	var req model.AddItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	if req.Count <= 0 {
		req.Count = 1
	}

	slot, err := h.service.AddItem(req.RoleID, req.ItemID, req.Count, req.IsBind)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "添加成功", "slot": slot})
}

// MoveItem 移动物品
func (h *Handler) MoveItem(c *gin.Context) {
	roleIdStr := c.Param("roleId")
	roleId, err := strconv.ParseUint(roleIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	var req struct {
		FromGrid int `json:"from_grid" binding:"required"`
		ToGrid   int `json:"to_grid" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	if err := h.service.MoveItem(roleId, req.FromGrid, req.ToGrid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "移动成功"})
}

// SplitItem 拆分物品
func (h *Handler) SplitItem(c *gin.Context) {
	roleIdStr := c.Param("roleId")
	roleId, err := strconv.ParseUint(roleIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	var req model.SplitItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	if err := h.service.SplitItem(roleId, req.GridIndex, req.Count); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "拆分成功"})
}

// UseItem 使用物品
func (h *Handler) UseItem(c *gin.Context) {
	roleIdStr := c.Param("roleId")
	roleId, err := strconv.ParseUint(roleIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	var req model.UseItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	if err := h.service.UseItem(roleId, req.GridIndex); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "使用成功"})
}

// DiscardItem 丢弃物品
func (h *Handler) DiscardItem(c *gin.Context) {
	roleIdStr := c.Param("roleId")
	roleId, err := strconv.ParseUint(roleIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	var req struct {
		GridIndex int `json:"grid_index" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	if err := h.service.DiscardItem(roleId, req.GridIndex); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "丢弃成功"})
}

// SellItem 出售物品
func (h *Handler) SellItem(c *gin.Context) {
	roleIdStr := c.Param("roleId")
	roleId, err := strconv.ParseUint(roleIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	var req struct {
		GridIndex int `json:"grid_index" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	price, err := h.service.SellItem(roleId, req.GridIndex)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "出售成功", "price": price})
}

// DestroyItem 销毁物品
func (h *Handler) DestroyItem(c *gin.Context) {
	roleIdStr := c.Param("roleId")
	roleId, err := strconv.ParseUint(roleIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	var req struct {
		GridIndex int `json:"grid_index" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	if err := h.service.DestroyItem(roleId, req.GridIndex); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "销毁成功"})
}

// ClearBag 清空背包
func (h *Handler) ClearBag(c *gin.Context) {
	roleIdStr := c.Param("roleId")
	roleId, err := strconv.ParseUint(roleIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	if err := h.service.ClearBag(roleId); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "清空成功"})
}

// GetEquippedItems 获取已装备列表
func (h *Handler) GetEquippedItems(c *gin.Context) {
	roleIdStr := c.Param("roleId")
	roleId, err := strconv.ParseUint(roleIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	equips, err := h.service.GetEquippedItems(roleId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "获取装备列表失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": equips})
}

// GetEquippedItemByType 获取指定位置装备
func (h *Handler) GetEquippedItemByType(c *gin.Context) {
	roleIdStr := c.Param("roleId")
	roleId, err := strconv.ParseUint(roleIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	typeStr := c.Param("type")
	equipType, err := strconv.ParseUint(typeStr, 10, 8)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的装备位置"})
		return
	}

	equip, err := h.service.GetEquipmentByType(roleId, uint8(equipType))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "该位置没有装备"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": equip})
}

// EquipItem 穿戴装备
func (h *Handler) EquipItem(c *gin.Context) {
	roleIdStr := c.Param("roleId")
	roleId, err := strconv.ParseUint(roleIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	var req model.EquipItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	if err := h.service.EquipItem(roleId, req.BagItemID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "穿戴成功"})
}

// UnequipItem 卸下装备
func (h *Handler) UnequipItem(c *gin.Context) {
	roleIdStr := c.Param("roleId")
	roleId, err := strconv.ParseUint(roleIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的角色ID"})
		return
	}

	var req struct {
		EquipType uint8 `json:"equip_type" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	if err := h.service.UnequipItem(roleId, req.EquipType); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "卸下成功"})
}
