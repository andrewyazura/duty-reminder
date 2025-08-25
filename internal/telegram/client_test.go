package telegram

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/andrewyazura/duty-reminder/internal/config"
)

type mockHandler struct {
	handler http.HandlerFunc
}

func (m *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.handler(w, r)
}

func getTestClient(t *testing.T) (*Client, *mockHandler, func()) {
	t.Helper()

	mock := &mockHandler{}
	server := httptest.NewServer(mock)

	testConfig := config.TelegramConfig{
		BaseURL:  server.URL,
		APIToken: "test-token",
		Timeout:  5 * time.Second,
	}

	testClient := NewClient(&testConfig)

	teardownFunc := func() {
		server.Close()
	}

	return testClient, mock, teardownFunc
}

func TestSendMessage(t *testing.T) {
	client, handler, teardown := getTestClient(t)
	defer teardown()

	ctx := context.Background()

	t.Run("success, simple sendMessage", func(t *testing.T) {
		want := sendMessagePayload{
			ChatID: 12345,
			Text:   "test message",
		}

		handler.handler = func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("got %s method, want %s", r.Method, "POST")
			}

			if !strings.HasSuffix(r.URL.Path, "/sendMessage") {
				t.Errorf("got endpoint %s, want %s", r.URL.Path, "/sendMessage")
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("failed to read request body: %v", err)
			}

			var payload map[string]any
			if err := json.Unmarshal(body, &payload); err != nil {
				t.Fatalf("failed to unmarshal request body: %v", err)
			}

			if got := int(payload["chat_id"].(float64)); got != want.ChatID {
				t.Errorf("chat_id is %d, want %d", got, want.ChatID)
			}

			if got := payload["text"]; got != want.Text {
				t.Errorf("text is %s, want %s", got, want.Text)
			}

			if _, ok := payload["reply_parameters"]; ok {
				t.Errorf("expected reply_parameters to be ommited")
			}
		}

		err := client.SendMessage(ctx, want.ChatID, want.Text)

		if err != nil {
			t.Errorf("SendMessage() returned an error: %v", err)
		}
	})

	t.Run("success, sendMessage with reply parameters", func(t *testing.T) {
		want := replyParameters{123, 12345}

		handler.handler = func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			var payload map[string]any
			json.Unmarshal(body, &payload)

			got := payload["reply_parameters"].(map[string]int)
			if got == nil {
				t.Fatalf("expected reply_parameters to be present")
			}

			if v := got["message_id"]; v != want.MessageID {
				t.Errorf("got message_id %d, want %d", v, want.MessageID)
			}

			if v := got["chat_id"]; v != want.ChatID {
				t.Errorf("got chat_id %d, want %d", v, want.ChatID)
			}
		}

		err := client.SendMessage(ctx, 1, "t", WithReplyParameters(want.MessageID, want.ChatID))

		if err != nil {
			t.Errorf("SendMessage() returned an error: %v", err)
		}
	})
}
