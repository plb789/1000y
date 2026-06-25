package gamemap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	common "game-server/Common"
	monster "game-server/GameService/Monster"

	"github.com/gin-gonic/gin"
)

// Handler HTTP请求处理
type Handler struct {
	service    *Service
	instanceID uint32
	httpClient *http.Client
}

// NewHandler 创建Handler实例
func NewHandler(instanceID uint32) *Handler {
	return &Handler{
		service:    GetService(), // 使用全局单例
		instanceID: instanceID,
		httpClient: &http.Client{Timeout: 10 * time.Second},
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
		mapGroup.POST("/move", h.Move)                  // 移动

		// 位置验证
		mapGroup.GET("/:id/can_move/:x/:y", h.CanMove) // 检查移动是否允许

		// NPC和怪物
		mapGroup.GET("/:id/npcs", h.GetNPCs)         // 获取地图NPC
		mapGroup.GET("/:id/monsters", h.GetMonsters) // 获取地图怪物

		// ★ 按区块加载地图数据（方案B按需加载）
		mapGroup.GET("/:id/chunk_info", h.GetChunkInfo)   // 获取区块划分信息（必须放在 :id/chunk 之前避免冲突）
		mapGroup.GET("/:id/chunk/:cx/:cy", h.GetMapChunk) // 获取指定区块的二进制数据
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
		X      int    `json:"x"` // 可选，为0时从数据库获取
		Y      int    `json:"y"` // 可选，为0时从数据库获取
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		// 读取body用于调试
		body, _ := io.ReadAll(c.Request.Body)
		log.Printf("EnterMap: 参数绑定失败 - %v, body=%s", err, string(body))
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	log.Printf("EnterMap请求: roleID=%d, mapID=%d, x=%d, y=%d", req.RoleID, req.MapID, req.X, req.Y)

	// 如果X或Y为0，从数据库获取位置
	tileX, tileY := req.X, req.Y

	if err := h.service.EnterMap(req.RoleID, req.MapID, tileX, tileY); err != nil {
		log.Printf("EnterMap: 玩家 %d 进入地图 %d 失败 - %v", req.RoleID, req.MapID, err)
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	// 返回实际的坐标（可能是从数据库读取的）
	actualX, actualY := tileX, tileY
	if tileX == 0 && tileY == 0 {
		// 从数据库读取了位置
		position, err := common.DBGetRolePosition(req.RoleID)
		if err == nil && position != nil {
			actualX, actualY = position.X, position.Y
		}
	}

	// 获取玩家信息用于广播
	lm, _ := h.service.GetLoadedMap(req.MapID)
	if lm != nil {
		lm.mu.RLock()
		if player, ok := lm.Players[req.RoleID]; ok {
			actualX, actualY = player.X, player.Y
		}
		lm.mu.RUnlock()
	}

	// 获取怪物列表（用于客户端显示）
	monsterList := h.getMonsterListForMap(req.MapID)
	log.Printf("EnterMap: 地图%d 怪物数量=%d", req.MapID, len(monsterList))

	c.JSON(http.StatusOK, gin.H{
		"code":         200,
		"msg":          "进入成功",
		"x":            actualX,
		"y":            actualY,
		"monster_list": monsterList,
	})
}

// LeaveMap 离开地图（保存下线位置）
func (h *Handler) LeaveMap(c *gin.Context) {
	var req struct {
		RoleID uint64 `json:"role_id" binding:"required"`
		MapID  uint32 `json:"map_id" binding:"required"`
		X      int    `json:"x"` // 可选：客户端提供的坐标
		Y      int    `json:"y"` // 可选：客户端提供的坐标
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	// 如果客户端提供了坐标，优先使用客户端坐标
	if req.X != 0 || req.Y != 0 {
		if err := h.service.LeaveMapWithPosition(req.RoleID, req.MapID, req.X, req.Y); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
			return
		}
	} else {
		// 否则使用服务端保存的坐标
		if err := h.service.LeaveMap(req.RoleID, req.MapID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
			return
		}
	}

	log.Printf("💾 玩家 %d 离开地图 %d，位置 (%d, %d)", req.RoleID, req.MapID, req.X, req.Y)
	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "离开成功，位置已保存"})
}

// Move 处理玩家移动
func (h *Handler) Move(c *gin.Context) {
	var req struct {
		RoleID uint64 `json:"role_id" binding:"required"`
		MapID  uint32 `json:"map_id" binding:"required"`
		X      int    `json:"x" binding:"required"`
		Y      int    `json:"y" binding:"required"`
	}

	// 读取原始请求体用于调试
	bodyBytes, _ := io.ReadAll(c.Request.Body)
	log.Printf("📥 Move: 收到原始请求体: %s", string(bodyBytes))

	// 重新设置请求体（因为已经读取了）
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("❌ Move: 参数绑定失败 - %v", err)
		log.Printf("❌ Move: 期望格式: {role_id: number, map_id: number, x: number, y: number}")
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "请求参数错误"})
		return
	}

	log.Printf("✅ Move: 参数绑定成功 - roleID=%d, mapID=%d, x=%d, y=%d",
		req.RoleID, req.MapID, req.X, req.Y)

	// 移动玩家
	result, err := h.service.MovePlayer(req.RoleID, req.MapID, req.X, req.Y)
	if err != nil {
		log.Printf("Move: 移动失败 - roleID=%d, mapID=%d, x=%d, y=%d, error=%v",
			req.RoleID, req.MapID, req.X, req.Y, err)
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	// 如果移动成功，广播给同地图玩家
	if result.Success {
		log.Printf("Move: 移动成功 - roleID=%d, mapID=%d, x=%d, y=%d",
			req.RoleID, req.MapID, req.X, req.Y)
		go h.broadcastMove(req.MapID, req.RoleID, req.X, req.Y)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"msg":     "success",
		"success": result.Success,
		"x":       result.X,
		"y":       result.Y,
	})
}

// broadcastMove 广播移动消息给同地图玩家
func (h *Handler) broadcastMove(mapID uint32, roleID uint64, x, y int) {
	moveData := map[string]interface{}{
		"role_id": roleID,
		"x":       x,
		"y":       y,
		"map_id":  mapID,
	}

	// 使用消息总线广播
	if common.GlobalMessageBus != nil && common.GlobalMessageBus.IsAvailable() {
		go func() {
			if err := common.GlobalMessageBus.Publish("map_move", moveData); err != nil {
				log.Printf("消息总线广播失败，降级到HTTP: %v", err)
				h.fallbackBroadcastMove(moveData)
			}
		}()
	} else {
		h.fallbackBroadcastMove(moveData)
	}
}

// fallbackBroadcastMove 降级到HTTP广播
func (h *Handler) fallbackBroadcastMove(moveData map[string]interface{}) {
	gatewayURL := common.GetGatewayServiceURL()
	if gatewayURL == "" {
		log.Printf("Gateway URL未配置，跳过移动广播")
		return
	}

	jsonData, err := json.Marshal(moveData)
	if err != nil {
		log.Printf("序列化移动数据失败: %v", err)
		return
	}

	go func() {
		resp, err := h.httpClient.Post(gatewayURL+"/internal/broadcast_map", "application/json", bytes.NewReader(jsonData))
		if err != nil {
			log.Printf("广播移动消息失败: %v", err)
			return
		}
		defer resp.Body.Close()
	}()
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

	// 检查目标地图是否由当前服务处理
	if h.isMapHandledLocally(req.ToMapID) {
		// 本地传送
		if err := h.service.TeleportPlayer(req.RoleID, req.FromMapID, req.ToMapID, req.X, req.Y); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "传送成功"})
	} else {
		// 跨服务传送
		err := h.teleportToRemoteService(req.RoleID, req.FromMapID, req.ToMapID, req.X, req.Y)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "跨服传送成功"})
	}
}

// isMapHandledLocally 检查地图是否由当前服务处理
func (h *Handler) isMapHandledLocally(mapID uint32) bool {
	return common.HandleMapInRegistry(h.instanceID, mapID)
}

// teleportToRemoteService 跨服务传送
func (h *Handler) teleportToRemoteService(roleID uint64, fromMapID, toMapID uint32, x, y int) error {
	// 1. 查询目标地图所在的服务实例
	targetInst := common.GetInstanceByMapID(toMapID)
	if targetInst == nil {
		return fmt.Errorf("目标地图 %d 未找到处理服务", toMapID)
	}

	log.Printf("跨服传送: 玩家 %d 从地图 %d 传送至地图 %d(服务: %s)", roleID, fromMapID, toMapID, targetInst.URL)

	// 2. 获取玩家数据
	role, err := common.DBRoleGet(roleID)
	if err != nil || role == nil {
		return fmt.Errorf("获取玩家数据失败: %v", err)
	}

	// 3. 离开原地图
	if err := h.service.LeaveMap(roleID, fromMapID); err != nil {
		return fmt.Errorf("离开原地图失败: %v", err)
	}

	// 4. 构造传送请求数据
	teleportReq := map[string]interface{}{
		"role_id":     roleID,
		"from_map_id": fromMapID,
		"to_map_id":   toMapID,
		"x":           x,
		"y":           y,
		"player_data": map[string]interface{}{
			"id":     role.ID,
			"name":   role.Name,
			"level":  role.Level,
			"hp":     role.Hp,
			"max_hp": role.MaxHp,
			"mp":     role.Mp,
			"max_mp": role.MaxMp,
			"x":      role.MapX,
			"y":      role.MapY,
			"map_id": role.MapID,
		},
	}

	// 5. 调用目标服务的传送接口
	jsonData, err := json.Marshal(teleportReq)
	if err != nil {
		return fmt.Errorf("序列化传送数据失败: %v", err)
	}

	resp, err := h.httpClient.Post(targetInst.URL+"/api/map/teleport", "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("调用目标服务传送接口失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("目标服务传送失败，状态码: %d", resp.StatusCode)
	}

	log.Printf("跨服传送成功: 玩家 %d 已传送到地图 %d", roleID, toMapID)
	return nil
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

// GetMonsters 获取地图怪物（返回运行时实例数据）
func (h *Handler) GetMonsters(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的地图ID"})
		return
	}

	// 调用 Monster Service 获取运行时实例（包含坐标、HP等动态数据）
	monsterSvc := monster.GetService()
	if monsterSvc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "怪物服务未初始化"})
		return
	}

	monsters := monsterSvc.GetMonstersByMap(uint32(id))
	if monsters == nil || len(monsters) == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": []interface{}{}})
		return
	}

	// 统一转换为客户端友好的格式（避免 Go 结构体默认的大写字段名）
	result := make([]interface{}, 0, len(monsters))
	for _, m := range monsters {
		if m.Status == 4 { // 跳过死亡怪物
			continue
		}
		result = append(result, map[string]interface{}{
			"id":      m.ID,
			"base_id": m.BaseID,
			"name":    m.Name,
			"x":       m.X,
			"y":       m.Y,
			"hp":      m.CurrentHP,
			"max_hp":  m.MaxHP,
			"level":   m.Level,
			"type":    m.Type,
			"status":  m.Status,
		})
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "success", "data": result})
}

// getMonsterListForMap 获取指定地图的怪物列表（用于EnterMap响应）
func (h *Handler) getMonsterListForMap(mapID uint32) []interface{} {
	monsterSvc := monster.GetService()
	if monsterSvc == nil {
		return nil
	}

	monsters := monsterSvc.GetMonstersByMap(mapID)
	if monsters == nil || len(monsters) == 0 {
		return nil
	}

	result := make([]interface{}, 0, len(monsters))
	for _, m := range monsters {
		if m.Status == 4 { // 跳过死亡怪物
			continue
		}
		result = append(result, map[string]interface{}{
			"id":      m.ID,
			"base_id": m.BaseID,
			"name":    m.Name,
			"x":       m.X,
			"y":       m.Y,
			"hp":      m.CurrentHP,
			"max_hp":  m.MaxHP,
			"level":   m.Level,
			"type":    m.Type,
			"status":  m.Status,
		})
	}

	return result
}

// ★ GetChunkInfo 获取地图的区块划分信息
// 前端在加载大地图前调用此接口，决定是否启用分块模式
//
// 响应示例：
//
//	{
//	  "code": 200,
//	  "data": {
//	    "map_id": 1,
//	    "width": 3000,           // 地图宽度（瓦片数）
//	    "height": 3000,          // 地图高度（瓦片数）
//	    "chunk_size": 64,        // 区块尺寸
//	    "chunk_cols": 47,        // 区块列数
//	    "chunk_rows": 47,        // 区块行数
//	    "total_chunks": 2209,    // 总区块数
//	    "recommend_chunk_mode": true  // 是否建议启用分块模式
//	  }
//	}
func (h *Handler) GetChunkInfo(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的地图ID"})
		return
	}

	// 获取地图配置
	mapBase, err := h.service.GetMapBase(uint32(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "地图不存在"})
		return
	}

	// 获取已加载的 GameMap（含瓦片数据）
	gameMap := GetGameMap(mapBase.MapFile)
	if gameMap == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "msg": "地图文件未加载"})
		return
	}

	chunkSize := 64 // 与前端约定
	chunkCols, chunkRows := gameMap.GetChunkSize(chunkSize)
	totalChunks := chunkCols * chunkRows
	totalTiles := int(gameMap.Width) * int(gameMap.Height)

	// 超过 20 万瓦片建议启用分块模式
	recommendChunkMode := totalTiles > 200000

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": gin.H{
			"map_id":               id,
			"width":                gameMap.Width,
			"height":               gameMap.Height,
			"chunk_size":           chunkSize,
			"chunk_cols":           chunkCols,
			"chunk_rows":           chunkRows,
			"total_chunks":         totalChunks,
			"total_tiles":          totalTiles,
			"recommend_chunk_mode": recommendChunkMode,
		},
	})
}

// ★ GetMapChunk 获取指定区块的二进制瓦片数据
// 用于前端按需加载，避免一次性下载整张地图文件
//
// 路由：GET /api/map/:id/chunk/:cx/:cy
// 响应：二进制数据（application/octet-stream）
//
//	格式：[width:u16 LE][height:u16 LE][每瓦片5字节: low:u16 LE, high:u16 LE, attr:u8]
//
// 前端通过 fetch 获取 ArrayBuffer 后按此格式解码
func (h *Handler) GetMapChunk(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的地图ID"})
		return
	}

	cxStr := c.Param("cx")
	cx, err := strconv.Atoi(cxStr)
	if err != nil || cx < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的区块X坐标"})
		return
	}

	cyStr := c.Param("cy")
	cy, err := strconv.Atoi(cyStr)
	if err != nil || cy < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的区块Y坐标"})
		return
	}

	// 获取地图配置
	mapBase, err := h.service.GetMapBase(uint32(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "地图不存在"})
		return
	}

	// 获取已加载的 GameMap
	gameMap := GetGameMap(mapBase.MapFile)
	if gameMap == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "msg": "地图文件未加载"})
		return
	}

	// 获取区块数据
	chunkSize := 64
	data := gameMap.GetChunkData(cx, cy, chunkSize)
	if data == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "msg": "区块超出地图范围"})
		return
	}

	// 返回二进制数据
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", strconv.Itoa(len(data)))
	c.Header("Cache-Control", "public, max-age=86400") // 缓存1天（区块数据不变）
	c.Data(http.StatusOK, "application/octet-stream", data)
}
