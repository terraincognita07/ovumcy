package api

import (
	"time"

	"github.com/terraincognita07/lume/internal/models"
	"github.com/terraincognita07/lume/internal/services"
)

func (handler *Handler) deleteDailyLogByDate(userID uint, day time.Time) error {
	dayStart, dayEnd := dayRange(day, handler.location)
	return handler.db.Where("user_id = ? AND date >= ? AND date < ?", userID, dayStart, dayEnd).Delete(&models.DailyLog{}).Error
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
