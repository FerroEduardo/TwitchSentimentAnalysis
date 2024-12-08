package kafka

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"go.uber.org/zap"

	testutils "chat-reader/internal/testutils"
)

var logger *zap.SugaredLogger

func init() {
	logger = zap.NewNop().Sugar()
}

func TestClient(t *testing.T) {
	kafkaContainer, broker, err := testutils.StartKafkaContainer()
	require.NoError(t, err)
	defer testcontainers.CleanupContainer(t, kafkaContainer)

	err = testutils.CreateTopics(*broker)
	require.NoError(t, err)

	t.Setenv("KAFKA_BROKER_HOST", *broker)

	kafkaClient := NewKafkaClient(logger)
	defer kafkaClient.Cleanup()

	require.Zero(t, kafkaClient.BufferCount())

	kafkaClient.AsyncProduce(context.Background(), []byte("hi"))
	require.Equal(t, int64(1), kafkaClient.BufferCount())

	kafkaClient.Flush(context.Background())
	require.Zero(t, kafkaClient.BufferCount())

	messages, err := testutils.ConsumeTopic(*broker, "messages")
	require.NoError(t, err)
	require.Len(t, messages, 1)

	require.Equal(t, "hi", messages[0])
}

func TestClientWithDisconnection(t *testing.T) {
	kafkaContainer, broker, err := testutils.StartKafkaContainer()
	require.NoError(t, err)
	defer testcontainers.CleanupContainer(t, kafkaContainer)

	err = testutils.CreateTopics(*broker)
	require.NoError(t, err)

	t.Setenv("KAFKA_BROKER_HOST", *broker)

	kafkaClient := NewKafkaClient(logger)
	defer kafkaClient.Cleanup()

	require.Zero(t, kafkaClient.BufferCount())

	kafkaClient.AsyncProduce(context.Background(), []byte("hi1"))
	require.Equal(t, int64(1), kafkaClient.BufferCount())

	kafkaClient.Flush(context.Background())
	require.Zero(t, kafkaClient.BufferCount())

	// Simulate disconnection
	second := time.Duration(time.Second)
	kafkaContainer.Stop(context.Background(), &second)

	kafkaClient.AsyncProduce(context.Background(), []byte("hi2"))
	require.Equal(t, int64(1), kafkaClient.BufferCount())

	// Flush when broker is off
	ctx, stop := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer stop()
	kafkaClient.Flush(ctx)
	require.Equal(t, int64(1), kafkaClient.BufferCount())

	// Another flush after the broker is ready again
	kafkaContainer.Start(context.Background())
	kafkaClient.Flush(context.Background())
	require.Zero(t, kafkaClient.BufferCount())

	messages, err := testutils.ConsumeTopic(*broker, "messages")
	require.NoError(t, err)
	require.Len(t, messages, 2)

	require.Equal(t, "hi1", messages[0])
	require.Equal(t, "hi2", messages[1])
}

func TestClientWithoutHost(t *testing.T) {
	t.Run("without env", func(t *testing.T) {
		require.Panics(t, func() {
			NewKafkaClient(logger)
		})
	})
	t.Run("empty env", func(t *testing.T) {
		require.Panics(t, func() {
			t.Setenv("KAFKA_BROKER_HOST", "")
			NewKafkaClient(logger)
		})
	})
}
