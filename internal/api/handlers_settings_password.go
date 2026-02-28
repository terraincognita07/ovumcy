package api

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/ovumcy/internal/services"
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

	handler.ensureDependencies()
	if err := handler.settingsService.ValidatePasswordChange(
		user.PasswordHash,
		input.CurrentPassword,
		input.NewPassword,
		input.ConfirmPassword,
	); err != nil {
		switch {
		case errors.Is(err, services.ErrSettingsPasswordChangeInvalidInput):
			return handler.respondSettingsError(c, fiber.StatusBadRequest, "invalid settings input")
		case errors.Is(err, services.ErrSettingsPasswordMismatch):
			return handler.respondSettingsError(c, fiber.StatusBadRequest, "password mismatch")
		case errors.Is(err, services.ErrSettingsInvalidCurrentPassword):
			return handler.respondSettingsError(c, fiber.StatusUnauthorized, "invalid current password")
		case errors.Is(err, services.ErrSettingsNewPasswordMustDiffer):
			return handler.respondSettingsError(c, fiber.StatusBadRequest, "new password must differ")
		case errors.Is(err, services.ErrSettingsWeakPassword):
			return handler.respondSettingsError(c, fiber.StatusBadRequest, "weak password")
		default:
			return apiError(c, fiber.StatusInternalServerError, "failed to validate password")
		}
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to secure password")
	}

	if err := handler.settingsService.UpdatePassword(user.ID, string(passwordHash), false); err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to update password")
	}

	if acceptsJSON(c) {
		return c.JSON(fiber.Map{"ok": true})
	}
	handler.setFlashCookie(c, FlashPayload{SettingsSuccess: "password_changed"})
	return redirectOrJSON(c, "/settings")
}
