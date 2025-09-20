package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/andrewyazura/duty-reminder/internal/config"
	"github.com/andrewyazura/duty-reminder/internal/domain"
	"github.com/andrewyazura/duty-reminder/internal/eventbus"
	"github.com/andrewyazura/duty-reminder/internal/storage"
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
	h := event.(*domain.Household)

	err := s.uow.Execute(ctx, func(repo storage.HouseholdRepository) error {
		household, err := repo.FindByID(ctx, h.TelegramID)

		if err != nil {
			return err
		}

		if len(household.Members) == 0 {
			return nil
		}

		m := household.PopCurrentMember()
		s.client.SendMessage(
			ctx,
			household.TelegramID,
			fmt.Sprintf(
				"It's [%s](tg://user?id=%d)'s turn to clean this week",
				m.Name,
				m.TelegramID,
			),
		)

		err = repo.SaveWithMembers(ctx, household)

		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		s.logger.Error("something went wrong", "error", err)
	}
}
