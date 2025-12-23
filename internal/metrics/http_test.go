package metrics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewHTTPHandler(t *testing.T) {
	reg := NewRegistry(100)
	handler := NewHTTPHandler(reg)

	if handler == nil {
		t.Fatal("NewHTTPHandler returned nil")
	}
}

func TestHTTPHandler_ServeHTTP(t *testing.T) {
	t.Run("empty registry", func(t *testing.T) {
		reg := NewRegistry(100)
		handler := NewHTTPHandler(reg)

		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status code = %d, want %d", rec.Code, http.StatusOK)
		}

		contentType := rec.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("content-type = %s, want application/json", contentType)
		}

		var resp payload
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(resp.Endpoints) != 0 {
			t.Errorf("endpoints length = %d, want 0", len(resp.Endpoints))
		}
	})

	t.Run("registry with data", func(t *testing.T) {
		reg := NewRegistry(100)

		// Record some metrics
		reg.Record("GET", "/api/users", 100*time.Millisecond, 200)
		reg.Record("GET", "/api/users", 150*time.Millisecond, 200)
		reg.Record("POST", "/api/posts", 200*time.Millisecond, 201)

		handler := NewHTTPHandler(reg)

		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status code = %d, want %d", rec.Code, http.StatusOK)
		}

		var resp payload
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(resp.Endpoints) != 2 {
			t.Errorf("endpoints length = %d, want 2", len(resp.Endpoints))
		}

		// Verify response contains expected data
		foundUsers := false
		foundPosts := false

		for _, ep := range resp.Endpoints {
			if ep.Method == "GET" && ep.Path == "/api/users" {
				foundUsers = true
				if ep.Count != 2 {
					t.Errorf("GET /api/users count = %d, want 2", ep.Count)
				}
				if ep.LastStatus != 200 {
					t.Errorf("GET /api/users lastStatus = %d, want 200", ep.LastStatus)
				}
				if ep.SampleSize != 2 {
					t.Errorf("GET /api/users sampleSize = %d, want 2", ep.SampleSize)
				}
				// Average should be (100 + 150) / 2 = 125
				if ep.AverageMs < 124 || ep.AverageMs > 126 {
					t.Errorf("GET /api/users averageMs = %f, want ~125", ep.AverageMs)
				}
			}

			if ep.Method == "POST" && ep.Path == "/api/posts" {
				foundPosts = true
				if ep.Count != 1 {
					t.Errorf("POST /api/posts count = %d, want 1", ep.Count)
				}
				if ep.LastStatus != 201 {
					t.Errorf("POST /api/posts lastStatus = %d, want 201", ep.LastStatus)
				}
			}
		}

		if !foundUsers {
			t.Error("GET /api/users not found in response")
		}
		if !foundPosts {
			t.Error("POST /api/posts not found in response")
		}
	})

	t.Run("verify average calculation", func(t *testing.T) {
		reg := NewRegistry(100)

		// Record metrics with known values
		durations := []time.Duration{
			100 * time.Millisecond,
			200 * time.Millisecond,
			300 * time.Millisecond,
		}

		for _, d := range durations {
			reg.Record("GET", "/test", d, 200)
		}

		handler := NewHTTPHandler(reg)

		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		var resp payload
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(resp.Endpoints) != 1 {
			t.Fatalf("endpoints length = %d, want 1", len(resp.Endpoints))
		}

		ep := resp.Endpoints[0]
		// Average should be (100 + 200 + 300) / 3 = 200
		expectedAvg := 200.0
		if ep.AverageMs != expectedAvg {
			t.Errorf("averageMs = %f, want %f", ep.AverageMs, expectedAvg)
		}
	})

	t.Run("verify timestamp is set", func(t *testing.T) {
		reg := NewRegistry(100)
		handler := NewHTTPHandler(reg)

		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		rec := httptest.NewRecorder()

		before := time.Now()
		handler.ServeHTTP(rec, req)
		after := time.Now()

		var resp payload
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.GeneratedAt.Before(before) || resp.GeneratedAt.After(after) {
			t.Errorf("generatedAt = %v, want between %v and %v", resp.GeneratedAt, before, after)
		}
	})

	t.Run("concurrent requests", func(t *testing.T) {
		reg := NewRegistry(100)
		handler := NewHTTPHandler(reg)

		// Add some data
		reg.Record("GET", "/test", 100*time.Millisecond, 200)

		// Make concurrent requests
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func() {
				req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)

				if rec.Code != http.StatusOK {
					t.Errorf("status code = %d, want %d", rec.Code, http.StatusOK)
				}

				done <- true
			}()
		}

		// Wait for all requests
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

func BenchmarkHTTPHandler_ServeHTTP(b *testing.B) {
	reg := NewRegistry(1000)

	// Populate with data
	for i := 0; i < 100; i++ {
		reg.Record("GET", "/api/test", 100*time.Millisecond, 200)
	}

	handler := NewHTTPHandler(reg)
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}
