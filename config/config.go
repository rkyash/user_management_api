package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Log      LogConfig
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessExpiry  int // minutes
	RefreshExpiry int // days
}

type LogConfig struct {
	Level string
	File  string
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	viper.SetDefault("server.port", "8080")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("jwt.accessExpiry", 15) // 15 minutes
	viper.SetDefault("jwt.refreshExpiry", 7) // 7 days
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.file", "logs/app.log")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
