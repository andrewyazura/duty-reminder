// Package config
package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Telegram TelegramConfig
}

type ServerConfig struct {
	Port                string
	TelegramRouteSecret string
}

type DatabaseConfig struct{}

type TelegramConfig struct {
	HeaderSecret string
	APIToken     string
}

func NewConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	config := &Config{
		Server: ServerConfig{
			Port:                "8080",
			TelegramRouteSecret: "secret",
		},
		Database: DatabaseConfig{},
		Telegram: TelegramConfig{
			HeaderSecret: "secret",
			APIToken:     "token",
		},
	}

	if v := os.Getenv("SERVER_PORT"); v != "" {
		config.Server.Port = v
	}

	if v := os.Getenv("SERVER_TELEGRAM_ROUTE_SECRET"); v != "" {
		config.Server.TelegramRouteSecret = v
	}

	if v := os.Getenv("TELEGRAM_HEADER_SECRET"); v != "" {
		config.Telegram.HeaderSecret = v
	}

	if v := os.Getenv("TELEGRAM_API_TOKEN"); v != "" {
		config.Telegram.APIToken = v
	}

	return config, nil
}
