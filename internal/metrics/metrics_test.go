package metrics

import (
	"testing"
	"time"
)

func TestNewCircularBuffer(t *testing.T) {
	tests := []struct {
		name string
		size int
	}{
		{"small buffer", 5},
		{"medium buffer", 100},
		{"large buffer", 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := NewCircularBuffer(tt.size)
			if buf == nil {
				t.Fatal("NewCircularBuffer returned nil")
			}
			if buf.size != tt.size {
				t.Errorf("size = %d, want %d", buf.size, tt.size)
			}
			if buf.count != 0 {
				t.Errorf("count = %d, want 0", buf.count)
			}
			if buf.head != 0 {
				t.Errorf("head = %d, want 0", buf.head)
			}
			if len(buf.values) != tt.size {
				t.Errorf("values length = %d, want %d", len(buf.values), tt.size)
			}
		})
	}
}

func TestCircularBuffer_Append(t *testing.T) {
	t.Run("append to empty buffer", func(t *testing.T) {
		buf := NewCircularBuffer(5)
		buf.Append(100 * time.Millisecond)

		if buf.count != 1 {
			t.Errorf("count = %d, want 1", buf.count)
		}
		if buf.head != 1 {
			t.Errorf("head = %d, want 1", buf.head)
		}
	})

	t.Run("append multiple values", func(t *testing.T) {
		buf := NewCircularBuffer(5)
		values := []time.Duration{
			10 * time.Millisecond,
			20 * time.Millisecond,
			30 * time.Millisecond,
		}

		for _, v := range values {
			buf.Append(v)
		}

		if buf.count != 3 {
			t.Errorf("count = %d, want 3", buf.count)
		}
		if buf.head != 3 {
			t.Errorf("head = %d, want 3", buf.head)
		}
	})

	t.Run("append beyond capacity", func(t *testing.T) {
		size := 3
		buf := NewCircularBuffer(size)

		for i := 0; i < 10; i++ {
			buf.Append(time.Duration(i) * time.Millisecond)
		}

		if buf.count != size {
			t.Errorf("count = %d, want %d", buf.count, size)
		}
		if buf.head != 10%size {
			t.Errorf("head = %d, want %d", buf.head, 10%size)
		}
	})
}

func TestCircularBuffer_SnapshotValues(t *testing.T) {
	t.Run("snapshot empty buffer", func(t *testing.T) {
		buf := NewCircularBuffer(5)
		snapshot := buf.SnapshotValues()

		if snapshot != nil {
			t.Errorf("snapshot = %v, want nil", snapshot)
		}
	})

	t.Run("snapshot partial buffer", func(t *testing.T) {
		buf := NewCircularBuffer(5)
		values := []time.Duration{
			10 * time.Millisecond,
			20 * time.Millisecond,
			30 * time.Millisecond,
		}

		for _, v := range values {
			buf.Append(v)
		}

		snapshot := buf.SnapshotValues()
		if len(snapshot) != len(values) {
			t.Fatalf("snapshot length = %d, want %d", len(snapshot), len(values))
		}

		for i, v := range values {
			if snapshot[i] != v {
				t.Errorf("snapshot[%d] = %v, want %v", i, snapshot[i], v)
			}
		}
	})

	t.Run("snapshot full buffer", func(t *testing.T) {
		size := 3
		buf := NewCircularBuffer(size)

		for i := 0; i < size; i++ {
			buf.Append(time.Duration(i) * time.Millisecond)
		}

		snapshot := buf.SnapshotValues()
		if len(snapshot) != size {
			t.Fatalf("snapshot length = %d, want %d", len(snapshot), size)
		}

		for i := 0; i < size; i++ {
			expected := time.Duration(i) * time.Millisecond
			if snapshot[i] != expected {
				t.Errorf("snapshot[%d] = %v, want %v", i, snapshot[i], expected)
			}
		}
	})

	t.Run("snapshot wrapped buffer", func(t *testing.T) {
		size := 3
		buf := NewCircularBuffer(size)

		// Fill buffer and wrap around
		for i := 0; i < 7; i++ {
			buf.Append(time.Duration(i) * time.Millisecond)
		}

		snapshot := buf.SnapshotValues()
		if len(snapshot) != size {
			t.Fatalf("snapshot length = %d, want %d", len(snapshot), size)
		}

		// Should contain last 3 values: 4, 5, 6
		expected := []time.Duration{
			4 * time.Millisecond,
			5 * time.Millisecond,
			6 * time.Millisecond,
		}

		for i, exp := range expected {
			if snapshot[i] != exp {
				t.Errorf("snapshot[%d] = %v, want %v", i, snapshot[i], exp)
			}
		}
	})
}

func TestNewRegistry(t *testing.T) {
	tests := []struct {
		name string
		size int
	}{
		{"small registry", 10},
		{"medium registry", 100},
		{"large registry", 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := NewRegistry(tt.size)
			if reg == nil {
				t.Fatal("NewRegistry returned nil")
			}
			if reg.size != tt.size {
				t.Errorf("size = %d, want %d", reg.size, tt.size)
			}
			if reg.metrics == nil {
				t.Error("metrics map is nil")
			}
			if len(reg.metrics) != 0 {
				t.Errorf("metrics length = %d, want 0", len(reg.metrics))
			}
		})
	}
}

func TestRegistry_Record(t *testing.T) {
	t.Run("record single metric", func(t *testing.T) {
		reg := NewRegistry(100)
		reg.Record("GET", "/api/users", 100*time.Millisecond, 200)

		if len(reg.metrics) != 1 {
			t.Errorf("metrics count = %d, want 1", len(reg.metrics))
		}

		key := EndpointKey{Method: "GET", Path: "/api/users"}
		m, ok := reg.metrics[key]
		if !ok {
			t.Fatal("metric not found")
		}

		if m.count != 1 {
			t.Errorf("count = %d, want 1", m.count)
		}
		if m.lastStatus != 200 {
			t.Errorf("lastStatus = %d, want 200", m.lastStatus)
		}
	})

	t.Run("record multiple metrics for same endpoint", func(t *testing.T) {
		reg := NewRegistry(100)

		for i := 0; i < 5; i++ {
			reg.Record("GET", "/api/users", 100*time.Millisecond, 200)
		}

		if len(reg.metrics) != 1 {
			t.Errorf("metrics count = %d, want 1", len(reg.metrics))
		}

		key := EndpointKey{Method: "GET", Path: "/api/users"}
		m, ok := reg.metrics[key]
		if !ok {
			t.Fatal("metric not found")
		}

		if m.count != 5 {
			t.Errorf("count = %d, want 5", m.count)
		}
	})

	t.Run("record metrics for different endpoints", func(t *testing.T) {
		reg := NewRegistry(100)

		endpoints := []struct {
			method string
			path   string
			status int
		}{
			{"GET", "/api/users", 200},
			{"POST", "/api/users", 201},
			{"GET", "/api/posts", 200},
			{"DELETE", "/api/users", 204},
		}

		for _, ep := range endpoints {
			reg.Record(ep.method, ep.path, 100*time.Millisecond, ep.status)
		}

		if len(reg.metrics) != len(endpoints) {
			t.Errorf("metrics count = %d, want %d", len(reg.metrics), len(endpoints))
		}

		for _, ep := range endpoints {
			key := EndpointKey{Method: ep.method, Path: ep.path}
			m, ok := reg.metrics[key]
			if !ok {
				t.Errorf("metric not found for %s %s", ep.method, ep.path)
				continue
			}

			if m.lastStatus != ep.status {
				t.Errorf("lastStatus = %d, want %d", m.lastStatus, ep.status)
			}
		}
	})

	t.Run("concurrent recording", func(t *testing.T) {
		reg := NewRegistry(100)

		// Run concurrent recordings
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func() {
				for j := 0; j < 100; j++ {
					reg.Record("GET", "/api/test", 100*time.Millisecond, 200)
				}
				done <- true
			}()
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}

		key := EndpointKey{Method: "GET", Path: "/api/test"}
		m, ok := reg.metrics[key]
		if !ok {
			t.Fatal("metric not found")
		}

		if m.count != 1000 {
			t.Errorf("count = %d, want 1000", m.count)
		}
	})
}

func TestRegistry_Snapshot(t *testing.T) {
	t.Run("snapshot empty registry", func(t *testing.T) {
		reg := NewRegistry(100)
		snapshot := reg.Snapshot()

		if len(snapshot) != 0 {
			t.Errorf("snapshot length = %d, want 0", len(snapshot))
		}
	})

	t.Run("snapshot with data", func(t *testing.T) {
		reg := NewRegistry(100)

		reg.Record("GET", "/api/users", 100*time.Millisecond, 200)
		reg.Record("GET", "/api/users", 150*time.Millisecond, 200)
		reg.Record("POST", "/api/posts", 200*time.Millisecond, 201)

		snapshot := reg.Snapshot()

		if len(snapshot) != 2 {
			t.Errorf("snapshot length = %d, want 2", len(snapshot))
		}

		key1 := EndpointKey{Method: "GET", Path: "/api/users"}
		values1, ok := snapshot[key1]
		if !ok {
			t.Error("GET /api/users not found in snapshot")
		}
		if len(values1) != 2 {
			t.Errorf("GET /api/users values length = %d, want 2", len(values1))
		}

		key2 := EndpointKey{Method: "POST", Path: "/api/posts"}
		values2, ok := snapshot[key2]
		if !ok {
			t.Error("POST /api/posts not found in snapshot")
		}
		if len(values2) != 1 {
			t.Errorf("POST /api/posts values length = %d, want 1", len(values2))
		}
	})
}

func TestRegistry_Summaries(t *testing.T) {
	t.Run("summaries empty registry", func(t *testing.T) {
		reg := NewRegistry(100)
		summaries := reg.Summaries()

		if len(summaries) != 0 {
			t.Errorf("summaries length = %d, want 0", len(summaries))
		}
	})

	t.Run("summaries with data", func(t *testing.T) {
		reg := NewRegistry(100)

		reg.Record("GET", "/api/users", 100*time.Millisecond, 200)
		reg.Record("GET", "/api/users", 150*time.Millisecond, 200)
		reg.Record("POST", "/api/posts", 200*time.Millisecond, 201)

		summaries := reg.Summaries()

		if len(summaries) != 2 {
			t.Errorf("summaries length = %d, want 2", len(summaries))
		}

		// Find GET /api/users summary
		var getUsersSummary *EndpointSummary
		for i := range summaries {
			if summaries[i].Method == "GET" && summaries[i].Path == "/api/users" {
				getUsersSummary = &summaries[i]
				break
			}
		}

		if getUsersSummary == nil {
			t.Fatal("GET /api/users summary not found")
		}

		if getUsersSummary.Count != 2 {
			t.Errorf("count = %d, want 2", getUsersSummary.Count)
		}
		if getUsersSummary.LastStatus != 200 {
			t.Errorf("lastStatus = %d, want 200", getUsersSummary.LastStatus)
		}
	})
}

func BenchmarkCircularBuffer_Append(b *testing.B) {
	buf := NewCircularBuffer(1000)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Append(time.Duration(i) * time.Millisecond)
	}
}

func BenchmarkRegistry_Record(b *testing.B) {
	reg := NewRegistry(1000)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		reg.Record("GET", "/api/test", 100*time.Millisecond, 200)
	}
}

func BenchmarkRegistry_Snapshot(b *testing.B) {
	reg := NewRegistry(1000)

	// Populate with data
	for i := 0; i < 100; i++ {
		reg.Record("GET", "/api/test", 100*time.Millisecond, 200)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = reg.Snapshot()
	}
}
