package main

// Coded by Tahsin Mert Mutlu – Real Time Code Intelligence Gateway (RCIG)

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/example/rcig/internal/alert"
	"github.com/example/rcig/internal/analyzer"
	"github.com/example/rcig/internal/metrics"
	"github.com/example/rcig/internal/proxy"
)

type config struct {
	ListenAddr string
	TargetURL  string
}

func loadConfig() config {
	listen := os.Getenv("RCIG_LISTEN_ADDR")
	if listen == "" {
		listen = ":8080"
	}

	target := os.Getenv("RCIG_TARGET_URL")
	if target == "" {
		log.Fatal("RCIG_TARGET_URL environment variable is required (e.g.: http://localhost:9000)")
	}

	return config{
		ListenAddr: listen,
		TargetURL:  target,
	}
}

func main() {
	cfg := loadConfig()

	targetURL, err := url.Parse(cfg.TargetURL)
	if err != nil {
		log.Fatalf("failed to parse target URL: %v", err)
	}

	rp := httputil.NewSingleHostReverseProxy(targetURL)

	registry := metrics.NewRegistry(100)

	alertSink := alert.NewConsoleSink()

	an := analyzer.NewAnalyzer(registry, alertSink, analyzer.Config{
		TickInterval:     5 * time.Second,
		OlderWindowSize:  80,
		RecentWindowSize: 20,
		DriftThreshold:   0.40,
	})
	go an.Run()

	handler := proxy.NewHandler(rp, registry)

	mux := http.NewServeMux()
	mux.Handle("/rcig/metrics", metrics.NewHTTPHandler(registry))
	mux.Handle("/rcig/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	mux.Handle("/", handler)

	server := &http.Server{
		Addr:         cfg.ListenAddr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("RCIG proxy listening on %s, target: %s", cfg.ListenAddr, cfg.TargetURL)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)
}
