package api

import (
	"time"

	"github.com/terraincognita07/lume/internal/models"
	"github.com/terraincognita07/lume/internal/services"
)

func (handler *Handler) deleteDailyLogByDate(userID uint, day time.Time) error {
	dayStart := dateAtLocation(day, handler.location)
	dayKeys := dayLookupKeys(dayStart, handler.location)

	candidates := make([]models.DailyLog, 0)
	if err := handler.db.
		Select("id", "date").
		Where("user_id = ? AND substr(date, 1, 10) IN ?", userID, dayKeys).
		Order("date DESC, id DESC").
		Find(&candidates).Error; err != nil {
		return err
	}

	ids := make([]uint, 0, len(candidates))
	for _, candidate := range candidates {
		if sameCalendarDayAtLocation(candidate.Date, dayStart, handler.location) {
			ids = append(ids, candidate.ID)
		}
	}
	if len(ids) == 0 {
		return nil
	}

	return handler.db.Where("id IN ?", ids).Delete(&models.DailyLog{}).Error
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
