package api

import (
	"time"

	"github.com/terraincognita07/lume/internal/models"
	"gorm.io/gorm"
)

func (handler *Handler) dailyLogQueryForUser(userID uint) *gorm.DB {
	return handler.db.Model(&models.DailyLog{}).Where("user_id = ?", userID)
}

func (handler *Handler) applyDailyLogDateRange(query *gorm.DB, from *time.Time, to *time.Time) *gorm.DB {
	if from != nil {
		query = query.Where("date >= ?", dayStorageKey(*from, handler.location))
	}
	if to != nil {
		query = query.Where("date < ?", nextDayStorageKey(*to, handler.location))
	}
	return query
}

func (handler *Handler) dailyLogRangeQueryForUser(userID uint, from *time.Time, to *time.Time) *gorm.DB {
	return handler.applyDailyLogDateRange(handler.dailyLogQueryForUser(userID), from, to)
}
