// Package eventbus
package eventbus

import (
	"log/slog"
	"sync"
)

type Event any
type EventType string
type Handler func(Event)

type EventBus struct {
	table map[EventType][]Handler
	lock  sync.RWMutex
}

func NewEventBus() *EventBus {
	return &EventBus{
		table: make(map[EventType][]Handler),
	}
}

func (eb *EventBus) Subscribe(eventType EventType, handler Handler) {
	eb.lock.Lock()
	defer eb.lock.Unlock()

	eb.table[eventType] = append(eb.table[eventType], handler)
}

func (eb *EventBus) Publish(eventType EventType, event Event) {
	eb.lock.RLock()
	handlersToCall := make([]Handler, 0, len(eb.table[eventType]))
	handlersToCall = append(handlersToCall, eb.table[eventType]...)
	eb.lock.RUnlock()

	for _, handler := range handlersToCall {
		go func(h Handler) {
			defer func() {
				if r := recover(); r != nil {
					slog.Error("recovered error: %v", r)
				}
			}()

			h(event)
		}(handler)
	}
}
