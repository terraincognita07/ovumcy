package api

import "github.com/gofiber/fiber/v2"

func (handler *Handler) ShowSettings(c *fiber.Ctx) error {
	user, handled, err := handler.currentUserOrRedirectToLogin(c)
	if err != nil {
		return err
	}
	if handled {
		return nil
	}

	data, errorMessage, err := handler.buildSettingsPageData(c, user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(errorMessage)
	}

	return handler.render(c, "settings", data)
}
