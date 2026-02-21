package api

import (
	"time"

	"github.com/terraincognita07/lume/internal/models"
)

func isOwnerUser(user *models.User) bool {
	return user != nil && user.Role == models.RoleOwner
}

func isPartnerUser(user *models.User) bool {
	return user != nil && user.Role == models.RolePartner
}

func sanitizeLogForViewer(user *models.User, entry models.DailyLog) models.DailyLog {
	if isPartnerUser(user) {
		return sanitizeLogForPartner(entry)
	}
	return entry
}

func sanitizeLogsForViewer(user *models.User, logs []models.DailyLog) {
	if !isPartnerUser(user) {
		return
	}
	for index := range logs {
		logs[index] = sanitizeLogForPartner(logs[index])
	}
}

func (handler *Handler) fetchSymptomsForViewer(user *models.User) ([]models.SymptomType, error) {
	if !isOwnerUser(user) {
		return []models.SymptomType{}, nil
	}
	return handler.fetchSymptoms(user.ID)
}

func (handler *Handler) fetchDayLogForViewer(user *models.User, day time.Time) (models.DailyLog, []models.SymptomType, error) {
	logEntry, err := handler.fetchLogByDate(user.ID, day)
	if err != nil {
		return models.DailyLog{}, nil, err
	}

	symptoms, err := handler.fetchSymptomsForViewer(user)
	if err != nil {
		return models.DailyLog{}, nil, err
	}

	return sanitizeLogForViewer(user, logEntry), symptoms, nil
}
