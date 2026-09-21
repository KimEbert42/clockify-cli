package api

import (
	"time"
)

type RetryPolicy struct {
	MaxRetries  int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	RetryableFn func(statusCode int, err error) bool
}

var DefaultRetryPolicy = RetryPolicy{
	MaxRetries: 5,
	BaseDelay:  1 * time.Second,
	MaxDelay:   30 * time.Second,
	RetryableFn: func(statusCode int, err error) bool {
		if err != nil {
			return true
		}
		if statusCode == 429 || statusCode >= 500 {
			return true
		}
		return false
	},
}

func (p RetryPolicy) Delay(attempt int) time.Duration {
	delay := p.BaseDelay * time.Duration(1<<uint(attempt))
	if delay > p.MaxDelay {
		delay = p.MaxDelay
	}
	return delay
}
