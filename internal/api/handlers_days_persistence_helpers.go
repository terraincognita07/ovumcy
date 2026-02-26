package api

import (
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
	"github.com/terraincognita07/ovumcy/internal/services"
)

var (
	errDayEntryLoadFailed   = services.ErrDayEntryLoadFailed
	errDayEntryCreateFailed = services.ErrDayEntryCreateFailed
	errDayEntryUpdateFailed = services.ErrDayEntryUpdateFailed
	errDeleteDayFailed      = services.ErrDeleteDayFailed
	errSyncLastPeriodFailed = services.ErrSyncLastPeriodFailed
)

func (handler *Handler) upsertDayEntry(userID uint, dayStart time.Time, payload dayPayload, cleanIDs []uint) (models.DailyLog, bool, error) {
	handler.ensureDependencies()
	return handler.dayService.UpsertDayEntry(userID, dayStart, services.DayEntryInput{
		IsPeriod:   payload.IsPeriod,
		Flow:       payload.Flow,
		Notes:      payload.Notes,
		SymptomIDs: cleanIDs,
	}, handler.location)
}

func (handler *Handler) deleteDayAndRefreshLastPeriod(userID uint, day time.Time) error {
	handler.ensureDependencies()
	return handler.dayService.DeleteDayAndRefreshLastPeriod(userID, day, handler.location)
}
