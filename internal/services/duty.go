package services

import (
	"context"
	"log/slog"

	"github.com/andrewyazura/duty-reminder/internal/config"
	"github.com/andrewyazura/duty-reminder/internal/domain"
	"github.com/andrewyazura/duty-reminder/internal/eventbus"
	"github.com/andrewyazura/duty-reminder/internal/telegram"
)

type DutyService struct {
	config *config.TelegramConfig
	client *telegram.Client
	logger *slog.Logger
	uow    UnitOfWork
}

func NewDutyService(
	bus *eventbus.EventBus,
	config *config.TelegramConfig,
	logger *slog.Logger,
	uow UnitOfWork,
) *DutyService {
	s := &DutyService{
		config: config,
		client: telegram.NewClient(config, logger),
		logger: logger,
		uow:    uow,
	}

	bus.Subscribe("NotifyHousehold", s.NotifyHousehold)
	return s
}

func (s DutyService) NotifyHousehold(ctx context.Context, event eventbus.Event) {
	h := event.(domain.Household)

	s.client.SendMessage(ctx, h.TelegramID, "Go clean")
}
