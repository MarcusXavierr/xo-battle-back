package room

import (
	"errors"
	"sync"
)

type RoomManager struct {
	mu    *sync.Mutex
	rooms map[string]*Room
}

func NewRoomManager() *RoomManager {
	return &RoomManager{
		mu:    &sync.Mutex{},
		rooms: make(map[string]*Room),
	}
}

func (rm *RoomManager) CreateRoom(name string) error {
	room := &Room{
		players: [2]*Player{},
		events:  make(chan Event),
	}

	go room.Run()

	rm.mu.Lock()
	if _, ok := rm.rooms[name]; ok {
		rm.mu.Unlock()
		return errors.New("The room already exists")
	}
	rm.rooms[name] = room
	rm.mu.Unlock()
	return nil
}

func (rm *RoomManager) JoinRoom(roomName string, player *Player) error {
	rm.mu.Lock()
	room, ok := rm.rooms[roomName]
	rm.mu.Unlock()

	if !ok {
		return errors.New("room not found")
	}

	return room.AddPlayer(player)
}
