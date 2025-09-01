package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/andrewyazura/duty-reminder/internal/config"
	"github.com/andrewyazura/duty-reminder/internal/eventbus"
	"github.com/andrewyazura/duty-reminder/internal/scheduler"
	"github.com/andrewyazura/duty-reminder/internal/server"
	"github.com/andrewyazura/duty-reminder/internal/services"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	config, err := config.NewConfig()
	if err != nil {
		logger.Error("couldn't build config", "error", err)
	}

	pool, err := pgxpool.New(context.Background(), config.Database.URL)
	if err != nil {
		logger.Error("couldn't start a db connection pool", "error", err)
	}
	defer pool.Close()

	uow := services.NewPostgresUnitOfWork(pool)
	eventBus := eventbus.NewEventBus(logger)

	services.NewTelegramService(eventBus, &config.Telegram, logger, uow)
	services.NewDutyService(eventBus, &config.Telegram, logger, uow)

	scheduler, err := scheduler.New(eventBus, logger, uow)
	if err != nil {
		panic(err)
	}

	scheduler.Start()
	defer scheduler.Shutdown()

	server := server.NewServer(config.Server, config.Telegram, logger, eventBus)

	logger.Info("starting server on port 8080")

	if err := http.ListenAndServe(":"+config.Server.Port, server); err != nil {
		logger.Error("server failed to start")
		os.Exit(1)
	}
}
