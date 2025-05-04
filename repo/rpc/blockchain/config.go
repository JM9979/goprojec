package blockchain

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"ginproject/middleware/conf"
	"ginproject/middleware/log"
)

// RPCConfig 表示区块链节点RPC配置
type RPCConfig struct {
	URL      string        // 节点RPC服务URL
	User     string        // 认证用户名
	Password string        // 认证密码
	Timeout  time.Duration // 调用超时时间
}

var (
	// 全局RPC配置
	rpcConfig     *RPCConfig
	rpcConfigOnce sync.Once
	rpcConfigMu   sync.RWMutex
)

// GetRPCConfig 获取区块链节点RPC配置
func GetRPCConfig() *RPCConfig {
	rpcConfigOnce.Do(func() {
		initRPCConfig()
	})

	rpcConfigMu.RLock()
	defer rpcConfigMu.RUnlock()
	return rpcConfig
}

// GetConfig 获取区块链节点RPC配置 (兼容旧接口)
func GetConfig() *RPCConfig {
	return GetRPCConfig()
}

// 初始化RPC配置
func initRPCConfig() {
	// 从全局配置中获取TBCNode配置
	tbcNodeConfig := conf.GetTBCNodeConfig()
	if tbcNodeConfig == nil {
		log.Error("无法获取TBCNode配置")
		return
	}

	// 创建RPC配置
	cfg := &RPCConfig{
		URL:      tbcNodeConfig.URL,
		User:     tbcNodeConfig.User,
		Password: tbcNodeConfig.Password,
		Timeout:  30 * time.Second, // 默认超时时间
	}

	// 验证必要的配置项
	if cfg.URL == "" {
		log.Error("缺少RPC URL配置")
	}

	if cfg.User == "" || cfg.Password == "" {
		log.Warn("RPC认证信息不完整，可能需要在某些环境中进行认证")
	}

	rpcConfigMu.Lock()
	rpcConfig = cfg
	rpcConfigMu.Unlock()

	log.Infof("区块链节点RPC配置初始化完成: %s", cfg.URL)
}

// ValidateConfig 验证配置的合法性
func ValidateConfig(cfg *RPCConfig) error {
	if cfg == nil {
		return errors.New("RPC配置不能为空")
	}

	if cfg.URL == "" {
		return errors.New("RPC URL不能为空")
	}

	// 用户名和密码可以为空，但需要发出警告
	if cfg.User == "" || cfg.Password == "" {
		log.Warn("RPC认证信息不完整，某些操作可能需要认证")
	}

	return nil
}

// UpdateConfig 更新RPC配置
func UpdateConfig(cfg *RPCConfig) error {
	if err := ValidateConfig(cfg); err != nil {
		return fmt.Errorf("RPC配置无效: %w", err)
	}

	rpcConfigMu.Lock()
	rpcConfig = cfg
	rpcConfigMu.Unlock()

	log.Info("区块链节点RPC配置已更新")
	return nil
}

// ReloadConfig 从全局配置重新加载RPC配置
func ReloadConfig() error {
	tbcNodeConfig := conf.GetTBCNodeConfig()
	if tbcNodeConfig == nil {
		return errors.New("无法获取TBCNode配置")
	}

	cfg := &RPCConfig{
		URL:      tbcNodeConfig.URL,
		User:     tbcNodeConfig.User,
		Password: tbcNodeConfig.Password,
		Timeout:  30 * time.Second,
	}

	if err := ValidateConfig(cfg); err != nil {
		return err
	}

	rpcConfigMu.Lock()
	rpcConfig = cfg
	rpcConfigMu.Unlock()

	log.Info("区块链节点RPC配置已重新加载")
	return nil
}
