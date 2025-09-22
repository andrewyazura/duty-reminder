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

	err := s.uow.ExecuteTransaction(ctx, func(repo storage.HouseholdRepository) error {
		household, err := repo.FindByID(ctx, h.TelegramID)

		if err != nil {
			return err
		}

		if len(household.Members) == 0 {
			return nil
		}

		m := household.PopCurrentMember()
		s.client.SendMessage(
			household.TelegramID,
			fmt.Sprintf(
				"🧹 It's [%s](tg://user?id=%d)'s turn to clean",
				m.Name,
				m.TelegramID,
			),
		).Execute(ctx)

		if checklist := household.Checklist; checklist != nil {
			keyboard := telegram.InlineKeyboard{}

			for _, item := range checklist {
				keyboard = append(keyboard,
					[]*telegram.InlineKeyboardButton{
						{
							Text: item,
							CallbackData: fmt.Sprintf(
								"update_checklist:%s",
								item,
							),
						},
					},
				)
			}

			s.client.SendMessage(
				household.TelegramID,
				"List of stuff to complete:",
			).WithInlineKeyboardMarkup(keyboard).Execute(ctx)
		}

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
