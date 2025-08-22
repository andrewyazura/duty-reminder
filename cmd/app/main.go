package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/andrewyazura/duty-reminder/internal/config"
	"github.com/andrewyazura/duty-reminder/internal/eventbus"
	"github.com/andrewyazura/duty-reminder/internal/routes"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	config, err := config.NewConfig()
	if err != nil {
		logger.Error("couldn't build config", "error", err)
	}

	eventBus := eventbus.NewEventBus()
	server := routes.NewServer(config.Server, config.Telegram, logger, eventBus)

	logger.Info("starting server on port 8080")

	if err := http.ListenAndServe(":"+config.Server.Port, server); err != nil {
		logger.Error("server failed to start")
		os.Exit(1)
	}
}
