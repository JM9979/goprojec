package conf

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// 默认配置文件路径
const DefaultConfigPath = "./conf/conf.yaml"

// BaseConfigManager 基础配置管理器结构
type BaseConfigManager struct {
	Mu           sync.RWMutex
	Config       interface{} // 存储配置
	Viper        *viper.Viper
	ConfigPath   string
	LastLoadTime time.Time // 最后加载时间
	WatchEnabled bool      // 监视状态
}

// Manager 提供并发安全的配置管理
type Manager struct {
	BaseConfigManager
}

// 全局配置管理器实例
var globalManager *Manager
var once sync.Once

// GetManager 获取全局配置管理器实例（单例模式）
func GetManager() *Manager {
	once.Do(func() {
		globalManager = &Manager{
			BaseConfigManager: BaseConfigManager{
				ConfigPath: DefaultConfigPath,
				Viper:      viper.New(),
			},
		}
		// 初始加载配置
		if err := globalManager.LoadConfig(); err != nil {
			fmt.Printf("初始化配置失败: %v\n", err)
		}
	})
	return globalManager
}

// LoadTBCConfig 从指定路径加载TBCConfig
func LoadTBCConfig(path string) (*interface{}, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var config interface{}
	if err := v.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// LoadTBCConfigWithDefault 从指定路径加载TBCConfig，如果失败则返回默认配置
func LoadTBCConfigWithDefault(path string) *interface{} {
	config, err := LoadTBCConfig(path)
	if err != nil {
		return nil
	}
	return config
}

// LoadConfig 加载配置
func (m *Manager) LoadConfig() error {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	v := m.Viper
	v.SetConfigFile(m.ConfigPath)
	v.SetConfigType("yaml")

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		// 如果配置文件不存在，使用默认配置
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			m.Config = nil
			m.LastLoadTime = time.Now()
			return nil
		}
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 将配置映射到结构体
	var vcfg interface{}
	if err := v.Unmarshal(&vcfg); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	m.Config = vcfg
	m.LastLoadTime = time.Now()
	return nil
}

// GetConfig 安全地获取配置
func (m *Manager) GetConfig(val *atomic.Value) {
	m.Mu.RLock()
	defer m.Mu.RUnlock()
	conf := val.Load()
	err := m.Viper.Unmarshal(conf)
	if err != nil {
		fmt.Printf("解析配置文件失败: %v\n", err)
	}
	val.Store(conf)
}

// EnableWatch 启用配置文件监视
func (m *Manager) EnableWatch(callback func()) error {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	if m.WatchEnabled {
		return nil
	}

	m.Viper.WatchConfig()
	m.Viper.OnConfigChange(func(e fsnotify.Event) {
		if err := m.LoadConfig(); err != nil {
			fmt.Printf("重新加载配置失败: %v\n", err)
			return
		}

		if callback != nil {
			callback()
		}
	})

	m.WatchEnabled = true
	return nil
}

// DisableWatch 禁用配置文件监视
func (m *Manager) DisableWatch() {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	if m.WatchEnabled {
		m.Viper = viper.New()
		m.Viper.SetConfigFile(m.ConfigPath)
		if err := m.LoadConfig(); err != nil {
			fmt.Printf("重新加载配置失败: %v\n", err)
		}
		m.WatchEnabled = false
	}
}

// SetConfigPath 设置配置文件路径
func (m *Manager) SetConfigPath(path string) {
	m.Mu.Lock()
	defer m.Mu.Unlock()
	m.ConfigPath = path
}

// GetConfigPath 获取当前使用的配置文件路径
func (m *Manager) GetConfigPath() string {
	m.Mu.RLock()
	defer m.Mu.RUnlock()
	return m.ConfigPath
}

// GetViper 获取底层的Viper实例
func (m *Manager) GetViper() *viper.Viper {
	m.Mu.RLock()
	defer m.Mu.RUnlock()
	return m.Viper
}

// Get 通过键获取配置值
func (m *Manager) Get(key string) interface{} {
	return m.Viper.Get(key)
}

// GetString 获取字符串配置值
func (m *Manager) GetString(key string) string {
	return m.Viper.GetString(key)
}

// GetInt 获取整数配置值
func (m *Manager) GetInt(key string) int {
	return m.Viper.GetInt(key)
}

// GetBool 获取布尔配置值
func (m *Manager) GetBool(key string) bool {
	return m.Viper.GetBool(key)
}

// GetFloat64 获取浮点数配置值
func (m *Manager) GetFloat64(key string) float64 {
	return m.Viper.GetFloat64(key)
}

// GetStringSlice 获取字符串切片配置值
func (m *Manager) GetStringSlice(key string) []string {
	return m.Viper.GetStringSlice(key)
}

// GetStringMap 获取字符串映射配置值
func (m *Manager) GetStringMap(key string) map[string]interface{} {
	return m.Viper.GetStringMap(key)
}

// GetStringMapString 获取字符串映射字符串配置值
func (m *Manager) GetStringMapString(key string) map[string]string {
	return m.Viper.GetStringMapString(key)
}

// GetDuration 获取时间间隔配置值
func (m *Manager) GetDuration(key string) time.Duration {
	return m.Viper.GetDuration(key)
}

// IsSet 检查配置键是否已设置
func (m *Manager) IsSet(key string) bool {
	return m.Viper.IsSet(key)
}

// AllSettings 获取所有配置
func (m *Manager) AllSettings() map[string]interface{} {
	return m.Viper.AllSettings()
}

// ParseConfigVariables 解析配置字符串中的变量引用
// 支持 ${xxx.yyy} 格式的变量替换
func ParseConfigVariables(v *viper.Viper, text string) string {
	if text == "" {
		return text
	}

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
func (m *Manager) ParseLogPath() string {
	path := m.GetString("log.path")
	return ParseConfigVariables(m.Viper, path)
}
