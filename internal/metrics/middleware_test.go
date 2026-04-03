package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestPrometheusMiddleware_RecordsRequestCount(t *testing.T) {
	m := NewMetrics(true)
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(PrometheusMiddleware(m))
	r.Get("/hello/{name}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/hello/world", nil)
	r.ServeHTTP(httptest.NewRecorder(), req)

	count := testutil.ToFloat64(m.httpRequestsTotal.WithLabelValues("GET", "/hello/{name}", "200"))
	if count != 1 {
		t.Errorf("expected 1 request, got %v", count)
	}
}

func TestPrometheusMiddleware_RecordsDifferentStatusCodes(t *testing.T) {
	m := NewMetrics(true)
	r := chi.NewRouter()
	r.Use(PrometheusMiddleware(m))
	r.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	req := httptest.NewRequest("GET", "/hello", nil)
	r.ServeHTTP(httptest.NewRecorder(), req)

	count := testutil.ToFloat64(m.httpRequestsTotal.WithLabelValues("GET", "/hello", "404"))
	if count != 1 {
		t.Errorf("expected 1 request with status 404, got %v", count)
	}
}

func TestPrometheusMiddleware_RecordsLatency(t *testing.T) {
	m := NewMetrics(true)
	r := chi.NewRouter()
	r.Use(PrometheusMiddleware(m))
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	r.ServeHTTP(httptest.NewRecorder(), req)

	// Verify histogram recorded an observation by checking via Gather
	families, err := m.registry.Gather()
	if err != nil {
		t.Fatalf("failed to gather: %v", err)
	}
	found := false
	for _, fam := range families {
		if fam.GetName() == "http_request_duration_seconds" {
			for _, metric := range fam.GetMetric() {
				if metric.GetHistogram().GetSampleCount() >= 1 {
					found = true
				}
			}
		}
	}
	if !found {
		t.Error("expected at least 1 histogram observation")
	}
}

func TestPrometheusMiddleware_UsesRoutePattern(t *testing.T) {
	m := NewMetrics(true)
	r := chi.NewRouter()
	r.Use(PrometheusMiddleware(m))
	r.Get("/room/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Hit /room/abc123 — should record path as "/room/{id}", not "/room/abc123"
	req := httptest.NewRequest("GET", "/room/abc123", nil)
	r.ServeHTTP(httptest.NewRecorder(), req)

	count := testutil.ToFloat64(m.httpRequestsTotal.WithLabelValues("GET", "/room/{id}", "200"))
	if count != 1 {
		t.Errorf("expected path label to be /room/{id}, got count=%v", count)
	}
}
