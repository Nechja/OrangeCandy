package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/Nechja/OrangeCandy/mcp-server/dap"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	client := dap.NewClient(logger)

	projectPath := os.Args[1]
	fmt.Fprintf(os.Stderr, "=== Smoke test: launching %s ===\n", projectPath)

	if err := client.Launch(projectPath, nil); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "=== Launch succeeded, session is live ===\n")

	session := client.Session()

	// Wait for the stopAtEntry event
	fmt.Fprintf(os.Stderr, "=== Waiting for entry stop... ===\n")
	stop, err := session.WaitForStop()
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL waiting for stop: %v\n", err)
		client.Disconnect()
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "=== Stopped: reason=%s thread=%d ===\n", stop.Body.Reason, stop.Body.ThreadId)

	// Get stack trace
	frames, err := session.StackTrace(stop.Body.ThreadId, 5)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL stack trace: %v\n", err)
	} else {
		for i, f := range frames {
			file := ""
			if f.Source != nil {
				file = f.Source.Path
			}
			fmt.Fprintf(os.Stderr, "  [%d] %s at %s:%d\n", i, f.Name, file, f.Line)
		}
	}

	// Get locals from top frame
	if len(frames) > 0 {
		locals, err := session.Locals(frames[0].Id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "FAIL locals: %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "=== Locals (%d) ===\n", len(locals))
			for _, v := range locals {
				fmt.Fprintf(os.Stderr, "  %s (%s) = %s\n", v.Name, v.Type, v.Value)
			}
		}
	}

	fmt.Fprintf(os.Stderr, "=== Disconnecting ===\n")
	client.Disconnect()
	fmt.Fprintf(os.Stderr, "=== PASS ===\n")
}
