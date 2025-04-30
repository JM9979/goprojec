package repo

import (
	"fmt"

	"ginproject/middleware/conf"
	"ginproject/middleware/log"
)

// Global_init 全局初始化
func Global_init() error {

	// 加载配置
	if err := conf.LoadConfig(); err != nil {
		return fmt.Errorf("配置初始化失败: %w", err)
	}

	// 启用配置自动重载
	conf.EnableAutoReload(0) // 使用默认间隔时间

	// 初始化日志
	if err := log.InitLogger(conf.GetLogConfig()); err != nil {
		return fmt.Errorf("日志初始化失败: %w", err)
	}

	// 记录初始化成功日志
	log.Info("全局配置和日志初始化成功")

	return nil
}
