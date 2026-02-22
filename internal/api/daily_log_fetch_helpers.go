package api

import (
	"time"

	"github.com/terraincognita07/lume/internal/models"
)

func (handler *Handler) fetchLogsForUser(userID uint, from time.Time, to time.Time) ([]models.DailyLog, error) {
	logs := make([]models.DailyLog, 0)
	err := handler.dailyLogRangeQueryForUser(userID, &from, &to).
		Order("date ASC, id ASC").
		Find(&logs).Error
	return logs, err
}

func (handler *Handler) fetchAllLogsForUser(userID uint) ([]models.DailyLog, error) {
	logs := make([]models.DailyLog, 0)
	err := handler.db.Where("user_id = ?", userID).Order("date ASC").Find(&logs).Error
	return logs, err
}

func (handler *Handler) fetchLogByDate(userID uint, day time.Time) (models.DailyLog, error) {
	entry := models.DailyLog{}
	dayStart := dateAtLocation(day, handler.location)
	dayKey := dayStorageKey(day, handler.location)
	result := handler.db.
		Select("id", "user_id", "date", "is_period", "flow", "symptom_ids", "notes", "created_at", "updated_at").
		Where("user_id = ? AND substr(date, 1, 10) = ?", userID, dayKey).
		Order("date DESC, id DESC").
		Limit(1).
		Find(&entry)
	if result.Error != nil {
		return models.DailyLog{}, result.Error
	}
	if result.RowsAffected == 0 {
		return models.DailyLog{
			UserID:     userID,
			Date:       dayStart,
			Flow:       models.FlowNone,
			SymptomIDs: []uint{},
		}, nil
	}
	return entry, nil
}

func (handler *Handler) dayHasDataForDate(userID uint, day time.Time) (bool, error) {
	dayKey := dayStorageKey(day, handler.location)
	entries := make([]models.DailyLog, 0)
	if err := handler.db.
		Select("is_period", "flow", "symptom_ids", "notes").
		Where("user_id = ? AND substr(date, 1, 10) = ?", userID, dayKey).
		Find(&entries).Error; err != nil {
		return false, err
	}
	for _, entry := range entries {
		if dayHasData(entry) {
			return true, nil
		}
	}
	return false, nil
}
