package main

import (
	"errors"
	"sync"
	"time"
)

type router struct {
	services []service
	tb       []*tokenBucket
	current  int
	sync.Mutex
}

func newServiceRouter(config []serviceConfig) *router {
	r := router{
		services: make([]service, 0, len(config)),
		tb:       make([]*tokenBucket, 0, len(config)),
	}
	for _, entry := range config {
		r.services = append(r.services, service{entry.URL, entry.ReplyPath})
		burst := entry.Burst
		if burst == 0 {
			burst = entry.Rate
		}
		tb := newBucket(entry.Rate, time.Second*time.Duration(entry.Period), burst)
		r.tb = append(r.tb, tb)
	}
	return &r
}

func (r *router) get(now time.Time) (service, error) {
	r.Lock()
	for i := 0; i < len(r.tb); i++ {
		selected := (r.current + i) % len(r.tb)
		tb := r.tb[selected]
		tb.fill(now)
		if tb.consume() {
			r.current = selected
			svc := r.services[selected]
			r.Unlock()
			return svc, nil
		}
	}
	r.Unlock()
	return service{}, errors.New("no service available")
}
