package conf

import (
	"fmt"

	"github.com/spf13/viper"
)

type AppConfig struct {
	Port     int    `mapstructure:"server.port" yaml:"server.port"`
	LogPath  string `mapstructure:"log.path" yaml:"log.path"`
	LogLevel string `mapstructure:"log.level" yaml:"log.level"`
}

func InitConfig() (*AppConfig, error) {
	v := viper.New()

	// 配置文件设置
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	// 使用项目内conf目录（显式相对路径）
	v.AddConfigPath("./conf")

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("读取配置文件失败: %w", err)
		}
	}

	// 环境变量支持
	v.AutomaticEnv()

	// 默认值设置
	v.SetDefault("server.port", 8080)
	v.SetDefault("log.path", "./logs/app.log")
	v.SetDefault("log.level", "info")

	var cfg AppConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("配置解析失败: %w", err)
	}
	return &cfg, nil
}
