package analyzer

import (
	"math"
	"time"

	"github.com/example/rcig/internal/alert"
	"github.com/example/rcig/internal/metrics"
)

type Config struct {
	TickInterval     time.Duration
	OlderWindowSize  int
	RecentWindowSize int
	DriftThreshold   float64
}

type Analyzer struct {
	registry *metrics.Registry
	sink     *alert.ConsoleSink
	cfg      Config
}

func NewAnalyzer(registry *metrics.Registry, sink *alert.ConsoleSink, cfg Config) *Analyzer {
	return &Analyzer{
		registry: registry,
		sink:     sink,
		cfg:      cfg,
	}
}

func (a *Analyzer) Run() {
	ticker := time.NewTicker(a.cfg.TickInterval)
	defer ticker.Stop()

	for range ticker.C {
		a.runOnce()
	}
}

func (a *Analyzer) runOnce() {
	snap := a.registry.Snapshot()
	now := time.Now().UTC()

	for key, values := range snap {
		if len(values) < a.cfg.OlderWindowSize+a.cfg.RecentWindowSize {
			a.sink.ClearDrift(key.Method, key.Path)
			continue
		}

		olderVals := values[:a.cfg.OlderWindowSize]
		recentVals := values[len(values)-a.cfg.RecentWindowSize:]

		olderAvg := averageDuration(olderVals)
		recentAvg := averageDuration(recentVals)

		if olderAvg <= 0 {
			a.sink.ClearDrift(key.Method, key.Path)
			continue
		}

		increase := (float64(recentAvg) - float64(olderAvg)) / float64(olderAvg)
		if increase >= a.cfg.DriftThreshold {
			a.sink.EmitDriftAlert(alert.DriftAlert{
				Method:             key.Method,
				Path:               key.Path,
				PercentageIncrease: roundToTwoDecimals(increase * 100),
				OlderAvgMillis:     float64(olderAvg.Milliseconds()),
				RecentAvgMillis:    float64(recentAvg.Milliseconds()),
				Timestamp:          now.Format(time.RFC3339Nano),
			})
		} else {
			a.sink.ClearDrift(key.Method, key.Path)
		}
	}
}

func averageDuration(vs []time.Duration) time.Duration {
	if len(vs) == 0 {
		return 0
	}
	var sum int64
	for _, v := range vs {
		sum += int64(v)
	}
	return time.Duration(sum / int64(len(vs)))
}

func roundToTwoDecimals(v float64) float64 {
	return math.Round(v*100) / 100
}
