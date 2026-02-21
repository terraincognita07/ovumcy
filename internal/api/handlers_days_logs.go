package api

import (
	"errors"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/lume/internal/models"
	"gorm.io/gorm"
)

func (handler *Handler) GetDays(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return apiError(c, fiber.StatusUnauthorized, "unauthorized")
	}

	from, err := parseDayParam(c.Query("from"), handler.location)
	if err != nil {
		return apiError(c, fiber.StatusBadRequest, "invalid from date")
	}
	to, err := parseDayParam(c.Query("to"), handler.location)
	if err != nil {
		return apiError(c, fiber.StatusBadRequest, "invalid to date")
	}
	if to.Before(from) {
		return apiError(c, fiber.StatusBadRequest, "invalid range")
	}

	logs, err := handler.fetchLogsForUser(user.ID, from, to)
	if err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to fetch logs")
	}

	sanitizeLogsForViewer(user, logs)

	return c.JSON(logs)
}

func (handler *Handler) GetDay(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return apiError(c, fiber.StatusUnauthorized, "unauthorized")
	}

	day, err := parseDayParam(c.Params("date"), handler.location)
	if err != nil {
		return apiError(c, fiber.StatusBadRequest, "invalid date")
	}

	logEntry, err := handler.fetchLogByDate(user.ID, day)
	if err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to fetch day")
	}

	logEntry = sanitizeLogForViewer(user, logEntry)

	return c.JSON(logEntry)
}

func (handler *Handler) CheckDayExists(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return apiError(c, fiber.StatusUnauthorized, "unauthorized")
	}

	day, err := parseDayParam(c.Params("date"), handler.location)
	if err != nil {
		return apiError(c, fiber.StatusBadRequest, "invalid date")
	}

	exists, err := handler.dayHasDataForDate(user.ID, day)
	if err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to fetch day")
	}

	return c.JSON(fiber.Map{"exists": exists})
}

func (handler *Handler) UpsertDay(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return apiError(c, fiber.StatusUnauthorized, "unauthorized")
	}

	day, err := parseDayParam(c.Params("date"), handler.location)
	if err != nil {
		return apiError(c, fiber.StatusBadRequest, "invalid date")
	}

	payload, err := parseDayPayload(c)
	if err != nil {
		return apiError(c, fiber.StatusBadRequest, "invalid payload")
	}

	if !isValidFlow(payload.Flow) {
		return apiError(c, fiber.StatusBadRequest, "invalid flow value")
	}
	if payload.IsPeriod && payload.Flow == models.FlowNone {
		return apiError(c, fiber.StatusBadRequest, "period flow is required")
	}
	if !payload.IsPeriod {
		payload.Flow = models.FlowNone
	}

	if len(payload.Notes) > 2000 {
		payload.Notes = payload.Notes[:2000]
	}

	cleanIDs, err := handler.validateSymptomIDs(user.ID, payload.SymptomIDs)
	if err != nil {
		return apiError(c, fiber.StatusBadRequest, "invalid symptom ids")
	}

	dayStart, _ := dayRange(day, handler.location)
	dayKey := dayStorageKey(dayStart, handler.location)
	nextDayKey := nextDayStorageKey(dayStart, handler.location)
	wasPeriod := false
	autoPeriodFillEnabled := false
	periodLength := 5

	var entry models.DailyLog
	result := handler.db.
		Where("user_id = ? AND date >= ? AND date < ?", user.ID, dayKey, nextDayKey).
		Order("date DESC, id DESC").
		First(&entry)
	if result.Error == nil {
		wasPeriod = entry.IsPeriod
	}

	if payload.IsPeriod {
		settings := struct {
			PeriodLength   int
			AutoPeriodFill bool
		}{}
		if err := handler.db.Model(&models.User{}).
			Select("period_length", "auto_period_fill").
			First(&settings, user.ID).Error; err != nil {
			return apiError(c, fiber.StatusInternalServerError, "failed to load day")
		}
		if isValidOnboardingPeriodLength(settings.PeriodLength) {
			periodLength = settings.PeriodLength
		}
		autoPeriodFillEnabled = settings.AutoPeriodFill
	}

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		entry = models.DailyLog{
			UserID:   user.ID,
			Date:     dayStart,
			IsPeriod: payload.IsPeriod,
			Flow:     payload.Flow,
			Notes:    payload.Notes,
		}
		entry.SymptomIDs = cleanIDs
		if err := handler.db.Create(&entry).Error; err != nil {
			return apiError(c, fiber.StatusInternalServerError, "failed to create day")
		}
	} else if result.Error != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to load day")
	} else {
		entry.IsPeriod = payload.IsPeriod
		entry.Flow = payload.Flow
		entry.SymptomIDs = cleanIDs
		entry.Notes = payload.Notes
		if err := handler.db.Save(&entry).Error; err != nil {
			return apiError(c, fiber.StatusInternalServerError, "failed to update day")
		}
	}

	if payload.IsPeriod && autoPeriodFillEnabled && periodLength > 1 && !wasPeriod {
		previousDay := dayStart.AddDate(0, 0, -1)
		previousDayEntry, err := handler.fetchLogByDate(user.ID, previousDay)
		if err != nil {
			return apiError(c, fiber.StatusInternalServerError, "failed to load day")
		}

		hasRecentPeriod, err := handler.hasPeriodInRecentDays(user.ID, dayStart, 3)
		if err != nil {
			return apiError(c, fiber.StatusInternalServerError, "failed to load day")
		}

		if !previousDayEntry.IsPeriod && !hasRecentPeriod {
			if err := handler.autoFillFollowingPeriodDays(user.ID, dayStart, periodLength, payload.Flow); err != nil {
				return apiError(c, fiber.StatusInternalServerError, "failed to update day")
			}
		}
	}

	if err := handler.refreshUserLastPeriodStart(user.ID); err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to sync last period start")
	}

	if isHTMX(c) {
		c.Set("HX-Trigger", "calendar-day-updated")
		timestamp := time.Now().In(handler.location).Format("15:04")
		pattern := translateMessage(currentMessages(c), "common.saved_at")
		if pattern == "common.saved_at" {
			pattern = "Saved at %s"
		}
		message := fmt.Sprintf(pattern, timestamp)
		return c.SendString(fmt.Sprintf("<div class=\"status-ok status-transient\">%s</div>", template.HTMLEscapeString(message)))
	}

	return c.JSON(entry)
}

func (handler *Handler) hasPeriodInRecentDays(userID uint, day time.Time, lookbackDays int) (bool, error) {
	if lookbackDays <= 0 {
		return false, nil
	}

	for offset := 1; offset <= lookbackDays; offset++ {
		previousDay := day.AddDate(0, 0, -offset)
		entry, err := handler.fetchLogByDate(userID, previousDay)
		if err != nil {
			return false, err
		}
		if entry.IsPeriod {
			return true, nil
		}
	}

	return false, nil
}

func (handler *Handler) autoFillFollowingPeriodDays(userID uint, startDay time.Time, periodLength int, flow string) error {
	if periodLength <= 1 {
		return nil
	}

	for offset := 1; offset < periodLength; offset++ {
		targetDay := dateAtLocation(startDay.AddDate(0, 0, offset), handler.location)
		entry, err := handler.fetchLogByDate(userID, targetDay)
		if err != nil {
			return err
		}

		if entry.ID != 0 {
			if dayHasData(entry) && !entry.IsPeriod {
				break
			}
			if entry.IsPeriod {
				continue
			}

			entry.IsPeriod = true
			entry.Flow = flow
			if err := handler.db.Save(&entry).Error; err != nil {
				return err
			}
			continue
		}

		newEntry := models.DailyLog{
			UserID:     userID,
			Date:       targetDay,
			IsPeriod:   true,
			Flow:       flow,
			SymptomIDs: []uint{},
		}
		if err := handler.db.Create(&newEntry).Error; err != nil {
			return err
		}
	}

	return nil
}

func (handler *Handler) DeleteDailyLog(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return apiError(c, fiber.StatusUnauthorized, "unauthorized")
	}

	day, err := parseDayParam(c.Query("date"), handler.location)
	if err != nil {
		return apiError(c, fiber.StatusBadRequest, "invalid date")
	}

	if err := handler.deleteDailyLogByDate(user.ID, day); err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to delete day")
	}
	if err := handler.refreshUserLastPeriodStart(user.ID); err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to sync last period start")
	}

	source := strings.ToLower(strings.TrimSpace(c.Query("source")))
	if isHTMX(c) {
		c.Set("HX-Trigger", "calendar-day-updated")
		switch source {
		case "calendar":
			return handler.renderDayEditorPartial(c, user, day)
		case "dashboard":
			c.Set("HX-Redirect", "/dashboard")
			return c.SendStatus(fiber.StatusOK)
		default:
			return c.SendStatus(fiber.StatusNoContent)
		}
	}

	if source == "dashboard" {
		return redirectOrJSON(c, "/dashboard")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (handler *Handler) DeleteDay(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return apiError(c, fiber.StatusUnauthorized, "unauthorized")
	}

	day, err := parseDayParam(c.Params("date"), handler.location)
	if err != nil {
		return apiError(c, fiber.StatusBadRequest, "invalid date")
	}

	if err := handler.deleteDailyLogByDate(user.ID, day); err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to delete day")
	}
	if err := handler.refreshUserLastPeriodStart(user.ID); err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to sync last period start")
	}

	if isHTMX(c) {
		c.Set("HX-Trigger", "calendar-day-updated")
		return handler.renderDayEditorPartial(c, user, day)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
