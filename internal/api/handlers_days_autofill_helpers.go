package api

import (
	"time"

	"github.com/terraincognita07/lume/internal/models"
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
