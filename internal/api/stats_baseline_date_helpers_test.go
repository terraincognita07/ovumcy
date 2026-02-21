package api

import (
	"testing"
	"time"
)

func mustParseBaselineDay(t *testing.T, raw string) time.Time {
	t.Helper()

	parsed, err := time.ParseInLocation("2006-01-02", raw, time.UTC)
	if err != nil {
		t.Fatalf("parse baseline day %q: %v", raw, err)
	}
	return parsed
}
