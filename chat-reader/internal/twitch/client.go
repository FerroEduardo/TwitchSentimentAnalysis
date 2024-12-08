package twitch

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gempir/go-twitch-irc/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

var (
	channelMessagesReadCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "twitch_messages_read_total",
		},
		[]string{"channel"},
	)
)

type Message struct {
	ID        string `json:"id"`
	Message   string `json:"message"`
	Channel   string `json:"channel"`
	User      string `json:"user"`
	Timestamp int64  `json:"timestamp"`
}

func (m *Message) String() string {
	return fmt.Sprintf("Time: %v - ID: %v - User: %v - Channel: %v - Message: %v", time.Unix(m.Timestamp, 0), m.ID, m.User, m.Channel, m.Message)
}

type Client struct {
	client *twitch.Client
	logger *zap.SugaredLogger
}

func NewTwitchClient(messageChan chan *Message, logger *zap.SugaredLogger) *Client {
	channelsEnv, found := os.LookupEnv("TWITCH_CHANNELS")
	if !found {
		logger.Panic("Missing TWITCH_CHANNELS environment variable")
	}
	channels := strings.Split(channelsEnv, ",")

	for _, channel := range channels {
		if len(channel) == 0 {
			logger.Panicf("Invalid Twitch channel: %v", channel)
		}
	}

	if len(channels) == 0 {
		logger.Panic("Should have at least 1 Twitch channel")
	}

	client := twitch.NewAnonymousClient()

	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		go func() {
			// prevent trash
			if strings.HasPrefix(message.Message, "!") || strings.HasPrefix(message.Message, "@") || message.User.IsMod || message.User.IsBroadcaster {
				return
			}

			processedMessage := message.Message
			for _, v := range message.Emotes {
				processedMessage = strings.ReplaceAll(processedMessage, v.Name, "")
			}
			processedMessage = strings.TrimSpace(processedMessage)
			if len(processedMessage) == 0 {
				return
			}

			channelMessagesReadCounter.With(prometheus.Labels{"channel": message.Channel}).Inc()

			messageChan <- &Message{
				ID:        message.ID,
				Channel:   message.Channel,
				Message:   processedMessage,
				Timestamp: message.Time.Unix(),
				User:      message.User.Name,
			}
		}()
	})

	client.Join(channels...)

	go func() {
		logger.Info("Connecting Twitch client")
		err := client.Connect()
		if err != nil {
			if errors.Is(err, twitch.ErrClientDisconnected) {
				logger.Info("Twitch client disconnected")
				return
			}
			logger.Panic(err)
		}
	}()

	return &Client{
		client: client,
		logger: logger,
	}
}

func (c *Client) Cleanup() {
	if c.client != nil {
		c.logger.Info("Disconnecting Twitch client")
		err := c.client.Disconnect()
		if err != nil {
			c.logger.Panic(err)
		}
	}
}
