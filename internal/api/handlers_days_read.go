package api

import "github.com/gofiber/fiber/v2"

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
