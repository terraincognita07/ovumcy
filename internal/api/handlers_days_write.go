package api

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/lume/internal/models"
	"gorm.io/gorm"
)

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
	payload, err = normalizeDayPayload(payload)
	if err != nil {
		if errors.Is(err, errInvalidFlowValue) {
			return apiError(c, fiber.StatusBadRequest, "invalid flow value")
		}
		if errors.Is(err, errPeriodFlowRequired) {
			return apiError(c, fiber.StatusBadRequest, "period flow is required")
		}
		return apiError(c, fiber.StatusBadRequest, "invalid payload")
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
		periodLength, autoPeriodFillEnabled, err = handler.loadDayAutoFillSettings(user.ID)
		if err != nil {
			return apiError(c, fiber.StatusInternalServerError, "failed to load day")
		}
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

	if payload.IsPeriod {
		shouldAutoFill, err := handler.shouldAutoFillPeriodDays(user.ID, dayStart, wasPeriod, autoPeriodFillEnabled, periodLength)
		if err != nil {
			return apiError(c, fiber.StatusInternalServerError, "failed to load day")
		}
		if shouldAutoFill {
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
		return handler.sendDaySaveStatus(c)
	}

	return c.JSON(entry)
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
