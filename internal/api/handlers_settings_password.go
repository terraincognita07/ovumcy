package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/lume/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func (handler *Handler) ChangePassword(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return apiError(c, fiber.StatusUnauthorized, "unauthorized")
	}

	input, parseError := parseChangePasswordInput(c)
	if parseError != "" {
		return handler.respondSettingsError(c, fiber.StatusBadRequest, parseError)
	}

	validationStatus, validationError := validateChangePasswordInput(input, user)
	if validationStatus != 0 {
		return handler.respondSettingsError(c, validationStatus, validationError)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to secure password")
	}

	if err := handler.db.Model(&models.User{}).Where("id = ?", user.ID).Updates(map[string]any{
		"password_hash":        string(passwordHash),
		"must_change_password": false,
	}).Error; err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to update password")
	}

	if acceptsJSON(c) {
		return c.JSON(fiber.Map{"ok": true})
	}
	handler.setFlashCookie(c, FlashPayload{SettingsSuccess: "password_changed"})
	return redirectOrJSON(c, "/settings")
}
