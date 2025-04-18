package main

import (
	"fmt"
	"os"

	"GinProject/conf"
	"GinProject/middleware"
	"GinProject/service"

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
	// 初始化服务并注册路由
	helloService := service.NewHelloService()
	helloService.RegisterRoutes(r)
}
