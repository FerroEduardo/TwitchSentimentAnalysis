package metrics

import (
	"context"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type MetricsServer struct {
	logger *zap.SugaredLogger

	server *http.Server
}

func NewMetricsServer(logger *zap.SugaredLogger) *MetricsServer {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	server := &http.Server{Addr: ":8080", Handler: mux}

	go func() {
		logger.Info("Server started on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Server failed: %v", err)
		}
	}()

	return &MetricsServer{
		logger: logger,
		server: server,
	}
}

func (s *MetricsServer) Cleanup() {
	s.logger.Info("Closing server gracefully")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		s.logger.Fatalf("Server forced to shutdown: %v", err)
	}
}
