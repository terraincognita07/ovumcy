package api

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
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
	autoPeriodFillEnabled := false
	periodLength := 5

	if payload.IsPeriod {
		periodLength, autoPeriodFillEnabled, err = handler.loadDayAutoFillSettings(user.ID)
		if err != nil {
			return apiError(c, fiber.StatusInternalServerError, "failed to load day")
		}
	}

	entry, wasPeriod, err := handler.upsertDayEntry(user.ID, dayStart, payload, cleanIDs)
	if err != nil {
		return upsertDayPersistenceAPIError(c, err)
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

	if err := handler.deleteDayAndRefreshLastPeriod(user.ID, day); err != nil {
		return deleteDayPersistenceAPIError(c, err)
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

	if err := handler.deleteDayAndRefreshLastPeriod(user.ID, day); err != nil {
		return deleteDayPersistenceAPIError(c, err)
	}

	if isHTMX(c) {
		c.Set("HX-Trigger", "calendar-day-updated")
		return handler.renderDayEditorPartial(c, user, day)
	}

	return c.SendStatus(fiber.StatusNoContent)
}
