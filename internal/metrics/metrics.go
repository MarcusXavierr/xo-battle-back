package metrics

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	enabled bool

	// HTTP
	httpRequestDuration *prometheus.HistogramVec
	httpRequestsTotal   *prometheus.CounterVec

	// WebSocket
	wsMessagesSent      *prometheus.CounterVec
	wsMessagesReceived  *prometheus.CounterVec
	wsActiveConnections prometheus.Gauge
	wsConnectionsTotal  *prometheus.CounterVec

	// Game
	gameRoomsActive       prometheus.Gauge
	gameRoomsCreatedTotal prometheus.Counter

	registry *prometheus.Registry
	server   *http.Server
}

func NewMetrics(enabled bool) *Metrics {
	reg := prometheus.NewRegistry()
	reg.MustRegister(collectors.NewGoCollector())
	reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	m := &Metrics{
		enabled:  enabled,
		registry: reg,
		httpRequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: prometheus.DefBuckets,
		}, []string{"method", "path"}),
		httpRequestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests processed",
		}, []string{"method", "path", "status"}),
		wsMessagesSent: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "ws_messages_sent_total",
			Help: "Total WebSocket messages sent to clients",
		}, []string{"type"}),
		wsMessagesReceived: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "ws_messages_received_total",
			Help: "Total WebSocket messages received from clients",
		}, []string{"type"}),
		wsActiveConnections: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "ws_active_connections",
			Help: "Current number of open WebSocket connections",
		}),
		wsConnectionsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "ws_connections_total",
			Help: "Total WebSocket connection attempts",
		}, []string{"status"}),
		gameRoomsActive: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "game_rooms_active",
			Help: "Current number of active game rooms",
		}),
		gameRoomsCreatedTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "game_rooms_created_total",
			Help: "Total number of game rooms created",
		}),
	}

	reg.MustRegister(
		m.httpRequestDuration,
		m.httpRequestsTotal,
		m.wsMessagesSent,
		m.wsMessagesReceived,
		m.wsActiveConnections,
		m.wsConnectionsTotal,
		m.gameRoomsActive,
		m.gameRoomsCreatedTotal,
	)

	return m
}

func (m *Metrics) Registry() *prometheus.Registry {
	if m == nil {
		return nil
	}
	return m.registry
}

func (m *Metrics) IncGameRoomsCreated() {
	if m == nil || !m.enabled {
		return
	}
	m.gameRoomsCreatedTotal.Inc()
}

func (m *Metrics) IncGameRoomsActive() {
	if m == nil || !m.enabled {
		return
	}
	m.gameRoomsActive.Inc()
}

func (m *Metrics) DecGameRoomsActive() {
	if m == nil || !m.enabled {
		return
	}
	m.gameRoomsActive.Dec()
}

func (m *Metrics) IncWSMessagesSent(msgType string) {
	if m == nil || !m.enabled {
		return
	}
	m.wsMessagesSent.WithLabelValues(msgType).Inc()
}

func (m *Metrics) IncWSMessagesReceived(msgType string) {
	if m == nil || !m.enabled {
		return
	}
	m.wsMessagesReceived.WithLabelValues(msgType).Inc()
}

func (m *Metrics) IncWSActiveConnections() {
	if m == nil || !m.enabled {
		return
	}
	m.wsActiveConnections.Inc()
}

func (m *Metrics) DecWSActiveConnections() {
	if m == nil || !m.enabled {
		return
	}
	m.wsActiveConnections.Dec()
}

func (m *Metrics) ObserveHTTPRequestDuration(method, path string, duration float64) {
	if m == nil || !m.enabled {
		return
	}
	m.httpRequestDuration.WithLabelValues(method, path).Observe(duration)
}

func (m *Metrics) IncHTTPRequestsTotal(method, path, status string) {
	if m == nil || !m.enabled {
		return
	}
	m.httpRequestsTotal.WithLabelValues(method, path, status).Inc()
}

func (m *Metrics) IncWSConnectionsTotal(status string) {
	if m == nil || !m.enabled {
		return
	}
	m.wsConnectionsTotal.WithLabelValues(status).Inc()
}

func (m *Metrics) Handler() http.Handler {
	if m == nil {
		return http.NotFoundHandler()
	}
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

func (m *Metrics) StartServer(port string) {
	if m == nil || !m.enabled {
		return
	}

	if port == "" {
		port = "9091"
	}
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", m.Handler())
	m.server = &http.Server{Addr: ":" + port, Handler: metricsMux}

	go func() {
		log.Printf("Starting metrics server on port %s", port)
		if err := m.server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Metrics server error: %v", err)
		}
	}()
}

func (m *Metrics) StopServer(ctx context.Context) error {
	if m == nil || m.server == nil {
		return nil
	}
	return m.server.Shutdown(ctx)
}
