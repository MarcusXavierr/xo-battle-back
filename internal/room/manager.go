package room

import (
	"errors"
	"log"
	"sync"

	"github.com/MarcusXavierr/xo-battle-back/internal/metrics"
)

type WebsocketError error
type RoomManager struct {
	mu      *sync.Mutex
	rooms   map[string]*Room
	delete  chan string
	metrics *metrics.Metrics
}

var (
	RoomNotFoundError WebsocketError = errors.New("room_not_found")
	RoomFullError     WebsocketError = errors.New("room_full")
)

func NewRoomManager(m *metrics.Metrics) *RoomManager {
	if m == nil {
		panic("metrics is nil")
	}
	return &RoomManager{
		mu:      &sync.Mutex{},
		rooms:   make(map[string]*Room),
		delete:  make(chan string),
		metrics: m,
	}
}

func (rm *RoomManager) CreateRoom(name string) error {
	room := &Room{
		players: [2]*Player{},
		events:  make(chan Event),
	}

	go room.Run(rm.delete, name)

	rm.mu.Lock()
	if _, ok := rm.rooms[name]; ok {
		rm.mu.Unlock()
		return errors.New("The room already exists")
	}
	rm.rooms[name] = room
	rm.mu.Unlock()
	rm.metrics.IncGameRoomsCreated()
	rm.metrics.IncGameRoomsActive()
	return nil
}

func (rm *RoomManager) JoinRoom(roomName string, player *Player, preferredType string) error {
	rm.mu.Lock()
	room, ok := rm.rooms[roomName]
	rm.mu.Unlock()

	if !ok {
		return RoomNotFoundError
	}

	return room.AddPlayer(player, preferredType)
}

func (rm *RoomManager) RoomDeleter() {
	for roomName := range rm.delete {
		log.Println("deleting room: ", roomName)
		rm.mu.Lock()
		delete(rm.rooms, roomName)
		rm.mu.Unlock()
		rm.metrics.DecGameRoomsActive()
	}
}
