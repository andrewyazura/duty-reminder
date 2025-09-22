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
		c.logger.Error("failed to marshal request json", "endpoint", endpoint, "error", err)
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
		c.logger.Error("failed to create http request", "endpoint", endpoint, "error", err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	c.logger.Debug(
		"sending telegram api request",
		"endpoint",
		endpoint,
		"body",
		string(jsonData),
	)

	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.Error("http request failed", "endpoint", endpoint, "error", err)
		return nil, err
	}
	defer resp.Body.Close()

	c.logger.Debug("received telegram api response", "status_code", resp.StatusCode)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("failed to read response body", "endpoint", endpoint, "error", err)
		return nil, err
	}

	var result Result
	if err := json.Unmarshal(respBody, &result); err != nil {
		c.logger.Error("failed to decode response body", "endpoint", endpoint, "body", string(respBody), "error", err)
		return nil, err
	}

	if !result.Ok {
		err := fmt.Errorf("telegram api error: %s", result.Description)
		c.logger.Error("telegram api returned an error", "endpoint", endpoint, "error", err)
		return nil, err
	}

	return result.Result, nil
}

func (c *Client) SendMessage(chatID int64, text string) *SendMessageBuilder {
	return &SendMessageBuilder{
		client: c,
		payload: sendMessagePayload{
			ChatID: chatID,
			Text:   text,
		},
	}
}

func (c *Client) EditMessageReplyMarkup(
	chatID int64,
	messageID int64,
) *EditMessageReplyMarkupBuilder {
	return &EditMessageReplyMarkupBuilder{
		client: c,
		payload: editMessageReplyMarkupPayload{
			ChatID:    chatID,
			MessageID: messageID,
		},
	}
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

type SendMessageBuilder struct {
	client  *Client
	payload sendMessagePayload
}

func (b *SendMessageBuilder) WithParseMode(parseMode string) *SendMessageBuilder {
	b.payload.ParseMode = &parseMode
	return b
}

func (b *SendMessageBuilder) WithReplyParameters(messageID int64, chatID int64) *SendMessageBuilder {
	b.payload.ReplyParameters = &replyParameters{MessageID: messageID, ChatID: chatID}
	return b
}

func (b *SendMessageBuilder) WithInlineKeyboardMarkup(markup InlineKeyboard) *SendMessageBuilder {
	b.payload.ReplyMarkup = &replyMarkup{InlineKeyboard: markup}
	return b
}

func (b *SendMessageBuilder) Execute(ctx context.Context) error {
	_, err := b.client.postJSON(ctx, "sendMessage", b.payload)
	return err
}

type EditMessageReplyMarkupBuilder struct {
	client  *Client
	payload editMessageReplyMarkupPayload
}

func (b *EditMessageReplyMarkupBuilder) WithInlineKeyboardMarkup(
	markup InlineKeyboard,
) *EditMessageReplyMarkupBuilder {
	b.payload.ReplyMarkup = &replyMarkup{InlineKeyboard: markup}
	return b
}

func (b *EditMessageReplyMarkupBuilder) Execute(ctx context.Context) error {
	_, err := b.client.postJSON(ctx, "editMessageReplyMarkup", b.payload)
	return err
}
