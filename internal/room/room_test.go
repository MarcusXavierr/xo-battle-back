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

func TestAddPlayer_SecondPlayerGetsOpposite(t *testing.T) {
	r := newTestRoom()
	p1 := &Player{name: "Alice", send: make(chan []byte, 2)}
	p2 := &Player{name: "Bob", send: make(chan []byte, 2)}

	r.AddPlayer(p1, "o")
	r.AddPlayer(p2, "x") // preference ignored

	if p2.kind != PlayerX {
		// p1 is O, so p2 must be X
		t.Fatalf("expected X, got %s", string(p2.kind))
	}
}

func TestAddPlayer_SecondPlayerIgnoresPreference(t *testing.T) {
	r := newTestRoom()
	p1 := &Player{name: "Alice", send: make(chan []byte, 2)}
	p2 := &Player{name: "Bob", send: make(chan []byte, 2)}

	r.AddPlayer(p1, "x")
	r.AddPlayer(p2, "x") // wants X but should get O

	if p2.kind != PlayerO {
		t.Fatalf("expected O, got %s", string(p2.kind))
	}
}
