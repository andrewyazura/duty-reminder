// Package eventbus
package eventbus

import "sync"

type EventType string
type Event struct {
	Type EventType
}

type Handler func(Event)

type EventBus struct {
	table map[EventType][]Handler
	lock    sync.RWMutex
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

func (eb *EventBus) Publish(event Event) {
	eb.lock.RLock()
	handlers := eb.table[event.Type]
	eb.lock.RUnlock()

	for _, h := range handlers {
		h(event)
	}
}
