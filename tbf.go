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
	diff := float64(now.Sub(t.ts).Nanoseconds())
	if diff > 0 {
		t.capacity += rate * diff / 1e9

		if t.capacity > burst {
			t.capacity = burst
		}

		t.ts = now
	} else {
		log.Printf("Time not increasing: prev=%+v now=%+v diff=%+v ns", t.ts, now, diff)
	}

	// fulfill request
	if amount <= t.capacity {
		t.capacity -= amount
		return true
	}

	return false
}
