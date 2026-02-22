package services

import (
	"testing"
	"time"
)

func TestPredictCycleWindow_ShortCycleLongPeriod(t *testing.T) {
	t.Parallel()

	periodStart := mustParseDay(t, "2026-02-10")
	ovulationDate, fertilityStart, fertilityEnd := PredictCycleWindow(periodStart, 15, 10, 14)

	if got := ovulationDate.Format("2006-01-02"); got != "2026-02-21" {
		t.Fatalf("expected ovulation date 2026-02-21, got %s", got)
	}
	if got := fertilityStart.Format("2006-01-02"); got != "2026-02-20" {
		t.Fatalf("expected fertility start 2026-02-20, got %s", got)
	}
	if got := fertilityEnd.Format("2006-01-02"); got != "2026-02-22" {
		t.Fatalf("expected fertility end 2026-02-22, got %s", got)
	}

	assertCyclePredictionInvariants(t, periodStart, 15, 10, ovulationDate, fertilityStart, fertilityEnd)
}

func TestPredictCycleWindow_NormalCycle(t *testing.T) {
	t.Parallel()

	periodStart := mustParseDay(t, "2026-02-10")
	ovulationDate, fertilityStart, fertilityEnd := PredictCycleWindow(periodStart, 28, 5, 14)

	if got := ovulationDate.Format("2006-01-02"); got != "2026-02-24" {
		t.Fatalf("expected ovulation date 2026-02-24, got %s", got)
	}
	if got := fertilityStart.Format("2006-01-02"); got != "2026-02-19" {
		t.Fatalf("expected fertility start 2026-02-19, got %s", got)
	}
	if got := fertilityEnd.Format("2006-01-02"); got != "2026-02-25" {
		t.Fatalf("expected fertility end 2026-02-25, got %s", got)
	}

	assertCyclePredictionInvariants(t, periodStart, 28, 5, ovulationDate, fertilityStart, fertilityEnd)
}

func TestPredictCycleWindow_InvariantsAcrossRanges(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		cycleLength  int
		periodLength int
	}{
		{name: "short cycle long period", cycleLength: 15, periodLength: 10},
		{name: "short cycle medium period", cycleLength: 20, periodLength: 8},
		{name: "regular cycle", cycleLength: 28, periodLength: 5},
		{name: "long cycle", cycleLength: 35, periodLength: 6},
		{name: "max cycle", cycleLength: 90, periodLength: 1},
	}

	periodStart := mustParseDay(t, "2026-02-10")
	for _, testCase := range cases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			ovulationDate, fertilityStart, fertilityEnd := PredictCycleWindow(
				periodStart,
				testCase.cycleLength,
				testCase.periodLength,
				14,
			)
			assertCyclePredictionInvariants(
				t,
				periodStart,
				testCase.cycleLength,
				testCase.periodLength,
				ovulationDate,
				fertilityStart,
				fertilityEnd,
			)
		})
	}
}

func assertCyclePredictionInvariants(t *testing.T, periodStart time.Time, cycleLength int, periodLength int, ovulationDate time.Time, fertilityStart time.Time, fertilityEnd time.Time) {
	t.Helper()

	periodEnd := periodStart.AddDate(0, 0, periodLength-1)
	nextPeriodStart := periodStart.AddDate(0, 0, cycleLength)

	if ovulationDate.IsZero() {
		t.Fatalf("ovulation date must not be zero")
	}
	if !ovulationDate.After(periodEnd) {
		t.Fatalf("ovulation %s must be after period end %s", ovulationDate.Format("2006-01-02"), periodEnd.Format("2006-01-02"))
	}
	if !ovulationDate.Before(nextPeriodStart) {
		t.Fatalf("ovulation %s must be before next period start %s", ovulationDate.Format("2006-01-02"), nextPeriodStart.Format("2006-01-02"))
	}

	if fertilityStart.IsZero() || fertilityEnd.IsZero() {
		return
	}
	if !fertilityStart.After(periodEnd) {
		t.Fatalf("fertility start %s must be after period end %s", fertilityStart.Format("2006-01-02"), periodEnd.Format("2006-01-02"))
	}
	if fertilityStart.After(fertilityEnd) {
		t.Fatalf("fertility start %s must not be after fertility end %s", fertilityStart.Format("2006-01-02"), fertilityEnd.Format("2006-01-02"))
	}
	if !fertilityEnd.Before(nextPeriodStart) {
		t.Fatalf("fertility end %s must be before next period start %s", fertilityEnd.Format("2006-01-02"), nextPeriodStart.Format("2006-01-02"))
	}
}
