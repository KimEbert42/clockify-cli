package api

import (
	"sync"
	"time"
)

type RateLimiter struct {
	mu       sync.Mutex
	tokens   int
	maxBurst int
	lastTime time.Duration
	rate     time.Duration
}

func NewRateLimiter(requestsPerSecond int) *RateLimiter {
	return &RateLimiter{
		tokens:   requestsPerSecond,
		maxBurst: requestsPerSecond,
		rate:     time.Second / time.Duration(requestsPerSecond),
	}
}

func (rl *RateLimiter) Take() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if rl.tokens > 0 {
		rl.tokens--
		return
	}

	time.Sleep(rl.rate)
	rl.tokens = rl.maxBurst - 1
}
