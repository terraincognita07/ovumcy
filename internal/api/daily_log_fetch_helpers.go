package api

import (
	"time"

	"github.com/terraincognita07/lume/internal/models"
)

func (handler *Handler) fetchLogsForUser(userID uint, from time.Time, to time.Time) ([]models.DailyLog, error) {
	fromStart, _ := dayRange(from, handler.location)
	_, toEnd := dayRange(to, handler.location)
	fromKeyLower := dayStorageKey(from.AddDate(0, 0, -1), handler.location)
	toKeyUpper := dayStorageKey(to.AddDate(0, 0, 1), handler.location)

	candidates := make([]models.DailyLog, 0)
	err := handler.dailyLogQueryForUser(userID).
		Where(
			"(julianday(date) >= julianday(?) AND julianday(date) < julianday(?)) OR (substr(date, 1, 10) >= ? AND substr(date, 1, 10) <= ?)",
			fromStart,
			toEnd,
			fromKeyLower,
			toKeyUpper,
		).
		Order("date ASC, id ASC").
		Find(&candidates).Error
	if err != nil {
		return nil, err
	}

	fromLocal := dateAtLocation(from, handler.location)
	toLocal := dateAtLocation(to, handler.location)
	logs := make([]models.DailyLog, 0, len(candidates))
	for _, candidate := range candidates {
		candidateDay := dateAtLocation(candidate.Date, handler.location)
		if candidateDay.Before(fromLocal) || candidateDay.After(toLocal) {
			continue
		}
		logs = append(logs, candidate)
	}
	return logs, nil
}

func (handler *Handler) fetchAllLogsForUser(userID uint) ([]models.DailyLog, error) {
	logs := make([]models.DailyLog, 0)
	err := handler.db.Where("user_id = ?", userID).Order("date ASC").Find(&logs).Error
	return logs, err
}

func (handler *Handler) fetchLogByDate(userID uint, day time.Time) (models.DailyLog, error) {
	dayStart := dateAtLocation(day, handler.location)
	dayKeys := dayLookupKeys(day, handler.location)
	candidates := make([]models.DailyLog, 0)
	result := handler.db.
		Select("id", "user_id", "date", "is_period", "flow", "symptom_ids", "notes", "created_at", "updated_at").
		Where("user_id = ? AND substr(date, 1, 10) IN ?", userID, dayKeys).
		Order("date DESC, id DESC").
		Find(&candidates)
	if result.Error != nil {
		return models.DailyLog{}, result.Error
	}
	for _, candidate := range candidates {
		if sameCalendarDayAtLocation(candidate.Date, dayStart, handler.location) {
			return candidate, nil
		}
	}
	if len(candidates) == 0 {
		return models.DailyLog{
			UserID:     userID,
			Date:       dayStart,
			Flow:       models.FlowNone,
			SymptomIDs: []uint{},
		}, nil
	}
	return models.DailyLog{
		UserID:     userID,
		Date:       dayStart,
		Flow:       models.FlowNone,
		SymptomIDs: []uint{},
	}, nil
}

func (handler *Handler) dayHasDataForDate(userID uint, day time.Time) (bool, error) {
	dayStart := dateAtLocation(day, handler.location)
	dayKeys := dayLookupKeys(dayStart, handler.location)
	entries := make([]models.DailyLog, 0)
	if err := handler.db.
		Select("date", "is_period", "flow", "symptom_ids", "notes").
		Where("user_id = ? AND substr(date, 1, 10) IN ?", userID, dayKeys).
		Find(&entries).Error; err != nil {
		return false, err
	}
	for _, entry := range entries {
		if !sameCalendarDayAtLocation(entry.Date, dayStart, handler.location) {
			continue
		}
		if dayHasData(entry) {
			return true, nil
		}
	}
	return false, nil
}
