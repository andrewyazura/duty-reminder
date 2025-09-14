// Package telegram
package telegram

type Update struct {
	UpdateID int      `json:"update_id"`
	Message  *Message `json:"message,omitempty"`
}

type Message struct {
	MessageID      int             `json:"message_id"`
	NewChatMembers []User          `json:"new_chat_members"`
	Chat           Chat            `json:"chat"`
	From           User            `json:"from"`
	Text           string          `json:"text"`
	Entities       []MessageEntity `json:"entities"`
}

type User struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type Chat struct {
	ID   int    `json:"id"`
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
	ChatID          int              `json:"chat_id"`
	Text            string           `json:"text"`
	ReplyParameters *replyParameters `json:"reply_parameters,omitempty"`
}

type replyParameters struct {
	MessageID int
	ChatID    int
}

type SendMessageOption func(*sendMessagePayload)

func WithReplyParameters(messageID int, chatID int) SendMessageOption {
	return func(p *sendMessagePayload) {
		p.ReplyParameters = &replyParameters{MessageID: messageID, ChatID: chatID}
	}
}
