package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/kversion"
)

func main() {
	broker, found := os.LookupEnv("KAFKA_BROKER_HOST")
	if !found {
		panic("Missing KAFKA_BROKER_HOST environment variable")
	}
	seeds := []string{broker}
	var adminClient *kadm.Client
	{
		client, err := kgo.NewClient(
			kgo.SeedBrokers(seeds...),

			// Do not try to send requests newer than 2.4.0 to avoid breaking changes in the request struct.
			// Sometimes there are breaking changes for newer versions where more properties are required to set.
			kgo.MaxVersions(kversion.V2_4_0()),
		)
		if err != nil {
			panic(err)
		}
		defer client.Close()

		tentatives := int64(0)
		for {
			err := client.Ping(context.Background())
			if err == nil {
				break
			}
			tentatives++
			if tentatives > 5 {
				panic(fmt.Sprint("Failed to connect to Kafka broker:", err))
			}
			toSleep := time.Duration(tentatives) * time.Second
			fmt.Printf("Sleeping %s due to failed ping\n", toSleep)
			time.Sleep(toSleep)
		}

		adminClient = kadm.NewClient(client)
	}

	topicsResult, err := adminClient.ListTopics(context.Background())
	if err != nil {
		panic(fmt.Errorf("failed to list topics: %v", err))
	}

	if topicsResult.Has("messages") {
		fmt.Println("Topic messages already exists")
		return
	}

	partitions := int32(1)
	replicationFactor := int16(1)
	res, err := adminClient.CreateTopic(
		context.Background(),
		partitions,
		replicationFactor,
		map[string]*string{
			"delete.retention.ms": strPtr("60000"),
		}, "messages",
	)
	if err != nil {
		panic(fmt.Errorf("failed to create topic: %v", err))
	}
	fmt.Printf("Successfully created topic %v\n", res.Topic)
}

func strPtr(v string) *string {
	return &v
}
