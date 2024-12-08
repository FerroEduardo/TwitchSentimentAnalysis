package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"website/internal/database"
	"website/internal/logger"
	"website/internal/server"
	"website/internal/service"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger := logger.NewLogger()
	defer logger.Sync()

	conn := database.NewDatabaseConnection(logger.Named("database"))
	resultsService := service.NewResultsService(conn, logger.Named("results-service"))

	server.Start(ctx, logger, resultsService)
}
