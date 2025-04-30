package main

import (
	"os"

	"ginproject/middleware/log"
	"ginproject/repo"
	"ginproject/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var router *gin.Engine

func init() {
	router = gin.New()
	router.Use(gin.Recovery())
}

func main() {
	// 全局初始化
	if err := repo.Global_init(); err != nil {
		log.Error("全局初始化失败", zap.Error(err))
		os.Exit(1)
	}

	// 注册路由
	registerRoutes(router)

	// 创建HTTP服务器并启动
	srv := service.CreateServer(router)
	srv.Start()
}

func registerRoutes(r *gin.Engine) {
	// 初始化服务
	helloService := service.NewHelloService()

	// 添加新的时间服务路由
	r.GET("/current_time", service.NewCurrentTimeService().GetCurrentTimeService)

	// 注册 /say/hello 路由
	r.POST("/say/hello", helloService.HelloHandler)
}
