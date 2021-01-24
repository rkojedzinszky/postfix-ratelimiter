package main

import (
	"log"
	"sync"
	"time"
)

type tbf struct {
	mu       sync.Mutex
	ts       time.Time
	capacity float64
}

func (t *tbf) get(rate, burst, amount float64) bool {
	now := time.Now()

	t.mu.Lock()
	defer t.mu.Unlock()

	// replenish token-bucket
	diff := now.Sub(t.ts).Seconds()
	if diff < 0 {
		log.Printf("Time went backwards: from %+v to %+v, diff=%+v", t.ts, now, diff)
	}

	t.capacity += rate * diff

	// Fixup for negative capacity
	if t.capacity < 0 {
		t.capacity = 0
	}

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
