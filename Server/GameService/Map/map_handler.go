package gamemap

import (
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
	mapGroup := r.Group("/api/map")
	{
		// 地图信息
		mapGroup.GET("/list", h.GetAllMaps)             // 获取所有地图
		mapGroup.GET("/:id", h.GetMap)                  // 获取地图详情
		mapGroup.GET("/:id/info", h.GetMapInfo)         // 获取地图信息(含在线人数)
		mapGroup.GET("/:id/players", h.GetPlayersOnMap) // 获取地图上玩家列表

		// 地图操作
		mapGroup.POST("/:id/load", h.LoadMap)           // 加载地图
		mapGroup.POST("/:id/collision", h.SetCollision) // 设置碰撞
		mapGroup.POST("/enter", h.EnterMap)             // 进入地图
		mapGroup.POST("/leave", h.LeaveMap)             // 离开地图
		mapGroup.POST("/teleport", h.Teleport)          // 传送

		// 位置验证
		mapGroup.GET("/:id/can_move/:x/:y", h.CanMove) // 检查移动是否允许

		// NPC和怪物
		mapGroup.GET("/:id/npcs", h.GetNPCs)         // 获取地图NPC
		mapGroup.GET("/:id/monsters", h.GetMonsters) // 获取地图怪物
	}
}

// GetAllMaps 获取所有地图
func (h *Handler) GetAllMaps(c *gin.Context) {
	maps, err := h.service.GetAllMaps()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "获取地图列表失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": maps})
}

// GetMap 获取地图详情
func (h *Handler) GetMap(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的地图ID"})
		return
	}

	mapBase, err := h.service.GetMapBase(uint32(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "地图不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": mapBase})
}

// GetMapInfo 获取地图信息(含在线人数)
func (h *Handler) GetMapInfo(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的地图ID"})
		return
	}

	mapBase, playerCount, err := h.service.GetMapInfo(uint32(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "地图不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": gin.H{
			"map":          mapBase,
			"player_count": playerCount,
		},
	})
}

// GetPlayersOnMap 获取地图上玩家列表
func (h *Handler) GetPlayersOnMap(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的地图ID"})
		return
	}

	// 获取视野范围内的玩家,默认全图
	players := h.service.GetPlayersInView(uint32(id), 0, 0, 10000)
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": players})
}

// LoadMap 加载地图
func (h *Handler) LoadMap(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的地图ID"})
		return
	}

	lm, err := h.service.LoadMap(uint32(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "加载成功",
		"data": gin.H{
			"map_id":      lm.MapData.ID,
			"name":        lm.MapData.Name,
			"tile_width":  lm.TileWidth,
			"tile_height": lm.TileHeight,
			"width":       lm.MapData.Width,
			"height":      lm.MapData.Height,
		},
	})
}

// SetCollision 设置碰撞
func (h *Handler) SetCollision(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的地图ID"})
		return
	}

	var req struct {
		X       int  `json:"x" binding:"required"`
		Y       int  `json:"y" binding:"required"`
		Blocked bool `json:"blocked"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	if err := h.service.SetCollision(uint32(id), req.X, req.Y, req.Blocked); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "设置成功"})
}

// EnterMap 进入地图
func (h *Handler) EnterMap(c *gin.Context) {
	var req struct {
		RoleID uint64 `json:"role_id" binding:"required"`
		MapID  uint32 `json:"map_id" binding:"required"`
		X      int    `json:"x" binding:"required"`
		Y      int    `json:"y" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	if err := h.service.EnterMap(req.RoleID, req.MapID, req.X, req.Y); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "进入成功"})
}

// LeaveMap 离开地图
func (h *Handler) LeaveMap(c *gin.Context) {
	var req struct {
		RoleID uint64 `json:"role_id" binding:"required"`
		MapID  uint32 `json:"map_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	if err := h.service.LeaveMap(req.RoleID, req.MapID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "离开成功"})
}

// Teleport 传送
func (h *Handler) Teleport(c *gin.Context) {
	var req struct {
		RoleID    uint64 `json:"role_id" binding:"required"`
		FromMapID uint32 `json:"from_map_id" binding:"required"`
		ToMapID   uint32 `json:"to_map_id" binding:"required"`
		X         int    `json:"x" binding:"required"`
		Y         int    `json:"y" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	if err := h.service.TeleportPlayer(req.RoleID, req.FromMapID, req.ToMapID, req.X, req.Y); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "传送成功"})
}

// CanMove 检查移动是否允许
func (h *Handler) CanMove(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的地图ID"})
		return
	}

	xStr := c.Param("x")
	x, err := strconv.Atoi(xStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的X坐标"})
		return
	}

	yStr := c.Param("y")
	y, err := strconv.Atoi(yStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的Y坐标"})
		return
	}

	canMove := h.service.CanMove(uint32(id), x, y)
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "can_move": canMove})
}

// GetNPCs 获取地图NPC
func (h *Handler) GetNPCs(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的地图ID"})
		return
	}

	npcs, err := h.service.GetNPCsByMap(uint32(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "获取NPC列表失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": npcs})
}

// GetMonsters 获取地图怪物
func (h *Handler) GetMonsters(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的地图ID"})
		return
	}

	monsters, err := h.service.GetMonstersByMap(uint32(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "获取怪物列表失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": monsters})
}
