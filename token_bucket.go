package main

import "time"

type tokenBucket struct {
	step     float64
	burst    float64
	capacity float64
	lastFill time.Time
}

func newBucket(rate int, unit time.Duration, burst int) *tokenBucket {
	return &tokenBucket{float64(unit) / float64(rate), float64(burst), float64(burst), time.Now()}
}

func (b *tokenBucket) fill(now time.Time) {
	tokens := float64(now.Sub(b.lastFill)) / b.step
	if tokens > 0 {
		b.capacity += tokens
		b.lastFill = now
		if b.capacity > b.burst {
			b.capacity = b.burst
		}
	}
}

func (b *tokenBucket) consume() bool {
	if b.capacity > 0 {
		b.capacity--
		return true
	}
	return false
}
