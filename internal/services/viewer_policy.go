package services

import "github.com/terraincognita07/ovumcy/internal/models"

func IsOwnerUser(user *models.User) bool {
	return user != nil && user.Role == models.RoleOwner
}

func IsPartnerUser(user *models.User) bool {
	return user != nil && user.Role == models.RolePartner
}

func SanitizeLogForPartner(entry models.DailyLog) models.DailyLog {
	entry.Notes = ""
	entry.SymptomIDs = []uint{}
	return entry
}

func SanitizeLogForViewer(user *models.User, entry models.DailyLog) models.DailyLog {
	if IsPartnerUser(user) {
		return SanitizeLogForPartner(entry)
	}
	return entry
}

func SanitizeLogsForViewer(user *models.User, logs []models.DailyLog) {
	if !IsPartnerUser(user) {
		return
	}
	for index := range logs {
		logs[index] = SanitizeLogForPartner(logs[index])
	}
}

func ShouldExposeSymptomsForViewer(user *models.User) bool {
	return IsOwnerUser(user)
}
