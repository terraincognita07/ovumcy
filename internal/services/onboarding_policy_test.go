package services

import (
	"errors"
	"testing"
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
)

func TestOnboardingDateBounds_UsesYearStartWhenWithinFirstSixtyDays(t *testing.T) {
	location, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		t.Fatalf("load location: %v", err)
	}

	now := time.Date(2026, time.February, 15, 18, 45, 0, 0, location)
	minDate, maxDate := OnboardingDateBounds(now, location)

	expectedMin := time.Date(2026, time.January, 1, 0, 0, 0, 0, location)
	expectedMax := time.Date(2026, time.February, 15, 0, 0, 0, 0, location)
	if !minDate.Equal(expectedMin) {
		t.Fatalf("expected min date %s, got %s", expectedMin.Format(time.RFC3339), minDate.Format(time.RFC3339))
	}
	if !maxDate.Equal(expectedMax) {
		t.Fatalf("expected max date %s, got %s", expectedMax.Format(time.RFC3339), maxDate.Format(time.RFC3339))
	}
}

func TestOnboardingDateBounds_UsesRollingSixtyDaysAfterWindow(t *testing.T) {
	location, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		t.Fatalf("load location: %v", err)
	}

	now := time.Date(2026, time.April, 15, 9, 10, 0, 0, location)
	minDate, maxDate := OnboardingDateBounds(now, location)

	expectedMin := time.Date(2026, time.February, 14, 0, 0, 0, 0, location)
	expectedMax := time.Date(2026, time.April, 15, 0, 0, 0, 0, location)
	if !minDate.Equal(expectedMin) {
		t.Fatalf("expected min date %s, got %s", expectedMin.Format(time.RFC3339), minDate.Format(time.RFC3339))
	}
	if !maxDate.Equal(expectedMax) {
		t.Fatalf("expected max date %s, got %s", expectedMax.Format(time.RFC3339), maxDate.Format(time.RFC3339))
	}
}

func TestValidateStep1StartDate_RejectsRequiredAndOutOfRange(t *testing.T) {
	location, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		t.Fatalf("load location: %v", err)
	}

	service := NewOnboardingService(nil)
	now := time.Date(2026, time.April, 15, 10, 0, 0, 0, location)

	if err := service.ValidateStep1StartDate(time.Time{}, now, location); !errors.Is(err, ErrOnboardingStartDateRequired) {
		t.Fatalf("expected ErrOnboardingStartDateRequired, got %v", err)
	}

	tooOld := time.Date(2026, time.February, 13, 0, 0, 0, 0, location)
	if err := service.ValidateStep1StartDate(tooOld, now, location); !errors.Is(err, ErrOnboardingStartDateOutOfRange) {
		t.Fatalf("expected ErrOnboardingStartDateOutOfRange for old date, got %v", err)
	}

	future := now.AddDate(0, 0, 1)
	if err := service.ValidateStep1StartDate(future, now, location); !errors.Is(err, ErrOnboardingStartDateOutOfRange) {
		t.Fatalf("expected ErrOnboardingStartDateOutOfRange for future date, got %v", err)
	}
}

func TestValidateStep1StartDate_AcceptsBoundaries(t *testing.T) {
	location, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		t.Fatalf("load location: %v", err)
	}

	service := NewOnboardingService(nil)
	now := time.Date(2026, time.April, 15, 10, 0, 0, 0, location)
	minDate, maxDate := OnboardingDateBounds(now, location)

	if err := service.ValidateStep1StartDate(minDate, now, location); err != nil {
		t.Fatalf("expected nil error for min boundary, got %v", err)
	}
	if err := service.ValidateStep1StartDate(maxDate, now, location); err != nil {
		t.Fatalf("expected nil error for max boundary, got %v", err)
	}
}

func TestResolveCycleAndPeriodDefaults(t *testing.T) {
	t.Run("keeps valid values", func(t *testing.T) {
		cycleLength, periodLength := ResolveCycleAndPeriodDefaults(29, 6)
		if cycleLength != 29 || periodLength != 6 {
			t.Fatalf("expected 29/6, got %d/%d", cycleLength, periodLength)
		}
	})

	t.Run("falls back for invalid values", func(t *testing.T) {
		cycleLength, periodLength := ResolveCycleAndPeriodDefaults(120, 0)
		if cycleLength != models.DefaultCycleLength || periodLength != models.DefaultPeriodLength {
			t.Fatalf("expected defaults %d/%d, got %d/%d", models.DefaultCycleLength, models.DefaultPeriodLength, cycleLength, periodLength)
		}
	})
}
