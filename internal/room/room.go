package room

import (
	"encoding/json"
	"errors"
)

type EventKind string

const (
	EventMessage    EventKind = "message"
	EventDisconnect EventKind = "disconnect"
)

type Event struct {
	kind   EventKind
	player *Player
	data   []byte
}

type Room struct {
	players [2]*Player
	events  chan Event
}

func (r *Room) AddPlayer(player *Player) error {
	if r.players[0] == nil {
		r.players[0] = player
		go player.readLoop(r)
		go player.writeLoop()
		return nil
	}

	if r.players[1] == nil {
		r.players[1] = player
		go player.readLoop(r)
		go player.writeLoop()

		r.notifyPlayerJoined(r.players[0], 2)
		r.notifyPlayerJoined(r.players[1], 1)
		return nil
	}

	return errors.New("The room is full. Get out!")
}

func (r *Room) notifyPlayerJoined(player *Player, oponentOrder int) {
	oponent := r.opponent(player)
	msg, _ := json.Marshal(PlayerJoinedMsg{
		Type:       MsgPlayerJoined,
		Name:       oponent.name,
		PlayerType: oponent.kind,
		Order:      oponentOrder,
	})

	player.send <- msg
}

func (r *Room) Run() {
	for event := range r.events {
		switch event.kind {
		case EventMessage:
			other := r.opponent(event.player)
			if other == nil {
				continue
			}
			other.send <- event.data

		case EventDisconnect:
			event.player.send <- []byte("GoodBye")
			other := r.opponent(event.player)
			if other == nil {
				return
			}
			msg, _ := json.Marshal(PlayerDisconnectedMsg{Type: MsgPlayerDisconnected})
			other.send <- msg
			close(other.send)
			return
		}
	}
}

func (r *Room) opponent(player *Player) *Player {
	if r.players[0] == player {
		return r.players[1]
	}
	return r.players[0]
}
