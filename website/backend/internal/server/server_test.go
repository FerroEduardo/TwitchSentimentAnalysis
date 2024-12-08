package server

import (
	"context"
	"testing"
	"time"
	"website/internal/database"
	"website/internal/service"
	"website/internal/testutils"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"go.uber.org/zap"
)

var logger *zap.SugaredLogger

func init() {
	logger = zap.NewNop().Sugar()
}

func TestServer(t *testing.T) {
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

	service := service.NewResultsService(conn, logger)

	serverCtx, cancel := context.WithCancel(ctx)
	serverSignal := make(chan bool)
	go func() {
		Start(serverCtx, logger, service)
		serverSignal <- true
	}()

	var ws *websocket.Conn
	for i := 0; i < 5; i++ {
		ws, _, err = websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
		if err == nil {
			break
		}
		time.Sleep(time.Millisecond * 50)
	}
	require.NoError(t, err)
	defer ws.Close()

	cancel()
	<-serverSignal
}

func TestServerEvents(t *testing.T) {
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

	service := service.NewResultsService(conn, logger)

	serverCtx, cancel := context.WithCancel(ctx)
	serverSignal := make(chan bool)
	go func() {
		Start(serverCtx, logger, service)
		serverSignal <- true
	}()

	var ws *websocket.Conn
	for i := 0; i < 5; i++ {
		ws, _, err = websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
		if err == nil {
			break
		}
		time.Sleep(time.Millisecond * 100)
	}
	require.NoError(t, err)
	defer ws.Close()

	events := []string{"messages", "results"}
	eventsFound := map[string]bool{}

	for i := 0; i < 10; i++ {
		var event Event
		err = ws.ReadJSON(&event)
		require.NoError(t, err)
		require.NotEmpty(t, event.Event)

		if _, found := eventsFound[event.Event]; !found {
			eventsFound[event.Event] = true
		}
		if len(eventsFound) == len(events) {
			break
		}
	}
	if len(eventsFound) != len(events) {
		t.Log("Not all events were returned")
		t.Fail()
	}

	cancel()
	<-serverSignal
}
