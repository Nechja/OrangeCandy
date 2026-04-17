package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/Nechja/OrangeCandy/mcp-server/dap"
	"github.com/Nechja/OrangeCandy/mcp-server/snapshot"
	"github.com/Nechja/OrangeCandy/mcp-server/store"
	"github.com/Nechja/OrangeCandy/mcp-server/web"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var version = "dev"

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	client := dap.NewClient(logger)
	snapper := snapshot.NewBuilder(client)
	st := store.New()
	hub := web.NewHub(logger)

	st.OnEvent = func(eventType string, data any) {
		hub.Broadcast(web.Event{Type: eventType, Data: data})
	}

	client.OnOutput = func(category, text string) {
		st.RecordOutput(category, text)
	}

	webServer := web.NewServer(st, hub, logger, 9119)
	if err := webServer.Start(); err != nil {
		slog.Error("web server failed to start", "error", err)
	}

	server := buildServer(client, snapper, st, hub, logger)

	slog.Info("OrangeCandy MCP debug server starting", "version", version)
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
