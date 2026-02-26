package api

import (
	"time"
)

func (handler *Handler) deleteDailyLogByDate(userID uint, day time.Time) error {
	handler.ensureDependencies()
	return handler.dayService.DeleteDailyLogByDate(userID, day, handler.location)
}

func (handler *Handler) refreshUserLastPeriodStart(userID uint) error {
	handler.ensureDependencies()
	return handler.dayService.RefreshUserLastPeriodStart(userID, handler.location)
}
