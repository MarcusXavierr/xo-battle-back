package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestNewMetrics_RegistersAllMetrics(t *testing.T) {
	m := NewMetrics(true)

	// Verify all struct fields are initialized
	if m.httpRequestDuration == nil {
		t.Error("HTTPRequestDuration is nil")
	}
	if m.httpRequestsTotal == nil {
		t.Error("HTTPRequestsTotal is nil")
	}
	if m.wsMessagesSent == nil {
		t.Error("WSMessagesSent is nil")
	}
	if m.wsMessagesReceived == nil {
		t.Error("WSMessagesReceived is nil")
	}
	if m.wsActiveConnections == nil {
		t.Error("WSActiveConnections is nil")
	}
	if m.wsConnectionsTotal == nil {
		t.Error("WSConnectionsTotal is nil")
	}
	if m.gameRoomsActive == nil {
		t.Error("GameRoomsActive is nil")
	}
	if m.gameRoomsCreatedTotal == nil {
		t.Error("GameRoomsCreatedTotal is nil")
	}
	if m.registry == nil {
		t.Error("Registry is nil")
	}

	// Verify registry can gather (proves all collectors are properly registered)
	_, err := m.registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}
}

func TestMetrics_CanIncrementCounters(t *testing.T) {
	m := NewMetrics(true)

	m.wsMessagesSent.WithLabelValues("move").Inc()
	m.wsMessagesReceived.WithLabelValues("move").Inc()
	m.gameRoomsCreatedTotal.Inc()

	if count := testutil.ToFloat64(m.wsMessagesSent.WithLabelValues("move")); count != 1 {
		t.Errorf("expected ws_messages_sent_total{type=move} = 1, got %v", count)
	}
	if count := testutil.ToFloat64(m.wsMessagesReceived.WithLabelValues("move")); count != 1 {
		t.Errorf("expected ws_messages_received_total{type=move} = 1, got %v", count)
	}
	if count := testutil.ToFloat64(m.gameRoomsCreatedTotal); count != 1 {
		t.Errorf("expected game_rooms_created_total = 1, got %v", count)
	}
}

func TestMetrics_CanSetGauges(t *testing.T) {
	m := NewMetrics(true)

	m.wsActiveConnections.Inc()
	m.wsActiveConnections.Inc()
	m.gameRoomsActive.Inc()

	if val := testutil.ToFloat64(m.wsActiveConnections); val != 2 {
		t.Errorf("expected ws_active_connections = 2, got %v", val)
	}
	if val := testutil.ToFloat64(m.gameRoomsActive); val != 1 {
		t.Errorf("expected game_rooms_active = 1, got %v", val)
	}

	m.wsActiveConnections.Dec()
	m.gameRoomsActive.Dec()

	if val := testutil.ToFloat64(m.wsActiveConnections); val != 1 {
		t.Errorf("expected ws_active_connections = 1 after Dec, got %v", val)
	}
	if val := testutil.ToFloat64(m.gameRoomsActive); val != 0 {
		t.Errorf("expected game_rooms_active = 0 after Dec, got %v", val)
	}
}
