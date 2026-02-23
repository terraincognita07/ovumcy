package api

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/ovumcy/internal/models"
)

func (handler *Handler) CalendarDayPanel(c *fiber.Ctx) error {
	user, handled, err := currentUserOrUnauthorized(c)
	if err != nil {
		return err
	}
	if handled {
		return nil
	}

	day, err := parseDayParam(c.Params("date"), handler.location)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("invalid date")
	}

	return handler.renderDayEditorPartial(c, user, day)
}

func (handler *Handler) renderDayEditorPartial(c *fiber.Ctx, user *models.User, day time.Time) error {
	language, messages, now := handler.currentPageViewContext(c)
	payload, errorMessage, err := handler.buildDayEditorPartialData(user, language, messages, day, now)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(errorMessage)
	}
	return handler.renderPartial(c, "day_editor_partial", payload)
}
