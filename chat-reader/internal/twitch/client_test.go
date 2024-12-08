package twitch

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var logger *zap.SugaredLogger

func init() {
	logger = zap.NewNop().Sugar()
}

func TestClient(t *testing.T) {
	t.Setenv("TWITCH_CHANNELS", "gaules,xqc,kaicenat,piratesoftware,summit1g")
	messageChan := make(chan *Message)

	client := NewTwitchClient(messageChan, logger)
	assert.NotNil(t, client)

	timeout := 10 * time.Second
	ctx, stop := context.WithTimeout(context.Background(), timeout)
	defer stop()

	for {
		select {
		case message := <-messageChan:
			assert.NotNil(t, message)

			assert.NotPanics(t, func() {
				client.Cleanup()
			})
			close(messageChan)

			return
		case <-ctx.Done():
			t.Fatalf("Did not receive any message after %v", timeout)
		}
	}
}

func TestClient2(t *testing.T) {
	t.Setenv("TWITCH_CHANNELS", "gaules,xqc,kaicenat,piratesoftware,summit1g")
	messageChan := make(chan *Message)

	client := NewTwitchClient(messageChan, logger)
	assert.NotNil(t, client)

	timeout := 10 * time.Second
	ctx, stop := context.WithTimeout(context.Background(), timeout)
	defer stop()

	for {
		select {
		case message := <-messageChan:
			assert.NotNil(t, message)

			assert.NotPanics(t, func() {
				client.Cleanup()
			})
			close(messageChan)

			return
		case <-ctx.Done():
			t.Fatalf("Did not receive any message after %v", timeout)
		}
	}
}

func TestClientWithoutChannel(t *testing.T) {
	messageChan := make(chan *Message)
	t.Run("without env", func(t *testing.T) {
		assert.Panics(t, func() {
			NewTwitchClient(messageChan, logger)
		})
	})
	t.Run("empty env", func(t *testing.T) {
		t.Setenv("TWITCH_CHANNELS", "")

		assert.Panics(t, func() {
			NewTwitchClient(messageChan, logger)
		})
	})
}
