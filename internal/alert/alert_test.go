package alert

import (
	"testing"
	"time"
)

func TestNewConsoleSink(t *testing.T) {
	sink := NewConsoleSink()

	if sink == nil {
		t.Fatal("NewConsoleSink returned nil")
	}
	if sink.activeDrift == nil {
		t.Error("activeDrift map is nil")
	}
	if len(sink.activeDrift) != 0 {
		t.Errorf("activeDrift length = %d, want 0", len(sink.activeDrift))
	}
}

func TestMakeKey(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		path     string
		expected string
	}{
		{"simple GET", "GET", "/api/users", "GET /api/users"},
		{"POST request", "POST", "/api/posts", "POST /api/posts"},
		{"root path", "GET", "/", "GET /"},
		{"nested path", "DELETE", "/api/v1/users/123", "DELETE /api/v1/users/123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := makeKey(tt.method, tt.path)
			if result != tt.expected {
				t.Errorf("makeKey(%s, %s) = %s, want %s", tt.method, tt.path, result, tt.expected)
			}
		})
	}
}

func TestConsoleSink_EmitDriftAlert(t *testing.T) {
	t.Run("emit first alert", func(t *testing.T) {
		sink := NewConsoleSink()

		alert := DriftAlert{
			Method:             "GET",
			Path:               "/api/test",
			PercentageIncrease: 50.0,
			OlderAvgMillis:     100.0,
			RecentAvgMillis:    150.0,
			Timestamp:          time.Now().UTC().Format(time.RFC3339Nano),
		}

		// Should not panic
		sink.EmitDriftAlert(alert)

		// Verify alert is tracked
		key := makeKey(alert.Method, alert.Path)
		if !sink.activeDrift[key] {
			t.Error("alert not tracked in activeDrift")
		}
	})

	t.Run("deduplicate alert", func(t *testing.T) {
		sink := NewConsoleSink()

		alert := DriftAlert{
			Method:             "GET",
			Path:               "/api/test",
			PercentageIncrease: 50.0,
			OlderAvgMillis:     100.0,
			RecentAvgMillis:    150.0,
			Timestamp:          time.Now().UTC().Format(time.RFC3339Nano),
		}

		// Emit first time
		sink.EmitDriftAlert(alert)

		// Emit second time - should be deduplicated
		sink.EmitDriftAlert(alert)

		// Still should only have one entry
		key := makeKey(alert.Method, alert.Path)
		if !sink.activeDrift[key] {
			t.Error("alert not tracked in activeDrift")
		}
	})

	t.Run("set default type", func(t *testing.T) {
		sink := NewConsoleSink()

		alert := DriftAlert{
			Method:             "GET",
			Path:               "/api/test",
			PercentageIncrease: 50.0,
			OlderAvgMillis:     100.0,
			RecentAvgMillis:    150.0,
			// Type is empty
		}

		// Should set default type
		sink.EmitDriftAlert(alert)

		key := makeKey(alert.Method, alert.Path)
		if !sink.activeDrift[key] {
			t.Error("alert not tracked in activeDrift")
		}
	})

	t.Run("set default timestamp", func(t *testing.T) {
		sink := NewConsoleSink()

		alert := DriftAlert{
			Method:             "GET",
			Path:               "/api/test",
			PercentageIncrease: 50.0,
			OlderAvgMillis:     100.0,
			RecentAvgMillis:    150.0,
			Type:               "latency_drift",
			// Timestamp is empty
		}

		// Should set default timestamp
		sink.EmitDriftAlert(alert)

		key := makeKey(alert.Method, alert.Path)
		if !sink.activeDrift[key] {
			t.Error("alert not tracked in activeDrift")
		}
	})

	t.Run("concurrent alerts", func(t *testing.T) {
		sink := NewConsoleSink()

		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func(n int) {
				alert := DriftAlert{
					Method:             "GET",
					Path:               "/api/test",
					PercentageIncrease: 50.0,
					OlderAvgMillis:     100.0,
					RecentAvgMillis:    150.0,
					Timestamp:          time.Now().UTC().Format(time.RFC3339Nano),
				}
				sink.EmitDriftAlert(alert)
				done <- true
			}(i)
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}

		// Should still only have one entry due to deduplication
		key := makeKey("GET", "/api/test")
		if !sink.activeDrift[key] {
			t.Error("alert not tracked in activeDrift")
		}
	})
}

func TestConsoleSink_ClearDrift(t *testing.T) {
	t.Run("clear existing drift", func(t *testing.T) {
		sink := NewConsoleSink()

		// First emit an alert
		alert := DriftAlert{
			Method:             "GET",
			Path:               "/api/test",
			PercentageIncrease: 50.0,
			OlderAvgMillis:     100.0,
			RecentAvgMillis:    150.0,
			Timestamp:          time.Now().UTC().Format(time.RFC3339Nano),
		}
		sink.EmitDriftAlert(alert)

		// Verify it's tracked
		key := makeKey("GET", "/api/test")
		if !sink.activeDrift[key] {
			t.Fatal("alert not tracked before clear")
		}

		// Clear the drift
		sink.ClearDrift("GET", "/api/test")

		// Verify it's removed
		if sink.activeDrift[key] {
			t.Error("alert still tracked after clear")
		}
	})

	t.Run("clear non-existent drift", func(t *testing.T) {
		sink := NewConsoleSink()

		// Should not panic
		sink.ClearDrift("GET", "/api/test")
	})

	t.Run("concurrent clear", func(t *testing.T) {
		sink := NewConsoleSink()

		// Emit an alert
		alert := DriftAlert{
			Method:             "GET",
			Path:               "/api/test",
			PercentageIncrease: 50.0,
			OlderAvgMillis:     100.0,
			RecentAvgMillis:    150.0,
			Timestamp:          time.Now().UTC().Format(time.RFC3339Nano),
		}
		sink.EmitDriftAlert(alert)

		// Clear concurrently
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func() {
				sink.ClearDrift("GET", "/api/test")
				done <- true
			}()
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}

		// Verify it's removed
		key := makeKey("GET", "/api/test")
		if sink.activeDrift[key] {
			t.Error("alert still tracked after concurrent clear")
		}
	})
}

func TestDriftAlert(t *testing.T) {
	t.Run("create valid alert", func(t *testing.T) {
		alert := DriftAlert{
			Type:               "latency_drift",
			Method:             "GET",
			Path:               "/api/users",
			PercentageIncrease: 45.67,
			OlderAvgMillis:     100.0,
			RecentAvgMillis:    145.67,
			Timestamp:          time.Now().UTC().Format(time.RFC3339Nano),
		}

		if alert.Type != "latency_drift" {
			t.Errorf("Type = %s, want latency_drift", alert.Type)
		}
		if alert.Method != "GET" {
			t.Errorf("Method = %s, want GET", alert.Method)
		}
		if alert.Path != "/api/users" {
			t.Errorf("Path = %s, want /api/users", alert.Path)
		}
		if alert.PercentageIncrease != 45.67 {
			t.Errorf("PercentageIncrease = %f, want 45.67", alert.PercentageIncrease)
		}
	})
}

func BenchmarkConsoleSink_EmitDriftAlert(b *testing.B) {
	sink := NewConsoleSink()

	alert := DriftAlert{
		Method:             "GET",
		Path:               "/api/test",
		PercentageIncrease: 50.0,
		OlderAvgMillis:     100.0,
		RecentAvgMillis:    150.0,
		Timestamp:          time.Now().UTC().Format(time.RFC3339Nano),
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Clear between iterations to avoid deduplication
		sink.ClearDrift(alert.Method, alert.Path)
		sink.EmitDriftAlert(alert)
	}
}
