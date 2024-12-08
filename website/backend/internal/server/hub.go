package server

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	clientsGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "broadcast_hub_clients_total",
	})

	messagesCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "broadcast_hub_messages_total",
	})
)

type Client struct {
	conn *websocket.Conn
	send chan []byte
}

func (c *Client) String() string {
	return fmt.Sprintf("Client - %v", c.conn.RemoteAddr())
}

// Handles connected clients and broadcasts messages
type broadcastHub struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte

	mu *sync.Mutex

	logger *zap.SugaredLogger
}

func NewBroadcastHub(logger *zap.SugaredLogger) *broadcastHub {
	return &broadcastHub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
		mu:         &sync.Mutex{},
		logger:     logger,
	}
}

// Starts event loop
func (h *broadcastHub) Start() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.logger.Debug("Registering client", client)
			h.clients[client] = true
			clientsGauge.Inc()
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				h.logger.Debug("Unregistering client", client)
				delete(h.clients, client)
				close(client.send)
				clientsGauge.Dec()
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.Lock()
			messagesCounter.Inc()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					delete(h.clients, client)
					close(client.send)
				}
			}
			h.mu.Unlock()
		}
	}
}

func (h *broadcastHub) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.logger.Info("Stopping broadcastHub")

	for client := range h.clients {
		h.logger.Debugf("Closing connection for client: %s", client)
		if err := client.conn.Close(); err != nil {
			h.logger.Errorf("Error closing client connection: %v", err)
		}
	}

	h.logger.Info("broadcastHub stopped")
}

// Handles the WebSocket connection for a single client
func handleClient(hub *broadcastHub, conn *websocket.Conn, logger *zap.SugaredLogger) {
	client := &Client{conn: conn, send: make(chan []byte)}
	hub.register <- client

	defer func() {
		hub.unregister <- client
		err := conn.Close()
		if err != nil && !errors.Is(err, net.ErrClosed) {
			logger.Error("failed to close client connection: ", err)
		}
	}()

	// Start a goroutine to write messages to the client
	go func() {
		for message := range client.send {
			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				logger.Error("failed to send message to client: ", err)
				return
			}
		}
	}()

	for {
		// Read loop to keep the connection open
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}

func wsHandler(hub *broadcastHub, logger *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Errorf("Failed to upgrade connection: %v", err)
			return
		}
		handleClient(hub, conn, logger.Named("client-handler"))
	}
}
