package service

import (
	"sync"
	"time"
)

type RateLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	entries map[string][]time.Time
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		limit:   limit,
		window:  window,
		entries: make(map[string][]time.Time),
	}
}

func (l *RateLimiter) Allow(key string) bool {
	now := time.Now()
	cutoff := now.Add(-l.window)

	l.mu.Lock()
	defer l.mu.Unlock()

	hits := l.entries[key]
	keep := hits[:0]
	for _, hit := range hits {
		if hit.After(cutoff) {
			keep = append(keep, hit)
		}
	}
	if len(keep) >= l.limit {
		l.entries[key] = keep
		return false
	}
	keep = append(keep, now)
	l.entries[key] = keep

	for entryKey, entryHits := range l.entries {
		if len(entryHits) == 0 || entryHits[len(entryHits)-1].Before(cutoff) {
			delete(l.entries, entryKey)
		}
	}
	return true
}
