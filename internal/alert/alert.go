package alert

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type DriftAlert struct {
	Type              string  `json:"type"`
	Method            string  `json:"method"`
	Path              string  `json:"path"`
	PercentageIncrease float64 `json:"percentage_increase"`
	OlderAvgMillis    float64 `json:"older_avg_ms"`
	RecentAvgMillis   float64 `json:"recent_avg_ms"`
	Timestamp         string  `json:"timestamp"`
}

type Sink interface {
	EmitDriftAlert(a DriftAlert)
}

type ConsoleSink struct {
	mu         sync.Mutex
	activeDrift map[string]bool
}

func NewConsoleSink() *ConsoleSink {
	return &ConsoleSink{
		activeDrift: make(map[string]bool),
	}
}

func makeKey(method, path string) string {
	return method + " " + path
}

func (c *ConsoleSink) EmitDriftAlert(a DriftAlert) {
	key := makeKey(a.Method, a.Path)

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.activeDrift[key] {
		return
	}
	c.activeDrift[key] = true

	if a.Type == "" {
		a.Type = "latency_drift"
	}
	if a.Timestamp == "" {
		a.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	}

	data, err := json.Marshal(a)
	if err != nil {
		fmt.Printf(`{"type":"latency_drift","error":"marshal_failed","detail":%q}`+"\n", err.Error())
		return
	}
	fmt.Println(string(data))
}

func (c *ConsoleSink) ClearDrift(method, path string) {
	key := makeKey(method, path)

	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.activeDrift, key)
}


