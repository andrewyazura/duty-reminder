package main

type Household struct {
	Members       []*Member
	CurrentMember int
	Crontab       string
}

func NewHousehold() Household {
	return Household{
		Members:       []*Member{},
		CurrentMember: 0,
		Crontab:       "",
	}
}

func (h *Household) AddMember(m Member) {
	h.Members = append(h.Members, &m)
}

func (h *Household) RemoveMember(m Member) {
	for i, hm := range h.Members {
		if m.TelegramID == hm.TelegramID {
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
	TelegramID int
}
