package dap

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	godap "github.com/google/go-dap"
)

type Session struct {
	transport  *Transport
	logger     *slog.Logger

	pending   map[int]chan godap.Message
	pendingMu sync.Mutex

	stopCh chan *godap.StoppedEvent

	termCh chan struct{}

	breakpoints   map[string][]int
	breakpointsMu sync.Mutex

	OnOutput func(category, text string)

	done chan struct{}
}

func NewSession(transport *Transport, logger *slog.Logger) *Session {
	return &Session{
		transport:   transport,
		logger:      logger,
		pending:     make(map[int]chan godap.Message),
		stopCh:      make(chan *godap.StoppedEvent, 1),
		termCh:      make(chan struct{}, 1),
		breakpoints: make(map[string][]int),
		done:        make(chan struct{}),
	}
}

func (s *Session) StartReceiveLoop() {
	go func() {
		defer close(s.done)
		for {
			msg, err := s.transport.Receive()
			if err != nil {
				s.logger.Info("dap receive loop ended", "error", err)
				return
			}
			s.dispatch(msg)
		}
	}()
}

func (s *Session) dispatch(msg godap.Message) {
	switch m := msg.(type) {
	case godap.ResponseMessage:
		resp := m.GetResponse()
		s.pendingMu.Lock()
		ch, ok := s.pending[resp.RequestSeq]
		if ok {
			delete(s.pending, resp.RequestSeq)
		}
		s.pendingMu.Unlock()

		if ok {
			ch <- msg
		} else {
			s.logger.Warn("unmatched response", "request_seq", resp.RequestSeq)
		}

	case *godap.StoppedEvent:
		s.logger.Info("stopped", "reason", m.Body.Reason, "thread", m.Body.ThreadId)
		select {
		case s.stopCh <- m:
		default:
			s.logger.Warn("dropped stopped event, channel full")
		}

	case *godap.TerminatedEvent:
		s.logger.Info("debuggee terminated")
		s.signalTerminated()

	case *godap.ExitedEvent:
		s.logger.Info("debuggee exited", "code", m.Body.ExitCode)
		s.signalTerminated()

	case *godap.InitializedEvent:
		s.logger.Info("debuggee initialized")

	case *godap.OutputEvent:
		s.logger.Info("output", "category", m.Body.Category, "text", m.Body.Output)
		if s.OnOutput != nil {
			s.OnOutput(m.Body.Category, m.Body.Output)
		}

	default:
		s.logger.Debug("unhandled dap message", "type", fmt.Sprintf("%T", msg))
	}
}

func (s *Session) signalTerminated() {
	select {
	case s.termCh <- struct{}{}:
	default:
	}
}

func (s *Session) sendRequest(msg godap.Message, seq int) (godap.Message, error) {
	ch := make(chan godap.Message, 1)
	s.pendingMu.Lock()
	s.pending[seq] = ch
	s.pendingMu.Unlock()

	if err := s.transport.Send(msg); err != nil {
		s.pendingMu.Lock()
		delete(s.pending, seq)
		s.pendingMu.Unlock()
		return nil, fmt.Errorf("send failed: %w", err)
	}

	select {
	case resp := <-ch:
		return s.checkResponse(resp)
	case <-s.done:
		return nil, fmt.Errorf("session closed while waiting for response")
	}
}

func (s *Session) checkResponse(msg godap.Message) (godap.Message, error) {
	if errResp, ok := msg.(*godap.ErrorResponse); ok {
		detail := errResp.Message
		if errResp.Body.Error != nil {
			detail = errResp.Body.Error.Format
		}
		return nil, fmt.Errorf("DAP error (command=%s): %s", errResp.Command, detail)
	}

	if resp, ok := msg.(godap.ResponseMessage); ok {
		r := resp.GetResponse()
		if !r.Success {
			return nil, fmt.Errorf("DAP request failed (command=%s): %s", r.Command, r.Message)
		}
	}

	return msg, nil
}

func (s *Session) Initialize() error {
	seq := s.transport.NextSeq()
	req := &godap.InitializeRequest{
		Request: godap.Request{
			ProtocolMessage: godap.ProtocolMessage{Seq: seq, Type: "request"},
			Command:         "initialize",
		},
		Arguments: godap.InitializeRequestArguments{
			ClientID:                            "orangecandy",
			ClientName:                          "OrangeCandy Debug Server",
			AdapterID:                           "coreclr",
			LinesStartAt1:                       true,
			ColumnsStartAt1:                     true,
			PathFormat:                          "path",
			SupportsVariableType:                true,
			SupportsRunInTerminalRequest:        false,
		},
	}

	_, err := s.sendRequest(req, seq)
	if err != nil {
		return fmt.Errorf("initialize: %w", err)
	}

	return nil
}

func (s *Session) Launch(program string, args []string) error {
	launchArgs := map[string]any{
		"program":     program,
		"stopAtEntry": true,
	}
	if len(args) > 0 {
		launchArgs["args"] = args
	}

	argsJson, err := json.Marshal(launchArgs)
	if err != nil {
		return fmt.Errorf("marshal launch args: %w", err)
	}

	seq := s.transport.NextSeq()
	req := &godap.LaunchRequest{
		Request: godap.Request{
			ProtocolMessage: godap.ProtocolMessage{Seq: seq, Type: "request"},
			Command:         "launch",
		},
		Arguments: argsJson,
	}

	_, err = s.sendRequest(req, seq)
	if err != nil {
		return fmt.Errorf("launch: %w", err)
	}

	return nil
}

func (s *Session) ConfigurationDone() error {
	seq := s.transport.NextSeq()
	req := &godap.ConfigurationDoneRequest{
		Request: godap.Request{
			ProtocolMessage: godap.ProtocolMessage{Seq: seq, Type: "request"},
			Command:         "configurationDone",
		},
	}

	_, err := s.sendRequest(req, seq)
	if err != nil {
		return fmt.Errorf("configurationDone: %w", err)
	}

	return nil
}

func (s *Session) SetBreakpoint(file string, line int) error {
	s.breakpointsMu.Lock()
	lines := s.breakpoints[file]
	for _, l := range lines {
		if l == line {
			s.breakpointsMu.Unlock()
			return nil // already set
		}
	}
	s.breakpoints[file] = append(lines, line)
	allLines := make([]int, len(s.breakpoints[file]))
	copy(allLines, s.breakpoints[file])
	s.breakpointsMu.Unlock()

	return s.sendBreakpoints(file, allLines)
}

func (s *Session) sendBreakpoints(file string, lines []int) error {
	bps := make([]godap.SourceBreakpoint, len(lines))
	for i, l := range lines {
		bps[i] = godap.SourceBreakpoint{Line: l}
	}

	seq := s.transport.NextSeq()
	req := &godap.SetBreakpointsRequest{
		Request: godap.Request{
			ProtocolMessage: godap.ProtocolMessage{Seq: seq, Type: "request"},
			Command:         "setBreakpoints",
		},
		Arguments: godap.SetBreakpointsArguments{
			Source:      godap.Source{Path: file},
			Breakpoints: bps,
		},
	}

	_, err := s.sendRequest(req, seq)
	if err != nil {
		return fmt.Errorf("setBreakpoints: %w", err)
	}

	return nil
}

func (s *Session) Continue(threadId int) (*godap.StoppedEvent, error) {
	seq := s.transport.NextSeq()
	req := &godap.ContinueRequest{
		Request: godap.Request{
			ProtocolMessage: godap.ProtocolMessage{Seq: seq, Type: "request"},
			Command:         "continue",
		},
		Arguments: godap.ContinueArguments{
			ThreadId: threadId,
		},
	}

	_, err := s.sendRequest(req, seq)
	if err != nil {
		return nil, fmt.Errorf("continue: %w", err)
	}

	select {
	case stop := <-s.stopCh:
		return stop, nil
	case <-s.termCh:
		return nil, fmt.Errorf("process terminated")
	case <-s.done:
		return nil, fmt.Errorf("session closed")
	}
}

func (s *Session) StackTrace(threadId int, maxFrames int) ([]godap.StackFrame, error) {
	seq := s.transport.NextSeq()
	req := &godap.StackTraceRequest{
		Request: godap.Request{
			ProtocolMessage: godap.ProtocolMessage{Seq: seq, Type: "request"},
			Command:         "stackTrace",
		},
		Arguments: godap.StackTraceArguments{
			ThreadId: threadId,
			Levels:   maxFrames,
		},
	}

	resp, err := s.sendRequest(req, seq)
	if err != nil {
		return nil, fmt.Errorf("stackTrace: %w", err)
	}

	r, ok := resp.(*godap.StackTraceResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type for stackTrace: %T", resp)
	}
	return r.Body.StackFrames, nil
}

func (s *Session) Scopes(frameId int) ([]godap.Scope, error) {
	seq := s.transport.NextSeq()
	req := &godap.ScopesRequest{
		Request: godap.Request{
			ProtocolMessage: godap.ProtocolMessage{Seq: seq, Type: "request"},
			Command:         "scopes",
		},
		Arguments: godap.ScopesArguments{
			FrameId: frameId,
		},
	}

	resp, err := s.sendRequest(req, seq)
	if err != nil {
		return nil, fmt.Errorf("scopes: %w", err)
	}

	r, ok := resp.(*godap.ScopesResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type for scopes: %T", resp)
	}
	return r.Body.Scopes, nil
}

func (s *Session) Variables(variablesRef int) ([]godap.Variable, error) {
	seq := s.transport.NextSeq()
	req := &godap.VariablesRequest{
		Request: godap.Request{
			ProtocolMessage: godap.ProtocolMessage{Seq: seq, Type: "request"},
			Command:         "variables",
		},
		Arguments: godap.VariablesArguments{
			VariablesReference: variablesRef,
		},
	}

	resp, err := s.sendRequest(req, seq)
	if err != nil {
		return nil, fmt.Errorf("variables: %w", err)
	}

	r, ok := resp.(*godap.VariablesResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type for variables: %T", resp)
	}
	return r.Body.Variables, nil
}

func (s *Session) Locals(frameId int) ([]godap.Variable, error) {
	scopes, err := s.Scopes(frameId)
	if err != nil {
		return nil, err
	}

	var locals []godap.Variable
	for _, scope := range scopes {
		if scope.Expensive {
			continue
		}
		vars, err := s.Variables(scope.VariablesReference)
		if err != nil {
			return nil, err
		}
		locals = append(locals, vars...)
	}

	return locals, nil
}

func (s *Session) Evaluate(expression string, frameId int) (string, string, error) {
	seq := s.transport.NextSeq()
	req := &godap.EvaluateRequest{
		Request: godap.Request{
			ProtocolMessage: godap.ProtocolMessage{Seq: seq, Type: "request"},
			Command:         "evaluate",
		},
		Arguments: godap.EvaluateArguments{
			Expression: expression,
			FrameId:    frameId,
			Context:    "watch",
		},
	}

	resp, err := s.sendRequest(req, seq)
	if err != nil {
		return "", "", fmt.Errorf("evaluate: %w", err)
	}

	r, ok := resp.(*godap.EvaluateResponse)
	if !ok {
		return "", "", fmt.Errorf("unexpected response type for evaluate: %T", resp)
	}

	return r.Body.Result, r.Body.Type, nil
}

func (s *Session) SetExceptionBreakpoints(filters []string) error {
	seq := s.transport.NextSeq()
	req := &godap.SetExceptionBreakpointsRequest{
		Request: godap.Request{
			ProtocolMessage: godap.ProtocolMessage{Seq: seq, Type: "request"},
			Command:         "setExceptionBreakpoints",
		},
		Arguments: godap.SetExceptionBreakpointsArguments{
			Filters: filters,
		},
	}

	_, err := s.sendRequest(req, seq)
	if err != nil {
		return fmt.Errorf("setExceptionBreakpoints: %w", err)
	}

	return nil
}

func (s *Session) ExceptionInfo(threadId int) (*godap.ExceptionInfoResponseBody, error) {
	seq := s.transport.NextSeq()
	req := &godap.ExceptionInfoRequest{
		Request: godap.Request{
			ProtocolMessage: godap.ProtocolMessage{Seq: seq, Type: "request"},
			Command:         "exceptionInfo",
		},
		Arguments: godap.ExceptionInfoArguments{
			ThreadId: threadId,
		},
	}

	resp, err := s.sendRequest(req, seq)
	if err != nil {
		return nil, fmt.Errorf("exceptionInfo: %w", err)
	}

	r, ok := resp.(*godap.ExceptionInfoResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type for exceptionInfo: %T", resp)
	}
	return &r.Body, nil
}

func (s *Session) Disconnect() error {
	seq := s.transport.NextSeq()
	req := &godap.DisconnectRequest{
		Request: godap.Request{
			ProtocolMessage: godap.ProtocolMessage{Seq: seq, Type: "request"},
			Command:         "disconnect",
		},
		Arguments: &godap.DisconnectArguments{
			TerminateDebuggee: true,
		},
	}

	_ = s.transport.Send(req)
	return nil
}

func (s *Session) WaitForStop() (*godap.StoppedEvent, error) {
	select {
	case stop := <-s.stopCh:
		return stop, nil
	case <-s.termCh:
		return nil, fmt.Errorf("process terminated")
	case <-s.done:
		return nil, fmt.Errorf("session closed")
	}
}

func (s *Session) Done() <-chan struct{} {
	return s.done
}
