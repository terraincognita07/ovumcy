package api

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/lume/internal/models"
)

func isChangePasswordErrorMessage(message string) bool {
	switch strings.ToLower(strings.TrimSpace(message)) {
	case "invalid settings input", "password mismatch", "invalid current password", "new password must differ", "weak password":
		return true
	default:
		return false
	}
}

func pickFirstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func settingsStatusFromFlashOrQuery(c *fiber.Ctx, flash FlashPayload) string {
	return pickFirstNonEmpty(
		flash.SettingsSuccess,
		c.Query("success"),
		c.Query("status"),
	)
}

func settingsErrorSourceFromFlashOrQuery(c *fiber.Ctx, flash FlashPayload) string {
	return pickFirstNonEmpty(
		flash.SettingsError,
		c.Query("error"),
	)
}

func classifySettingsErrorSource(errorSource string) (string, string) {
	translatedErrorKey := authErrorTranslationKey(errorSource)
	if isChangePasswordErrorMessage(errorSource) && translatedErrorKey != "" {
		return "", translatedErrorKey
	}
	return translatedErrorKey, ""
}

func (handler *Handler) buildSettingsViewData(c *fiber.Ctx, user *models.User, flash FlashPayload) (fiber.Map, error) {
	messages := currentMessages(c)
	status := settingsStatusFromFlashOrQuery(c, flash)
	errorKey := ""
	changePasswordErrorKey := ""
	if status == "" {
		errorKey, changePasswordErrorKey = classifySettingsErrorSource(settingsErrorSourceFromFlashOrQuery(c, flash))
	}

	persisted := models.User{}
	if err := handler.db.Select("cycle_length", "period_length", "auto_period_fill").First(&persisted, user.ID).Error; err != nil {
		return nil, err
	}

	cycleLength := persisted.CycleLength
	if !isValidOnboardingCycleLength(cycleLength) {
		cycleLength = models.DefaultCycleLength
	}
	periodLength := persisted.PeriodLength
	if !isValidOnboardingPeriodLength(periodLength) {
		periodLength = models.DefaultPeriodLength
	}
	autoPeriodFill := persisted.AutoPeriodFill
	user.CycleLength = cycleLength
	user.PeriodLength = periodLength
	user.AutoPeriodFill = autoPeriodFill

	data := fiber.Map{
		"Title":                  localizedPageTitle(messages, "meta.title.settings", "Lume | Settings"),
		"CurrentUser":            user,
		"ErrorKey":               errorKey,
		"ChangePasswordErrorKey": changePasswordErrorKey,
		"SuccessKey":             settingsStatusTranslationKey(status),
		"CycleLength":            cycleLength,
		"PeriodLength":           periodLength,
		"AutoPeriodFill":         autoPeriodFill,
	}

	if user.Role == models.RoleOwner {
		totalEntries, firstDate, lastDate, err := handler.fetchExportSummary(user.ID)
		if err != nil {
			return nil, err
		}
		data["ExportTotalEntries"] = int(totalEntries)
		data["HasExportData"] = totalEntries > 0
		data["ExportDateFrom"] = firstDate
		data["ExportDateTo"] = lastDate
	}

	return data, nil
}
