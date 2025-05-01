package main

import (
	"os"

	"ginproject/middleware/log"
	"ginproject/middleware/trace"
	"ginproject/repo"
	"ginproject/service"
	tbcapi "ginproject/service/tbc_api"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var router *gin.Engine

func init() {
	router = gin.New()
	router.Use(gin.Recovery())
	// 添加trace中间件
	router.Use(trace.GinMiddleware())
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
	// 创建API路由组，设置前缀
	apiGroup := r.Group("/v1/tbc/main")

	// 添加健康检查端点
	apiGroup.GET("/health", tbcapi.NewTbcApiService().HealthCheck)
}
