package reader

import (
	"context"
	"sync"
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

func TestReader(t *testing.T) {
	// Setup Kafka
	kafkaContainer, broker, err := testutils.StartKafkaContainer()
	require.NoError(t, err)
	defer testcontainers.CleanupContainer(t, kafkaContainer)

	err = testutils.CreateTopics(*broker)
	require.NoError(t, err)

	t.Setenv("KAFKA_BROKER_HOST", *broker)

	// Setup Twitch
	t.Setenv("TWITCH_CHANNELS", "gaules,xqc,kaicenat,piratesoftware,summit1g")

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	wg.Add(1)
	go func() {
		Start(ctx, logger)
		wg.Done()
	}()

	tentatives := int64(0)
	for {
		messages, err := testutils.ConsumeTopic(*broker, "messages")
		require.NoError(t, err)
		if len(messages) > 0 {
			break
		}

		tentatives++
		if tentatives > 5 {
			t.Error("Did not produce any message")
			break
		}
		time.Sleep(time.Second)
	}

	cancel()
	wg.Wait()
}
