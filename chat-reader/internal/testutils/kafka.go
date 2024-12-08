package test

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"strconv"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/pkg/errors"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/kversion"
)

func StartKafkaContainer() (testcontainers.Container, *string, error) {
	brokerPort := strconv.Itoa(GetFreePort())
	controllerPort := strconv.Itoa(GetFreePort())

	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "apache/kafka-native:latest",
		WaitingFor:   wait.ForLog("Kafka Server started"),
		ExposedPorts: []string{brokerPort + "/tcp", controllerPort + "/tcp"},
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.PortBindings = map[nat.Port][]nat.PortBinding{
				nat.Port(brokerPort + "/tcp"):     {{HostIP: "0.0.0.0", HostPort: brokerPort}},
				nat.Port(controllerPort + "/tcp"): {{HostIP: "0.0.0.0", HostPort: controllerPort}},
			}
		},
		Env: map[string]string{
			"KAFKA_NODE_ID":                                  "1",
			"KAFKA_PROCESS_ROLES":                            "broker,controller",
			"KAFKA_LISTENERS":                                fmt.Sprintf("PLAINTEXT://0.0.0.0:%v,CONTROLLER://0.0.0.0:%v", brokerPort, controllerPort),
			"KAFKA_ADVERTISED_LISTENERS":                     fmt.Sprintf("PLAINTEXT://localhost:%v", brokerPort),
			"KAFKA_CONTROLLER_LISTENER_NAMES":                "CONTROLLER",
			"KAFKA_LISTENER_SECURITY_PROTOCOL_MAP":           "CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT",
			"KAFKA_CONTROLLER_QUORUM_VOTERS":                 fmt.Sprintf("1@localhost:%v", controllerPort),
			"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR":         "1",
			"KAFKA_TRANSACTION_STATE_LOG_REPLICATION_FACTOR": "1",
			"KAFKA_TRANSACTION_STATE_LOG_MIN_ISR":            "1",
			"KAFKA_GROUP_INITIAL_REBALANCE_DELAY_MS":         "0",
			"KAFKA_NUM_PARTITIONS":                           "1",
		},
	}
	kafkaContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, nil, err
	}

	host, err := kafkaContainer.Host(context.Background())
	if err != nil {
		return nil, nil, err
	}

	port, err := kafkaContainer.MappedPort(ctx, nat.Port(brokerPort))
	if err != nil {
		return nil, nil, err
	}

	broker := fmt.Sprintf("%v:%v", host, port.Port())

	return kafkaContainer, &broker, nil
}

func CreateTopics(broker string) error {
	seeds := []string{broker}
	var adminClient *kadm.Client
	{
		client, err := kgo.NewClient(
			kgo.SeedBrokers(seeds...),
			kgo.MaxVersions(kversion.V2_4_0()),
		)
		if err != nil {
			return err
		}
		defer client.Close()

		err = client.Ping(context.Background())
		if err != nil {
			return err
		}

		adminClient = kadm.NewClient(client)
	}

	_, err := adminClient.CreateTopic(context.Background(), 1, 1,
		map[string]*string{
			"delete.retention.ms": kadm.StringPtr("60000"),
		}, "messages",
	)
	if err != nil {
		return errors.Errorf("failed to create topic: %v", err)
	}

	return nil
}

func GetFreePort() (port int) {
	var a *net.TCPAddr
	var err error
	if a, err = net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", a); err == nil {
			defer l.Close()
			return l.Addr().(*net.TCPAddr).Port
		}
	}
	panic(err)
}

func ConsumeTopic(broker string, topic string) ([]string, error) {
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(broker),
		kgo.ConsumerGroup(strconv.Itoa(rand.Int())),
		kgo.ConsumeTopics(topic),
	)
	if err != nil {
		return nil, err
	}

	defer cl.Close()

	fetches := cl.PollFetches(context.Background())
	if errs := fetches.Errors(); len(errs) > 0 {
		return nil, errors.New(fmt.Sprint(errs))
	}

	messages := []string{}
	iter := fetches.RecordIter()
	for !iter.Done() {
		record := iter.Next()
		messages = append(messages, string(record.Value))
	}

	return messages, nil
}
