package api

import (
	"time"

	"github.com/terraincognita07/lume/internal/models"
	"github.com/terraincognita07/lume/internal/services"
)

func (handler *Handler) deleteDailyLogByDate(userID uint, day time.Time) error {
	dayKey := dayStorageKey(day, handler.location)
	return handler.db.Where("user_id = ? AND substr(date, 1, 10) = ?", userID, dayKey).Delete(&models.DailyLog{}).Error
}

func (handler *Handler) refreshUserLastPeriodStart(userID uint) error {
	periodLogs := make([]models.DailyLog, 0)
	if err := handler.db.
		Select("date", "is_period").
		Where("user_id = ? AND is_period = ?", userID, true).
		Order("date ASC").
		Find(&periodLogs).Error; err != nil {
		return err
	}

	starts := services.DetectCycleStarts(periodLogs)
	if len(starts) == 0 {
		return handler.db.Model(&models.User{}).Where("id = ?", userID).Update("last_period_start", nil).Error
	}

	latest := dateAtLocation(starts[len(starts)-1], handler.location)
	return handler.db.Model(&models.User{}).Where("id = ?", userID).Update("last_period_start", latest).Error
}
