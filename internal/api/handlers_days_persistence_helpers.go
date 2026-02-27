package api

import (
	"time"

	"github.com/terraincognita07/ovumcy/internal/services"
)

var (
	errDayEntryLoadFailed   = services.ErrDayEntryLoadFailed
	errDayEntryCreateFailed = services.ErrDayEntryCreateFailed
	errDayEntryUpdateFailed = services.ErrDayEntryUpdateFailed
	errDeleteDayFailed      = services.ErrDeleteDayFailed
	errSyncLastPeriodFailed = services.ErrSyncLastPeriodFailed
)

func (handler *Handler) deleteDayAndRefreshLastPeriod(userID uint, day time.Time) error {
	handler.ensureDependencies()
	return handler.dayService.DeleteDayAndRefreshLastPeriod(userID, day, handler.location)
}
