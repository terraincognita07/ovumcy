package services

import (
	"testing"
	"time"
)

func TestCalcOvulationDay(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		cycleLength  int
		periodLength int
		wantDay      int
		wantExact    bool
	}{
		{name: "regular cycle", cycleLength: 28, periodLength: 5, wantDay: 14, wantExact: true},
		{name: "short cycle approximate", cycleLength: 15, periodLength: 7, wantDay: 8, wantExact: false},
		{name: "incompatible short cycle", cycleLength: 15, periodLength: 8, wantDay: 0, wantExact: false},
		{name: "incompatible long period", cycleLength: 21, periodLength: 14, wantDay: 0, wantExact: false},
		{name: "long cycle long period", cycleLength: 35, periodLength: 14, wantDay: 21, wantExact: true},
	}

	for _, testCase := range cases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			gotDay, gotExact := CalcOvulationDay(testCase.cycleLength, testCase.periodLength)
			if gotDay != testCase.wantDay {
				t.Fatalf("expected ovulation day %d, got %d", testCase.wantDay, gotDay)
			}
			if gotExact != testCase.wantExact {
				t.Fatalf("expected exact=%v, got %v", testCase.wantExact, gotExact)
			}
		})
	}
}

func TestPredictCycleWindow_IncompatibleValues(t *testing.T) {
	t.Parallel()

	periodStart := mustParseDay(t, "2026-02-10")
	ovulationDate, fertilityStart, fertilityEnd, exact, calculable := PredictCycleWindow(periodStart, 15, 10, 14)

	if calculable {
		t.Fatalf("expected incompatible values to be non-calculable")
	}
	if exact {
		t.Fatalf("expected exact=false for non-calculable prediction")
	}
	if !ovulationDate.IsZero() || !fertilityStart.IsZero() || !fertilityEnd.IsZero() {
		t.Fatalf("expected zero dates for incompatible values, got ov=%s fs=%s fe=%s",
			ovulationDate.Format("2006-01-02"),
			fertilityStart.Format("2006-01-02"),
			fertilityEnd.Format("2006-01-02"))
	}
}

func TestPredictCycleWindow_ApproximateForShortRemaining(t *testing.T) {
	t.Parallel()

	periodStart := mustParseDay(t, "2026-02-10")
	ovulationDate, fertilityStart, fertilityEnd, exact, calculable := PredictCycleWindow(periodStart, 15, 7, 14)

	if !calculable {
		t.Fatalf("expected calculable prediction for remaining=8")
	}
	if exact {
		t.Fatalf("expected exact=false for approximate prediction")
	}
	if got := ovulationDate.Format("2006-01-02"); got != "2026-02-17" {
		t.Fatalf("expected ovulation date 2026-02-17, got %s", got)
	}
	if got := fertilityStart.Format("2006-01-02"); got != "2026-02-17" {
		t.Fatalf("expected fertility start 2026-02-17, got %s", got)
	}
	if got := fertilityEnd.Format("2006-01-02"); got != "2026-02-18" {
		t.Fatalf("expected fertility end 2026-02-18, got %s", got)
	}

	assertCyclePredictionInvariants(t, periodStart, 15, 7, ovulationDate, fertilityStart, fertilityEnd)
}

func TestPredictCycleWindow_NormalCycle(t *testing.T) {
	t.Parallel()

	periodStart := mustParseDay(t, "2026-02-10")
	ovulationDate, fertilityStart, fertilityEnd, exact, calculable := PredictCycleWindow(periodStart, 28, 5, 14)

	if !calculable {
		t.Fatalf("expected calculable prediction for regular cycle")
	}
	if !exact {
		t.Fatalf("expected exact prediction for regular cycle")
	}
	if got := ovulationDate.Format("2006-01-02"); got != "2026-02-23" {
		t.Fatalf("expected ovulation date 2026-02-23, got %s", got)
	}
	if got := fertilityStart.Format("2006-01-02"); got != "2026-02-18" {
		t.Fatalf("expected fertility start 2026-02-18, got %s", got)
	}
	if got := fertilityEnd.Format("2006-01-02"); got != "2026-02-24" {
		t.Fatalf("expected fertility end 2026-02-24, got %s", got)
	}

	assertCyclePredictionInvariants(t, periodStart, 28, 5, ovulationDate, fertilityStart, fertilityEnd)
}

func TestPredictCycleWindow_InvariantsAcrossRanges(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name           string
		cycleLength    int
		periodLength   int
		wantCalculable bool
	}{
		{name: "incompatible", cycleLength: 15, periodLength: 10, wantCalculable: false},
		{name: "approximate", cycleLength: 15, periodLength: 7, wantCalculable: true},
		{name: "regular", cycleLength: 28, periodLength: 5, wantCalculable: true},
		{name: "long period", cycleLength: 35, periodLength: 14, wantCalculable: true},
		{name: "max cycle", cycleLength: 90, periodLength: 14, wantCalculable: true},
	}

	periodStart := mustParseDay(t, "2026-02-10")
	for _, testCase := range cases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			ovulationDate, fertilityStart, fertilityEnd, _, calculable := PredictCycleWindow(
				periodStart,
				testCase.cycleLength,
				testCase.periodLength,
				14,
			)
			if calculable != testCase.wantCalculable {
				t.Fatalf("expected calculable=%v, got %v", testCase.wantCalculable, calculable)
			}
			if !calculable {
				if !ovulationDate.IsZero() || !fertilityStart.IsZero() || !fertilityEnd.IsZero() {
					t.Fatalf("expected zero dates for non-calculable prediction")
				}
				return
			}
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
