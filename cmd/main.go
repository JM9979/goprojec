package main

import (
	"fmt"
	"os"

	"ginproject/conf"
	"ginproject/middleware"
	"ginproject/service"

	"github.com/gin-gonic/gin"
	"golang.org/x/exp/slog"
)

func main() {
	cfg, err := conf.InitConfig()
	if err != nil {
		slog.Error("配置初始化失败", "error", err)
		os.Exit(1)
	}

	// 初始化日志
	if err := middleware.InitLogger(cfg); err != nil {
		slog.Error("日志初始化失败", "error", err)
		os.Exit(1)
	}

	router := gin.New()
	router.Use(
		middleware.GinLogger(),
		middleware.GinRecovery(true),
	)

	// 注册路由
	registerRoutes(router)

	slog.Info("服务启动", "port", cfg.Port)
	if err := router.Run(fmt.Sprintf(":%d", cfg.Port)); err != nil {
		slog.Error("服务启动失败", "error", err)
		os.Exit(1)
	}
}

func registerRoutes(r *gin.Engine) {
	// 初始化服务
	helloService := service.NewHelloService()

	// 添加新的时间服务路由
	r.GET("/current_time", service.NewCurrentTimeService().GetCurrentTimeService)

	// 注册 /say/hello 路由
	r.POST("/say/hello", helloService.HelloHandler)
}
