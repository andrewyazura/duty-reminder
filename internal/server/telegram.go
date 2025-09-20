package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/andrewyazura/duty-reminder/internal/config"
	"github.com/andrewyazura/duty-reminder/internal/eventbus"
	"github.com/andrewyazura/duty-reminder/internal/telegram"
)

type TelegramWebhookHandler struct {
	headerSecret string
	eventBus     *eventbus.EventBus
}

func NewTelegramWebhookHandler(
	config config.TelegramConfig,
	logger *slog.Logger,
	bus *eventbus.EventBus,
) *TelegramWebhookHandler {
	return &TelegramWebhookHandler{
		eventBus:     bus,
		headerSecret: config.HeaderSecret,
	}
}

func (h *TelegramWebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	token := r.Header.Get("X-Telegram-Bot-Api-Secret-Token")

	if token == "" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if token != h.headerSecret {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var update telegram.Update
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	h.eventBus.Publish(context.Background(), "TelegramUpdate", update)
	w.WriteHeader(http.StatusOK)
}
