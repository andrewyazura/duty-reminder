package domain

import "testing"

func TestAddMember(t *testing.T) {
	h := NewHousehold(-1234567898765)
	alice := Member{Name: "Alice", TelegramID: 1}
	bob := Member{Name: "Bob", TelegramID: 2}

	h.AddMember(&alice)

	if gotLen := len(h.Members); gotLen != 1 {
		t.Fatalf("members list has length %d, want %d", gotLen, 1)
	}

	if gotAlice := *h.Members[0]; gotAlice != alice {
		t.Fatalf("got [ %v ], want [ %v ]", gotAlice, alice)
	}

	h.AddMember(&bob)

	if gotLen := len(h.Members); gotLen != 2 {
		t.Fatalf("members list has length %d, want %d", gotLen, 2)
	}

	if gotBob := *h.Members[1]; gotBob != bob {
		t.Fatalf("got [ <alice>, %v ], want [ <alice>, %v ]", gotBob, bob)
	}
}

func TestRemoveMember(t *testing.T) {
	h := NewHousehold(-1234567898765)

	alice := Member{Name: "Alice", TelegramID: 1}
	bob := Member{Name: "Bob", TelegramID: 2}

	h.AddMember(&alice)
	h.AddMember(&bob)

	h.RemoveMember(alice.TelegramID)

	for _, m := range h.Members {
		if m.Name == "Alice" {
			t.Fatalf("Alice still in household")
		}
	}

	h.AddMember(&alice)
	h.RemoveMember(bob.TelegramID)

	for _, m := range h.Members {
		if m.Name == "Bob" {
			t.Fatalf("Bob still in household")
		}
	}

	h.RemoveMember(alice.TelegramID)

	if gotLen := len(h.Members); gotLen != 0 {
		t.Fatalf("members list has length %d, want %d", gotLen, 0)
	}
}

func TestPopCurrentMember(t *testing.T) {
	h := NewHousehold(-1234567898765)

	alice := Member{Name: "Alice", TelegramID: 1}
	bob := Member{Name: "Bob", TelegramID: 2}
	charlie := Member{Name: "Charlie", TelegramID: 3}

	h.AddMember(&alice)
	h.AddMember(&bob)
	h.AddMember(&charlie)

	if gotCurrent := h.PopCurrentMember(); *gotCurrent != alice {
		t.Fatalf("popped %v, want %v", gotCurrent, alice)
	}

	if gotCurrent := h.PopCurrentMember(); *gotCurrent != bob {
		t.Fatalf("popped %v, want %v", gotCurrent, bob)
	}

	if gotCurrent := h.PopCurrentMember(); *gotCurrent != charlie {
		t.Fatalf("popped %v, want %v", gotCurrent, charlie)
	}

	if gotCurrent := h.PopCurrentMember(); *gotCurrent != alice {
		t.Fatalf("popped %v, want %v", gotCurrent, alice)
	}
}
