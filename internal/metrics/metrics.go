package metrics

import (
	"sync"
	"time"
)

type EndpointKey struct {
	Method string
	Path   string
}

type CircularBuffer struct {
	values []time.Duration
	size   int
	count  int
	head   int
}

func NewCircularBuffer(size int) *CircularBuffer {
	return &CircularBuffer{
		values: make([]time.Duration, size),
		size:   size,
	}
}

func (b *CircularBuffer) Append(v time.Duration) {
	b.values[b.head] = v
	b.head = (b.head + 1) % b.size
	if b.count < b.size {
		b.count++
	}
}

func (b *CircularBuffer) SnapshotValues() []time.Duration {
	if b.count == 0 {
		return nil
	}
	out := make([]time.Duration, b.count)
	start := (b.head - b.count + b.size) % b.size
	for i := 0; i < b.count; i++ {
		idx := (start + i) % b.size
		out[i] = b.values[idx]
	}
	return out
}

type EndpointMetrics struct {
	key        EndpointKey
	buffer     *CircularBuffer
	count      int64
	lastStatus int
}

type Registry struct {
	mu      sync.RWMutex
	size    int
	metrics map[EndpointKey]*EndpointMetrics
}

func NewRegistry(size int) *Registry {
	return &Registry{
		size:    size,
		metrics: make(map[EndpointKey]*EndpointMetrics),
	}
}

func (r *Registry) Record(method, path string, d time.Duration, status int) {
	key := EndpointKey{Method: method, Path: path}

	r.mu.Lock()
	defer r.mu.Unlock()

	m, ok := r.metrics[key]
	if !ok {
		m = &EndpointMetrics{
			key:    key,
			buffer: NewCircularBuffer(r.size),
		}
		r.metrics[key] = m
	}
	m.buffer.Append(d)
	m.count++
	m.lastStatus = status
}

func (r *Registry) Snapshot() map[EndpointKey][]time.Duration {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make(map[EndpointKey][]time.Duration, len(r.metrics))
	for k, m := range r.metrics {
		out[k] = m.buffer.SnapshotValues()
	}
	return out
}

type EndpointSummary struct {
	Method     string
	Path       string
	Count      int64
	LastStatus int
}

func (r *Registry) Summaries() []EndpointSummary {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]EndpointSummary, 0, len(r.metrics))
	for _, m := range r.metrics {
		out = append(out, EndpointSummary{
			Method:     m.key.Method,
			Path:       m.key.Path,
			Count:      m.count,
			LastStatus: m.lastStatus,
		})
	}
	return out
}


