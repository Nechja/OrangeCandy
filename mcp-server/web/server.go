package web

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"

	"github.com/Nechja/OrangeCandy/mcp-server/store"
	"github.com/coder/websocket"
)

//go:embed dist/*
var distFS embed.FS

type Server struct {
	hub    *Hub
	store  *store.Store
	logger *slog.Logger
	port   int
}

func NewServer(st *store.Store, hub *Hub, logger *slog.Logger, port int) *Server {
	return &Server{
		hub:    hub,
		store:  st,
		logger: logger,
		port:   port,
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/ws", s.handleWebSocket)

	mux.HandleFunc("/api/session", s.handleSession)
	mux.HandleFunc("/api/timeline", s.handleTimeline)
	mux.HandleFunc("/api/flow", s.handleFlow)
	mux.HandleFunc("/api/output", s.handleOutput)
	mux.HandleFunc("POST /api/observe", s.handleObservePost)
	mux.HandleFunc("GET /api/observe", s.handleObserveGet)

	dist, err := fs.Sub(distFS, "dist")
	if err != nil {
		return fmt.Errorf("embedded fs: %w", err)
	}
	mux.Handle("/", http.FileServer(http.FS(dist)))

	addr := fmt.Sprintf(":%d", s.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", addr, err)
	}

	s.logger.Info("debug UI available", "url", fmt.Sprintf("http://localhost:%d", s.port))

	go func() {
		if err := http.Serve(listener, mux); err != nil {
			s.logger.Error("http server error", "error", err)
		}
	}()

	return nil
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		s.logger.Error("websocket accept failed", "error", err)
		return
	}

	ctx, cancel := context.WithCancel(r.Context())
	s.hub.Add(conn, cancel)
	defer s.hub.Remove(conn)

	s.hub.Broadcast(Event{
		Type: "session",
		Data: s.store.Info(),
	})

	for {
		_, _, err := conn.Read(ctx)
		if err != nil {
			return
		}
	}
}

func (s *Server) handleSession(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.store.Info())
}

func (s *Server) handleTimeline(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.store.Timeline(0, 0))
}

func (s *Server) handleFlow(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.store.Flow())
}

func (s *Server) handleOutput(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.store.Output(0))
}

func (s *Server) handleObservePost(w http.ResponseWriter, r *http.Request) {
	var events []store.ObserveEvent
	if err := json.NewDecoder(r.Body).Decode(&events); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	count := s.store.RecordObservations(events)
	writeJSON(w, map[string]any{"received": count, "observing": s.store.IsObserving()})
}

func (s *Server) handleObserveGet(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.store.ObserveTrace(0, 0))
}

func writeJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(data)
}
