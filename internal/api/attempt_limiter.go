package api

import (
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

type attemptLimiter struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
}

func newAttemptLimiter() *attemptLimiter {
	return &attemptLimiter{
		attempts: make(map[string][]time.Time),
	}
}

func (limiter *attemptLimiter) tooManyRecent(key string, now time.Time, limit int, window time.Duration) bool {
	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	pruned := limiter.pruneLocked(key, now, window)
	return len(pruned) >= limit
}

func (limiter *attemptLimiter) addFailure(key string, now time.Time, window time.Duration) {
	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	pruned := limiter.pruneLocked(key, now, window)
	pruned = append(pruned, now)
	limiter.attempts[key] = pruned
}

func (limiter *attemptLimiter) reset(key string) {
	limiter.mu.Lock()
	defer limiter.mu.Unlock()
	delete(limiter.attempts, key)
}

func (limiter *attemptLimiter) pruneLocked(key string, now time.Time, window time.Duration) []time.Time {
	values := limiter.attempts[key]
	if len(values) == 0 {
		return []time.Time{}
	}

	threshold := now.Add(-window)
	pruned := make([]time.Time, 0, len(values))
	for _, value := range values {
		if value.After(threshold) {
			pruned = append(pruned, value)
		}
	}

	if len(pruned) == 0 {
		delete(limiter.attempts, key)
		return []time.Time{}
	}

	limiter.attempts[key] = pruned
	return pruned
}

func requestLimiterKey(c *fiber.Ctx) string {
	key := strings.TrimSpace(c.IP())
	if key == "" {
		return "unknown"
	}
	return key
}
