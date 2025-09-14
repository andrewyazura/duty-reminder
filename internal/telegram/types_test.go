package telegram

import "testing"

func TestMessageEntity_Text(t *testing.T) {
	message := Message{
		Text: "/register@reminder_of_duty_bot",
	}

	entity := MessageEntity{
		Type:   "bot_command",
		Offset: 0,
		Length: 30,
	}

	command := entity.Text(&message)
	if command != "register" {
		t.Fatalf("got %s, want %s", command, "register")
	}
}
