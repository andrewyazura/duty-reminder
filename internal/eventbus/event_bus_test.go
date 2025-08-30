package eventbus

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
)

func TestSubscribe(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	eb := NewEventBus(logger)

	a := func(ctx context.Context, e Event) {}
	eb.Subscribe("event-1", a)
	eb.Subscribe("event-1", a)

	handlers, ok := eb.handlers["event-1"]
	if !ok || len(handlers) != 2 {
		t.Fatalf("no 'event-1' handlers registered, want 2")
	}

	eb.Subscribe("event-2", a)
	eb.Subscribe("event-3", a)

	if gotLen := len(eb.handlers); gotLen != 3 {
		t.Fatalf("%d event types in table, expected %d", gotLen, 3)
	}
}

func TestPublish(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	eb := NewEventBus(logger)
	var count atomic.Int32

	var wg sync.WaitGroup
	wg.Add(2)

	handler := func(ctx context.Context, e Event) {
		defer wg.Done()
		count.Add(1)
	}

	eb.Subscribe("event-1", handler)
	eb.Subscribe("event-1", handler)
	eb.Subscribe("event-2", handler)

	eb.Publish(context.Background(), "event-1", struct{}{})
	wg.Wait()

	if gotCount := count.Load(); gotCount != 2 {
		t.Fatalf("after publishing, count is %d, want %d", gotCount, 2)
	}
}
