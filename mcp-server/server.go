package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/Nechja/OrangeCandy/mcp-server/dap"
	"github.com/Nechja/OrangeCandy/mcp-server/snapshot"
	"github.com/Nechja/OrangeCandy/mcp-server/store"
	"github.com/Nechja/OrangeCandy/mcp-server/web"
	godap "github.com/google/go-dap"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func buildServer(client *dap.Client, snapper *snapshot.Builder, st *store.Store, hub *web.Hub, logger *slog.Logger) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "orangecandy-debug",
		Version: "0.2.0",
	}, nil)

	registerLaunch(server, client, snapper, st, logger)
	registerRestart(server, client, snapper, st, logger)
	registerStop(server, client, st, logger)

	registerSetBreakpoint(server, client, st, logger)
	registerContinue(server, client, snapper, st, logger)

	registerAddWatch(server, client, st, logger)
	registerRemoveWatch(server, st, logger)
	registerGetWatches(server, st, logger)

	registerDiagnose(server, client, snapper, st, logger)

	registerSessionInfo(server, st, logger)
	registerHistory(server, st, logger)
	registerTimeline(server, st, logger)
	registerFlow(server, st, logger)
	registerOutput(server, st, logger)

	registerShowSource(server, st, hub, logger)
	registerShowStop(server, st, hub, logger)

	registerObserveStart(server, st, logger)
	registerObserveStop(server, st, logger)
	registerObserveTrace(server, st, logger)
	registerObserveSearch(server, st, logger)

	return server
}

type launchArgs struct {
	ProjectPath string   `json:"project_path" jsonschema:"Path to the .NET project directory"`
	Args        []string `json:"args,omitempty" jsonschema:"Command line arguments passed to the .NET app"`
}

func registerLaunch(server *mcp.Server, client *dap.Client, snapper *snapshot.Builder, st *store.Store, logger *slog.Logger) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "launch",
		Description: "Launch a .NET project under the debugger. Returns a debug snapshot at the entry point.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args launchArgs) (*mcp.CallToolResult, any, error) {
		logger.Info("tool:launch", "project", args.ProjectPath)

		st.Reset(args.ProjectPath, args.Args)
		st.RecordLifecycle("launch", map[string]any{"project": args.ProjectPath, "args": args.Args})

		if err := client.Launch(args.ProjectPath, args.Args); err != nil {
			st.SetState(store.StateDead)
			return toolError(err), nil, nil
		}

		session := client.Session()
		stop, err := session.WaitForStop()
		if err != nil {
			st.SetState(store.StateDead)
			return toolError(fmt.Errorf("waiting for entry stop: %w", err)), nil, nil
		}

		snap, err := snapper.Capture(stop)
		if err != nil {
			return toolError(fmt.Errorf("snapshot: %w", err)), nil, nil
		}

		record := st.RecordStop(snap, stop.Body.ThreadId)
		evaluateWatches(client, st, record.Index, stop.Body.ThreadId, logger)
		return toolJSON(record), nil, nil
	})
}

type restartArgs struct {
	Args []string `json:"args,omitempty" jsonschema:"Command line arguments for the relaunch"`
}

func registerRestart(server *mcp.Server, client *dap.Client, snapper *snapshot.Builder, st *store.Store, logger *slog.Logger) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "restart",
		Description: "Restart the current debug session. Kills the running process, rebuilds, and relaunches with optional new args. Returns a debug snapshot at the entry point.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args restartArgs) (*mcp.CallToolResult, any, error) {
		logger.Info("tool:restart", "args", args.Args)

		lastProject := client.LastProject()
		if lastProject == "" {
			return toolError(fmt.Errorf("no previous session to restart — use launch first")), nil, nil
		}

		_ = client.Disconnect()
		st.Reset(lastProject, args.Args)
		st.RecordLifecycle("restart", map[string]any{"project": lastProject, "args": args.Args})

		if err := client.Launch(lastProject, args.Args); err != nil {
			st.SetState(store.StateDead)
			return toolError(err), nil, nil
		}

		session := client.Session()
		stop, err := session.WaitForStop()
		if err != nil {
			st.SetState(store.StateDead)
			return toolError(fmt.Errorf("waiting for entry stop: %w", err)), nil, nil
		}

		snap, err := snapper.Capture(stop)
		if err != nil {
			return toolError(fmt.Errorf("snapshot: %w", err)), nil, nil
		}

		record := st.RecordStop(snap, stop.Body.ThreadId)
		evaluateWatches(client, st, record.Index, stop.Body.ThreadId, logger)
		return toolJSON(record), nil, nil
	})
}

type stopArgs struct{}

func registerStop(server *mcp.Server, client *dap.Client, st *store.Store, logger *slog.Logger) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "stop",
		Description: "Stop the debug session and terminate the process.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args stopArgs) (*mcp.CallToolResult, any, error) {
		logger.Info("tool:stop")

		if err := client.Disconnect(); err != nil {
			return toolError(err), nil, nil
		}

		st.SetState(store.StateDead)
		st.RecordLifecycle("terminated", nil)
		return toolJSON(map[string]string{"status": "stopped"}), nil, nil
	})
}

type breakpointArgs struct {
	File string `json:"file" jsonschema:"Source file path"`
	Line int    `json:"line" jsonschema:"Line number"`
}

func registerSetBreakpoint(server *mcp.Server, client *dap.Client, st *store.Store, logger *slog.Logger) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "set_breakpoint",
		Description: "Set a breakpoint at a file and line.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args breakpointArgs) (*mcp.CallToolResult, any, error) {
		logger.Info("tool:set_breakpoint", "file", args.File, "line", args.Line)

		session := client.Session()
		if session == nil {
			return toolError(fmt.Errorf("no active debug session")), nil, nil
		}

		if err := session.SetBreakpoint(args.File, args.Line); err != nil {
			return toolError(err), nil, nil
		}

		st.RecordBreakpoint(args.File, args.Line)
		st.RecordAction("set_breakpoint", map[string]any{"file": args.File, "line": args.Line})

		return toolJSON(map[string]any{
			"status": "set",
			"file":   args.File,
			"line":   args.Line,
		}), nil, nil
	})
}

type continueArgs struct{}

func registerContinue(server *mcp.Server, client *dap.Client, snapper *snapshot.Builder, st *store.Store, logger *slog.Logger) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "continue_execution",
		Description: "Continue execution until next breakpoint, exception, or exit. Returns a full debug snapshot with locals, call stack, and source context.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args continueArgs) (*mcp.CallToolResult, any, error) {
		logger.Info("tool:continue")

		session := client.Session()
		if session == nil {
			return toolError(fmt.Errorf("no active debug session")), nil, nil
		}

		st.SetState(store.StateRunning)
		st.RecordAction("continue", nil)

		stop, err := session.Continue(1)
		if err != nil {
			st.SetState(store.StateDead)
			return toolError(err), nil, nil
		}

		snap, err := snapper.Capture(stop)
		if err != nil {
			return toolError(fmt.Errorf("snapshot: %w", err)), nil, nil
		}

		record := st.RecordStop(snap, stop.Body.ThreadId)
		evaluateWatches(client, st, record.Index, stop.Body.ThreadId, logger)
		return toolJSON(record), nil, nil
	})
}

type addWatchArgs struct {
	Expression string `json:"expression" jsonschema:"C# expression to evaluate at every stop"`
}

func registerAddWatch(server *mcp.Server, client *dap.Client, st *store.Store, logger *slog.Logger) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "add_watch",
		Description: "Add a watch expression that gets evaluated at every debug stop. Returns the watch ID.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args addWatchArgs) (*mcp.CallToolResult, any, error) {
		logger.Info("tool:add_watch", "expression", args.Expression)

		id := st.AddWatch(args.Expression)
		st.RecordAction("add_watch", map[string]any{"expression": args.Expression, "id": id})

		session := client.Session()
		info := st.Info()
		if session != nil && info.Current != nil {
			frames, err := session.StackTrace(info.Current.ThreadId, 1)
			if err == nil && len(frames) > 0 {
				value, typ, err := session.Evaluate(args.Expression, frames[0].Id)
				evalErr := ""
				if err != nil {
					evalErr = err.Error()
				}
				st.UpdateWatch(id, info.Current.Index, value, typ, evalErr)
			}
		}

		return toolJSON(map[string]any{"id": id, "expression": args.Expression}), nil, nil
	})
}

type removeWatchArgs struct {
	Id int `json:"id" jsonschema:"Watch ID to remove"`
}

func registerRemoveWatch(server *mcp.Server, st *store.Store, logger *slog.Logger) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "remove_watch",
		Description: "Remove a watch expression by ID.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args removeWatchArgs) (*mcp.CallToolResult, any, error) {
		logger.Info("tool:remove_watch", "id", args.Id)

		if !st.RemoveWatch(args.Id) {
			return toolError(fmt.Errorf("watch %d not found", args.Id)), nil, nil
		}

		return toolJSON(map[string]string{"status": "removed"}), nil, nil
	})
}

type getWatchesArgs struct{}

func registerGetWatches(server *mcp.Server, st *store.Store, logger *slog.Logger) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_watches",
		Description: "Get all watch expressions and their current values.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args getWatchesArgs) (*mcp.CallToolResult, any, error) {
		logger.Info("tool:get_watches")
		return toolJSON(st.GetWatches()), nil, nil
	})
}

type diagnoseArgs struct{}

func registerDiagnose(server *mcp.Server, client *dap.Client, snapper *snapshot.Builder, st *store.Store, logger *slog.Logger) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "diagnose",
		Description: "When stopped on an exception, answers 'what went wrong?' — returns exception type, message, inner exceptions, locals at the crash site, call stack, and source context. Call this after the debugger stops on an exception.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args diagnoseArgs) (*mcp.CallToolResult, any, error) {
		logger.Info("tool:diagnose")

		session := client.Session()
		if session == nil {
			return toolError(fmt.Errorf("no active debug session")), nil, nil
		}

		info := st.Info()
		if info.Current == nil {
			return toolError(fmt.Errorf("not stopped — nothing to diagnose")), nil, nil
		}

		current := info.Current
		snap := current.Snapshot
		if snap.Reason != "exception" {
			return toolError(fmt.Errorf("not stopped on an exception (stopped for: %s) — diagnose is for crashes", snap.Reason)), nil, nil
		}

		timelineCtx := fmt.Sprintf("exception occurred at event #%d, after %d total stops", info.EventCount-1, info.StopCount)

		fakeStop := &godap.StoppedEvent{}
		fakeStop.Body.Reason = "exception"
		fakeStop.Body.ThreadId = current.ThreadId

		diag, err := snapper.CaptureDiagnosis(fakeStop, timelineCtx)
		if err != nil {
			return toolError(fmt.Errorf("diagnosis failed: %w", err)), nil, nil
		}

		detail := map[string]any{
			"location": fmt.Sprintf("%s:%d", diag.CrashLocation.File, diag.CrashLocation.Line),
		}
		if diag.Exception != nil {
			detail["exception_type"] = diag.Exception.Type
			detail["message"] = diag.Exception.Message
		}
		st.RecordAction("diagnose", detail)

		return toolJSON(diag), nil, nil
	})
}

type sessionInfoArgs struct{}

func registerSessionInfo(server *mcp.Server, st *store.Store, logger *slog.Logger) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "session_info",
		Description: "Get the current debug session state: lifecycle status, current stop point, breakpoints and their hit counts.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args sessionInfoArgs) (*mcp.CallToolResult, any, error) {
		logger.Info("tool:session_info")
		return toolJSON(st.Info()), nil, nil
	})
}

type historyArgs struct {
	Start int `json:"start,omitempty" jsonschema:"Start index (0-based, default 0)"`
	End   int `json:"end,omitempty" jsonschema:"End index (exclusive, default all)"`
}

func registerHistory(server *mcp.Server, st *store.Store, logger *slog.Logger) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "history",
		Description: "Get past debug snapshots. Each entry includes the full snapshot with locals, call stack, and source at that stop point. Use start/end to paginate.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args historyArgs) (*mcp.CallToolResult, any, error) {
		logger.Info("tool:history", "start", args.Start, "end", args.End)
		return toolJSON(st.History(args.Start, args.End)), nil, nil
	})
}

type timelineArgs struct {
	Start int `json:"start,omitempty" jsonschema:"Start index (0-based, default 0)"`
	End   int `json:"end,omitempty" jsonschema:"End index (exclusive, default all)"`
}

func registerTimeline(server *mcp.Server, st *store.Store, logger *slog.Logger) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "timeline",
		Description: "Get the full debug session narrative — every action taken and every stop hit, in order. Shows the complete story: launches, breakpoints set, continues, stops with snapshots, and termination.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args timelineArgs) (*mcp.CallToolResult, any, error) {
		logger.Info("tool:timeline", "start", args.Start, "end", args.End)
		return toolJSON(st.Timeline(args.Start, args.End)), nil, nil
	})
}

type flowArgs struct{}

func registerFlow(server *mcp.Server, st *store.Store, logger *slog.Logger) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "flow",
		Description: "Get a condensed execution timeline showing the path through the code. Each entry is a function name, file, line, and stop reason — a visual trace of the debug session.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args flowArgs) (*mcp.CallToolResult, any, error) {
		logger.Info("tool:flow")
		return toolJSON(st.Flow()), nil, nil
	})
}

type outputArgs struct {
	Last int `json:"last,omitempty" jsonschema:"Number of recent output lines to return (default all)"`
}

func registerOutput(server *mcp.Server, st *store.Store, logger *slog.Logger) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "debug_output",
		Description: "Get captured stdout/stderr output from the debuggee process.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args outputArgs) (*mcp.CallToolResult, any, error) {
		logger.Info("tool:output", "last", args.Last)
		return toolJSON(st.Output(args.Last)), nil, nil
	})
}

type showSourceArgs struct {
	File     string `json:"file" jsonschema:"Absolute or relative file path"`
	Line     int    `json:"line,omitempty" jsonschema:"Center the view on this line (default: 1)"`
	Radius   int    `json:"radius,omitempty" jsonschema:"Lines above and below to show (default: 15)"`
}

type showSourceResult struct {
	File       string   `json:"file"`
	StartLine  int      `json:"start_line"`
	Lines      []string `json:"lines"`
	CenterLine int      `json:"center_line"`
}

func registerShowSource(server *mcp.Server, st *store.Store, hub *web.Hub, logger *slog.Logger) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "show_source",
		Description: "Show source code from a file. The code is displayed in the debug UI for the user to see. Use this to point the user at specific code during discussion or investigation.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args showSourceArgs) (*mcp.CallToolResult, any, error) {
		logger.Info("tool:show_source", "file", args.File, "line", args.Line)

		if args.Line <= 0 { args.Line = 1 }
		if args.Radius <= 0 { args.Radius = 15 }

		lines := readFileLines(args.File, args.Line, args.Radius)
		if lines == nil {
			return toolError(fmt.Errorf("could not read file: %s", args.File)), nil, nil
		}

		startLine := max(1, args.Line - args.Radius)

		result := showSourceResult{
			File:       args.File,
			StartLine:  startLine,
			Lines:      lines,
			CenterLine: args.Line,
		}

		hub.Broadcast(web.Event{Type: "show_source", Data: result})

		st.RecordAction("show_source", map[string]any{
			"file": args.File,
			"line": args.Line,
		})

		return toolJSON(result), nil, nil
	})
}

type showStopArgs struct {
	Index int `json:"index" jsonschema:"Stop index from the history (0-based)"`
}

func registerShowStop(server *mcp.Server, st *store.Store, hub *web.Hub, logger *slog.Logger) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "show_stop",
		Description: "Navigate the debug UI to a specific past stop point. Shows the snapshot (source, locals, call stack) from that moment in the debug session. Use this to point the user at a specific moment in the debug timeline.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args showStopArgs) (*mcp.CallToolResult, any, error) {
		logger.Info("tool:show_stop", "index", args.Index)

		stops := st.History(args.Index, args.Index+1)
		if len(stops) == 0 {
			return toolError(fmt.Errorf("stop #%d not found", args.Index)), nil, nil
		}

		record := stops[0]

		hub.Broadcast(web.Event{Type: "show_stop", Data: record})

		st.RecordAction("show_stop", map[string]any{
			"index":    args.Index,
			"function": record.Snapshot.StoppedAt.Function,
			"file":     record.Snapshot.StoppedAt.File,
			"line":     record.Snapshot.StoppedAt.Line,
		})

		return toolJSON(record), nil, nil
	})
}

type observeStartArgs struct{}

func registerObserveStart(server *mcp.Server, st *store.Store, logger *slog.Logger) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "observe_start",
		Description: "Start observing method calls in the .NET app. The app must have OrangeCandy.Observe installed. Events stream to the UI in real time.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args observeStartArgs) (*mcp.CallToolResult, any, error) {
		logger.Info("tool:observe_start")
		st.StartObserving()
		st.RecordLifecycle("observe_start", nil)
		return toolJSON(map[string]string{"status": "observing"}), nil, nil
	})
}

type observeStopArgs struct{}

func registerObserveStop(server *mcp.Server, st *store.Store, logger *slog.Logger) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "observe_stop",
		Description: "Stop observing method calls. Returns the total number of events captured.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args observeStopArgs) (*mcp.CallToolResult, any, error) {
		logger.Info("tool:observe_stop")
		count := st.StopObserving()
		st.RecordLifecycle("observe_stop", map[string]any{"event_count": count})
		return toolJSON(map[string]any{"status": "stopped", "event_count": count}), nil, nil
	})
}

type observeTraceArgs struct {
	Start int `json:"start,omitempty" jsonschema:"Start index (0-based, default 0)"`
	End   int `json:"end,omitempty" jsonschema:"End index (exclusive, default all)"`
}

func registerObserveTrace(server *mcp.Server, st *store.Store, logger *slog.Logger) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "observe_trace",
		Description: "Get the captured method call trace. Each event shows interface, method, arguments, return value, duration, and call depth.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args observeTraceArgs) (*mcp.CallToolResult, any, error) {
		logger.Info("tool:observe_trace", "start", args.Start, "end", args.End)
		return toolJSON(st.ObserveTrace(args.Start, args.End)), nil, nil
	})
}

type observeSearchArgs struct {
	Method string `json:"method" jsonschema:"Method or interface name to search for"`
}

func registerObserveSearch(server *mcp.Server, st *store.Store, logger *slog.Logger) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "observe_search",
		Description: "Search the observation trace for calls to a specific method or interface.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args observeSearchArgs) (*mcp.CallToolResult, any, error) {
		logger.Info("tool:observe_search", "method", args.Method)
		return toolJSON(st.ObserveSearch(args.Method)), nil, nil
	})
}

func readFileLines(file string, centerLine int, radius int) []string {
	f, err := os.Open(file)
	if err != nil {
		return nil
	}
	defer f.Close()

	startLine := max(1, centerLine - radius)
	endLine := centerLine + radius

	var lines []string
	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		if lineNum >= startLine && lineNum <= endLine {
			lines = append(lines, scanner.Text())
		}
		if lineNum > endLine {
			break
		}
	}
	return lines
}

func evaluateWatches(client *dap.Client, st *store.Store, stopIndex int, threadId int, logger *slog.Logger) {
	session := client.Session()
	if session == nil {
		return
	}

	watches := st.WatchExpressions()
	if len(watches) == 0 {
		return
	}

	frames, err := session.StackTrace(threadId, 1)
	if err != nil || len(frames) == 0 {
		return
	}

	frameId := frames[0].Id
	for _, w := range watches {
		value, typ, err := session.Evaluate(w.Expression, frameId)
		evalErr := ""
		if err != nil {
			evalErr = err.Error()
			value = ""
			typ = ""
		}
		st.UpdateWatch(w.Id, stopIndex, value, typ, evalErr)
	}
}

func toolJSON(v any) *mcp.CallToolResult {
	data, _ := json.Marshal(v)
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(data)},
		},
	}
}

func toolError(err error) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Error: %s", err.Error())},
		},
		IsError: true,
	}
}
