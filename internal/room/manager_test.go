package room

import (
	"testing"
	"time"

	"github.com/MarcusXavierr/xo-battle-back/internal/metrics"
)

func TestRoomDeleter_HandlesMultipleDeletions(t *testing.T) {
	m := metrics.NewMetrics(true)
	rm := NewRoomManager(m)
	go rm.RoomDeleter()

	rm.CreateRoom("room1")
	rm.CreateRoom("room2")

	rm.delete <- "room1"
	rm.delete <- "room2"

	waitForRoomCount(t, rm, 0)
}

func waitForRoomCount(t *testing.T, rm *RoomManager, want int) {
	t.Helper()
	for i := 0; i < 100; i++ {
		rm.mu.Lock()
		got := len(rm.rooms)
		rm.mu.Unlock()
		if got == want {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	rm.mu.Lock()
	got := len(rm.rooms)
	rm.mu.Unlock()
	t.Fatalf("expected %d rooms, got %d", want, got)
}

func TestCreateRoom(t *testing.T) {
	m := metrics.NewMetrics(true)
	rm := NewRoomManager(m)

	err := rm.CreateRoom("test-room")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteRoom(t *testing.T) {
	m := metrics.NewMetrics(true)
	rm := NewRoomManager(m)
	go rm.RoomDeleter()

	rm.CreateRoom("doomed-room")

	rm.delete <- "doomed-room"

	waitForRoomCount(t, rm, 0)
}
