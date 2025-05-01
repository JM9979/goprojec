package conf

import (
	"fmt"
	"sync"
	"time"

	"ginproject/entity/config"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// 默认配置文件路径
const DefaultConfigPath = "./conf/conf.yaml"

// ConfigManager 提供并发安全的配置管理
type ConfigManager struct {
	mu             sync.RWMutex
	config         *config.TBCConfig
	viper          *viper.Viper
	configPath     string
	lastLoadTime   time.Time
	reloadEnabled  bool
	reloadInterval time.Duration
	watchEnabled   bool
	watchCallback  func()
}

// 全局配置管理器实例
var globalManager *ConfigManager
var once sync.Once

// GetManager 获取全局配置管理器实例（单例模式）
func GetManager() *ConfigManager {
	once.Do(func() {
		globalManager = &ConfigManager{
			configPath:     DefaultConfigPath,
			reloadEnabled:  false,
			reloadInterval: 30 * time.Second,
			watchEnabled:   false,
			viper:          viper.New(),
		}
		// 初始加载配置
		if err := globalManager.LoadConfig(); err != nil {
			fmt.Printf("初始化配置失败: %v\n", err)
		}
	})
	return globalManager
}

// LoadConfig 加载配置
func (m *ConfigManager) LoadConfig() error {
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

// GetConfig 安全地获取配置（自动检查是否需要重新加载）
func (m *ConfigManager) GetConfig() *config.TBCConfig {
	// 如果启用了自动重载并且已经超过了重载时间间隔
	if m.reloadEnabled && time.Since(m.lastLoadTime) > m.reloadInterval {
		if err := m.LoadConfig(); err != nil {
			fmt.Printf("重新加载配置失败: %v\n", err)
		}
	}

	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// EnableAutoReload 启用配置自动重载
func (m *ConfigManager) EnableAutoReload(interval time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.reloadEnabled = true
	if interval > 0 {
		m.reloadInterval = interval
	}

	// 如果未启用Viper的配置监视，则启用
	m.enableWatch()
}

// DisableAutoReload 禁用配置自动重载
func (m *ConfigManager) DisableAutoReload() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.reloadEnabled = false
}

// enableWatch 启用Viper配置文件监视（内部方法）
func (m *ConfigManager) enableWatch() {
	// 如果已经启用了监视，则直接返回
	if m.watchEnabled {
		return
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
		if m.watchCallback != nil {
			m.watchCallback()
		}
	})

	m.watchEnabled = true
}

// EnableWatch 启用配置文件监视并设置回调
func (m *ConfigManager) EnableWatch(callback func()) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.watchCallback = callback
	m.enableWatch()
}

// SetConfigPath 设置配置文件路径
func (m *ConfigManager) SetConfigPath(path string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.configPath = path
}

// GetConfigPath 获取当前使用的配置文件路径
func (m *ConfigManager) GetConfigPath() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.configPath
}

// GetViper 获取底层的Viper实例
func (m *ConfigManager) GetViper() *viper.Viper {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.viper
}

// DisableWatch 禁用配置文件监视
func (m *ConfigManager) DisableWatch() {
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
		m.watchCallback = nil
	}
}
