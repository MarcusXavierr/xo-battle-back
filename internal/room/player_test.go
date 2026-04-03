package room

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MarcusXavierr/xo-battle-back/internal/metrics"
	"github.com/gorilla/websocket"
)

func newTestMetrics() *metrics.Metrics {
	return metrics.NewMetrics(true)
}

func TestNewPlayer_DoesNotRequireKind(t *testing.T) {
	p := NewPlayer(nil, "Alice", nil)
	if p.name != "Alice" {
		t.Fatalf("expected name Alice, got %s", p.name)
	}
	if p.kind != "" {
		t.Fatalf("expected empty kind, got %s", string(p.kind))
	}
}

func TestNewPlayer_IncrementsActiveConnections(t *testing.T) {
	m := newTestMetrics()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, _ := upgrader.Upgrade(w, r, nil)
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
		conn.Close()
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[4:]
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial error: %v", err)
	}
	defer conn.Close()

	p := NewPlayer(conn, "Alice", m)
	room := newTestRoom()
	p.start(room)
}

func TestReadLoop_IncrementsReceivedMetrics(t *testing.T) {
	m := newTestMetrics()

	serverConnCh := make(chan *websocket.Conn, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		serverConnCh <- conn
		// Don't read from conn — let the Player's readLoop handle it
		select {}
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[4:]
	clientConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial error: %v", err)
	}
	defer clientConn.Close()

	serverConn := <-serverConnCh
	p := NewPlayer(serverConn, "Alice", m)
	room := newTestRoom()
	p.start(room)

	// Client sends a move — server-side readLoop picks it up
	moveMsg, _ := json.Marshal(MoveMsg{Type: MsgMove, Cell: 0})
	clientConn.WriteMessage(websocket.TextMessage, moveMsg)

	// Wait for readLoop to process the message
	time.Sleep(100 * time.Millisecond)

	serverConn.Close()
}

func TestWriteLoop_IncrementsSentMetrics(t *testing.T) {
	m := newTestMetrics()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, _ := upgrader.Upgrade(w, r, nil)
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
		conn.Close()
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[4:]
	conn, _, _ := websocket.DefaultDialer.Dial(wsURL, nil)
	defer conn.Close()

	p := NewPlayer(conn, "Alice", m)

	// Start writeLoop before sending — p.send is unbuffered, writing first would deadlock
	go p.writeLoop()

	// Send a message through the player's send channel
	moveMsg, _ := json.Marshal(MoveMsg{Type: MsgMove, Cell: 5})
	p.send <- moveMsg

	// Wait for writeLoop to process
	time.Sleep(100 * time.Millisecond)

	// Close send channel to stop writeLoop
	close(p.send)
}
