package tracker

import (
	"fmt"
	"sync"
)

type Tracker struct {
	total   int
	current int
	mu      sync.RWMutex
}

func NewTracker() *Tracker {
	return &Tracker{
		total:   0,
		current: 0,
		mu:      sync.RWMutex{},
	}
}

func (t *Tracker) AddTotal(n int) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if n <= 0 {
		return fmt.Errorf("n must be > 0")
	}

	t.total += n
	return nil
}

func (t *Tracker) AddCurrent() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.current++
}

func (t *Tracker) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.total = 0
	t.current = 0
}

func (t *Tracker) State() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return fmt.Sprintf("scraping progres [%d/%d]", t.current, t.total)
}
