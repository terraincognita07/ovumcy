package api

import "github.com/gofiber/fiber/v2"

func (handler *Handler) UpdateCycleSettings(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return apiError(c, fiber.StatusUnauthorized, "unauthorized")
	}

	input, parseError := parseCycleSettingsInput(c)
	if parseError != "" {
		return handler.respondSettingsError(c, fiber.StatusBadRequest, parseError)
	}

	if err := handler.saveCycleSettings(user.ID, input); err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to update cycle settings")
	}

	applyCycleSettings(user, input)

	if acceptsJSON(c) {
		return c.JSON(fiber.Map{"ok": true})
	}
	handler.setFlashCookie(c, FlashPayload{SettingsSuccess: "cycle_updated"})
	return redirectOrJSON(c, "/settings")
}
