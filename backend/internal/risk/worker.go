package risk

import (
	"sync"
	"time"
)

// Throttle ensures per-trip risk evaluation runs at most once per interval.
type Throttle struct {
	mu   sync.Mutex
	last map[int64]time.Time
	min  time.Duration
}

func NewThrottle(min time.Duration) *Throttle {
	if min <= 0 {
		min = time.Second
	}
	return &Throttle{last: map[int64]time.Time{}, min: min}
}

func (t *Throttle) Allow(tripID int64, now time.Time) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	prev, ok := t.last[tripID]
	if ok && now.Sub(prev) < t.min {
		return false
	}
	t.last[tripID] = now
	return true
}
