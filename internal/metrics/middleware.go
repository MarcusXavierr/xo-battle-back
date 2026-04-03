package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func PrometheusMiddleware(m *Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			routePattern := chi.RouteContext(r.Context()).RoutePattern()
			elapsed := time.Since(start).Seconds()
			status := strconv.Itoa(ww.Status())

			m.ObserveHTTPRequestDuration(r.Method, routePattern, elapsed)
			m.IncHTTPRequestsTotal(r.Method, routePattern, status)
		})
	}
}
