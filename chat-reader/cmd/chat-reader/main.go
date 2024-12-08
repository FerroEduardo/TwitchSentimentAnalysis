package main

import (
	"chat-reader/internal/logger"
	"chat-reader/internal/reader"
	"context"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger := logger.NewLogger()
	defer logger.Sync()

	reader.Start(ctx, logger)
}
