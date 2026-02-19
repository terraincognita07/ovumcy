package api

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/lume/internal/services"
)

func (handler *Handler) GetStatsOverview(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return apiError(c, fiber.StatusUnauthorized, "unauthorized")
	}

	now := time.Now().In(handler.location)
	logs, err := handler.fetchLogsForUser(user.ID, now.AddDate(-2, 0, 0), now)
	if err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to fetch stats")
	}

	stats := services.BuildCycleStats(logs, now, handler.lutealPhaseDays)
	stats = handler.applyUserCycleBaseline(user, logs, stats, now)
	return c.JSON(stats)
}
