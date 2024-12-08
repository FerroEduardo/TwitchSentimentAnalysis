package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"go.uber.org/zap"

	"website/internal/testutils"
)

var logger *zap.SugaredLogger

func init() {
	logger = zap.NewNop().Sugar()
}

func TestClient(t *testing.T) {
	ctx := context.Background()

	postgresContainer, err := testutils.StartPostgresContainer()
	require.NoError(t, err)
	defer testcontainers.CleanupContainer(t, postgresContainer)

	t.Setenv("DATABASE_DSN", postgresContainer.MustConnectionString(ctx))
	conn := NewDatabaseConnection(logger)
	err = conn.Ping(ctx)
	require.NoError(t, err)

	_, err = conn.Exec(ctx, "INSERT INTO results (channel, \"user\", \"message_id\", \"timestamp\", message, sentiment_positive, sentiment_neutral, sentiment_negative ) VALUES ( 'channel1', 'user123', 'msg-001', '2024-12-01T14:30:00+00:00', 'This is a sample message.', 0.8, 0.1, 0.1 );")
	require.NoError(t, err)

	row := conn.QueryRow(ctx, "SELECT channel, \"user\" FROM results")

	var channel string
	var user string
	err = row.Scan(&channel, &user)
	require.NoError(t, err)

	require.Equal(t, "channel1", channel)
	require.Equal(t, "user123", user)
}
