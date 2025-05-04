package config

type TBCConfig struct {
	Server    ServerConfig    `yaml:"server"`
	Log       LogConfig       `yaml:"log"`
	DB        DBConfig        `yaml:"db"`
	TBCNode   TBCNodeConfig   `yaml:"tbcnode"`
	ElectrumX ElectrumXConfig `yaml:"electrumx"`
}

type ServerConfig struct {
	Name string `yaml:"name"`
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type LogConfig struct {
	Path  string `yaml:"path"`
	Level string `yaml:"level"`
}

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
	URL      string `yaml:"url"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

// ElectrumXConfig ElectrumX RPC客户端配置
type ElectrumXConfig struct {
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	Timeout    int    `yaml:"timeout"`
	RetryCount int    `yaml:"retry_count"`
}
