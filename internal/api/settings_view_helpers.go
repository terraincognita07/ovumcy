package api

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/ovumcy/internal/models"
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
	language := currentLanguage(c)
	status := settingsStatusFromFlashOrQuery(c, flash)
	errorKey := ""
	changePasswordErrorKey := ""
	if status == "" {
		errorKey, changePasswordErrorKey = classifySettingsErrorSource(settingsErrorSourceFromFlashOrQuery(c, flash))
	}

	handler.ensureDependencies()
	persisted, err := handler.settingsService.LoadSettings(user.ID)
	if err != nil {
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
	user.LastPeriodStart = persisted.LastPeriodStart

	lastPeriodStart := ""
	if persisted.LastPeriodStart != nil {
		lastPeriodStart = dateAtLocation(*persisted.LastPeriodStart, handler.location).Format("2006-01-02")
	}
	minCycleStart, today := currentYearDateBounds(time.Now().In(handler.location), handler.location)

	data := fiber.Map{
		"Title":                  localizedPageTitle(messages, "meta.title.settings", "Ovumcy | Settings"),
		"CurrentUser":            user,
		"ErrorKey":               errorKey,
		"ChangePasswordErrorKey": changePasswordErrorKey,
		"SuccessKey":             settingsStatusTranslationKey(status),
		"CycleLength":            cycleLength,
		"PeriodLength":           periodLength,
		"AutoPeriodFill":         autoPeriodFill,
		"LastPeriodStart":        lastPeriodStart,
		"TodayISO":               today.Format("2006-01-02"),
		"CycleStartMinISO":       minCycleStart.Format("2006-01-02"),
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
		displayFrom := firstDate
		if parsedFrom, parseErr := parseDayParam(firstDate, handler.location); parseErr == nil {
			displayFrom = localizedDateDisplay(language, parsedFrom)
		}
		displayTo := lastDate
		if parsedTo, parseErr := parseDayParam(lastDate, handler.location); parseErr == nil {
			displayTo = localizedDateDisplay(language, parsedTo)
		}
		data["ExportDateFromDisplay"] = displayFrom
		data["ExportDateToDisplay"] = displayTo
	}

	return data, nil
}
