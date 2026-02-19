package api

import (
	"testing"
	"time"
)

func TestAttemptLimiterWindowAndReset(t *testing.T) {
	t.Parallel()

	limiter := newAttemptLimiter()
	key := "127.0.0.1"
	window := time.Hour
	now := time.Now().UTC()

	limiter.addFailure(key, now.Add(-2*time.Hour), window)
	if limiter.tooManyRecent(key, now, 1, window) {
		t.Fatal("expected old attempt to be pruned from active window")
	}

	limiter.addFailure(key, now.Add(-30*time.Minute), window)
	if !limiter.tooManyRecent(key, now, 1, window) {
		t.Fatal("expected one recent attempt to hit limit 1")
	}

	limiter.reset(key)
	if limiter.tooManyRecent(key, now, 1, window) {
		t.Fatal("expected no attempts after reset")
	}
}
