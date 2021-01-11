package main

import (
	"sync"
	"time"
)

type tbf struct {
	mu       sync.Mutex
	ts       time.Time
	capacity float64
}

func (t *tbf) get(rate, burst, amount float64) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	// replenish token-bucket
	now := time.Now()
	t.capacity += rate * now.Sub(t.ts).Seconds()
	if t.capacity > burst {
		t.capacity = burst
	}
	t.ts = now

	// fulfill request
	if amount <= t.capacity {
		t.capacity -= amount
		return true
	}

	return false
}
