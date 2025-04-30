package conf

import (
	"time"

	"ginproject/entity/config"
)

// 以下是简化配置访问的辅助函数

// GetConfig 获取当前配置
func GetConfig() *config.TBCConfig {
	return GetManager().GetConfig()
}

// LoadConfig 强制重新加载配置
func LoadConfig() error {
	return GetManager().LoadConfig()
}

// EnableAutoReload 启用自动重新加载配置
func EnableAutoReload(interval time.Duration) {
	GetManager().EnableAutoReload(interval)
}

// DisableAutoReload 禁用自动重新加载配置
func DisableAutoReload() {
	GetManager().DisableAutoReload()
}

// SetConfigPath 设置配置文件路径
func SetConfigPath(path string) {
	GetManager().SetConfigPath(path)
}

// GetConfigPath 获取配置文件路径
func GetConfigPath() string {
	return GetManager().GetConfigPath()
}

// GetServerConfig 获取服务器配置
func GetServerConfig() *config.ServerConfig {
	return &GetManager().GetConfig().Server
}

// GetLogConfig 获取日志配置
func GetLogConfig() *config.LogConfig {
	return &GetManager().GetConfig().Log
}

// GetDBConfig 获取数据库配置
func GetDBConfig() *config.DBConfig {
	return &GetManager().GetConfig().DB
}

// 以下是一些常用配置项的快捷访问方法

// GetServerPort 获取服务器端口
func GetServerPort() int {
	return GetManager().GetConfig().Server.Port
}

// GetLogPath 获取日志路径
func GetLogPath() string {
	return GetManager().GetConfig().Log.Path
}

// GetLogLevel 获取日志级别
func GetLogLevel() string {
	return GetManager().GetConfig().Log.Level
}
