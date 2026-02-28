package services

import "strings"

type SettingsErrorTarget string

const (
	SettingsErrorTargetGeneral        SettingsErrorTarget = "general"
	SettingsErrorTargetChangePassword SettingsErrorTarget = "change_password"
)

type NotificationService struct{}

func NewNotificationService() *NotificationService {
	return &NotificationService{}
}

func (service *NotificationService) ResolveSettingsStatus(flashSuccess string, querySuccess string, queryStatus string) string {
	return firstNonEmptyTrimmed(flashSuccess, querySuccess, queryStatus)
}

func (service *NotificationService) ResolveSettingsErrorSource(flashError string, queryError string) string {
	return firstNonEmptyTrimmed(flashError, queryError)
}

func (service *NotificationService) ClassifySettingsErrorSource(errorSource string) SettingsErrorTarget {
	if service.IsChangePasswordErrorMessage(errorSource) {
		return SettingsErrorTargetChangePassword
	}
	return SettingsErrorTargetGeneral
}

func (service *NotificationService) IsChangePasswordErrorMessage(message string) bool {
	switch strings.ToLower(strings.TrimSpace(message)) {
	case "invalid settings input", "password mismatch", "invalid current password", "new password must differ", "weak password":
		return true
	default:
		return false
	}
}

func firstNonEmptyTrimmed(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}
