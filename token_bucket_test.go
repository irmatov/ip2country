package main

import (
	"testing"
	"time"
)

func TestTokenBucket(t *testing.T) {
	now := time.Now()
	b := tokenBucket{float64(100 * time.Millisecond), 50, 0, now}
	if b.consume() {
		t.Fatal("consume from empty bucket succeeds")
	}
	// skip a second. we should have 10 tokens available
	now = now.Add(time.Second)
	b.fill(now)
	for i := 0; i < 10; i++ {
		if !b.consume() {
			t.Fatal("failed to consume enough tokens")
		}
	}
	if b.consume() {
		t.Fatal("consume from empty bucket succeeds")
	}
	// skip two seconds, we should have 20 tokens
	now = now.Add(time.Second * 2)
	b.fill(now)
	for i := 0; i < 20; i++ {
		if !b.consume() {
			t.Fatal("failed to consume enough tokens")
		}
	}
	if b.consume() {
		t.Fatal("consume from empty bucket succeeds")
	}
	// skip 10 seconds, we should have 50 tokens (capped at burst)
	now = now.Add(time.Second * 10)
	b.fill(now)
	for i := 0; i < 50; i++ {
		if !b.consume() {
			t.Fatal("failed to consume enough tokens")
		}
	}
	if b.consume() {
		t.Fatal("consume from empty bucket succeeds")
	}
}
