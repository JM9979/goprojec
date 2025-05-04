package service

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ginproject/middleware/conf"
	"ginproject/middleware/log"
	"ginproject/repo/db"

	"github.com/gin-gonic/gin"
)

type Server struct {
	server *http.Server
}

// CreateServer 创建HTTP服务器
func CreateServer(r *gin.Engine) *Server {
	// 获取服务器配置
	serverConfig := conf.GetServerConfig()
	addr := fmt.Sprintf("%s:%d", serverConfig.Host, serverConfig.Port)
	return &Server{
		server: &http.Server{
			Addr:    addr,
			Handler: r,
		},
	}
}

// Start 启动服务并处理优雅关闭
func (h *Server) Start() {
	// 在单独的goroutine中启动服务器
	go func() {
		log.Info("服务启动成功", "地址:", h.server.Addr)
		if err := h.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("服务启动失败", "错误:", err)
			os.Exit(1)
		}
	}()

	// 等待中断信号并优雅地关闭服务器
	h.WaitForInterruptAndShutdown()
}

// WaitForInterruptAndShutdown 等待中断信号并优雅关闭服务器
func (h *Server) WaitForInterruptAndShutdown() {
	// 创建接收信号的通道
	quit := make(chan os.Signal, 1)
	// 监听 SIGINT 和 SIGTERM 信号
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	// 阻塞，直到接收到信号
	<-quit
	log.Info("正在关闭服务，断开连接...", "地址:", h.server.Addr)

	// 创建超时上下文，优雅关闭
	h.GracefulShutdown()
}

// GracefulShutdown 优雅关闭服务器
func (h *Server) GracefulShutdown() {
	// 设置5秒的超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := h.server.Shutdown(ctx); err != nil {
		log.Error("服务关闭时发生错误", "错误:", err, "地址:", h.server.Addr)
	}

	// 关闭数据库连接
	db.Close()

	log.Info("服务已安全关闭", "地址:", h.server.Addr)
}
