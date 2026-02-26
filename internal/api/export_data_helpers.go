package api

import (
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
)

func (handler *Handler) fetchExportData(userID uint, from *time.Time, to *time.Time) ([]models.DailyLog, map[uint]string, error) {
	handler.ensureDependencies()
	logs, err := handler.dayService.FetchLogsForOptionalRange(userID, from, to, handler.location)
	if err != nil {
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
	handler.ensureDependencies()
	logs, err := handler.dayService.FetchLogsForOptionalRange(userID, from, to, handler.location)
	if err != nil {
		return 0, "", "", err
	}
	if len(logs) == 0 {
		return 0, "", "", nil
	}

	first := logs[0].Date
	last := logs[0].Date
	for _, logEntry := range logs[1:] {
		if logEntry.Date.Before(first) {
			first = logEntry.Date
		}
		if logEntry.Date.After(last) {
			last = logEntry.Date
		}
	}

	return int64(len(logs)),
		dateAtLocation(first, handler.location).Format("2006-01-02"),
		dateAtLocation(last, handler.location).Format("2006-01-02"),
		nil
}
