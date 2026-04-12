package store

import "time"

type Watch struct {
	Id         int          `json:"id"`
	Expression string       `json:"expression"`
	Current    *WatchValue  `json:"current,omitempty"`
	History    []WatchValue `json:"history,omitempty"`
}

type WatchValue struct {
	StopIndex int       `json:"stop_index"`
	Timestamp time.Time `json:"timestamp"`
	Value     string    `json:"value"`
	Type      string    `json:"type"`
	Error     string    `json:"error,omitempty"`
}

func (s *Store) AddWatch(expression string) int {
	s.mu.Lock()

	id := len(s.watches) + 1
	s.watches = append(s.watches, Watch{
		Id:         id,
		Expression: expression,
	})

	watches := make([]Watch, len(s.watches))
	copy(watches, s.watches)
	s.mu.Unlock()

	s.emit("watches", watches)
	return id
}

func (s *Store) RemoveWatch(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, w := range s.watches {
		if w.Id == id {
			s.watches = append(s.watches[:i], s.watches[i+1:]...)
			return true
		}
	}
	return false
}

func (s *Store) GetWatches() []Watch {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Watch, len(s.watches))
	copy(result, s.watches)
	return result
}

func (s *Store) UpdateWatch(id int, stopIndex int, value, typ, evalErr string) {
	s.mu.Lock()

	for i := range s.watches {
		if s.watches[i].Id == id {
			wv := WatchValue{
				StopIndex: stopIndex,
				Timestamp: time.Now(),
				Value:     value,
				Type:      typ,
				Error:     evalErr,
			}
			s.watches[i].Current = &wv
			s.watches[i].History = append(s.watches[i].History, wv)
			break
		}
	}

	watches := make([]Watch, len(s.watches))
	copy(watches, s.watches)
	s.mu.Unlock()

	s.emit("watches", watches)
}

func (s *Store) WatchExpressions() []struct {
	Id         int
	Expression string
} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]struct {
		Id         int
		Expression string
	}, len(s.watches))
	for i, w := range s.watches {
		result[i].Id = w.Id
		result[i].Expression = w.Expression
	}
	return result
}
