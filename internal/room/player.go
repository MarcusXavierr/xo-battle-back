package room

import (
	"encoding/json"
	"log"

	"github.com/MarcusXavierr/xo-battle-back/internal/metrics"
	"github.com/gorilla/websocket"
)

type MessageType string
type PlayerType string

const (
	MsgPlayerJoined       MessageType = "player_joined"
	MsgPlayerDisconnected MessageType = "player_disconnected"
	MsgMove               MessageType = "move"
)

const (
	PlayerX PlayerType = "x"
	PlayerO PlayerType = "o"
)

type Player struct {
	conn    *websocket.Conn
	name    string
	kind    PlayerType
	send    chan []byte
	metrics *metrics.Metrics
}

// --- Messages
type PlayerJoinedMsg struct {
	Type       MessageType `json:"type"`
	Name       string      `json:"name"`
	PlayerType PlayerType  `json:"player_type"`
	Order      int         `json:"order"`
}

type PlayerDisconnectedMsg struct {
	Type MessageType `json:"type"`
}

type MoveMsg struct {
	Type MessageType `json:"type"`
	Cell int         `json:"cell"`
}

// --- End messages

func extractMessageType(msg []byte) string {
	var parsed struct {
		Type MessageType `json:"type"`
	}
	_ = json.Unmarshal(msg, &parsed)
	return string(parsed.Type)
}

func (p *Player) writeLoop() {
	for msg := range p.send {
		p.conn.WriteMessage(websocket.TextMessage, msg)
		p.metrics.IncWSMessagesSent(extractMessageType(msg))
	}
}

func (p *Player) readLoop(room *Room) {
	defer func() {
		p.metrics.DecWSActiveConnections()
	}()

	for {
		_, msg, err := p.conn.ReadMessage()
		if err != nil {
			log.Printf("Read error: %v\n", err)
			room.events <- Event{kind: "disconnect", player: p}
			return
		}
		p.metrics.IncWSMessagesReceived(extractMessageType(msg))
		room.events <- Event{kind: "message", player: p, data: msg}
	}
}

func NewPlayer(conn *websocket.Conn, name string, m *metrics.Metrics) *Player {
	return &Player{
		conn:    conn,
		name:    name,
		send:    make(chan []byte),
		metrics: m,
	}
}

func (p *Player) SetKind(kind PlayerType) {
	p.kind = kind
}

func (p *Player) start(room *Room) {
	if p.conn == nil {
		return
	}
	p.metrics.IncWSActiveConnections()
	go p.readLoop(room)
	go p.writeLoop()
}
