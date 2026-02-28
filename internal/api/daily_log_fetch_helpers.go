package api

import (
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
)

func (handler *Handler) fetchLogsForUser(userID uint, from time.Time, to time.Time) ([]models.DailyLog, error) {
	handler.ensureDependencies()
	return handler.dayService.FetchLogsForUser(userID, from, to, handler.location)
}

func (handler *Handler) fetchLogByDate(userID uint, day time.Time) (models.DailyLog, error) {
	handler.ensureDependencies()
	return handler.dayService.FetchLogByDate(userID, day, handler.location)
}

func (handler *Handler) dayHasDataForDate(userID uint, day time.Time) (bool, error) {
	handler.ensureDependencies()
	return handler.dayService.DayHasDataForDate(userID, day, handler.location)
}
