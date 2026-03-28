package room

import "testing"

func newTestRoom() *Room {
	return &Room{
		players: [2]*Player{},
		events:  make(chan Event, 10),
	}
}

func TestAddPlayer_FirstPlayerDefaultsToX(t *testing.T) {
	r := newTestRoom()
	p := NewPlayer(nil, "Alice")

	err := r.AddPlayer(p, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.kind != PlayerX {
		t.Fatalf("expected X, got %s", string(p.kind))
	}
}

func TestAddPlayer_FirstPlayerChoosesO(t *testing.T) {
	r := newTestRoom()
	p := NewPlayer(nil, "Alice")

	err := r.AddPlayer(p, "o")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.kind != PlayerO {
		t.Fatalf("expected O, got %s", string(p.kind))
	}
}

func TestAddPlayer_FirstPlayerInvalidType(t *testing.T) {
	r := newTestRoom()
	p := NewPlayer(nil, "Alice")

	err := r.AddPlayer(p, "z")
	if err == nil {
		t.Fatal("expected error for invalid player type")
	}
}
