package main

import (
	"sync/atomic"
	"testing"
)

func TestSubscribe(t *testing.T) {
	eb := NewEventBus()

	eb.Subscribe("event-1", func(e Event) {})
	eb.Subscribe("event-1", func(e Event) {})

	handlers, ok := eb.table["event-1"]
	if !ok || len(handlers) != 2 {
		t.Fatalf("no 'event-1' handlers registered, want 2")
	}

	eb.Subscribe("event-2", func(e Event) {})
	eb.Subscribe("event-3", func(e Event) {})

	if gotLen := len(eb.table); gotLen != 3 {
		t.Fatalf("%d event types in table, expected %d", gotLen, 3)
	}
}

func TestPublish(t *testing.T) {
	eb := NewEventBus()
	var count atomic.Int32

	eb.Subscribe("event-1", func(e Event) {
		count.Add(1)
	})
	eb.Subscribe("event-1", func(e Event) {
		count.Add(1)
	})
	eb.Subscribe("event-2", func(e Event) {
		count.Add(1)
	})

	eb.Publish(Event{Type: "event-1"})

	if gotCount := count.Load(); gotCount != 2 {
		t.Fatalf("after publishing, count is %d, want %d", gotCount, 2)
	}
}
