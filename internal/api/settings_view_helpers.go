package api

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/ovumcy/internal/models"
	"github.com/terraincognita07/ovumcy/internal/services"
)

func resolveSettingsErrorKeys(notificationService *services.NotificationService, errorSource string) (string, string) {
	translatedErrorKey := authErrorTranslationKey(errorSource)
	if translatedErrorKey == "" {
		return "", ""
	}
	if notificationService.ClassifySettingsErrorSource(errorSource) == services.SettingsErrorTargetChangePassword {
		return "", translatedErrorKey
	}
	return translatedErrorKey, ""
}

func (handler *Handler) buildSettingsViewData(c *fiber.Ctx, user *models.User, flash FlashPayload) (fiber.Map, error) {
	messages := currentMessages(c)
	language := currentLanguage(c)
	handler.ensureDependencies()

	status := handler.notificationService.ResolveSettingsStatus(
		flash.SettingsSuccess,
		c.Query("success"),
		c.Query("status"),
	)
	errorKey := ""
	changePasswordErrorKey := ""
	if status == "" {
		errorSource := handler.notificationService.ResolveSettingsErrorSource(flash.SettingsError, c.Query("error"))
		errorKey, changePasswordErrorKey = resolveSettingsErrorKeys(handler.notificationService, errorSource)
	}

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
