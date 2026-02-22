package api

import (
	"fmt"
	"html/template"

	"github.com/gofiber/fiber/v2"
)

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
	if isHTMX(c) {
		message := translateMessage(currentMessages(c), "settings.success.cycle_updated")
		if message == "settings.success.cycle_updated" {
			message = "Cycle settings updated successfully."
		}
		return c.SendString(fmt.Sprintf("<div class=\"status-ok status-transient\">%s</div>", template.HTMLEscapeString(message)))
	}

	handler.setFlashCookie(c, FlashPayload{SettingsSuccess: "cycle_updated"})
	return redirectOrJSON(c, "/settings")
}
