package conf

import (
	"fmt"
	"sync"
	"time"

	"ginproject/entity/config"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// ViperManager 提供基于Viper的配置管理
type ViperManager struct {
	mu           sync.RWMutex
	config       *config.TBCConfig
	viper        *viper.Viper
	configPath   string
	lastLoadTime time.Time
	watchEnabled bool
}

// 全局Viper配置管理器实例
var globalViperManager *ViperManager
var viperOnce sync.Once

// GetViperManager 获取全局Viper配置管理器实例（单例模式）
func GetViperManager() *ViperManager {
	viperOnce.Do(func() {
		globalViperManager = &ViperManager{
			configPath:   DefaultConfigPath,
			watchEnabled: false,
			viper:        viper.New(),
		}
		// 初始加载配置
		if err := globalViperManager.LoadConfig(); err != nil {
			fmt.Printf("初始化配置失败: %v\n", err)
		}
	})
	return globalViperManager
}

// LoadConfig 加载配置
func (m *ViperManager) LoadConfig() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	v := m.viper
	v.SetConfigFile(m.configPath)
	v.SetConfigType("yaml")

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		// 如果配置文件不存在，使用默认配置
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			m.config = &config.TBCConfig{}
			m.lastLoadTime = time.Now()
			return nil
		}
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 将配置映射到结构体
	config := &config.TBCConfig{}
	if err := v.Unmarshal(config); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	m.config = config
	m.lastLoadTime = time.Now()
	return nil
}

// GetConfig 安全地获取配置
func (m *ViperManager) GetConfig() *config.TBCConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// EnableWatch 启用配置文件监视
func (m *ViperManager) EnableWatch(callback func()) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 如果已经启用了监视，则直接返回
	if m.watchEnabled {
		return nil
	}

	// 启用配置文件监视
	m.viper.WatchConfig()
	m.viper.OnConfigChange(func(e fsnotify.Event) {
		// 重新加载配置
		err := m.LoadConfig()
		if err != nil {
			fmt.Printf("重新加载配置失败: %v\n", err)
			return
		}

		// 如果提供了回调函数，则调用
		if callback != nil {
			callback()
		}
	})

	m.watchEnabled = true
	return nil
}

// DisableWatch 禁用配置文件监视
func (m *ViperManager) DisableWatch() {
	// Viper不提供直接禁用监视的方法，需要创建新的实例
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.watchEnabled {
		// 创建新的Viper实例，不启用监视
		m.viper = viper.New()
		m.viper.SetConfigFile(m.configPath)
		// 重新加载配置
		if err := m.LoadConfig(); err != nil {
			fmt.Printf("重新加载配置失败: %v\n", err)
		}
		m.watchEnabled = false
	}
}

// SetConfigPath 设置配置文件路径
func (m *ViperManager) SetConfigPath(path string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.configPath = path
}

// GetConfigPath 获取当前使用的配置文件路径
func (m *ViperManager) GetConfigPath() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.configPath
}

// GetViper 获取底层的Viper实例
func (m *ViperManager) GetViper() *viper.Viper {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.viper
}

// Get 通过键获取配置值
func (m *ViperManager) Get(key string) interface{} {
	return m.viper.Get(key)
}

// GetString 获取字符串配置值
func (m *ViperManager) GetString(key string) string {
	return m.viper.GetString(key)
}

// GetInt 获取整数配置值
func (m *ViperManager) GetInt(key string) int {
	return m.viper.GetInt(key)
}

// GetBool 获取布尔配置值
func (m *ViperManager) GetBool(key string) bool {
	return m.viper.GetBool(key)
}

// GetFloat64 获取浮点数配置值
func (m *ViperManager) GetFloat64(key string) float64 {
	return m.viper.GetFloat64(key)
}

// GetStringSlice 获取字符串切片配置值
func (m *ViperManager) GetStringSlice(key string) []string {
	return m.viper.GetStringSlice(key)
}

// GetStringMap 获取字符串映射配置值
func (m *ViperManager) GetStringMap(key string) map[string]interface{} {
	return m.viper.GetStringMap(key)
}
