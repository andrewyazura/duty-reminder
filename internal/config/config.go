// Package config
package config

import (
	"encoding/json"
	"os"
	"time"
)

type Config struct {
	Server ServerConfig `json:"server"`
	DB     DBConfig     `json:"database"`
	App    AppConfig    `json:"app"`
}

type ServerConfig struct {
	Port         int           `json:"port"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
}

type DBConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type AppConfig struct {
	Environment string `json:"environment"`
	LogLevel    string `json:"log_level"`
}

func NewConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         1234,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
		DB: DBConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "duty-reminder-app",
			Password: "duty-reminder-app",
			Name:     "duty-reminder-db",
		},
		App: AppConfig{
			Environment: "dev",
			LogLevel:    "debug",
		},
	}
}

func LoadJSONConfigFile(config *Config, filename string) error {
	file, err := os.Open(filename)

	if err != nil {
		return err
	}

	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(config)
}
