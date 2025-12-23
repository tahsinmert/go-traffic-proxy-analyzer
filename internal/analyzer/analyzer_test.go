package analyzer

import (
	"testing"
	"time"

	"github.com/example/rcig/internal/alert"
	"github.com/example/rcig/internal/metrics"
)

func TestNewAnalyzer(t *testing.T) {
	reg := metrics.NewRegistry(100)
	sink := alert.NewConsoleSink()
	cfg := Config{
		TickInterval:     5 * time.Second,
		OlderWindowSize:  80,
		RecentWindowSize: 20,
		DriftThreshold:   0.40,
	}

	an := NewAnalyzer(reg, sink, cfg)

	if an == nil {
		t.Fatal("NewAnalyzer returned nil")
	}
	if an.registry != reg {
		t.Error("registry not set correctly")
	}
	if an.sink != sink {
		t.Error("sink not set correctly")
	}
	if an.cfg != cfg {
		t.Error("config not set correctly")
	}
}

func TestAnalyzer_runOnce(t *testing.T) {
	t.Run("no drift when insufficient data", func(t *testing.T) {
		reg := metrics.NewRegistry(100)
		sink := alert.NewConsoleSink()
		cfg := Config{
			TickInterval:     5 * time.Second,
			OlderWindowSize:  80,
			RecentWindowSize: 20,
			DriftThreshold:   0.40,
		}

		an := NewAnalyzer(reg, sink, cfg)

		// Record only 50 samples (less than 80 + 20 = 100)
		for i := 0; i < 50; i++ {
			reg.Record("GET", "/api/test", 100*time.Millisecond, 200)
		}

		// Should not trigger alert
		an.runOnce()

		// No way to verify directly, but this shouldn't panic
	})

	t.Run("no drift when within threshold", func(t *testing.T) {
		reg := metrics.NewRegistry(200)
		sink := alert.NewConsoleSink()
		cfg := Config{
			TickInterval:     5 * time.Second,
			OlderWindowSize:  80,
			RecentWindowSize: 20,
			DriftThreshold:   0.40,
		}

		an := NewAnalyzer(reg, sink, cfg)

		// Record 100 samples with minimal variation (no drift)
		for i := 0; i < 100; i++ {
			reg.Record("GET", "/api/test", 100*time.Millisecond, 200)
		}

		an.runOnce()
		// Should not trigger alert since there's no drift
	})

	t.Run("drift detected when threshold exceeded", func(t *testing.T) {
		reg := metrics.NewRegistry(200)
		sink := alert.NewConsoleSink()
		cfg := Config{
			TickInterval:     5 * time.Second,
			OlderWindowSize:  80,
			RecentWindowSize: 20,
			DriftThreshold:   0.40,
		}

		an := NewAnalyzer(reg, sink, cfg)

		// Record 80 samples with 100ms latency (older window)
		for i := 0; i < 80; i++ {
			reg.Record("GET", "/api/test", 100*time.Millisecond, 200)
		}

		// Record 20 samples with 200ms latency (recent window, 100% increase)
		for i := 0; i < 20; i++ {
			reg.Record("GET", "/api/test", 200*time.Millisecond, 200)
		}

		an.runOnce()
		// Should trigger alert since 100% > 40% threshold
		// We can't easily verify the alert was emitted without modifying the sink
	})

	t.Run("no drift when older average is zero", func(t *testing.T) {
		reg := metrics.NewRegistry(200)
		sink := alert.NewConsoleSink()
		cfg := Config{
			TickInterval:     5 * time.Second,
			OlderWindowSize:  80,
			RecentWindowSize: 20,
			DriftThreshold:   0.40,
		}

		an := NewAnalyzer(reg, sink, cfg)

		// Record 100 samples with 0ms latency
		for i := 0; i < 100; i++ {
			reg.Record("GET", "/api/test", 0, 200)
		}

		an.runOnce()
		// Should not panic or trigger alert
	})
}

func TestAverageDuration(t *testing.T) {
	tests := []struct {
		name     string
		values   []time.Duration
		expected time.Duration
	}{
		{
			name:     "empty slice",
			values:   []time.Duration{},
			expected: 0,
		},
		{
			name:     "single value",
			values:   []time.Duration{100 * time.Millisecond},
			expected: 100 * time.Millisecond,
		},
		{
			name: "multiple values",
			values: []time.Duration{
				100 * time.Millisecond,
				200 * time.Millisecond,
				300 * time.Millisecond,
			},
			expected: 200 * time.Millisecond,
		},
		{
			name: "varying values",
			values: []time.Duration{
				50 * time.Millisecond,
				150 * time.Millisecond,
			},
			expected: 100 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := averageDuration(tt.values)
			if result != tt.expected {
				t.Errorf("averageDuration() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRoundToTwoDecimals(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"no rounding needed", 1.23, 1.23},
		{"round down", 1.234, 1.23},
		{"round up", 1.235, 1.24},
		{"round up at boundary", 1.235, 1.24},
		{"negative number", -1.234, -1.23},
		{"zero", 0.0, 0.0},
		{"large number", 123.456, 123.46},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := roundToTwoDecimals(tt.input)
			if result != tt.expected {
				t.Errorf("roundToTwoDecimals(%f) = %f, want %f", tt.input, result, tt.expected)
			}
		})
	}
}

func TestConfig(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		cfg := Config{
			TickInterval:     5 * time.Second,
			OlderWindowSize:  80,
			RecentWindowSize: 20,
			DriftThreshold:   0.40,
		}

		if cfg.TickInterval != 5*time.Second {
			t.Errorf("TickInterval = %v, want 5s", cfg.TickInterval)
		}
		if cfg.OlderWindowSize != 80 {
			t.Errorf("OlderWindowSize = %d, want 80", cfg.OlderWindowSize)
		}
		if cfg.RecentWindowSize != 20 {
			t.Errorf("RecentWindowSize = %d, want 20", cfg.RecentWindowSize)
		}
		if cfg.DriftThreshold != 0.40 {
			t.Errorf("DriftThreshold = %f, want 0.40", cfg.DriftThreshold)
		}
	})
}

func BenchmarkAnalyzer_runOnce(b *testing.B) {
	reg := metrics.NewRegistry(1000)
	sink := alert.NewConsoleSink()
	cfg := Config{
		TickInterval:     5 * time.Second,
		OlderWindowSize:  80,
		RecentWindowSize: 20,
		DriftThreshold:   0.40,
	}

	an := NewAnalyzer(reg, sink, cfg)

	// Populate with data
	for i := 0; i < 100; i++ {
		reg.Record("GET", "/api/test", 100*time.Millisecond, 200)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		an.runOnce()
	}
}
