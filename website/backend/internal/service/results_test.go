package service

import (
	"context"
	"testing"
	"time"
	"website/internal/database"
	"website/internal/testutils"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"go.uber.org/zap"
)

var logger *zap.SugaredLogger
var now time.Time

func init() {
	logger = zap.NewNop().Sugar()
	now = time.Date(2024, 12, 1, 14, 0, 0, 0, time.UTC)
}

func TestGetLastHourChannelAverageResults(t *testing.T) {
	ctx := context.Background()

	postgresContainer, err := testutils.StartPostgresContainer()
	require.NoError(t, err)
	defer testcontainers.CleanupContainer(t, postgresContainer)

	dsn := postgresContainer.MustConnectionString(ctx)

	databaseChannels := 10
	messagesPerChannel := 2
	err = testutils.PopulateDatabase(dsn, databaseChannels, messagesPerChannel)
	require.NoError(t, err)

	t.Setenv("DATABASE_DSN", dsn)
	conn := database.NewDatabaseConnection(logger)
	require.NoError(t, err)

	service := ResultsService{
		conn:   conn,
		logger: logger,
	}

	require.NoError(t, err)

	results, err := service.GetLastHourChannelAverageResults(ctx, now)
	require.NoError(t, err)
	require.Len(t, results, databaseChannels)
	for _, channelResult := range results {
		require.Len(t, channelResult, messagesPerChannel)
	}
}

func TestGetLastResults(t *testing.T) {
	ctx := context.Background()

	postgresContainer, err := testutils.StartPostgresContainer()
	require.NoError(t, err)
	defer testcontainers.CleanupContainer(t, postgresContainer)

	dsn := postgresContainer.MustConnectionString(ctx)

	databaseChannels := 10
	messagesPerChannel := 2
	err = testutils.PopulateDatabase(dsn, databaseChannels, messagesPerChannel)
	require.NoError(t, err)

	t.Setenv("DATABASE_DSN", dsn)
	conn := database.NewDatabaseConnection(logger)
	require.NoError(t, err)

	service := ResultsService{
		conn:   conn,
		logger: logger,
	}

	results, err := service.GetLastResults(ctx, 5, now)
	require.NoError(t, err)
	require.NotEmpty(t, results)
	require.LessOrEqual(t, len(results), 5)
}
