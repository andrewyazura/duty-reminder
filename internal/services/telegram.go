// Package services
package services

import (
	"context"

	"github.com/andrewyazura/duty-reminder/internal/config"
	"github.com/andrewyazura/duty-reminder/internal/domain"
	"github.com/andrewyazura/duty-reminder/internal/eventbus"
	"github.com/andrewyazura/duty-reminder/internal/storage"
	"github.com/andrewyazura/duty-reminder/internal/telegram"
)

type TelegramService struct {
	bus    *eventbus.EventBus
	config *config.TelegramConfig
	client *telegram.Client
	uow    UnitOfWork
}

func NewTelegramService(bus *eventbus.EventBus, config *config.TelegramConfig, uow UnitOfWork) *TelegramService {
	return &TelegramService{
		bus:    bus,
		config: config,
		client: telegram.NewClient(config),
		uow:    uow,
	}
}

func (t TelegramService) HandleUpdate(ctx context.Context, event eventbus.Event) {
	update := event.(telegram.Update)
	message := update.Message

	// someone was added to a group
	if newMembers := message.NewChatMembers; newMembers != nil && message.Chat.Type == "group" {
		for _, m := range newMembers {

			// new member is the bot itself
			if m.ID == t.config.BotID {
				t.handleNewGroup(ctx, message)
			}
		}
	}

	if message.Entities != nil {
		for _, e := range message.Entities {
			if e.Type == "bot_command" {
				t.handleCommand(ctx, message, &e)
			}
		}
	}
}

func (t TelegramService) handleNewGroup(ctx context.Context, message *telegram.Message) {
	t.uow.Execute(ctx, func(repo storage.HouseholdRepository) error {
		_, err := repo.FindByID(ctx, message.Chat.ID)

		if err == nil {
			return nil
		}

		// request members
		members := make([]*domain.Member, 0)

		household := &domain.Household{
			TelegramID: message.Chat.ID,
			Members:    members,
		}

		repo.Create(ctx, household)
		t.bus.Publish(ctx, "HouseholdCreated", household)

		return nil
	})
}

func (t TelegramService) handleCommand(ctx context.Context, message *telegram.Message, entity *telegram.MessageEntity) {
	text := entity.Text(message)

	switch text {
	case "register":
		t.addUser(ctx, message)
	case "help":
		t.help(ctx, message)
	case "skip":
		t.skip(ctx, message)
	default:
		t.unknownCommand(ctx, message)
	}
}

func (t TelegramService) addUser(ctx context.Context, message *telegram.Message) {
	t.uow.Execute(ctx, func(repo storage.HouseholdRepository) error {
		household, err := repo.FindByID(ctx, message.Chat.ID)

		if err != nil {
			return err
		}

		user := message.From

		for _, m := range household.Members {
			if m.TelegramID == user.ID {
				t.client.SendMessage(
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
			Name:       user.FirstName + " " + user.SecondName,
		}

		household.AddMember(member)

		repo.Save(ctx, household)

		return nil
	})
}

func (t TelegramService) help(ctx context.Context, message *telegram.Message) {
	t.client.SendMessage(ctx, message.Chat.ID, "/register to register in the household")
}

func (t TelegramService) skip(ctx context.Context, message *telegram.Message) {
	t.client.SendMessage(ctx, message.Chat.ID, "/skip")
}

func (t TelegramService) unknownCommand(ctx context.Context, message *telegram.Message) {
	t.client.SendMessage(ctx, message.Chat.ID, "Unknown command")
}
