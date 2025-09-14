// Package services
package services

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/andrewyazura/duty-reminder/internal/config"
	"github.com/andrewyazura/duty-reminder/internal/domain"
	"github.com/andrewyazura/duty-reminder/internal/eventbus"
	"github.com/andrewyazura/duty-reminder/internal/storage"
	"github.com/andrewyazura/duty-reminder/internal/telegram"
	"github.com/robfig/cron/v3"
)

type TelegramService struct {
	bus    *eventbus.EventBus
	config *config.TelegramConfig
	client *telegram.Client
	logger *slog.Logger
	uow    UnitOfWork
}

func NewTelegramService(
	bus *eventbus.EventBus,
	config *config.TelegramConfig,
	logger *slog.Logger,
	uow UnitOfWork,
) *TelegramService {
	s := &TelegramService{
		bus:    bus,
		config: config,
		client: telegram.NewClient(config, logger),
		logger: logger,
		uow:    uow,
	}

	bus.Subscribe("TelegramUpdate", s.HandleUpdate)

	return s
}

func (s *TelegramService) HandleUpdate(
	ctx context.Context,
	event eventbus.Event,
) {
	update := event.(telegram.Update)
	s.logger.Debug("update received", "update", update)

	message := update.Message

	if t := message.Chat.Type; t != "group" && t != "supergroup" {
		s.client.SendMessage(ctx, message.Chat.ID, "Sorry, I only work in groups")
		return
	}

	// someone was added to a group
	if newMembers := message.NewChatMembers; newMembers != nil {
		for _, m := range newMembers {

			// new member is the bot itself
			if m.ID == s.config.BotID {
				s.handleNewGroup(ctx, message)
				return
			}
		}
	}

	if message.Entities != nil {
		for _, e := range message.Entities {
			if e.Type == "bot_command" {
				s.handleCommand(ctx, message, &e)
				return
			}
		}
	}
}

func (s *TelegramService) handleNewGroup(
	ctx context.Context,
	message *telegram.Message,
) {
	var household *domain.Household

	err := s.uow.Execute(ctx, func(repo storage.HouseholdRepository) error {
		_, err := repo.FindByID(ctx, message.Chat.ID)

		if err == nil {
			return nil
		}

		household = domain.NewHousehold(message.Chat.ID)

		err = repo.Create(ctx, household)
		if err != nil {
			return err
		}

		s.bus.Publish(ctx, "HouseholdCreated", household)

		return nil
	})

	if err != nil {
		s.logger.Error("something went wrong", "error", err)
		return
	}

	s.client.SendMessage(ctx, message.Chat.ID, fmt.Sprintf(`
		Hey! Group chat was successfully added.

		Your current schedule is %s

		To register as a member, please use /register
	`, household.Crontab))

	s.bus.Publish(ctx, "HouseholdCreated", household)
}

func (s *TelegramService) handleCommand(
	ctx context.Context,
	message *telegram.Message,
	entity *telegram.MessageEntity,
) {
	command := entity.Text(message)

	switch command {
	case "register":
		s.register(ctx, message)
	case "setSchedule":
		s.setSchedule(ctx, message)
	case "help":
		s.help(ctx, message)
	case "skip":
		s.skip(ctx, message)
	default:
		s.unknownCommand(ctx, message)
	}
}

func (s *TelegramService) register(
	ctx context.Context,
	message *telegram.Message,
) {
	err := s.uow.Execute(ctx, func(repo storage.HouseholdRepository) error {
		household, err := repo.FindByID(ctx, message.Chat.ID)

		if err != nil {
			return err
		}

		user := message.From

		for _, m := range household.Members {
			if m.TelegramID == user.ID {
				s.client.SendMessage(
					ctx,
					message.Chat.ID,
					"already registered",
					telegram.WithReplyParameters(message.MessageID, message.Chat.ID),
				)

				return nil
			}
		}

		member := &domain.Member{
			TelegramID: user.ID,
			Name:       user.FirstName + " " + user.LastName,
		}

		household.AddMember(member)
		repo.SaveWithMembers(ctx, household)

		return nil
	})

	s.client.SendMessage(
		ctx,
		message.Chat.ID,
		"successfully registered",
		telegram.WithReplyParameters(message.MessageID, message.Chat.ID),
	)

	if err != nil {
		s.logger.Error("something went wrong", "error", err)
	}
}

func (s *TelegramService) setSchedule(
	ctx context.Context,
	message *telegram.Message,
) {
	parts := strings.Split(message.Text, " ")

	if len(parts) == 1 {
		s.client.SendMessage(ctx, message.Chat.ID, "no args, usage: /setSchedule 0 9 * * 5")
		return
	}

	newCrontab := parts[1]
	cronParser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	if _, err := cronParser.Parse(newCrontab); err != nil {
		s.client.SendMessage(ctx, message.Chat.ID, "invalid crontab string, example: 0 9 * * 5")
		return
	}

	s.uow.Execute(ctx, func(repo storage.HouseholdRepository) error {
		h, err := repo.FindByID(ctx, message.Chat.ID)
		if err != nil {
			return err
		}

		h.Crontab = newCrontab

		err = repo.Save(ctx, h)
		if err != nil {
			return err
		}

		return nil
	})

	s.client.SendMessage(ctx, message.Chat.ID, "new crontab string saved")
}

func (s *TelegramService) help(ctx context.Context, message *telegram.Message) {
	s.client.SendMessage(ctx, message.Chat.ID, "/register to register in the household")
}

func (s *TelegramService) skip(ctx context.Context, message *telegram.Message) {
	s.client.SendMessage(ctx, message.Chat.ID, "/skip")
}

func (s *TelegramService) unknownCommand(ctx context.Context, message *telegram.Message) {
	s.client.SendMessage(ctx, message.Chat.ID, "Unknown command")
}
