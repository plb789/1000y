package main

import (
	"fmt"
	common "game-server/Common"
	battle "game-server/GameService/Battle"
	gamemap "game-server/GameService/Map"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
