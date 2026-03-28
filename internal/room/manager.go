package room

import (
	"errors"
	"log"
	"sync"
)

type WebsocketError error
type RoomManager struct {
	mu     *sync.Mutex
	rooms  map[string]*Room
	delete chan string
}

var (
	RoomNotFoundError WebsocketError = errors.New("room_not_found")
	RoomFullError     WebsocketError = errors.New("room_full")
)

func NewRoomManager() *RoomManager {
	return &RoomManager{
		mu:     &sync.Mutex{},
		rooms:  make(map[string]*Room),
		delete: make(chan string),
	}
}

func (rm *RoomManager) CreateRoom(name string) error {
	room := &Room{
		players: [2]*Player{},
		events:  make(chan Event),
	}

	go room.Run(rm.delete, name)
	go rm.RoomDeleter()

	rm.mu.Lock()
	if _, ok := rm.rooms[name]; ok {
		rm.mu.Unlock()
		return errors.New("The room already exists")
	}
	rm.rooms[name] = room
	rm.mu.Unlock()
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
	roomName := <-rm.delete
	log.Println("deleting room: ", roomName)
	rm.mu.Lock()
	delete(rm.rooms, roomName)
	rm.mu.Unlock()
}
