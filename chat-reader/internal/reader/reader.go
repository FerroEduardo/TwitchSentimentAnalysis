package reader

import (
	"chat-reader/internal/kafka"
	"chat-reader/internal/metrics"
	"chat-reader/internal/twitch"
	"context"
	"encoding/json"
	"time"

	"go.uber.org/zap"
)

func Start(ctx context.Context, logger *zap.SugaredLogger) {
	messageChan := make(chan *twitch.Message)
	kafkaClient := kafka.NewKafkaClient(logger.Named("kafka-client"))
	twitchClient := twitch.NewTwitchClient(messageChan, logger.Named("twitch-client"))
	metricsServer := metrics.NewMetricsServer(logger.Named("metrics-server"))

	flushTicker := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-ctx.Done():
			logger.Info("Shutting down...")
			twitchClient.Cleanup()
			kafkaClient.Cleanup()
			metricsServer.Cleanup()

			return
		case <-flushTicker.C:
			if kafkaClient.BufferCount() == 0 {
				logger.Debug("Skipping flush due to empty buffer...")
				continue
			}

			ctx, stop := context.WithTimeout(context.Background(), 1*time.Second)
			defer stop()
			kafkaClient.Flush(ctx)
		case message := <-messageChan:
			logger.Info(message)
			b, err := json.Marshal(message)
			if err != nil {
				logger.Error("failed to encode message to JSON", err)
				continue
			}

			if kafkaClient.BufferCount() > 100 {
				ctx, stop := context.WithTimeout(context.Background(), 1*time.Second)
				defer stop()
				kafkaClient.Flush(ctx)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			kafkaClient.AsyncProduce(ctx, b)
		}
	}
}
