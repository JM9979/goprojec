package config

type TBCConfig struct {
	Server ServerConfig `yaml:"server"`
	Log    LogConfig    `yaml:"log"`
	DB     DBConfig     `yaml:"db"`
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
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}
