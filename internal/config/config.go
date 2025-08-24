// Package config
package config

import (
	"log"
	"os"
	"strconv"
	"time"

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

type DatabaseConfig struct {
	URL string
}

type TelegramConfig struct {
	APIToken     string
	BaseURL      string
	BotID        int
	HeaderSecret string
	Timeout      time.Duration
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
			APIToken:     "token",
			BaseURL:      "https://api.telegram.org",
			BotID:        1234,
			HeaderSecret: "secret",
			Timeout:      30 * time.Second,
		},
	}

	if v := os.Getenv("SERVER_PORT"); v != "" {
		config.Server.Port = v
	}

	if v := os.Getenv("SERVER_TELEGRAM_ROUTE_SECRET"); v != "" {
		config.Server.TelegramRouteSecret = v
	}

	if v := os.Getenv("DATABASE_URL"); v != "" {
		config.Database.URL = v
	}

	if v := os.Getenv("TELEGRAM_API_TOKEN"); v != "" {
		config.Telegram.APIToken = v
	}

	if v := os.Getenv("TELEGRAM_BASE_URL"); v != "" {
		config.Telegram.BaseURL = v
	}

	if v := os.Getenv("TELEGRAM_BOT_ID"); v != "" {
		i, err := strconv.Atoi(v)
		if err != nil {
			log.Fatalf("invalid config param TELEGRAM_BOT_ID: %v", err)
		}

		config.Telegram.BotID = i
	}

	if v := os.Getenv("TELEGRAM_HEADER_SECRET"); v != "" {
		config.Telegram.HeaderSecret = v
	}

	if v := os.Getenv("TELEGRAM_TIMEOUT"); v != "" {
		i, err := strconv.Atoi(v)
		if err != nil {
			log.Fatalf("invalid config param TELEGRAM_TIMEOUT: %v", err)
		}

		config.Telegram.Timeout = time.Duration(i) * time.Second
	}

	return config, nil
}
