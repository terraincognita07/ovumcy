package api

import (
	"testing"
	"time"

	"github.com/terraincognita07/lume/internal/models"
)

func TestFetchExportDataDateRangeFiltersInclusiveBoundaries(t *testing.T) {
	t.Parallel()

	handler, database := newDataAccessTestHandler(t)
	user := createDataAccessTestUser(t, database, "export-range-data@example.com")

	logs := []models.DailyLog{
		{UserID: user.ID, Date: time.Date(2026, time.February, 10, 0, 0, 0, 0, time.UTC), Flow: models.FlowNone},
		{UserID: user.ID, Date: time.Date(2026, time.February, 11, 0, 0, 0, 0, time.UTC), Flow: models.FlowLight},
		{UserID: user.ID, Date: time.Date(2026, time.February, 12, 0, 0, 0, 0, time.UTC), Flow: models.FlowMedium},
	}
	if err := database.Create(&logs).Error; err != nil {
		t.Fatalf("create logs: %v", err)
	}

	from := time.Date(2026, time.February, 11, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, time.February, 11, 0, 0, 0, 0, time.UTC)

	t.Run("exact day range includes target day", func(t *testing.T) {
		filtered, _, err := handler.fetchExportData(user.ID, &from, &to)
		if err != nil {
			t.Fatalf("fetchExportData returned error: %v", err)
		}
		if len(filtered) != 1 {
			t.Fatalf("expected exactly one entry, got %d", len(filtered))
		}
		if filtered[0].Date.Format("2006-01-02") != "2026-02-11" {
			t.Fatalf("expected date 2026-02-11, got %s", filtered[0].Date.Format("2006-01-02"))
		}
	})

	t.Run("from only includes from and after", func(t *testing.T) {
		filtered, _, err := handler.fetchExportData(user.ID, &from, nil)
		if err != nil {
			t.Fatalf("fetchExportData returned error: %v", err)
		}
		if len(filtered) != 2 {
			t.Fatalf("expected two entries, got %d", len(filtered))
		}
		if filtered[0].Date.Format("2006-01-02") != "2026-02-11" || filtered[1].Date.Format("2006-01-02") != "2026-02-12" {
			t.Fatalf("unexpected dates: %s, %s", filtered[0].Date.Format("2006-01-02"), filtered[1].Date.Format("2006-01-02"))
		}
	})

	t.Run("to only includes up to and including day", func(t *testing.T) {
		filtered, _, err := handler.fetchExportData(user.ID, nil, &to)
		if err != nil {
			t.Fatalf("fetchExportData returned error: %v", err)
		}
		if len(filtered) != 2 {
			t.Fatalf("expected two entries, got %d", len(filtered))
		}
		if filtered[0].Date.Format("2006-01-02") != "2026-02-10" || filtered[1].Date.Format("2006-01-02") != "2026-02-11" {
			t.Fatalf("unexpected dates: %s, %s", filtered[0].Date.Format("2006-01-02"), filtered[1].Date.Format("2006-01-02"))
		}
	})
}

func TestFetchExportSummaryForRangeDateFiltersInclusiveBoundaries(t *testing.T) {
	t.Parallel()

	handler, database := newDataAccessTestHandler(t)
	user := createDataAccessTestUser(t, database, "export-range-summary@example.com")

	logs := []models.DailyLog{
		{UserID: user.ID, Date: time.Date(2026, time.February, 10, 0, 0, 0, 0, time.UTC), Flow: models.FlowNone},
		{UserID: user.ID, Date: time.Date(2026, time.February, 11, 0, 0, 0, 0, time.UTC), Flow: models.FlowLight},
		{UserID: user.ID, Date: time.Date(2026, time.February, 12, 0, 0, 0, 0, time.UTC), Flow: models.FlowMedium},
	}
	if err := database.Create(&logs).Error; err != nil {
		t.Fatalf("create logs: %v", err)
	}

	from := time.Date(2026, time.February, 11, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, time.February, 11, 0, 0, 0, 0, time.UTC)

	total, first, last, err := handler.fetchExportSummaryForRange(user.ID, &from, &to)
	if err != nil {
		t.Fatalf("fetchExportSummaryForRange returned error: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected total=1, got %d", total)
	}
	if first != "2026-02-11" || last != "2026-02-11" {
		t.Fatalf("expected range 2026-02-11..2026-02-11, got %s..%s", first, last)
	}
}
