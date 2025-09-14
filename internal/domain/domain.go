// Package domain
package domain

type Household struct {
	Checklist     []string
	Crontab       string
	CurrentMember int
	Members       []*Member
	TelegramID    int64
}

func NewHousehold(telegramID int64) *Household {
	return &Household{
		Checklist:     []string{},
		Crontab:       "0 9 * * 6", // at 9:00 on Saturday
		CurrentMember: 0,
		Members:       []*Member{},
		TelegramID:    telegramID,
	}
}

func (h *Household) AddMember(m *Member) {
	m.Order = len(h.Members)
	h.Members = append(h.Members, m)
}

func (h *Household) RemoveMember(telegramID int64) {
	for i, m := range h.Members {
		if telegramID == m.TelegramID {
			h.Members = append(h.Members[:i], h.Members[i+1:]...)
			return
		}
	}
}

func (h *Household) PopCurrentMember() *Member {
	m := h.Members[h.CurrentMember]
	h.CurrentMember++
	h.CurrentMember %= len(h.Members)

	return m
}

type Member struct {
	Name       string
	TelegramID int64
	Order      int
}
