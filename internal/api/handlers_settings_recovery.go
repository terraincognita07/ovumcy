package api

import (
	"github.com/gofiber/fiber/v2"
)

func (handler *Handler) RegenerateRecoveryCode(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return apiError(c, fiber.StatusUnauthorized, "unauthorized")
	}

	recoveryCode, recoveryHash, err := generateRecoveryCodeHash()
	if err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to create recovery code")
	}

	handler.ensureDependencies()
	if err := handler.settingsService.UpdateRecoveryCodeHash(user.ID, recoveryHash); err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to update recovery code")
	}

	if acceptsJSON(c) {
		return c.JSON(fiber.Map{
			"ok":            true,
			"recovery_code": recoveryCode,
		})
	}

	data, err := handler.buildSettingsViewData(c, user, FlashPayload{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("failed to load settings")
	}
	data["SuccessKey"] = "settings.success.recovery_code_regenerated"
	data["GeneratedRecoveryCode"] = recoveryCode
	return handler.render(c, "settings", data)
}
