package middleware

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"GinProject/conf"

	"github.com/gin-gonic/gin"
	"golang.org/x/exp/slog"
	"gopkg.in/natefinch/lumberjack.v2"
)

func InitLogger(cfg *conf.AppConfig) error {
	// 自动创建日志目录
	if cfg.LogPath != "" {
		if err := os.MkdirAll(filepath.Dir(cfg.LogPath), 0755); err != nil {
			return fmt.Errorf("创建日志目录失败: %w", err)
		}
	}

	// 配置日志输出
	output := []interface{}{os.Stdout}
	if cfg.LogPath != "" {
		output = append(output, &lumberjack.Logger{
			Filename:   cfg.LogPath,
			MaxSize:    100, // megabytes
			MaxBackups: 3,
			MaxAge:     7, // days
		})
	}

	level := slog.LevelInfo
	switch cfg.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	h := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	}).WithAttrs([]slog.Attr{
		slog.String("service", "GinProject"),
	})

	logger := slog.New(h)
	slog.SetDefault(logger)
	return nil
}

// GinLogger 创建Gin日志中间件
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		// 处理请求
		c.Next()

		// 记录日志
		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		logAttrs := []interface{}{
			"status", status,
			"latency", latency,
			"client_ip", clientIP,
			"method", method,
			"path", path,
		}

		switch {
		case status >= 500:
			slog.ErrorCtx(c.Request.Context(), "Server Error", logAttrs...)
		case status >= 400:
			slog.WarnCtx(c.Request.Context(), "Client Error", logAttrs...)
		default:
			slog.InfoCtx(c.Request.Context(), "Request processed", logAttrs...)
		}
	}
}

// GinRecovery 创建异常恢复中间件
func GinRecovery(stack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录panic日志
				slog.ErrorCtx(c.Request.Context(), "Recovered from panic",
					"error", err,
					"stack", getCallers(stack),
				)

				// 返回错误响应
				c.AbortWithStatusJSON(500, gin.H{
					"code":    500,
					"message": "Internal Server Error",
				})
			}
		}()
		c.Next()
	}
}

func getCallers(all bool) string {
	// 获取调用栈信息（实现略）
	return ""
}
