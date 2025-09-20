// Package server
package server

import (
	"bytes"
	"io"
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

	s.telegramHandler = NewTelegramWebhookHandler(
		telegramConfig,
		logger,
		bus,
	)

	s.registerRoutes()

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var bodyStr string
	if r.Body != nil {
		bodyBuf, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBuf))

		bodyStr = string(bodyBuf)
		if len(bodyBuf) > s.config.MaxLoggedBodySize {
			bodyStr = bodyStr[:s.config.MaxLoggedBodySize] + "... [truncated]"
		}
	}

	s.logger.Debug("incoming request",
		"method", r.Method,
		"path", r.URL.Path,
		"remote", r.RemoteAddr,
		"body", bodyStr,
	)

	s.router.ServeHTTP(w, r)
}

func (s *Server) registerRoutes() {
	s.router.HandleFunc("/", healthCheck)
	s.router.Handle("/telegram/"+s.config.TelegramRouteSecret, s.telegramHandler)
}

func healthCheck(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "ok")
}
