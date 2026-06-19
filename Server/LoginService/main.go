package main

import (
	"fmt"
	common "game-server/Common"
	auth "game-server/LoginService/Auth"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载全局配置
	if err := common.LoadConfig("./Config/Login.yaml"); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化服务客户端
	common.InitServiceClients()

	r := gin.Default()

	// 启用CORS
	r.Use(corsMiddleware())

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "login"})
	})

	// 登录接口
	r.POST("/api/login", handleLogin)

	// 注册接口
	r.POST("/api/register", handleRegister)

	// Token验证接口
	r.POST("/api/validate_token", handleValidateToken)

	// 登出接口
	r.POST("/api/logout", handleLogout)

	// 获取账号信息接口
	r.POST("/api/account_info", handleAccountInfo)

	log.Println("=================================")
	log.Println("  千年江湖 - 登录微服务启动")

	port := common.AppConfig.HTTPPort
	if port == 0 {
		port = 8084
	}

	log.Printf("  端口: %d", port)
	log.Println("=================================")

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), r); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}

// CORS中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// 登录处理
func handleLogin(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	code, uid, msg := auth.CheckLogin(req.Username, req.Password)
	if code != 0 {
		c.JSON(200, gin.H{"code": code, "msg": msg})
		return
	}

	token := auth.GenerateToken(uid)
	c.JSON(200, gin.H{
		"code":  0,
		"uid":   uid,
		"token": token,
		"msg":   msg,
	})
}

// 注册处理
func handleRegister(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required,min=4,max=20"`
		Password string `json:"password" binding:"required,min=6,max=20"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, gin.H{"code": -1, "msg": "参数错误: 用户名4-20位，密码6-20位"})
		return
	}

	code, uid, msg := auth.Register(auth.RegisterRequest{
		Username: req.Username,
		Password: req.Password,
	})

	if code != 0 {
		c.JSON(200, gin.H{"code": code, "msg": msg})
		return
	}

	token := auth.GenerateToken(uid)
	c.JSON(200, gin.H{
		"code":  0,
		"uid":   uid,
		"token": token,
		"msg":   msg,
	})
}

// Token验证处理
func handleValidateToken(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	// 验证Token
	uid, err := auth.ValidateToken(req.Token)
	if err != nil {
		c.JSON(200, gin.H{"code": -1, "msg": err.Error()})
		return
	}

	c.JSON(200, gin.H{"code": 0, "msg": "验证成功", "uid": uid})
}

// 登出处理
func handleLogout(c *gin.Context) {
	var req struct {
		UID   uint   `json:"uid"`
		Token string `json:"token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	auth.Logout(req.UID)
	c.JSON(200, gin.H{"code": 0, "msg": "登出成功"})
}

// 获取账号信息处理
func handleAccountInfo(c *gin.Context) {
	var req struct {
		UID uint `json:"uid"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, gin.H{"code": -1, "msg": "参数错误"})
		return
	}

	acc, err := auth.GetAccountInfo(req.UID)
	if err != nil {
		c.JSON(200, gin.H{"code": 1, "msg": "账号不存在"})
		return
	}

	c.JSON(200, gin.H{
		"code": 0,
		"msg":  "success",
		"data": gin.H{
			"uid":      acc.ID,
			"username": acc.Username,
			"status":   acc.Status,
		},
	})
}
