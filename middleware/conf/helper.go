package conf

import (
	"strings"
	"time"

	"ginproject/entity/config"

	"github.com/spf13/viper"
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

// EnableWatch 启用配置文件监视
func EnableWatch(callback func()) error {
	GetManager().EnableWatch(callback)
	return nil
}

// DisableWatch 禁用配置文件监视
func DisableWatch() {
	GetManager().DisableWatch()
}

// SetConfigPath 设置配置文件路径
func SetConfigPath(path string) {
	GetManager().SetConfigPath(path)
}

// GetConfigPath 获取配置文件路径
func GetConfigPath() string {
	return GetManager().GetConfigPath()
}

// GetViper 获取底层的Viper实例
func GetViper() *viper.Viper {
	return GetManager().GetViper()
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

// GetServerName 获取服务器名称
func GetServerName() string {
	return GetViper().GetString("server.name")
}

// GetServerHost 获取服务器主机
func GetServerHost() string {
	return GetViper().GetString("server.host")
}

// GetServerPort 获取服务器端口
func GetServerPort() int {
	return GetViper().GetInt("server.port")
}

// GetLogPath 获取日志路径
func GetLogPath() string {
	return GetViper().GetString("log.path")
}

// GetLogLevel 获取日志级别
func GetLogLevel() string {
	return GetViper().GetString("log.level")
}

// ParseConfigVariables 解析配置字符串中的变量引用
// 支持 ${xxx.yyy} 格式的变量替换
// 例如: ${server.name}, ${server.port}, ${log.level} 等
//
// 注：使用了Viper后，可以直接使用Viper的变量替换功能
func ParseConfigVariables(text string) string {
	if text == "" {
		return text
	}

	v := GetViper()

	// 使用viper的环境变量替换能力
	result := text

	// 匹配并替换 ${xxx.yyy} 格式的变量
	for {
		start := strings.Index(result, "${")
		if start == -1 {
			break
		}

		end := strings.Index(result[start:], "}")
		if end == -1 {
			break
		}

		end = start + end + 1

		// 提取变量名
		varName := result[start+2 : end-1]

		// 获取变量值
		varValue := v.GetString(varName)

		// 替换变量
		result = result[:start] + varValue + result[end:]
	}

	return result
}

// ParseLogPath 解析日志路径中的变量
func ParseLogPath() string {
	// 使用Viper内置的变量替换功能
	path := GetViper().GetString("log.path")
	return ParseConfigVariables(path)
}

// Get 通过键获取配置值
func Get(key string) interface{} {
	return GetViper().Get(key)
}

// GetString 获取字符串配置值
func GetString(key string) string {
	return GetViper().GetString(key)
}

// GetInt 获取整数配置值
func GetInt(key string) int {
	return GetViper().GetInt(key)
}

// GetBool 获取布尔配置值
func GetBool(key string) bool {
	return GetViper().GetBool(key)
}

// GetFloat64 获取浮点数配置值
func GetFloat64(key string) float64 {
	return GetViper().GetFloat64(key)
}

// GetDuration 获取时间间隔配置值
func GetDuration(key string) time.Duration {
	return GetViper().GetDuration(key)
}

// GetStringSlice 获取字符串切片配置值
func GetStringSlice(key string) []string {
	return GetViper().GetStringSlice(key)
}

// GetStringMap 获取字符串映射配置值
func GetStringMap(key string) map[string]interface{} {
	return GetViper().GetStringMap(key)
}

// GetStringMapString 获取字符串映射字符串配置值
func GetStringMapString(key string) map[string]string {
	return GetViper().GetStringMapString(key)
}

// IsSet 检查配置键是否已设置
func IsSet(key string) bool {
	return GetViper().IsSet(key)
}

// AllSettings 获取所有配置
func AllSettings() map[string]interface{} {
	return GetViper().AllSettings()
}
