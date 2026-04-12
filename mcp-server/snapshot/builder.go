package snapshot

import (
	"bufio"
	"os"

	"github.com/Nechja/OrangeCandy/mcp-server/dap"

	godap "github.com/google/go-dap"
)

type Snapshot struct {
	StoppedAt   Location `json:"stopped_at"`
	SourceLines []string `json:"source_window"`
	Locals      []Local  `json:"locals"`
	CallStack   []Frame  `json:"call_stack"`
	Reason      string   `json:"reason"`
}

type Location struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Function string `json:"function"`
}

type Local struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

type Frame struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

type Builder struct {
	client *dap.Client
}

func NewBuilder(client *dap.Client) *Builder {
	return &Builder{client: client}
}

func (b *Builder) Capture(stop *godap.StoppedEvent) (*Snapshot, error) {
	session := b.client.Session()
	if session == nil {
		return nil, nil
	}

	threadId := stop.Body.ThreadId

	frames, err := session.StackTrace(threadId, 5)
	if err != nil {
		return nil, err
	}

	snap := &Snapshot{
		Reason:    stop.Body.Reason,
		CallStack: make([]Frame, 0, len(frames)),
	}

	if len(frames) > 0 {
		top := frames[0]
		file := ""
		if top.Source != nil {
			file = top.Source.Path
		}
		snap.StoppedAt = Location{
			File:     file,
			Line:     top.Line,
			Function: top.Name,
		}

		locals, err := session.Locals(top.Id)
		if err == nil {
			snap.Locals = make([]Local, len(locals))
			for i, v := range locals {
				snap.Locals[i] = Local{Name: v.Name, Type: v.Type, Value: v.Value}
			}
		}

		if file != "" {
			snap.SourceLines = readSourceWindow(file, top.Line, 5)
		}
	}

	for _, f := range frames {
		file := ""
		if f.Source != nil {
			file = f.Source.Path
		}
		snap.CallStack = append(snap.CallStack, Frame{
			Function: f.Name,
			File:     file,
			Line:     f.Line,
		})
	}

	return snap, nil
}

func (b *Builder) CaptureDiagnosis(stop *godap.StoppedEvent, timelineContext string) (*Diagnosis, error) {
	session := b.client.Session()
	if session == nil {
		return nil, nil
	}

	threadId := stop.Body.ThreadId
	diag := &Diagnosis{
		TimelineContext: timelineContext,
	}

	exInfo, err := session.ExceptionInfo(threadId)
	if err == nil && exInfo != nil {
		diag.Exception = &ExceptionInfo{
			Type:    exInfo.ExceptionId,
			Message: exInfo.Description,
		}
		if exInfo.Details != nil {
			diag.Exception.Message = exInfo.Details.Message
			diag.Exception.Type = exInfo.Details.TypeName
			diag.Exception.StackTrace = exInfo.Details.StackTrace
			diag.Exception.Inner = convertInnerExceptions(exInfo.Details.InnerException)
		}
	}

	frames, err := session.StackTrace(threadId, 10)
	if err == nil && len(frames) > 0 {
		top := frames[0]
		file := ""
		if top.Source != nil {
			file = top.Source.Path
		}
		diag.CrashLocation = Location{
			File:     file,
			Line:     top.Line,
			Function: top.Name,
		}

		if file != "" {
			diag.SourceLines = readSourceWindow(file, top.Line, 5)
		}

		locals, err := session.Locals(top.Id)
		if err == nil {
			diag.Locals = make([]Local, len(locals))
			for i, v := range locals {
				diag.Locals[i] = Local{Name: v.Name, Type: v.Type, Value: v.Value}
			}
		}

		for _, f := range frames {
			file := ""
			if f.Source != nil {
				file = f.Source.Path
			}
			diag.CallStack = append(diag.CallStack, Frame{
				Function: f.Name,
				File:     file,
				Line:     f.Line,
			})
		}
	}

	return diag, nil
}

func convertInnerExceptions(inner []godap.ExceptionDetails) []ExceptionInfo {
	if len(inner) == 0 {
		return nil
	}
	result := make([]ExceptionInfo, len(inner))
	for i, e := range inner {
		result[i] = ExceptionInfo{
			Type:       e.TypeName,
			Message:    e.Message,
			StackTrace: e.StackTrace,
			Inner:      convertInnerExceptions(e.InnerException),
		}
	}
	return result
}

func readSourceWindow(file string, line int, radius int) []string {
	f, err := os.Open(file)
	if err != nil {
		return nil
	}
	defer f.Close()

	startLine := max(1, line-radius)
	endLine := line + radius

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
