package services

import (
	"errors"
	"strings"
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
)

var (
	ErrSettingsCycleLengthOutOfRange    = errors.New("settings cycle length out of range")
	ErrSettingsPeriodLengthOutOfRange   = errors.New("settings period length out of range")
	ErrSettingsPeriodLengthIncompatible = errors.New("settings period length incompatible with cycle length")
	ErrSettingsCycleStartDateInvalid    = errors.New("settings cycle start date invalid")
)

type CycleSettingsValidationInput struct {
	CycleLength        int
	PeriodLength       int
	AutoPeriodFill     bool
	LastPeriodStartRaw string
	LastPeriodStartSet bool
}

func (service *SettingsService) ValidateCycleSettings(input CycleSettingsValidationInput, now time.Time, location *time.Location) (CycleSettingsUpdate, error) {
	if !IsValidOnboardingCycleLength(input.CycleLength) {
		return CycleSettingsUpdate{}, ErrSettingsCycleLengthOutOfRange
	}
	if !IsValidOnboardingPeriodLength(input.PeriodLength) {
		return CycleSettingsUpdate{}, ErrSettingsPeriodLengthOutOfRange
	}

	ovulationDay, _ := CalcOvulationDay(input.CycleLength, input.PeriodLength)
	if ovulationDay <= 0 {
		return CycleSettingsUpdate{}, ErrSettingsPeriodLengthIncompatible
	}

	update := CycleSettingsUpdate{
		CycleLength:        input.CycleLength,
		PeriodLength:       input.PeriodLength,
		AutoPeriodFill:     input.AutoPeriodFill,
		LastPeriodStartSet: input.LastPeriodStartSet,
	}

	if !input.LastPeriodStartSet {
		return update, nil
	}

	rawDate := strings.TrimSpace(input.LastPeriodStartRaw)
	if rawDate == "" {
		update.LastPeriodStart = nil
		return update, nil
	}

	if location == nil {
		location = time.UTC
	}
	parsedDay, err := time.ParseInLocation("2006-01-02", rawDate, location)
	if err != nil {
		return CycleSettingsUpdate{}, ErrSettingsCycleStartDateInvalid
	}
	parsedDay = DateAtLocation(parsedDay, location)

	minCycleStart, today := SettingsCycleStartDateBounds(now, location)
	if parsedDay.Before(minCycleStart) || parsedDay.After(today) {
		return CycleSettingsUpdate{}, ErrSettingsCycleStartDateInvalid
	}

	update.LastPeriodStart = &parsedDay
	return update, nil
}

func (service *SettingsService) ApplyCycleSettings(user *models.User, update CycleSettingsUpdate) {
	if user == nil {
		return
	}

	user.CycleLength = update.CycleLength
	user.PeriodLength = update.PeriodLength
	user.AutoPeriodFill = update.AutoPeriodFill

	if !update.LastPeriodStartSet {
		return
	}
	if update.LastPeriodStart == nil {
		user.LastPeriodStart = nil
		return
	}

	day := *update.LastPeriodStart
	user.LastPeriodStart = &day
}

func SettingsCycleStartDateBounds(now time.Time, location *time.Location) (time.Time, time.Time) {
	if location == nil {
		location = time.UTC
	}
	today := DateAtLocation(now.In(location), location)
	minDate := time.Date(today.Year(), time.January, 1, 0, 0, 0, 0, location)
	return minDate, today
}
