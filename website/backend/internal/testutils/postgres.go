package testutils

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func StartPostgresContainer() (*postgres.PostgresContainer, error) {
	databaseName := "twitchsentimentanalysis"
	username := "user"
	password := "password"

	postgresContainer, err := postgres.Run(
		context.Background(),
		"postgres:16.6-alpine3.20",
		postgres.WithInitScripts(filepath.Join("../../../../message-analyzer/database.sql")),
		postgres.WithDatabase(databaseName),
		postgres.WithUsername(username),
		postgres.WithPassword(password),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)

	return postgresContainer, err
}

func PopulateDatabase(dsn string, channelsCount int, messagesPerChannel int) error {
	conn, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer conn.Close()

	for i := 0; i < channelsCount; i++ {
		channelName := fmt.Sprintf("channel%d", i)
		if err := insertMessagesForChannel(context.Background(), conn, channelName, i, messagesPerChannel); err != nil {
			return fmt.Errorf("failed to insert messages for channel %s: %w", channelName, err)
		}
	}

	return nil
}

func insertMessagesForChannel(ctx context.Context, conn *pgxpool.Pool, channelName string, channelIndex, messagesPerChannel int) error {
	for j := 0; j < messagesPerChannel; j++ {
		timestamp := time.Date(2024, 12, 1, 14, j, 0, 0, time.UTC)
		messageID := fmt.Sprintf("msg-%02d%02d", channelIndex, j)
		userName := fmt.Sprintf("user%d", channelIndex)

		_, err := conn.Exec(
			ctx,
			`INSERT INTO results 
			(channel, "user", "message_id", "timestamp", message, sentiment_positive, sentiment_neutral, sentiment_negative) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8);`,
			channelName, userName, messageID, timestamp, "sample message", 0.8, 0.1, 0.1,
		)
		if err != nil {
			return err
		}
	}
	return nil
}
