package repo

import (
	"fmt"

	"ginproject/middleware/conf"
	"ginproject/middleware/log"
	"ginproject/middleware/trace"
	"ginproject/repo/db"
	"ginproject/repo/rpc/blockchain"
	"ginproject/repo/rpc/electrumx"
)

// Global_init 全局初始化
func Global_init() error {

	// 加载配置
	if err := conf.LoadConfig(); err != nil {
		return fmt.Errorf("配置初始化失败: %w", err)
	}

	// 启用配置文件监视
	err := conf.EnableWatch(func() {
		// 配置更改时的回调函数
		log.Info("检测到配置文件变更，已重新加载")
	})
	if err != nil {
		return fmt.Errorf("启用配置监视失败: %w", err)
	}

	// 获取服务名称
	serverName := conf.GetServerName()

	// 获取并解析日志配置
	logConfig := conf.GetLogConfig()

	// 解析日志路径中的变量
	// 例如: ./logs/${server.name}.log 会被解析为 ./logs/ginproject.log
	logConfig.Path = conf.ParseLogPath()

	// 初始化日志
	if err := log.InitLogger(logConfig, serverName); err != nil {
		return fmt.Errorf("日志初始化失败: %w", err)
	}

	// 初始化追踪
	trace.InitTracer(serverName)

	// 初始化数据库连接
	if err := db.Init(); err != nil {
		return fmt.Errorf("数据库初始化失败: %w", err)
	}

	// 初始化区块链RPC客户端
	if err := blockchain.Init(); err != nil {
		log.Warnf("区块链RPC客户端初始化失败: %v", err)
	}

	// 初始化ElectrumX客户端
	if err := electrumx.Init(); err != nil {
		log.Warnf("ElectrumX客户端初始化失败: %v", err)
	}

	// 记录初始化成功日志
	log.Info("全局配置和日志初始化成功")

	return nil
}
