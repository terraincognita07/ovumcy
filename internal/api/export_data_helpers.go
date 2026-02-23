package api

import (
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
)

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
