package server

import (
	"context"
	"net/http"
	"time"
	"website/internal/service"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func Start(ctx context.Context, logger *zap.SugaredLogger, resultsService *service.ResultsService) {
	hub := NewBroadcastHub(logger.Named("broadcast-hub"))
	go hub.Start()
	defer hub.Stop()

	scheduler := NewScheduler(logger.Named("scheduler"), hub, 1*time.Second, resultsService)
	go scheduler.Start(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", wsHandler(hub, logger.Named("ws-handler")))
	mux.Handle("/", http.FileServer(http.Dir("./public")))
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{Addr: ":8080", Handler: mux}

	go func() {
		logger.Info("WebSocket server started on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Server failed: %v", err)
		}
	}()

	<-ctx.Done()

	logger.Info("Closing server gracefully")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}
}
