package api

import (
	"time"

	"github.com/terraincognita07/lume/internal/models"
	"github.com/terraincognita07/lume/internal/services"
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

func (handler *Handler) fetchExportData(userID uint, from *time.Time, to *time.Time) ([]models.DailyLog, map[uint]string, error) {
	logs := make([]models.DailyLog, 0)
	query := handler.dailyLogRangeQueryForUser(userID, from, to)
	if err := query.Order("date ASC").Find(&logs).Error; err != nil {
		return nil, nil, err
	}

	symptoms, err := handler.fetchSymptoms(userID)
	if err != nil {
		return nil, nil, err
	}

	symptomNames := make(map[uint]string, len(symptoms))
	for _, symptom := range symptoms {
		symptomNames[symptom.ID] = symptom.Name
	}

	return logs, symptomNames, nil
}

func (handler *Handler) fetchExportSummary(userID uint) (int64, string, string, error) {
	return handler.fetchExportSummaryForRange(userID, nil, nil)
}

func (handler *Handler) fetchExportSummaryForRange(userID uint, from *time.Time, to *time.Time) (int64, string, string, error) {
	var aggregate struct {
		Total     int64  `gorm:"column:total"`
		FirstDate string `gorm:"column:first_date"`
		LastDate  string `gorm:"column:last_date"`
	}

	if err := handler.dailyLogRangeQueryForUser(userID, from, to).
		Select("COUNT(*) AS total, MIN(date) AS first_date, MAX(date) AS last_date").
		Scan(&aggregate).Error; err != nil {
		return 0, "", "", err
	}
	if aggregate.Total == 0 || aggregate.FirstDate == "" || aggregate.LastDate == "" {
		return 0, "", "", nil
	}

	firstDate := aggregate.FirstDate
	if len(firstDate) > 10 {
		firstDate = firstDate[:10]
	}
	lastDate := aggregate.LastDate
	if len(lastDate) > 10 {
		lastDate = lastDate[:10]
	}

	return aggregate.Total,
		firstDate,
		lastDate,
		nil
}

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
	dayStart, dayEnd := dayRange(day, handler.location)
	dayKey := dayStorageKey(dayStart, handler.location)
	nextDayKey := dayStorageKey(dayEnd, handler.location)
	result := handler.db.
		Where("user_id = ? AND date >= ? AND date < ?", userID, dayKey, nextDayKey).
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

func (handler *Handler) deleteDailyLogByDate(userID uint, day time.Time) error {
	dayKey := dayStorageKey(day, handler.location)
	nextDayKey := nextDayStorageKey(day, handler.location)
	return handler.db.Where("user_id = ? AND date >= ? AND date < ?", userID, dayKey, nextDayKey).Delete(&models.DailyLog{}).Error
}

func (handler *Handler) dayHasDataForDate(userID uint, day time.Time) (bool, error) {
	dayKey := dayStorageKey(day, handler.location)
	nextDayKey := nextDayStorageKey(day, handler.location)
	entries := make([]models.DailyLog, 0)
	if err := handler.db.
		Select("is_period", "flow", "symptom_ids", "notes").
		Where("user_id = ? AND date >= ? AND date < ?", userID, dayKey, nextDayKey).
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
