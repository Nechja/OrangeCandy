package store

import (
	"strings"
	"time"
)

type ObserveEvent struct {
	TraceId     string   `json:"trace_id"`
	EventType   string   `json:"event_type"`
	Interface   string   `json:"interface"`
	Method      string   `json:"method"`
	Arguments   []string `json:"arguments,omitempty"`
	ReturnValue string   `json:"return_value,omitempty"`
	Exception   string   `json:"exception,omitempty"`
	DurationMs  int64    `json:"duration_ms,omitempty"`
	Depth       int      `json:"depth"`
	Timestamp   time.Time `json:"timestamp"`
}

func (s *Store) StartObserving() {
	s.mu.Lock()
	s.observing = true
	s.observations = nil
	s.mu.Unlock()

	s.emit("observe_status", map[string]any{"observing": true})
}

func (s *Store) StopObserving() int {
	s.mu.Lock()
	s.observing = false
	count := len(s.observations)
	s.mu.Unlock()

	s.emit("observe_status", map[string]any{"observing": false, "count": count})
	return count
}

func (s *Store) IsObserving() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.observing
}

func (s *Store) RecordObservation(evt ObserveEvent) {
	s.mu.Lock()
	if !s.observing {
		s.mu.Unlock()
		return
	}
	s.observations = append(s.observations, evt)
	s.mu.Unlock()

	s.emit("observe", evt)
}

func (s *Store) RecordObservations(events []ObserveEvent) int {
	s.mu.Lock()
	if !s.observing {
		s.mu.Unlock()
		return 0
	}
	s.observations = append(s.observations, events...)
	s.mu.Unlock()

	for _, evt := range events {
		s.emit("observe", evt)
	}
	return len(events)
}

func (s *Store) ObserveTrace(start, end int) []ObserveEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if start < 0 {
		start = 0
	}
	if end <= 0 || end > len(s.observations) {
		end = len(s.observations)
	}
	if start >= end {
		return nil
	}

	result := make([]ObserveEvent, end-start)
	copy(result, s.observations[start:end])
	return result
}

func (s *Store) ObserveSearch(method string) []ObserveEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	needle := strings.ToLower(method)
	var results []ObserveEvent
	for _, evt := range s.observations {
		if strings.Contains(strings.ToLower(evt.Method), needle) ||
			strings.Contains(strings.ToLower(evt.Interface), needle) {
			results = append(results, evt)
		}
	}
	return results
}
