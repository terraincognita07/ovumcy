package api

import (
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
)

func (handler *Handler) hasPeriodInRecentDays(userID uint, day time.Time, lookbackDays int) (bool, error) {
	if lookbackDays <= 0 {
		return false, nil
	}

	for offset := 1; offset <= lookbackDays; offset++ {
		previousDay := day.AddDate(0, 0, -offset)
		entry, err := handler.fetchLogByDate(userID, previousDay)
		if err != nil {
			return false, err
		}
		if entry.IsPeriod {
			return true, nil
		}
	}

	return false, nil
}

func (handler *Handler) autoFillFollowingPeriodDays(userID uint, startDay time.Time, periodLength int, flow string) error {
	if periodLength <= 1 {
		return nil
	}

	for offset := 1; offset < periodLength; offset++ {
		targetDay := dateAtLocation(startDay.AddDate(0, 0, offset), handler.location)
		entry, err := handler.fetchLogByDate(userID, targetDay)
		if err != nil {
			return err
		}

		if entry.ID != 0 {
			if dayHasData(entry) && !entry.IsPeriod {
				break
			}
			if entry.IsPeriod {
				continue
			}

			entry.IsPeriod = true
			entry.Flow = flow
			if err := handler.db.Save(&entry).Error; err != nil {
				return err
			}
			continue
		}

		newEntry := models.DailyLog{
			UserID:     userID,
			Date:       targetDay,
			IsPeriod:   true,
			Flow:       flow,
			SymptomIDs: []uint{},
		}
		if err := handler.db.Create(&newEntry).Error; err != nil {
			return err
		}
	}

	return nil
}

func (handler *Handler) loadDayAutoFillSettings(userID uint) (int, bool, error) {
	periodLength := 5
	settings := struct {
		PeriodLength   int
		AutoPeriodFill bool
	}{}

	if err := handler.db.Model(&models.User{}).
		Select("period_length", "auto_period_fill").
		First(&settings, userID).Error; err != nil {
		return periodLength, false, err
	}
	if isValidOnboardingPeriodLength(settings.PeriodLength) {
		periodLength = settings.PeriodLength
	}
	return periodLength, settings.AutoPeriodFill, nil
}

func (handler *Handler) shouldAutoFillPeriodDays(userID uint, dayStart time.Time, wasPeriod bool, autoPeriodFillEnabled bool, periodLength int) (bool, error) {
	if !autoPeriodFillEnabled || periodLength <= 1 || wasPeriod {
		return false, nil
	}

	previousDay := dayStart.AddDate(0, 0, -1)
	previousDayEntry, err := handler.fetchLogByDate(userID, previousDay)
	if err != nil {
		return false, err
	}

	hasRecentPeriod, err := handler.hasPeriodInRecentDays(userID, dayStart, 3)
	if err != nil {
		return false, err
	}

	return !previousDayEntry.IsPeriod && !hasRecentPeriod, nil
}
