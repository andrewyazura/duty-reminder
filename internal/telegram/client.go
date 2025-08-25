package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/andrewyazura/duty-reminder/internal/config"
)

type Response struct {
	Ok       bool `json:"ok"`
	Response User `json:"result"`
}

type Client struct {
	config *config.TelegramConfig
	client *http.Client
}

func NewClient(config *config.TelegramConfig) *Client {
	return &Client{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

func (c *Client) buildURL(endpoint string) string {
	return fmt.Sprintf("%s/bot%s/%s", c.config.BaseURL, c.config.APIToken, endpoint)
}

func (c *Client) postJSON(ctx context.Context, endpoint string, data any) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		c.buildURL(endpoint),
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var r Response
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return err
	}

	return nil
}

func (c *Client) SendMessage(ctx context.Context, chatID int, text string, opts ...SendMessageOption) error {
	payload := &sendMessagePayload{
		ChatID: chatID,
		Text:   text,
	}

	for _, opt := range opts {
		opt(payload)
	}

	err := c.postJSON(ctx, "sendMessage", payload)
	if err != nil {
		return nil
	}

	return nil
}
