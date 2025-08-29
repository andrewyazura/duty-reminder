package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/andrewyazura/duty-reminder/internal/config"
)

type Result struct {
	Ok          bool            `json:"ok"`
	Description string          `json:"description"`
	Result      json.RawMessage `json:"result,omitempty"`
}

type Client struct {
	config *config.TelegramConfig
	client *http.Client
	logger *slog.Logger
}

func NewClient(config *config.TelegramConfig, logger *slog.Logger) *Client {
	return &Client{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		logger: logger,
	}
}

func (c *Client) buildURL(endpoint string) string {
	return fmt.Sprintf("%s/bot%s/%s", c.config.BaseURL, c.config.APIToken, endpoint)
}

func (c *Client) postJSON(ctx context.Context, endpoint string, data any) (json.RawMessage, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		slog.Error("failed to marshal request json", "endpoint", endpoint, "error", err)
		return nil, err
	}

	url := c.buildURL(endpoint)
	reqBody := bytes.NewBuffer(jsonData)
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		url,
		reqBody,
	)
	if err != nil {
		slog.Error("failed to create http request", "endpoint", endpoint, "error", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	c.logger.Debug("sending telegram api request", "url", url, "body", reqBody)

	resp, err := c.client.Do(req)
	if err != nil {
		slog.Error("http request failed", "endpoint", endpoint, "error", err)
		return nil, err
	}
	defer resp.Body.Close()

	c.logger.Debug("received telegram api response", "status_code", resp.StatusCode)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("failed to read response body", "endpoint", endpoint, "error", err)
		return nil, err
	}

	var result Result
	if err := json.Unmarshal(respBody, &result); err != nil {
		slog.Error("failed to decode response body", "endpoint", endpoint, "body", string(respBody), "error", err)
		return nil, err
	}

	if !result.Ok {
		err := fmt.Errorf("telegram api error: %s", result.Description)
		slog.Error("telegram api returned an error", "endpoint", endpoint, "error", err)
		return nil, err
	}

	return result.Result, nil
}

func (c *Client) SendMessage(ctx context.Context, chatID int, text string, opts ...SendMessageOption) error {
	payload := &sendMessagePayload{
		ChatID: chatID,
		Text:   text,
	}

	for _, opt := range opts {
		opt(payload)
	}

	_, err := c.postJSON(ctx, "sendMessage", payload)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetMe(ctx context.Context) (*User, error) {
	rawResult, err := c.postJSON(ctx, "getMe", nil)
	if err != nil {
		return nil, err
	}

	var user User
	if err := json.Unmarshal(rawResult, &user); err != nil {
		c.logger.Error("failed to decode getMe result", "result", string(rawResult), "error", err)
		return nil, err
	}

	return &user, nil
}
