package ratelimit

import (
	"context"
	"math"
	"time"
)

// Limiter allows a burst of request during the defined duration
type Limiter struct {
	maxCount uint
	count    uint
	ticker   *time.Ticker
	tokens   chan struct{}
	ctx      context.Context
}

func (limiter *Limiter) run() {
	for {
		if limiter.count == 0 {
			<-limiter.ticker.C
			limiter.count = limiter.maxCount
		}

		select {
		case <-limiter.ctx.Done():
			limiter.ticker.Stop()
			return
		case limiter.tokens <- struct{}{}:
			limiter.count--
		case <-limiter.ticker.C:
			limiter.count = limiter.maxCount
		}
	}
}

// Take one token from the bucket
func (rateLimiter *Limiter) Take() {
	<-rateLimiter.tokens
}

// New creates a new limiter instance with the tokens amount and the interval
func New(ctx context.Context, max uint, duration time.Duration) *Limiter {
	limiter := &Limiter{
		maxCount: uint(max),
		count:    uint(max),
		ticker:   time.NewTicker(duration),
		tokens:   make(chan struct{}),
		ctx:      ctx,
	}
	go limiter.run()

	return limiter
}

// NewUnlimited create a bucket with approximated unlimited tokens
func NewUnlimited(ctx context.Context) *Limiter {
	limiter := &Limiter{
		maxCount: math.MaxUint,
		count:    math.MaxUint,
		ticker:   time.NewTicker(time.Millisecond),
		tokens:   make(chan struct{}),
		ctx:      ctx,
	}
	go limiter.run()

	return limiter
}
