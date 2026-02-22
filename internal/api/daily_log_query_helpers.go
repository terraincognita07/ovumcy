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
		fromStart, _ := dayRange(*from, handler.location)
		query = query.Where("date >= ?", fromStart)
	}
	if to != nil {
		_, toEnd := dayRange(*to, handler.location)
		query = query.Where("date < ?", toEnd)
	}
	return query
}

func (handler *Handler) dailyLogRangeQueryForUser(userID uint, from *time.Time, to *time.Time) *gorm.DB {
	return handler.applyDailyLogDateRange(handler.dailyLogQueryForUser(userID), from, to)
}
