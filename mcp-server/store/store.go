package store

import (
	"fmt"
	"sync"
	"time"

	"github.com/Nechja/OrangeCandy/mcp-server/snapshot"
)

type StopRecord struct {
	Index     int                `json:"index"`
	Timestamp time.Time          `json:"timestamp"`
	Snapshot  *snapshot.Snapshot `json:"snapshot"`
	ThreadId  int                `json:"thread_id"`
}

type Breakpoint struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	HitCount int    `json:"hit_count"`
}

type OutputLine struct {
	Timestamp time.Time `json:"timestamp"`
	Category  string    `json:"category"`
	Text      string    `json:"text"`
}

type TimelineEntry struct {
	Index     int            `json:"index"`
	Timestamp time.Time      `json:"timestamp"`
	Type      string         `json:"type"` // "launch", "restart", "stop", "action", "terminated", "output"
	Tool      string         `json:"tool,omitempty"`
	Detail    map[string]any `json:"detail,omitempty"`
	Snapshot  *snapshot.Snapshot `json:"snapshot,omitempty"`
}

type SessionState string

const (
	StateIdle     SessionState = "idle"
	StateLaunched SessionState = "launched"
	StateStopped  SessionState = "stopped_at_breakpoint"
	StateRunning  SessionState = "running"
	StateDead     SessionState = "terminated"
)

type SessionInfo struct {
	State       SessionState `json:"state"`
	ProjectPath string       `json:"project_path,omitempty"`
	Args        []string     `json:"args,omitempty"`
	StopCount   int          `json:"stop_count"`
	EventCount  int          `json:"event_count"`
	Current     *StopRecord  `json:"current,omitempty"`
	Breakpoints []Breakpoint `json:"breakpoints"`
}

type Store struct {
	mu          sync.RWMutex
	state       SessionState
	projectPath string
	args        []string
	stops       []StopRecord
	timeline    []TimelineEntry
	breakpoints map[string]*Breakpoint
	output      []OutputLine
	watches     []Watch

	OnEvent func(eventType string, data any)
}

func New() *Store {
	return &Store{
		state:       StateIdle,
		breakpoints: make(map[string]*Breakpoint),
	}
}

func (s *Store) emit(eventType string, data any) {
	if s.OnEvent != nil {
		s.OnEvent(eventType, data)
	}
}

func (s *Store) Reset(projectPath string, args []string) {
	s.mu.Lock()
	s.state = StateLaunched
	s.projectPath = projectPath
	s.args = args
	s.stops = nil
	s.timeline = nil
	s.breakpoints = make(map[string]*Breakpoint)
	s.output = nil
	s.mu.Unlock()

	s.emit("session", s.Info())
}

func (s *Store) SetState(state SessionState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = state
}

func (s *Store) RecordAction(tool string, detail map[string]any) {
	s.mu.Lock()
	entry := TimelineEntry{
		Index:     len(s.timeline),
		Timestamp: time.Now(),
		Type:      "action",
		Tool:      tool,
		Detail:    detail,
	}
	s.timeline = append(s.timeline, entry)
	s.mu.Unlock()

	s.emit("timeline", entry)
}

func (s *Store) RecordLifecycle(eventType string, detail map[string]any) {
	s.mu.Lock()
	entry := TimelineEntry{
		Index:     len(s.timeline),
		Timestamp: time.Now(),
		Type:      eventType,
		Detail:    detail,
	}
	s.timeline = append(s.timeline, entry)
	s.mu.Unlock()

	s.emit("timeline", entry)
}

func (s *Store) RecordStop(snap *snapshot.Snapshot, threadId int) *StopRecord {
	s.mu.Lock()

	s.state = StateStopped

	record := StopRecord{
		Index:     len(s.stops),
		Timestamp: time.Now(),
		Snapshot:  snap,
		ThreadId:  threadId,
	}
	s.stops = append(s.stops, record)

	entry := TimelineEntry{
		Index:     len(s.timeline),
		Timestamp: record.Timestamp,
		Type:      "stop",
		Snapshot:  snap,
		Detail: map[string]any{
			"reason":   snap.Reason,
			"function": snap.StoppedAt.Function,
			"file":     snap.StoppedAt.File,
			"line":     snap.StoppedAt.Line,
		},
	}
	s.timeline = append(s.timeline, entry)

	if snap.Reason == "breakpoint" {
		key := bpKey(snap.StoppedAt.File, snap.StoppedAt.Line)
		if bp, ok := s.breakpoints[key]; ok {
			bp.HitCount++
		}
	}

	s.mu.Unlock()

	s.emit("stop", record)
	s.emit("timeline", entry)

	return &record
}

func (s *Store) RecordBreakpoint(file string, line int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := bpKey(file, line)
	if _, exists := s.breakpoints[key]; !exists {
		s.breakpoints[key] = &Breakpoint{
			File: file,
			Line: line,
		}
	}
}

func (s *Store) RecordOutput(category, text string) {
	s.mu.Lock()
	line := OutputLine{
		Timestamp: time.Now(),
		Category:  category,
		Text:      text,
	}
	s.output = append(s.output, line)
	s.mu.Unlock()

	s.emit("output", line)
}

func (s *Store) Info() SessionInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	info := SessionInfo{
		State:       s.state,
		ProjectPath: s.projectPath,
		Args:        s.args,
		StopCount:   len(s.stops),
		EventCount:  len(s.timeline),
		Breakpoints: make([]Breakpoint, 0, len(s.breakpoints)),
	}

	if len(s.stops) > 0 {
		last := s.stops[len(s.stops)-1]
		info.Current = &last
	}

	for _, bp := range s.breakpoints {
		info.Breakpoints = append(info.Breakpoints, *bp)
	}

	return info
}

func (s *Store) History(start, end int) []StopRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if start < 0 {
		start = 0
	}
	if end <= 0 || end > len(s.stops) {
		end = len(s.stops)
	}
	if start >= end {
		return nil
	}

	result := make([]StopRecord, end-start)
	copy(result, s.stops[start:end])
	return result
}

func (s *Store) Timeline(start, end int) []TimelineEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if start < 0 {
		start = 0
	}
	if end <= 0 || end > len(s.timeline) {
		end = len(s.timeline)
	}
	if start >= end {
		return nil
	}

	result := make([]TimelineEntry, end-start)
	copy(result, s.timeline[start:end])
	return result
}

func (s *Store) Output(last int) []OutputLine {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if last <= 0 || last > len(s.output) {
		last = len(s.output)
	}

	start := len(s.output) - last
	result := make([]OutputLine, last)
	copy(result, s.output[start:])
	return result
}

func (s *Store) Flow() []FlowEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries := make([]FlowEntry, len(s.stops))
	for i, stop := range s.stops {
		entries[i] = FlowEntry{
			Index:    stop.Index,
			Function: stop.Snapshot.StoppedAt.Function,
			File:     stop.Snapshot.StoppedAt.File,
			Line:     stop.Snapshot.StoppedAt.Line,
			Reason:   stop.Snapshot.Reason,
		}
	}
	return entries
}

type FlowEntry struct {
	Index    int    `json:"index"`
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Reason   string `json:"reason"`
}

func bpKey(file string, line int) string {
	return fmt.Sprintf("%s:%d", file, line)
}
