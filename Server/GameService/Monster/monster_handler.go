package monster

import (
	"fmt"
	common "game-server/Common"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// MonsterHandler 怪物相关HTTP请求处理
type MonsterHandler struct{}

// RegisterRoutes 注册怪物相关路由
func (h *MonsterHandler) RegisterRoutes(r *gin.Engine) {
	monsterGroup := r.Group("/api/monster")
	{
		// GM命令：生成怪物
		monsterGroup.POST("/gm/spawn", h.GMSpawnMonster)

		// 查询地图怪物列表
		monsterGroup.GET("/list/:mapID", h.GetMapMonsters)

		// 查询所有怪物实例
		monsterGroup.GET("/all", h.GetAllMonsters)

		// ========== 新增：配置文件相关接口 ==========

		// 查询地图的生成点配置
		monsterGroup.GET("/spawn-config/:mapID", h.GetSpawnConfig)

		// 查询所有生成点配置
		monsterGroup.GET("/spawn-config/all", h.GetAllSpawnConfig)

		// 导出当前怪物为配置（用于保存到JSON）
		monsterGroup.POST("/export-spawn-config/:mapID", h.ExportSpawnConfig)

		// 重新加载生成点配置（热更新）
		monsterGroup.POST("/reload-spawn-config", h.ReloadSpawnConfig)
	}
}

// GMSpawnMonster GM生成怪物
// POST /api/monster/gm/spawn
// Body: {"base_id": 101, "map_id": 1, "x": 100, "y": 200, "count": 5}
func (h *MonsterHandler) GMSpawnMonster(c *gin.Context) {
	var req struct {
		BaseID uint32 `json:"base_id" binding:"required"` // 怪物模板ID
		MapID  uint32 `json:"map_id" binding:"required"`  // 地图ID
		X      int    `json:"x"`                          // X坐标（0=随机）
		Y      int    `json:"y"`                          // Y坐标（0=随机）
		Count  int    `json:"count"`                      // 数量（默认1）
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数错误: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 默认生成1个
	if req.Count <= 0 {
		req.Count = 1
	}

	// 最大限制20个，防止滥用
	if req.Count > 20 {
		req.Count = 20
	}

	monsters, err := GMSpawnMonster(req.BaseID, req.MapID, req.X, req.Y, req.Count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "生成失败: " + err.Error(),
			"data": nil,
		})
		return
	}

	// 返回生成的怪物信息
	var result []map[string]interface{}
	for _, m := range monsters {
		result = append(result, map[string]interface{}{
			"id":      m.ID,
			"base_id": m.BaseID,
			"name":    m.Name,
			"level":   m.Level,
			"x":       m.X,
			"y":       m.Y,
			"hp":      m.CurrentHP,
			"max_hp":  m.MaxHP,
			"status":  m.Status,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "成功生成" + strconv.Itoa(len(result)) + "个怪物",
		"data": result,
	})
}

// GetMapMonsters 获取指定地图的怪物列表
// GET /api/monster/list/:mapID
func (h *MonsterHandler) GetMapMonsters(c *gin.Context) {
	mapIDStr := c.Param("mapID")
	mapID, err := strconv.ParseUint(mapIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的地图ID",
		})
		return
	}

	svc := GetService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "怪物服务未初始化",
		})
		return
	}

	monsters := svc.GetMonstersByMap(uint32(mapID))

	var result []map[string]interface{}
	for _, m := range monsters {
		result = append(result, map[string]interface{}{
			"id":        m.ID,
			"base_id":   m.BaseID,
			"name":      m.Name,
			"level":     m.Level,
			"type":      m.Type,
			"x":         m.X,
			"y":         m.Y,
			"hp":        m.CurrentHP,
			"max_hp":    m.MaxHP,
			"status":    m.Status,
			"target_id": m.TargetID,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": gin.H{
			"map_id":   mapID,
			"count":    len(result),
			"monsters": result,
		},
	})
}

// GetAllMonsters 获取所有怪物实例（调试用）
// GET /api/monster/all
func (h *MonsterHandler) GetAllMonsters(c *gin.Context) {
	svc := GetService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "怪物服务未初始化",
		})
		return
	}

	allMonsters := svc.GetAllMonsters()

	var result []map[string]interface{}
	for _, m := range allMonsters {
		result = append(result, map[string]interface{}{
			"id":      m.ID,
			"base_id": m.BaseID,
			"name":    m.Name,
			"map_id":  m.MapID,
			"x":       m.X,
			"y":       m.Y,
			"hp":      m.CurrentHP,
			"max_hp":  m.MaxHP,
			"status":  m.Status,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": gin.H{
			"total":    len(result),
			"monsters": result,
		},
	})
}

// GetSpawnConfig 查询指定地图的生成点配置
// GET /api/monster/spawn-config/:mapID
func (h *MonsterHandler) GetSpawnConfig(c *gin.Context) {
	mapIDStr := c.Param("mapID")
	mapID, err := strconv.ParseUint(mapIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的地图ID",
		})
		return
	}

	config := common.GetMapSpawnConfig(uint32(mapID))
	if config == nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"msg":  "该地图没有配置生成点（将使用算法生成）",
			"data": gin.H{
				"map_id":       mapID,
				"spawn_points": []interface{}{},
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": *config,
	})
}

// GetAllSpawnConfig 查询所有地图的生成点配置
// GET /api/monster/spawn-config/all
func (h *MonsterHandler) GetAllSpawnConfig(c *gin.Context) {
	allConfigs := common.GetAllSpawnPoints()

	// 统计信息
	totalMaps := len(allConfigs)
	totalPoints := 0
	activePoints := 0
	for _, mapConfig := range allConfigs {
		for _, point := range mapConfig.SpawnPoints {
			totalPoints++
			if point.IsActive {
				activePoints++
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "success",
		"data": gin.H{
			"total_maps":    totalMaps,
			"total_points":  totalPoints,
			"active_points": activePoints,
			"configs":       allConfigs,
		},
	})
}

// ExportSpawnConfig 导出当前怪物为配置格式（用于保存到JSON）
// POST /api/monster/export-spawn-config/:mapID
func (h *MonsterHandler) ExportSpawnConfig(c *gin.Context) {
	mapIDStr := c.Param("mapID")
	mapID, err := strconv.ParseUint(mapIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "无效的地图ID",
		})
		return
	}

	svc := GetService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "怪物服务未初始化",
		})
		return
	}

	// 获取该地图的所有怪物
	monsters := svc.GetMonstersByMap(uint32(mapID))

	// 按base_id分组
	groups := make(map[uint32][]*MonsterInstance)
	for _, m := range monsters {
		groups[m.BaseID] = append(groups[m.BaseID], m)
	}

	// 转换为SpawnPointConfig格式
	var spawnPoints []common.SpawnPointConfig
	pointID := uint32(1000) // 起始ID

	for baseID, monsterList := range groups {
		// 获取怪物名称
		monsterConfig := common.GetMonsterConfig(baseID)
		name := "未知怪物"
		if monsterConfig != nil {
			name = monsterConfig.Name
		}

		// 计算中心坐标和半径
		sumX, sumY := 0, 0
		for _, m := range monsterList {
			sumX += m.X
			sumY += m.Y
		}
		centerX := sumX / len(monsterList)
		centerY := sumY / len(monsterList)

		// 计算最大距离作为半径
		maxDist := 0
		for _, m := range monsterList {
			dist := int(math.Sqrt(float64((m.X-centerX)*(m.X-centerX) + (m.Y-centerY)*(m.Y-centerY))))
			if dist > maxDist {
				maxDist = dist
			}
		}

		spawnPoints = append(spawnPoints, common.SpawnPointConfig{
			ID:            pointID,
			Name:          fmt.Sprintf("导出-%s-群", name),
			BaseMonsterID: baseID,
			MonsterName:   name,
			X:             centerX,
			Y:             centerY,
			Count:         len(monsterList),
			RespawnTime:   30, // 默认30秒
			SpawnRadius:   maxDist + 2,
			LevelRange:    [2]int{1, 99},
			IsActive:      true,
			Description:   fmt.Sprintf("从当前运行状态导出，包含%d个%s实例", len(monsterList), name),
		})

		pointID++
	}

	// 构建完整的MapSpawnConfig
	mapName := fmt.Sprintf("地图%d", mapID)
	mapCfg := common.GetMapConfig(uint32(mapID))
	if mapCfg != nil {
		mapName = mapCfg.Name
	}

	exportConfig := common.MapSpawnConfig{
		MapID:       uint32(mapID),
		MapName:     mapName,
		SpawnPoints: spawnPoints,
		GlobalSettings: common.MapSpawnGlobalSettings{
			MaxMonstersPerMap:    100,
			AutoRespawn:          true,
			RespawnCheckInterval: 10,
			DespawnDistance:      500,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  fmt.Sprintf("成功导出地图%d的%d个生成点配置", mapID, len(spawnPoints)),
		"data": exportConfig,
		"tip":  "可将此JSON保存到 Config/monster_spawns.json 对应地图下",
	})
}

// ReloadSpawnConfig 热重载生成点配置
// POST /api/monster/reload-spawn-config
func (h *MonsterHandler) ReloadSpawnConfig(c *gin.Context) {
	err := common.LoadGameConfig("./Config")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 500,
			"msg":  "重新加载配置失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "✅ 配置热更新成功！下次生成怪物将使用新配置",
		"data": gin.H{
			"reload_time": time.Now().Format("2006-01-02 15:04:05"),
			"maps_count":  len(common.GetAllSpawnPoints()),
		},
	})
}
