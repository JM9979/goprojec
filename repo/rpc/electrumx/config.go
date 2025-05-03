package electrumx

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"ginproject/middleware/conf"
	"ginproject/middleware/log"
)

// RPCConfig 表示ElectrumX节点RPC配置
type RPCConfig struct {
	Host     string        // ElectrumX服务器主机地址
	Port     int           // ElectrumX服务器端口
	UseTLS   bool          // 是否使用TLS加密连接
	Timeout  time.Duration // 调用超时时间
	Protocol string        // 使用的协议，默认为"TCP"
}

var (
	// 全局RPC配置
	rpcConfig     *RPCConfig
	rpcConfigOnce sync.Once
	rpcConfigMu   sync.RWMutex
)

// GetRPCConfig 获取ElectrumX节点RPC配置
func GetRPCConfig() *RPCConfig {
	rpcConfigOnce.Do(func() {
		initRPCConfig()
	})

	rpcConfigMu.RLock()
	defer rpcConfigMu.RUnlock()
	return rpcConfig
}

// 初始化RPC配置
func initRPCConfig() {
	// 从全局配置中获取ElectrumX配置
	electrumXConfig := conf.GetElectrumXConfig()
	if electrumXConfig == nil {
		log.Error("无法获取ElectrumX配置")
		return
	}

	// 创建RPC配置
	cfg := &RPCConfig{
		Host:     electrumXConfig.Host,
		Port:     electrumXConfig.Port,
		UseTLS:   false,            // 默认不使用TLS
		Timeout:  30 * time.Second, // 默认超时时间
		Protocol: "tcp",            // 默认协议
	}

	// 验证必要的配置项
	if cfg.Host == "" {
		log.Error("缺少ElectrumX Host配置")
	}

	if cfg.Port <= 0 {
		log.Error("缺少ElectrumX Port配置或配置无效")
	}

	rpcConfigMu.Lock()
	rpcConfig = cfg
	rpcConfigMu.Unlock()

	log.Infof("ElectrumX节点RPC配置初始化完成: %s:%d", cfg.Host, cfg.Port)
}

// ValidateConfig 验证配置的合法性
func ValidateConfig(cfg *RPCConfig) error {
	if cfg == nil {
		return errors.New("RPC配置不能为空")
	}

	if cfg.Host == "" {
		return errors.New("ElectrumX Host不能为空")
	}

	if cfg.Port <= 0 {
		return errors.New("ElectrumX Port必须大于0")
	}

	return nil
}

// UpdateConfig 更新RPC配置
func UpdateConfig(cfg *RPCConfig) error {
	if err := ValidateConfig(cfg); err != nil {
		return fmt.Errorf("ElectrumX RPC配置无效: %w", err)
	}

	rpcConfigMu.Lock()
	rpcConfig = cfg
	rpcConfigMu.Unlock()

	log.Info("ElectrumX节点RPC配置已更新")
	return nil
}

// ReloadConfig 从全局配置重新加载RPC配置
func ReloadConfig() error {
	electrumXConfig := conf.GetElectrumXConfig()
	if electrumXConfig == nil {
		return errors.New("无法获取ElectrumX配置")
	}

	cfg := &RPCConfig{
		Host:     electrumXConfig.Host,
		Port:     electrumXConfig.Port,
		UseTLS:   false, // 默认不使用TLS
		Timeout:  30 * time.Second,
		Protocol: "tcp",
	}

	if err := ValidateConfig(cfg); err != nil {
		return err
	}

	rpcConfigMu.Lock()
	rpcConfig = cfg
	rpcConfigMu.Unlock()

	log.Info("ElectrumX节点RPC配置已重新加载")
	return nil
}
