package proxy

import (
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/example/rcig/internal/metrics"
)

type Handler struct {
	proxy    *httputil.ReverseProxy
	registry *metrics.Registry
}

type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func NewHandler(proxy *httputil.ReverseProxy, registry *metrics.Registry) http.Handler {
	return &Handler{
		proxy:    proxy,
		registry: registry,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
	h.proxy.ServeHTTP(rec, r)

	latency := time.Since(start)
	h.registry.Record(r.Method, r.URL.Path, latency, rec.status)
}


