package api

import (
	"errors"

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
