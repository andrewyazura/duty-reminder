package main

import (
	"context"
	"fmt"
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
	config, err := config.NewConfig()
	if err != nil {
		slog.Error("couldn't build config", "error", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: config.LogLevel,
	}))
	slog.SetDefault(logger)

	pool, err := pgxpool.New(context.Background(), config.Database.URL)
	if err != nil {
		logger.Error("couldn't start a db connection pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	uow := services.NewPostgresUnitOfWork(pool)
	eventBus := eventbus.NewEventBus(logger)

	services.NewTelegramService(eventBus, &config.Telegram, logger, uow)
	services.NewDutyService(eventBus, &config.Telegram, logger, uow)

	s, err := scheduler.New(eventBus, logger, uow)
	if err != nil {
		panic(err)
	}

	s.Start()
	defer s.Shutdown()

	server := server.NewServer(config.Server, config.Telegram, logger, eventBus)

	logger.Info(fmt.Sprintf("starting server on port %s", config.Server.Port))

	if err := http.ListenAndServe(":"+config.Server.Port, server); err != nil {
		logger.Error("server failed to start")
		os.Exit(1)
	}
}
