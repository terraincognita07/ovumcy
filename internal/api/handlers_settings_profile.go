package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/lume/internal/models"
)

func (handler *Handler) UpdateProfile(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return apiError(c, fiber.StatusUnauthorized, "unauthorized")
	}

	input := profileSettingsInput{}
	if err := c.BodyParser(&input); err != nil {
		return handler.respondSettingsError(c, fiber.StatusBadRequest, "invalid profile input")
	}

	displayName, err := normalizeDisplayName(input.DisplayName)
	if err != nil {
		return handler.respondSettingsError(c, fiber.StatusBadRequest, err.Error())
	}

	if err := handler.db.Model(&models.User{}).Where("id = ?", user.ID).Update("display_name", displayName).Error; err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to update profile")
	}

	status := profileUpdateStatus(user.DisplayName, displayName)

	user.DisplayName = displayName

	if acceptsJSON(c) {
		return c.JSON(fiber.Map{
			"ok":           true,
			"display_name": displayName,
			"status":       status,
		})
	}
	handler.setFlashCookie(c, FlashPayload{SettingsSuccess: status})
	return redirectOrJSON(c, "/settings")
}

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
