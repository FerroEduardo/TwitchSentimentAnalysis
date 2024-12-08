package kafka

import (
	"context"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

var (
	messagesCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "kafka_messages_processed_total",
	})
)

type Client struct {
	client *kgo.Client
	logger *zap.SugaredLogger

	topic string
}

func NewKafkaClient(logger *zap.SugaredLogger) *Client {
	broker, found := os.LookupEnv("KAFKA_BROKER_HOST")
	if !found || len(broker) == 0 {
		logger.Panic("Missing or invalid KAFKA_BROKER_HOST environment variable")
	}

	seeds := []string{broker}
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(seeds...),
		kgo.ManualFlushing(),
	)
	if err != nil {
		logger.Panic(err)
	}

	tentatives := int64(0)
	for {
		err := cl.Ping(context.Background())
		if err == nil {
			break
		}
		tentatives++
		if tentatives > 5 {
			logger.Panic("Failed to connect to Kafka broker:", err)
		}
		toSleep := time.Duration(tentatives) * time.Second
		logger.Debugf("Sleeping %s due to failed ping", toSleep)
		time.Sleep(toSleep)
	}

	return &Client{
		client: cl,
		logger: logger,
		topic:  "messages",
	}
}

func (c *Client) Cleanup() {
	if c.client != nil {
		c.logger.Info("Closing Kafka client")
		c.Flush(context.Background())
		c.client.Close()
		c.logger.Info("Kafka client closed")
	}
}

func (c *Client) AsyncProduce(ctx context.Context, value []byte) {
	record := &kgo.Record{Topic: c.topic, Value: value}
	c.client.Produce(ctx, record, func(_ *kgo.Record, err error) {
		if err != nil {
			if err == context.DeadlineExceeded {
				c.logger.Debugf("Took to long to produce: %v", err)
			} else {
				c.logger.Errorf("Failed to produce record: %v\n", err)
			}
		} else {
			messagesCounter.Inc()
		}
	})
}

func (c *Client) Flush(ctx context.Context) {
	c.logger.Debug("Flushing Kafka messages")
	err := c.client.Flush(ctx)
	if err != nil {
		c.logger.Errorf("Failed to flush: %v\n", err)
	}
}

func (c *Client) BufferCount() int64 {
	return c.client.BufferedProduceRecords()
}
