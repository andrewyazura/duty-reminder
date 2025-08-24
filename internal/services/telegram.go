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

func (t TelegramService) HandleUpdate(event eventbus.Event) {
	update := event.(telegram.Update)
	message := update.Message

	// someone was added to a group
	if newMembers := message.NewChatMembers; newMembers != nil && message.Chat.Type == "group" {
		for _, m := range newMembers {

			// new member is the bot itself
			if m.ID == t.config.BotID {
				t.handleNewGroup(message)
			}
		}
	}

	if message.Entities != nil {
		for _, e := range message.Entities {
			if e.Type == "bot_command" {
				t.handleCommand(message, &e)
			}
		}
	}
}

func (t TelegramService) handleNewGroup(message *telegram.Message) {
	t.uow.Execute(context.Background(), func(repo storage.HouseholdRepository) error {
		_, err := repo.FindByID(context.Background(), message.Chat.ID)

		if err == nil {
			return nil
		}

		// request members
		members := make([]*domain.Member, 0)

		household := &domain.Household{
			TelegramID: message.Chat.ID,
			Members:    members,
		}

		repo.Create(context.Background(), household)
		t.bus.Publish("HouseholdCreated", household)

		return nil
	})
}

func (t TelegramService) handleCommand(message *telegram.Message, entity *telegram.MessageEntity) {
	text := entity.Text(message)

	switch text {
	case "register":
		t.addUser(message)
	case "help":
		t.help(message)
	case "skip":
		t.skip(message)
	default:
		t.unknownCommand(message)
	}
}

func (t TelegramService) addUser(message *telegram.Message) {
	t.uow.Execute(context.Background(), func(repo storage.HouseholdRepository) error {
		household, err := repo.FindByID(context.Background(), message.Chat.ID)

		if err != nil {
			return err
		}

		user := message.From

		for _, m := range household.Members {
			if m.TelegramID == user.ID {
				return nil
			}
		}

		member := &domain.Member{
			TelegramID: user.ID,
			Name:       user.FirstName + " " + user.SecondName,
		}

		household.AddMember(member)

		repo.Save(context.Background(), household)

		return nil
	})
}

func (t TelegramService) help(message *telegram.Message) {
	t.client.sendMessage(message.Chat.ID, "/register to register in the household")
}

func (t TelegramService) skip(message *telegram.Message) {
}

func (t TelegramService) unknownCommand(message *telegram.Message) {
	t.client.sendMessage(message.Chat.ID, "Unknown command")
}
