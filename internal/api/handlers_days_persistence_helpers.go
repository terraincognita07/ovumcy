package api

import (
	"errors"
	"time"

	"github.com/terraincognita07/lume/internal/models"
)

var (
	errDayEntryLoadFailed   = errors.New("load day entry failed")
	errDayEntryCreateFailed = errors.New("create day entry failed")
	errDayEntryUpdateFailed = errors.New("update day entry failed")
	errDeleteDayFailed      = errors.New("delete day failed")
	errSyncLastPeriodFailed = errors.New("sync last period failed")
)

func (handler *Handler) upsertDayEntry(userID uint, dayStart time.Time, payload dayPayload, cleanIDs []uint) (models.DailyLog, bool, error) {
	dayKeys := dayLookupKeys(dayStart, handler.location)

	wasPeriod := false
	var entry models.DailyLog
	candidates := make([]models.DailyLog, 0)
	result := handler.db.
		Where("user_id = ? AND substr(date, 1, 10) IN ?", userID, dayKeys).
		Order("date DESC, id DESC").
		Find(&candidates)

	if result.Error != nil {
		return models.DailyLog{}, false, errDayEntryLoadFailed
	}

	found := false
	for _, candidate := range candidates {
		if sameCalendarDayAtLocation(candidate.Date, dayStart, handler.location) {
			entry = candidate
			wasPeriod = candidate.IsPeriod
			found = true
			break
		}
	}

	if !found {
		entry = models.DailyLog{
			UserID:   userID,
			Date:     dayStart,
			IsPeriod: payload.IsPeriod,
			Flow:     payload.Flow,
			Notes:    payload.Notes,
		}
		entry.SymptomIDs = cleanIDs
		if err := handler.db.Create(&entry).Error; err != nil {
			return models.DailyLog{}, false, errDayEntryCreateFailed
		}
		return entry, wasPeriod, nil
	}

	entry.IsPeriod = payload.IsPeriod
	entry.Flow = payload.Flow
	entry.SymptomIDs = cleanIDs
	entry.Notes = payload.Notes
	if err := handler.db.Save(&entry).Error; err != nil {
		return models.DailyLog{}, false, errDayEntryUpdateFailed
	}
	return entry, wasPeriod, nil
}

func (handler *Handler) deleteDayAndRefreshLastPeriod(userID uint, day time.Time) error {
	if err := handler.deleteDailyLogByDate(userID, day); err != nil {
		return errDeleteDayFailed
	}
	if err := handler.refreshUserLastPeriodStart(userID); err != nil {
		return errSyncLastPeriodFailed
	}
	return nil
}
