package services

import (
	"errors"
	"testing"
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
)

func TestValidateCycleSettingsRejectsOutOfRangeAndIncompatibleValues(t *testing.T) {
	service := NewSettingsService(nil)
	now := time.Date(2026, time.February, 15, 12, 0, 0, 0, time.UTC)

	_, err := service.ValidateCycleSettings(CycleSettingsValidationInput{
		CycleLength:  14,
		PeriodLength: 5,
	}, now, time.UTC)
	if !errors.Is(err, ErrSettingsCycleLengthOutOfRange) {
		t.Fatalf("expected ErrSettingsCycleLengthOutOfRange, got %v", err)
	}

	_, err = service.ValidateCycleSettings(CycleSettingsValidationInput{
		CycleLength:  28,
		PeriodLength: 15,
	}, now, time.UTC)
	if !errors.Is(err, ErrSettingsPeriodLengthOutOfRange) {
		t.Fatalf("expected ErrSettingsPeriodLengthOutOfRange, got %v", err)
	}

	_, err = service.ValidateCycleSettings(CycleSettingsValidationInput{
		CycleLength:  20,
		PeriodLength: 13,
	}, now, time.UTC)
	if !errors.Is(err, ErrSettingsPeriodLengthIncompatible) {
		t.Fatalf("expected ErrSettingsPeriodLengthIncompatible, got %v", err)
	}
}

func TestValidateCycleSettingsLastPeriodStartRules(t *testing.T) {
	service := NewSettingsService(nil)
	now := time.Date(2026, time.February, 15, 12, 0, 0, 0, time.UTC)

	_, err := service.ValidateCycleSettings(CycleSettingsValidationInput{
		CycleLength:        28,
		PeriodLength:       5,
		LastPeriodStartSet: true,
		LastPeriodStartRaw: "invalid-date",
	}, now, time.UTC)
	if !errors.Is(err, ErrSettingsCycleStartDateInvalid) {
		t.Fatalf("expected ErrSettingsCycleStartDateInvalid for parse, got %v", err)
	}

	_, err = service.ValidateCycleSettings(CycleSettingsValidationInput{
		CycleLength:        28,
		PeriodLength:       5,
		LastPeriodStartSet: true,
		LastPeriodStartRaw: "2025-12-31",
	}, now, time.UTC)
	if !errors.Is(err, ErrSettingsCycleStartDateInvalid) {
		t.Fatalf("expected ErrSettingsCycleStartDateInvalid for out-of-range old date, got %v", err)
	}

	_, err = service.ValidateCycleSettings(CycleSettingsValidationInput{
		CycleLength:        28,
		PeriodLength:       5,
		LastPeriodStartSet: true,
		LastPeriodStartRaw: "2026-02-16",
	}, now, time.UTC)
	if !errors.Is(err, ErrSettingsCycleStartDateInvalid) {
		t.Fatalf("expected ErrSettingsCycleStartDateInvalid for out-of-range future date, got %v", err)
	}
}

func TestValidateCycleSettingsBuildsUpdate(t *testing.T) {
	service := NewSettingsService(nil)
	now := time.Date(2026, time.February, 15, 12, 0, 0, 0, time.UTC)

	update, err := service.ValidateCycleSettings(CycleSettingsValidationInput{
		CycleLength:        28,
		PeriodLength:       6,
		AutoPeriodFill:     true,
		LastPeriodStartSet: true,
		LastPeriodStartRaw: "2026-02-10",
	}, now, time.UTC)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if update.LastPeriodStart == nil || update.LastPeriodStart.Format("2006-01-02") != "2026-02-10" {
		t.Fatalf("expected normalized last_period_start 2026-02-10, got %#v", update.LastPeriodStart)
	}

	cleared, err := service.ValidateCycleSettings(CycleSettingsValidationInput{
		CycleLength:        28,
		PeriodLength:       6,
		AutoPeriodFill:     false,
		LastPeriodStartSet: true,
		LastPeriodStartRaw: "   ",
	}, now, time.UTC)
	if err != nil {
		t.Fatalf("expected nil error for clear, got %v", err)
	}
	if cleared.LastPeriodStart != nil {
		t.Fatalf("expected nil last_period_start, got %#v", cleared.LastPeriodStart)
	}
}

func TestApplyCycleSettings(t *testing.T) {
	service := NewSettingsService(nil)
	existingStart := time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC)
	user := &models.User{
		CycleLength:     30,
		PeriodLength:    7,
		AutoPeriodFill:  false,
		LastPeriodStart: &existingStart,
	}

	update := CycleSettingsUpdate{
		CycleLength:        28,
		PeriodLength:       5,
		AutoPeriodFill:     true,
		LastPeriodStartSet: true,
		LastPeriodStart:    nil,
	}
	service.ApplyCycleSettings(user, update)

	if user.CycleLength != 28 || user.PeriodLength != 5 || !user.AutoPeriodFill {
		t.Fatalf("expected updated cycle settings, got cycle=%d period=%d auto=%v", user.CycleLength, user.PeriodLength, user.AutoPeriodFill)
	}
	if user.LastPeriodStart != nil {
		t.Fatalf("expected cleared last period start, got %v", user.LastPeriodStart)
	}
}
