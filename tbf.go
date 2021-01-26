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
	newcap := t.capacity + float64(now.Sub(t.ts).Nanoseconds())*rate/1e9
	if newcap > t.capacity {
		t.capacity = newcap

		t.ts = now
	} else {
		log.Printf("Capacity not increasing: prev=(%+v, %+v) now=(%+v, %+v)", t.ts, t.capacity, now, newcap)
	}

	if t.capacity > burst {
		t.capacity = burst
	}

	// fulfill request
	if amount <= t.capacity {
		t.capacity -= amount
		return true
	}

	return false
}
