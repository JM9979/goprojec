package config

import (
	"ginproject/middleware/conf"
	"sync/atomic"
)

// 全局原子值存储配置
var globalConfig atomic.Value

func init() {
	globalConfig.Store(&TBCConfig{})
}

// TBCConfig 总配置结构
type TBCConfig struct {
	Server    ServerConfig    `yaml:"server"`
	Log       LogConfig       `yaml:"log"`
	DB        DBConfig        `yaml:"db"`
	TBCNode   TBCNodeConfig   `yaml:"tbcnode"`
	ElectrumX ElectrumXConfig `yaml:"electrumx"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Name string `yaml:"name"`
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// LogConfig 日志配置
type LogConfig struct {
	Path  string `yaml:"path"`
	Level string `yaml:"level"`
}

// DBConfig 数据库配置
type DBConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	Database     string `yaml:"database"`
	Charset      string `yaml:"charset"`
	MaxIdleConns int    `yaml:"maxidleconns"`
	MaxOpenConns int    `yaml:"maxopenconns"`
}

// TBCNodeConfig RPC客户端配置
type TBCNodeConfig struct {
	URL          string `yaml:"url"`
	User         string `yaml:"user"`
	Password     string `yaml:"password"`
	Timeout      int    `yaml:"timeout"`
	MaxIdleConns int    `yaml:"maxidleconns"`
	MaxOpenConns int    `yaml:"maxopenconns"`
}

// ElectrumXConfig ElectrumX RPC客户端配置
type ElectrumXConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	Timeout      int    `yaml:"timeout"`
	RetryCount   int    `yaml:"retry_count"`
	UseTLS       bool   `yaml:"use_tls"`
	Protocol     string `yaml:"protocol"`
	MaxIdleConns int    `yaml:"maxidleconns"`
	MaxOpenConns int    `yaml:"maxopenconns"`
}

// GetConfig 获取配置
func GetConfig() *TBCConfig {
	conf.GetManager().GetConfig(&globalConfig)
	return globalConfig.Load().(*TBCConfig)
}

// GetServerConfig 获取服务器配置
func (c *TBCConfig) GetServerConfig() *ServerConfig {
	return &c.Server
}

// GetLogConfig 获取日志配置
func (c *TBCConfig) GetLogConfig() *LogConfig {
	return &c.Log
}

// GetDBConfig 获取数据库配置
func (c *TBCConfig) GetDBConfig() *DBConfig {
	return &c.DB
}

// GetTBCNodeConfig 获取TBCNode配置
func (c *TBCConfig) GetTBCNodeConfig() *TBCNodeConfig {
	return &c.TBCNode
}

// GetElectrumXConfig 获取ElectrumX配置
func (c *TBCConfig) GetElectrumXConfig() *ElectrumXConfig {
	return &c.ElectrumX
}
