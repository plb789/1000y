package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	common "game-server/Common"
	battle "game-server/GameService/Battle"
	gamemap "game-server/GameService/Map"
	monster "game-server/GameService/Monster"

	"github.com/gin-gonic/gin"
)

var (
	instanceID  uint32
	handledMaps []uint32
)

func main() {
	// 加载全局配置
	if err := common.LoadConfig("./Config/Game.yaml"); err != nil {
		log.Fatal("加载配置失败:", err)
	}

	// 获取实例配置
	instanceID = common.AppConfig.InstanceID
	handledMaps = common.AppConfig.HandledMaps

	// 如果没有配置handled_maps，默认处理所有地图
	if len(handledMaps) == 0 {
		log.Println("警告: 未配置handled_maps，该实例将处理所有地图")
		for _, m := range common.GameConfig.Maps {
			handledMaps = append(handledMaps, m.ID)
		}
	}

	log.Printf("GameService实例配置: instanceID=%d, handledMaps=%v", instanceID, handledMaps)

	// 从JSON文件加载所有配置数据
	if err := common.LoadGameConfig("./Config"); err != nil {
		log.Fatal("加载配置数据失败:", err)
	}
	log.Printf("配置加载完成: %d个地图, %d个NPC, %d个怪物, %d个武学, %d个道具, %d个掉落组, %d个BUFF, %d个任务, %d个商店, %d个公告",
		len(common.GameConfig.Maps), len(common.GameConfig.NPCs),
		len(common.GameConfig.Monsters), len(common.GameConfig.Skills),
		len(common.GameConfig.Items), len(common.GameConfig.DropGroups),
		len(common.GameConfig.Buffs), len(common.GameConfig.Quests),
		len(common.GameConfig.Shops), len(common.GameConfig.Announcements))

	// 初始化服务客户端(连接DBService)
	common.InitServiceClients()

	// 初始化消息总线
	initMessageBus()

	// 初始化注册中心
	registryURL := common.AppConfig.Services.RegistryService
	if registryURL == "" {
		registryURL = common.AppConfig.Services.DBService
	}
	common.InitRegistry(registryURL)

	// 注册到服务发现中心
	serviceURL := fmt.Sprintf("http://localhost:%d", common.AppConfig.HTTPPort)
	if err := common.RegisterGameService(instanceID, serviceURL, handledMaps); err != nil {
		log.Printf("注册到服务中心失败: %v (继续启动)", err)
	}

	// 启动心跳保活
	if common.AppConfig.RegisterToRegistry {
		go startHeartbeat()
	}

	// 加载本实例负责的地图
	loadManagedMaps()

	// 初始化怪物服务（必须在loadManagedMaps之后）
	battleSvc := battle.NewService()
	monster.InitService(battleSvc)

	// ✅ 解决循环依赖：双向注入服务接口
	// 1. Monster → Battle: AI系统需要调用战斗服务进行攻击计算
	monster.SetBattleService(battleSvc)

	// 2. Monster → Map: AI系统需要调用地图服务进行碰撞检测
	monster.SetMapService(gamemap.GetService())

	// 3. Monster → EntityChecker: AI系统需要检测玩家/怪物实体碰撞
	entityChecker := &EntityCollisionCheckerImpl{
		monsterSvc: monster.GetService(),
		mapSvc:     gamemap.GetService(),
	}
	monster.SetEntityChecker(entityChecker)

	// 4. Map → EntityChecker: 地图服务需要检测怪物碰撞（用于玩家移动）
	gamemap.SetEntityChecker(entityChecker)

	// 4. Battle → Monster: 战斗系统需要获取怪物信息和通知AI
	battle.SetMonsterService(monster.GetService())
	battle.SetAIService(monster.GetAIService())

	// 5. Monster → PlayerPosition: AI系统需要获取玩家位置用于追踪目标
	monster.SetPlayerPositionFunc(func(targetID uint64) (int, int, bool) {
		// 遍历所有已加载地图查找玩家
		players := gamemap.GetService().GetAllPlayersInMap(0)
		for _, p := range players {
			if p.RoleID == targetID {
				return p.X, p.Y, true
			}
		}
		return 0, 0, false
	})

	// 为每个地图生成初始怪物
	for _, mapID := range handledMaps {
		monster.InitMapMonsters(mapID)
	}

	// 启动HTTP服务
	go startHTTPServer()

	// 等待退出信号
	waitForExit()
}

func loadManagedMaps() {
	// 创建地图ID集合用于快速查找
	mapSet := make(map[uint32]bool)
	for _, mapID := range handledMaps {
		mapSet[mapID] = true
	}

	// 加载本实例负责的地图
	for _, m := range common.GameConfig.Maps {
		if mapSet[m.ID] {
			// 先加载地图文件
			mapPath := "./Res/Map/" + m.MapFile
			err := gamemap.LoadMapFile(mapPath)
			if err != nil {
				log.Printf("地图[%s]文件加载失败: %v", m.Name, err)
			} else {
				log.Printf("地图[%s]文件加载成功 (instance=%d)", m.Name, instanceID)
			}

			// 再加载地图数据到内存（初始化Collision数组等）
			_, err = gamemap.GetService().LoadMap(m.ID)
			if err != nil {
				log.Printf("地图[%s]内存加载失败: %v", m.Name, err)
			} else {
				log.Printf("地图[%s]内存加载成功 (instance=%d)", m.Name, instanceID)
			}
		}
	}
}

// initMessageBus 初始化消息总线
func initMessageBus() {
	busConfig := common.AppConfig.MessageBus

	if busConfig.Type == "" {
		busConfig.Type = "http" // 默认使用HTTP
	}

	gatewayURL := common.AppConfig.Services.GatewayService
	if gatewayURL == "" {
		gatewayURL = "http://localhost:8080"
	}

	log.Printf("消息总线配置: type=%s, gateway=%s", busConfig.Type, gatewayURL)

	config := common.MessageBusConfig{
		Type:         busConfig.Type,
		RabbitMQURL:  busConfig.RabbitMQURL,
		KafkaBrokers: busConfig.KafkaBrokers,
		HTTPURL:      gatewayURL,
	}

	common.InitMessageBus(config)
}

func startHeartbeat() {
	tickerInterval := common.RegistryHeartbeatInterval
	if tickerInterval <= 0 {
		tickerInterval = 30
	}

	// 使用周期性定时器（Ticker），而不是单次定时器（Timer）
	t := common.NewTicker(time.Duration(tickerInterval) * time.Second)
	defer t.Stop()

	for {
		if err := common.UpdateHeartbeat(instanceID); err != nil {
			log.Printf("心跳更新失败: %v", err)
		}
		<-t.C
	}
}

func waitForExit() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Printf("GameService实例 %d 正在注销...", instanceID)
	common.UnregisterGameService(instanceID)
	os.Exit(0)
}

// ========== 实体碰撞检测器实现 ==========

// EntityCollisionCheckerImpl 实体碰撞检测器（用于怪物AI和玩家移动）
type EntityCollisionCheckerImpl struct {
	monsterSvc *monster.Service
	mapSvc     *gamemap.Service
}

// CheckEntityCollision 检查实体间碰撞（怪物-玩家、怪物-怪物）
// 参数:
//   - mapID: 地图ID
//   - x, y: 目标坐标
//   - excludeID: 排除的实体ID（通常是自身，0表示不排除任何实体）
//
// 返回值:
//   - bool: 是否有碰撞
//   - uint64: 碰撞的实体ID
func (e *EntityCollisionCheckerImpl) CheckEntityCollision(mapID uint32, x, y int, excludeID uint64) (bool, uint64) {
	// 1. 检查怪物碰撞
	if e.monsterSvc != nil {
		monsters := e.monsterSvc.GetAllMonsters()

		for _, m := range monsters {
			// 排除自身
			if m.ID == excludeID {
				continue
			}

			// 只检查同地图的怪物
			if m.MapID != mapID {
				continue
			}

			// 跳过死亡怪物
			if m.Status == 4 { // 死亡状态
				continue
			}

			// 计算距离（碰撞半径=0.8格，允许部分重叠但不完全重合）
			dx := float64(m.X - x)
			dy := float64(m.Y - y)
			distance := math.Sqrt(dx*dx + dy*dy)

			collisionRadius := 0.8 // 碰撞半径（格）
			if distance < collisionRadius {
				return true, m.ID
			}
		}
	}

	// 2. 检查玩家碰撞
	if e.mapSvc != nil {
		players := e.mapSvc.GetAllPlayersInMap(mapID)

		for _, p := range players {
			// 排除自身（如果excludeID是玩家ID）
			if p.RoleID == excludeID {
				continue
			}

			// 计算距离（玩家碰撞半径=0.8格）
			dx := float64(p.X - x)
			dy := float64(p.Y - y)
			distance := math.Sqrt(dx*dx + dy*dy)

			collisionRadius := 0.8 // 碰撞半径（格）
			if distance < collisionRadius {
				return true, p.RoleID
			}
		}
	}

	return false, 0
}

// CheckMonsterCollision 仅检查怪物碰撞（用于玩家移动验证）
func (e *EntityCollisionCheckerImpl) CheckMonsterCollision(mapID uint32, x, y int, playerID uint64) (bool, uint64) {
	if e.monsterSvc == nil {
		return false, 0
	}

	monsters := e.monsterSvc.GetAllMonsters()

	for _, m := range monsters {
		// 只检查同地图的怪物
		if m.MapID != mapID {
			continue
		}

		// 跳过死亡怪物
		if m.Status == 4 {
			continue
		}

		// 计算距离
		dx := float64(m.X - x)
		dy := float64(m.Y - y)
		distance := math.Sqrt(dx*dx + dy*dy)

		collisionRadius := 0.8
		if distance < collisionRadius {
			return true, m.ID
		}
	}

	return false, 0
}

// GetEntityPosition 获取实体位置
func (e *EntityCollisionCheckerImpl) GetEntityPosition(entityID uint64) (int, int, bool) {
	// 先尝试从怪物服务查找
	if e.monsterSvc != nil {
		m, exists := e.monsterSvc.GetMonster(entityID)
		if exists {
			return m.X, m.Y, true
		}
	}

	// 再尝试从地图服务查找玩家
	if e.mapSvc != nil {
		players := e.mapSvc.GetAllPlayersInMap(0) // 需要知道mapID，这里简化处理
		for _, p := range players {
			if p.RoleID == entityID {
				return p.X, p.Y, true
			}
		}
	}

	return 0, 0, false
}

func startHTTPServer() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// 启用CORS
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Account-ID")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// 注册地图路由
	mapHandler := gamemap.NewHandler(instanceID)
	mapHandler.RegisterRoutes(r)

	// 注册战斗路由
	gatewayURL := common.AppConfig.Services.GatewayService
	battleHandler := battle.NewHandler(gatewayURL)
	battleHandler.RegisterRoutes(r)

	// 设置怪物位置广播函数（通过Gateway同步给所有客户端）
	monster.SetPositionBroadcastFunc(func(positionMap map[uint32][]monster.MonsterPositionInfo) {
		for mapID, positions := range positionMap {
			if len(positions) == 0 {
				continue
			}

			// 构造广播数据
			data := gin.H{
				"map_id":    mapID,
				"monsters":  positions,
				"timestamp": time.Now().UnixMilli(),
			}

			// 通过Handler的broadcastToMap方法广播
			battleHandler.BroadcastMonsterPositions(mapID, data)
		}
	})

	// 注册怪物路由（包含GM命令）
	monsterHandler := &monster.MonsterHandler{}
	monsterHandler.RegisterRoutes(r)

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":       "ok",
			"instance_id":  instanceID,
			"handled_maps": handledMaps,
		})
	})

	// 实例信息
	r.GET("/api/instance/info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"instance_id":  instanceID,
			"handled_maps": handledMaps,
			"service_url":  fmt.Sprintf("http://localhost:%d", common.AppConfig.HTTPPort),
		})
	})

	// 检查地图是否由本实例处理
	r.GET("/api/instance/handles/:map_id", func(c *gin.Context) {
		var mapID uint32
		fmt.Sscanf(c.Param("map_id"), "%d", &mapID)

		handles := false
		for _, m := range handledMaps {
			if m == mapID {
				handles = true
				break
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"map_id":  mapID,
			"handles": handles,
		})
	})

	port := common.AppConfig.HTTPPort
	if port == 0 {
		port = 8082
	}

	log.Printf("游戏微服务 HTTP API 启动 :%d (instance=%d)", port, instanceID)
	if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatal("HTTP服务启动失败:", err)
	}
}
