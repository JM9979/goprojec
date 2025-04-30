package conf

import (
	"fmt"
	"os"
	"sync"
	"time"

	"ginproject/entity/config"

	"gopkg.in/yaml.v3"
)

// ConfigManager 提供并发安全的配置管理
type ConfigManager struct {
	mu             sync.RWMutex
	config         *config.TBCConfig
	configPath     string
	lastLoadTime   time.Time
	reloadEnabled  bool
	reloadInterval time.Duration
}

// 全局配置管理器实例
var globalManager *ConfigManager
var once sync.Once

// 默认配置文件路径
const DefaultConfigPath = "./conf/conf.yaml"

// GetManager 获取全局配置管理器实例（单例模式）
func GetManager() *ConfigManager {
	once.Do(func() {
		globalManager = &ConfigManager{
			configPath:     DefaultConfigPath,
			reloadEnabled:  false,
			reloadInterval: 30 * time.Second,
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

	config := &config.TBCConfig{}

	// 检查配置文件是否存在
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		// 如果配置文件不存在，使用默认配置
		m.config = config
		m.lastLoadTime = time.Now()
		return nil
	}

	// 读取配置文件
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析YAML
	if err := yaml.Unmarshal(data, config); err != nil {
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
}

// DisableAutoReload 禁用配置自动重载
func (m *ConfigManager) DisableAutoReload() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.reloadEnabled = false
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
