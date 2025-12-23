package metrics

import (
	"encoding/json"
	"net/http"
	"time"
)

type httpHandler struct {
	registry *Registry
}

type endpointMetrics struct {
	Method          string  `json:"method"`
	Path            string  `json:"path"`
	Count           int64   `json:"count"`
	LastStatus      int     `json:"last_status"`
	AverageMs       float64 `json:"average_ms"`
	SampleSize      int     `json:"sample_size"`
	LastUpdatedUnix int64   `json:"last_updated_unix"`
}

type payload struct {
	GeneratedAt time.Time        `json:"generated_at"`
	Endpoints   []endpointMetrics `json:"endpoints"`
}

func NewHTTPHandler(registry *Registry) http.Handler {
	return &httpHandler{registry: registry}
}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	snap := h.registry.Snapshot()
	summaries := h.registry.Summaries()

	now := time.Now().UTC()
	out := make([]endpointMetrics, 0, len(summaries))

	for _, s := range summaries {
		values := snap[EndpointKey{Method: s.Method, Path: s.Path}]
		var sum int64
		for _, v := range values {
			sum += int64(v.Milliseconds())
		}
		var avg float64
		if len(values) > 0 {
			avg = float64(sum) / float64(len(values))
		}

		out = append(out, endpointMetrics{
			Method:          s.Method,
			Path:            s.Path,
			Count:           s.Count,
			LastStatus:      s.LastStatus,
			AverageMs:       avg,
			SampleSize:      len(values),
			LastUpdatedUnix: now.Unix(),
		})
	}

	resp := payload{
		GeneratedAt: now,
		Endpoints:   out,
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	_ = enc.Encode(resp)
}



