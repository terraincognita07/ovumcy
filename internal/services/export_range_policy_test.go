package services

import (
	"errors"
	"testing"
	"time"
)

func TestParseExportRange(t *testing.T) {
	location := time.UTC

	t.Run("empty range", func(t *testing.T) {
		from, to, err := ParseExportRange("", "", location)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if from != nil || to != nil {
			t.Fatalf("expected nil from/to, got from=%v to=%v", from, to)
		}
	})

	t.Run("valid from and to", func(t *testing.T) {
		from, to, err := ParseExportRange("2026-02-10", "2026-02-20", location)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if from == nil || to == nil {
			t.Fatalf("expected non-nil range bounds")
		}
		if from.Format("2006-01-02") != "2026-02-10" || to.Format("2006-01-02") != "2026-02-20" {
			t.Fatalf("unexpected range: from=%s to=%s", from.Format("2006-01-02"), to.Format("2006-01-02"))
		}
	})

	t.Run("invalid from", func(t *testing.T) {
		_, _, err := ParseExportRange("not-a-date", "2026-02-20", location)
		if !errors.Is(err, ErrExportFromDateInvalid) {
			t.Fatalf("expected ErrExportFromDateInvalid, got %v", err)
		}
	})

	t.Run("invalid to", func(t *testing.T) {
		_, _, err := ParseExportRange("2026-02-10", "not-a-date", location)
		if !errors.Is(err, ErrExportToDateInvalid) {
			t.Fatalf("expected ErrExportToDateInvalid, got %v", err)
		}
	})

	t.Run("invalid range order", func(t *testing.T) {
		_, _, err := ParseExportRange("2026-02-20", "2026-02-10", location)
		if !errors.Is(err, ErrExportRangeInvalid) {
			t.Fatalf("expected ErrExportRangeInvalid, got %v", err)
		}
	})
}
