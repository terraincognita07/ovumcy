package api

import (
	"testing"
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
)

func TestDayHasData(t *testing.T) {
	t.Parallel()

	if dayHasData(models.DailyLog{}) {
		t.Fatal("empty day log should not have data")
	}
	if !dayHasData(models.DailyLog{IsPeriod: true}) {
		t.Fatal("period day should have data")
	}
	if !dayHasData(models.DailyLog{SymptomIDs: []uint{1}}) {
		t.Fatal("symptom day should have data")
	}
	if !dayHasData(models.DailyLog{Notes: "note"}) {
		t.Fatal("notes day should have data")
	}
	if !dayHasData(models.DailyLog{Flow: models.FlowLight}) {
		t.Fatal("flow day should have data")
	}
}

func TestSanitizeLogForPartner(t *testing.T) {
	t.Parallel()

	entry := models.DailyLog{
		IsPeriod:   true,
		Flow:       models.FlowHeavy,
		SymptomIDs: []uint{1, 2, 3},
		Notes:      "secret",
	}

	sanitized := sanitizeLogForPartner(entry)
	if sanitized.Notes != "" {
		t.Fatalf("expected notes to be cleared, got %q", sanitized.Notes)
	}
	if len(sanitized.SymptomIDs) != 0 {
		t.Fatalf("expected symptom IDs to be cleared, got %#v", sanitized.SymptomIDs)
	}
	if !sanitized.IsPeriod || sanitized.Flow != models.FlowHeavy {
		t.Fatal("non-private fields must stay unchanged")
	}
}

func TestExportSymptomHelpers(t *testing.T) {
	t.Parallel()

	if got := exportSymptomColumn("Mood swings"); got != "mood" {
		t.Fatalf("expected mood, got %q", got)
	}
	if got := csvYesNo(true); got != "Yes" {
		t.Fatalf("expected Yes, got %q", got)
	}
	if got := normalizeExportFlow("unknown"); got != models.FlowNone {
		t.Fatalf("expected flow none, got %q", got)
	}
}

func TestDateHelpers(t *testing.T) {
	t.Parallel()

	location := time.UTC
	base := time.Date(2026, time.February, 17, 13, 14, 15, 0, time.UTC)
	day := dateAtLocation(base, location)
	if day.Hour() != 0 || day.Minute() != 0 || day.Second() != 0 {
		t.Fatal("dateAtLocation should zero time component")
	}

	start, end := dayRange(base, location)
	if !sameCalendarDay(start, day) {
		t.Fatal("dayRange start should match normalized day")
	}
	if !end.Equal(start.AddDate(0, 0, 1)) {
		t.Fatal("dayRange end should be next day")
	}

	if !betweenCalendarDaysInclusive(day, start, end.AddDate(0, 0, -1)) {
		t.Fatal("expected day to be between inclusive bounds")
	}
}
