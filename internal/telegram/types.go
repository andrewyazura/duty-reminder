// Package telegram
package telegram

import (
	"encoding/json"
	"strings"
)

type Update struct {
	UpdateID int      `json:"update_id"`
	Message  *Message `json:"message,omitempty"`
}

type Message struct {
	MessageID      int64           `json:"message_id"`
	NewChatMembers []User          `json:"new_chat_members"`
	Chat           Chat            `json:"chat"`
	From           User            `json:"from"`
	Text           string          `json:"text"`
	Entities       []MessageEntity `json:"entities"`
}

type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type Chat struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
}

type MessageEntity struct {
	Type   string `json:"type"`
	Offset int    `json:"offset"`
	Length int    `json:"length"`
}

func (e MessageEntity) Text(m *Message) string {
	text := m.Text[e.Offset+1 : e.Offset+e.Length]

	if i := strings.Index(text, "@"); i > -1 {
		text = text[0:i]
	}

	return text
}

type sendMessagePayload struct {
	ChatID int64  `json:"chat_id"`
	Text   string `json:"text"`

	ParseMode       *string          `json:"parse_mode,omitempty"`
	ReplyParameters *replyParameters `json:"reply_parameters,omitempty"`
	ReplyMarkup     json.RawMessage  `json:"reply_markup,omitempty"`
}

type replyParameters struct {
	MessageID int64 `json:"message_id"`
	ChatID    int64 `json:"chat_id"`
}

type InlineKeyboardMarkup struct {
	Keyboard InlineKeyboard `json:"inline_keyboard"`
}

type InlineKeyboard [][]*InlineKeyboardButton

type InlineKeyboardButton struct {
	Text         string `json:"text"`
	URL          string `json:"url"`
	CallbackData string `json:"callback_data"`
}

type SendMessageOption func(*sendMessagePayload)

func WithParseMode(parseMode string) SendMessageOption {
	return func(p *sendMessagePayload) {
		p.ParseMode = &parseMode
	}
}

func WithReplyParameters(messageID int64, chatID int64) SendMessageOption {
	return func(p *sendMessagePayload) {
		p.ReplyParameters = &replyParameters{MessageID: messageID, ChatID: chatID}
	}
}

func WithInlineKeyboardMarkup(markup InlineKeyboard) SendMessageOption {
	return func(p *sendMessagePayload) {
		p.ReplyMarkup, _ = json.Marshal(InlineKeyboardMarkup{
			Keyboard: markup,
		})
	}
}
