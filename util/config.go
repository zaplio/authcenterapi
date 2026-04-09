package util

import (
	"time"

	"github.com/spf13/viper"
)

const ConfigName = "config"
const ConfigType = "yaml"

var Configuration Config

type Config struct {
	Server struct {
		Port            int           `mapstructure:"port"`
		Mode            string        `mapstructure:"mode"`
		ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
		Path            struct {
			Messages string `mapstructure:"messages"`
			Contacts string `mapstructure:"contacts"`
			Groups   string `mapstructure:"groups"`
		} `mapstructure:"path"`
	} `mapstructure:"server"`
	Logger struct {
		Dir        string `mapstructure:"dir"`
		FileName   string `mapstructure:"file_name"`
		MaxBackups int    `mapstructure:"max_backups"`
		MaxSize    int    `mapstructure:"max_size"`
		MaxAge     int    `mapstructure:"max_age"`
		Compress   bool   `mapstructure:"compress"`
		LocalTime  bool   `mapstructure:"local_time"`
		Level      string `mapstructure:"level"`
	} `mapstructure:"logger"`
	Postgres struct {
		Host     string   `mapstructure:"host"`
		Port     int      `mapstructure:"port"`
		Username string   `mapstructure:"username"`
		Password string   `mapstructure:"password"`
		Database string   `mapstructure:"database"`
		Options  []string `mapstructure:"options"`
	} `mapstructure:"postgres"`
	Redis struct {
		Host     string   `mapstructure:"host"`
		Port     int      `mapstructure:"port"`
		Username string   `mapstructure:"username"`
		Password string   `mapstructure:"password"`
		DB       int      `mapstructure:"db"`
		Options  []string `mapstructure:"options"`
	} `mapstructure:"redis"`
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig(path string) (cfg *Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName(ConfigName)
	viper.SetConfigType(ConfigType)

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	var config Config
	err = viper.Unmarshal(&config)
	Configuration = config
	return &config, nil
}
