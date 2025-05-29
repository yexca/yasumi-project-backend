package main

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"yasumiProject-Backend/config"
	"yasumiProject-Backend/database"
	"yasumiProject-Backend/log"
)

func main() {
	// Init Zap log
	log.Init()
	defer log.Sync() // 添加 Sync 函数包装 zap.Sync()

	// Init Config
	config.InitConfig()

	// Init Database
	database.InitDB()

	// Init Router
	//r := routes.InitRouter()

	r := gin.Default()

	// 测试路由
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	port := config.Config.Server.Port
	if port == "" {
		port = ":8080"
	}

	err := r.Run(":" + port)
	if err != nil {
		log.Fatal("服务启动失败：%v", zap.Error(err))
	}

	// start
	log.Info("服务器已启动")
}
