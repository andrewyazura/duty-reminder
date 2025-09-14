package telegram

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
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
		APIToken: "ABC123",
		Timeout:  5 * time.Second,
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	testClient := NewClient(&testConfig, logger)

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
			ChatID: -1234567898765,
			Text:   "test message",
		}

		handler.handler = func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("got %s method, want %s", r.Method, "POST")
			}

			if !strings.HasSuffix(r.URL.Path, "/sendMessage") {
				t.Errorf("got endpoint %s, want %s", r.URL.Path, "/sendMessage")
			}

			var got sendMessagePayload
			if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
				t.Fatalf("failed to unmarshal request body: %v", err)
			}

			if got.ChatID != want.ChatID {
				t.Errorf("chat_id is %d, want %d", got.ChatID, want.ChatID)
			}

			if got.Text != want.Text {
				t.Errorf("text is %s, want %s", got.Text, want.Text)
			}

			if got.ReplyParameters != nil {
				t.Errorf("expected reply_parameters to be omitted")
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, `{"ok": true, "result": {}}`)
		}

		err := client.SendMessage(ctx, want.ChatID, want.Text)

		if err != nil {
			t.Errorf("SendMessage() returned an error: %v", err)
		}
	})

	t.Run("success, sendMessage with reply parameters", func(t *testing.T) {
		want := replyParameters{123, 12345}

		handler.handler = func(w http.ResponseWriter, r *http.Request) {
			var gotPayload sendMessagePayload
			if err := json.NewDecoder(r.Body).Decode(&gotPayload); err != nil {
				t.Fatalf("failed to unmarshal request body: %v", err)
			}

			got := gotPayload.ReplyParameters
			if got == nil {
				t.Fatalf("expected reply_parameters to be present")
			}

			if got.MessageID != want.MessageID {
				t.Errorf("got message_id %d, want %d", got.MessageID, want.MessageID)
			}

			if got.ChatID != want.ChatID {
				t.Errorf("got chat_id %d, want %d", got.ChatID, want.ChatID)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, `{"ok": true, "result": {}}`)
		}

		err := client.SendMessage(ctx, 1, "t", WithReplyParameters(want.MessageID, want.ChatID))

		if err != nil {
			t.Errorf("SendMessage() returned an error: %v", err)
		}
	})
}

func TestGetMe(t *testing.T) {
	client, handler, teardown := getTestClient(t)
	defer teardown()

	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		handler.handler = func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, `{
				"ok": true,
				"result": {
					"id": -1234567898765,
					"username": "test_bot",
					"first_name": "test_first_name",
					"last_name": "test_last_name"
				}
			}`)
		}

		wantUser := User{
			ID:        -1234567898765,
			Username:  "test_bot",
			FirstName: "test_first_name",
			LastName:  "test_last_name",
		}

		gotUser, err := client.GetMe(ctx)
		if err != nil {
			t.Errorf("GetMe() returned an error: %v", err)
		}

		if !reflect.DeepEqual(gotUser, &wantUser) {
			t.Errorf("got %v, want %v", gotUser, &wantUser)
		}
	})
}
