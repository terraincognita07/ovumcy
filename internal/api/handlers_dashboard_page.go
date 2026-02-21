package api

import "github.com/gofiber/fiber/v2"

func (handler *Handler) ShowDashboard(c *fiber.Ctx) error {
	user, handled, err := handler.currentUserOrRedirectToLogin(c)
	if err != nil {
		return err
	}
	if handled {
		return nil
	}

	language, messages, now := handler.currentPageViewContext(c)
	data, errorMessage, err := handler.buildDashboardViewData(user, language, messages, now)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(errorMessage)
	}

	return handler.render(c, "dashboard", data)
}
