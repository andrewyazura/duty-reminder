package eventbus

import (
	"sync"
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

	var wg sync.WaitGroup
	wg.Add(2)

	handler := func(e Event) {
		defer wg.Done()
		count.Add(1)
	}

	eb.Subscribe("event-1", handler)
	eb.Subscribe("event-1", handler)
	eb.Subscribe("event-2", handler)

	eb.Publish("event-1", struct{}{})
	wg.Wait()

	if gotCount := count.Load(); gotCount != 2 {
		t.Fatalf("after publishing, count is %d, want %d", gotCount, 2)
	}
}
