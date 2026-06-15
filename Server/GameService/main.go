package main

import (
	"game-server/Common"
	"game-server/DBService/mysql"
	"game-server/DBService/redis"
	gamemap "game-server/GameService/Map"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载全局配置
	if err := common.LoadConfig("../../Config/Game.yaml"); err != nil {
		log.Fatal("加载配置失败:", err)
	}

	if err := mysql.Init(); err != nil {
		log.Fatal("数据库初始化失败:", err)
	}
	redis.Init()

	// 加载默认地图
	err := gamemap.LoadMapFile("../../Res/Map/001.map")
	if err != nil {
		log.Println("地图加载失败(继续启动):", err)
	}

	// 启动HTTP服务(提供游戏API)
	go startHTTPServer()

	log.Println("===== 游戏微服务启动完成 =====")
	select {}
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
	mapHandler := gamemap.NewHandler()
	mapHandler.RegisterRoutes(r)

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	log.Println("游戏微服务 HTTP API 启动 :8082")
	if err := r.Run(":8082"); err != nil {
		log.Fatal("HTTP服务启动失败:", err)
	}
}
