package room

import (
	"encoding/json"
	"errors"
	"strings"
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

func (r *Room) AddPlayer(player *Player, preferredType string) error {
	if r.players[0] == nil {
		kind, err := resolveFirstPlayerType(preferredType)
		if err != nil {
			return err
		}
		player.SetKind(kind)
		r.players[0] = player
		player.start(r)
		return nil
	}

	if r.players[1] == nil {
		player.SetKind(oppositeType(r.players[0].kind))
		r.players[1] = player
		player.start(r)

		r.notifyPlayerJoined(r.players[0], 2)
		r.notifyPlayerJoined(r.players[1], 1)
		return nil
	}

	return errors.New("The room is full. Get out!")
}

func resolveFirstPlayerType(preferred string) (PlayerType, error) {
	switch strings.ToLower(preferred) {
	case "x":
		return PlayerX, nil
	case "o":
		return PlayerO, nil
	case "":
		return PlayerX, nil
	default:
		return "", errors.New("invalid player type")
	}
}

func oppositeType(kind PlayerType) PlayerType {
	if kind == PlayerX {
		return PlayerO
	}
	return PlayerX
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

func (r *Room) Run(deleteItself chan string, roomName string) {
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
				deleteItself <- roomName
				return
			}
			// Right now we don't let the other player wait for a new opponent, we just kick them out.
			// We can change this later if we want to.
			msg, _ := json.Marshal(PlayerDisconnectedMsg{Type: MsgPlayerDisconnected})
			other.send <- msg
			close(other.send)
			deleteItself <- roomName
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
