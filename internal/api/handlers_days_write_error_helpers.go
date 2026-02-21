package api

import (
	"errors"

	"github.com/gofiber/fiber/v2"
)

func upsertDayPersistenceAPIError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, errDayEntryLoadFailed):
		return apiError(c, fiber.StatusInternalServerError, "failed to load day")
	case errors.Is(err, errDayEntryCreateFailed):
		return apiError(c, fiber.StatusInternalServerError, "failed to create day")
	case errors.Is(err, errDayEntryUpdateFailed):
		return apiError(c, fiber.StatusInternalServerError, "failed to update day")
	default:
		return apiError(c, fiber.StatusInternalServerError, "failed to update day")
	}
}

func deleteDayPersistenceAPIError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, errDeleteDayFailed):
		return apiError(c, fiber.StatusInternalServerError, "failed to delete day")
	case errors.Is(err, errSyncLastPeriodFailed):
		return apiError(c, fiber.StatusInternalServerError, "failed to sync last period start")
	default:
		return apiError(c, fiber.StatusInternalServerError, "failed to delete day")
	}
}
