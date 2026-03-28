package room

import "testing"

func TestNewPlayer_DoesNotRequireKind(t *testing.T) {
	p := NewPlayer(nil, "Alice")
	if p.name != "Alice" {
		t.Fatalf("expected name Alice, got %s", p.name)
	}
	if p.kind != "" {
		t.Fatalf("expected empty kind, got %s", string(p.kind))
	}
}
