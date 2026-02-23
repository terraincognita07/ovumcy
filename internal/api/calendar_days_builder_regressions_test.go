package api

import (
	"testing"
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
	"github.com/terraincognita07/ovumcy/internal/services"
)

func TestBuildCalendarDaysUsesLatestLogPerDateDeterministically(t *testing.T) {
	handler := &Handler{location: time.UTC}
	monthStart := time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC)
	now := time.Date(2026, time.February, 20, 0, 0, 0, 0, time.UTC)

	logs := []models.DailyLog{
		{
			ID:       20,
			Date:     time.Date(2026, time.February, 17, 20, 0, 0, 0, time.UTC),
			IsPeriod: false,
			Flow:     models.FlowNone,
		},
		{
			ID:       10,
			Date:     time.Date(2026, time.February, 17, 8, 0, 0, 0, time.UTC),
			IsPeriod: true,
			Flow:     models.FlowMedium,
		},
		{
			ID:       30,
			Date:     time.Date(2026, time.February, 18, 9, 0, 0, 0, time.UTC),
			IsPeriod: true,
			Flow:     models.FlowMedium,
		},
		{
			ID:       31,
			Date:     time.Date(2026, time.February, 18, 9, 0, 0, 0, time.UTC),
			IsPeriod: false,
			Flow:     models.FlowNone,
		},
	}

	days := handler.buildCalendarDays(monthStart, logs, services.CycleStats{}, now)

	day17 := findCalendarDayByDateString(t, days, "2026-02-17")
	if day17.IsPeriod {
		t.Fatalf("expected 2026-02-17 period=false from latest log, got true")
	}

	day18 := findCalendarDayByDateString(t, days, "2026-02-18")
	if day18.IsPeriod {
		t.Fatalf("expected 2026-02-18 period=false from highest id tie-breaker, got true")
	}
}

func TestBuildCalendarDaysProjectsOvulationIntoFutureCycles(t *testing.T) {
	handler := &Handler{location: time.UTC}
	monthStart := time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC)
	now := time.Date(2026, time.February, 23, 0, 0, 0, 0, time.UTC)

	stats := services.CycleStats{
		MedianCycleLength:    28,
		AveragePeriodLength:  5,
		LastPeriodStart:      time.Date(2026, time.February, 10, 0, 0, 0, 0, time.UTC),
		NextPeriodStart:      time.Date(2026, time.March, 10, 0, 0, 0, 0, time.UTC),
		OvulationDate:        time.Date(2026, time.February, 23, 0, 0, 0, 0, time.UTC),
		FertilityWindowStart: time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC),
		FertilityWindowEnd:   time.Date(2026, time.February, 24, 0, 0, 0, 0, time.UTC),
	}

	days := handler.buildCalendarDays(monthStart, nil, stats, now)

	ovulationDay := findCalendarDayByDateString(t, days, "2026-03-23")
	if !ovulationDay.IsOvulation {
		t.Fatalf("expected projected ovulation marker on 2026-03-23")
	}
	if ovulationDay.IsFertility {
		t.Fatalf("expected ovulation day to not be marked as fertile tag state")
	}
	if ovulationDay.IsPredicted {
		t.Fatalf("did not expect ovulation day to be marked as predicted period")
	}
	if ovulationDay.BadgeClass != "calendar-tag calendar-tag-ovulation" {
		t.Fatalf("expected ovulation badge class, got %q", ovulationDay.BadgeClass)
	}
}
