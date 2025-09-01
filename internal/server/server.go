// Package server
package server

import (
	"io"
	"log"
	"log/slog"
	"net/http"

	"github.com/andrewyazura/duty-reminder/internal/config"
	"github.com/andrewyazura/duty-reminder/internal/eventbus"
)

type Server struct {
	config          config.ServerConfig
	logger          *slog.Logger
	bus             *eventbus.EventBus
	router          *http.ServeMux
	telegramHandler *TelegramWebhookHandler
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("%v\n", r)
	s.router.ServeHTTP(w, r)
}

func NewServer(
	config config.ServerConfig,
	telegramConfig config.TelegramConfig,
	logger *slog.Logger,
	bus *eventbus.EventBus,
) *Server {
	s := &Server{
		config: config,
		logger: logger,
		bus:    bus,
		router: http.NewServeMux(),
	}

	s.registerRoutes()
	s.telegramHandler = NewTelegramWebhookHandler(
		telegramConfig,
		logger,
		bus,
	)

	return s
}

func (s *Server) registerRoutes() {
	s.router.HandleFunc("/", healthCheck)
	s.router.Handle("/telegram/"+s.config.TelegramRouteSecret, s.telegramHandler)
}

func healthCheck(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "ok")
}
