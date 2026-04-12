package web

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/coder/websocket"
)

type Event struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

type Hub struct {
	logger  *slog.Logger
	clients map[*websocket.Conn]context.CancelFunc
	mu      sync.RWMutex
}

func NewHub(logger *slog.Logger) *Hub {
	return &Hub{
		logger:  logger,
		clients: make(map[*websocket.Conn]context.CancelFunc),
	}
}

func (h *Hub) Add(conn *websocket.Conn, cancel context.CancelFunc) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[conn] = cancel
	h.logger.Info("websocket client connected", "total", len(h.clients))
}

func (h *Hub) Remove(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if cancel, ok := h.clients[conn]; ok {
		cancel()
		delete(h.clients, conn)
	}
	h.logger.Info("websocket client disconnected", "total", len(h.clients))
}

func (h *Hub) Broadcast(event Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	data, err := json.Marshal(event)
	if err != nil {
		h.logger.Error("failed to marshal event", "error", err)
		return
	}

	for conn := range h.clients {
		err := conn.Write(context.Background(), websocket.MessageText, data)
		if err != nil {
			h.logger.Debug("failed to write to client", "error", err)
		}
	}
}
