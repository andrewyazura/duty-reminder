// Package eventbus
package eventbus

import (
	"context"
	"log/slog"
	"runtime/debug"
	"sync"
)

type Event any
type EventType string
type Handler func(context.Context, Event)

type EventBus struct {
	handlers map[EventType][]Handler
	lock     sync.RWMutex
	logger   *slog.Logger
}

func NewEventBus(logger *slog.Logger) *EventBus {
	return &EventBus{
		handlers: make(map[EventType][]Handler),
		logger:   logger,
	}
}

func (eb *EventBus) Subscribe(eventType EventType, handler Handler) {
	eb.lock.Lock()
	defer eb.lock.Unlock()

	eb.handlers[eventType] = append(eb.handlers[eventType], handler)
	eb.logger.Debug("new handler registered", "event", eventType)
}

func (eb *EventBus) Publish(ctx context.Context, eventType EventType, event Event) {
	eb.lock.RLock()
	handlersToCall := make([]Handler, 0, len(eb.handlers[eventType]))
	handlersToCall = append(handlersToCall, eb.handlers[eventType]...)
	eb.lock.RUnlock()

	eb.logger.Info("new event published", "event", eventType, "handlers", len(handlersToCall))
	for _, handler := range handlersToCall {
		go func(h Handler) {
			defer func() {
				if err := recover(); err != nil {
					slog.Error(
						"panic recovered",
						"error", err,
						"stack", string(debug.Stack()),
					)
				}
			}()

			h(ctx, event)
		}(handler)
	}
}
