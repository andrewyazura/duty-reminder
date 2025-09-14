// Package telegram
package telegram

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
	return m.Text[e.Offset : e.Offset+e.Length]
}

type sendMessagePayload struct {
	ChatID          int64            `json:"chat_id"`
	Text            string           `json:"text"`
	ReplyParameters *replyParameters `json:"reply_parameters,omitempty"`
}

type replyParameters struct {
	MessageID int64
	ChatID    int64
}

type SendMessageOption func(*sendMessagePayload)

func WithReplyParameters(messageID int64, chatID int64) SendMessageOption {
	return func(p *sendMessagePayload) {
		p.ReplyParameters = &replyParameters{MessageID: messageID, ChatID: chatID}
	}
}
