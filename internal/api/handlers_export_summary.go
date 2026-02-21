package api

import "github.com/gofiber/fiber/v2"

func (handler *Handler) ExportSummary(c *fiber.Ctx) error {
	user, from, to, status, message := handler.exportUserAndRange(c)
	if status != 0 {
		return apiError(c, status, message)
	}

	totalEntries, firstDate, lastDate, err := handler.fetchExportSummaryForRange(user.ID, from, to)
	if err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to fetch logs")
	}

	return c.JSON(fiber.Map{
		"total_entries": int(totalEntries),
		"has_data":      totalEntries > 0,
		"date_from":     firstDate,
		"date_to":       lastDate,
	})
}
