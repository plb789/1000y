package main

import (
	"game-server/Common"
	"game-server/DBService/mysql"
	"game-server/DBService/redis"
	"game-server/DBService/model"
	"log"
)

func main() {
	// 加载全局配置
	err := common.LoadConfig("../../Config/DB.yaml")
	if err != nil {
		log.Fatal("加载配置失败:", err)
	}

	// 初始化MySQL
	err = mysql.Init()
	if err != nil {
		log.Fatal("MySQL初始化失败:", err)
	}
	// 自动建表
	err = mysql.DB.AutoMigrate(&model.Account{}, &model.Role{})
	if err != nil {
		log.Fatal("建表失败:", err)
	}

	// 初始化Redis
	redis.Init()

	log.Println("===== 数据微服务启动完成 =====")
	select {} // 常驻
}
