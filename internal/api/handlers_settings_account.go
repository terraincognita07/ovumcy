package api

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/lume/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func (handler *Handler) ChangePassword(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return apiError(c, fiber.StatusUnauthorized, "unauthorized")
	}

	input := changePasswordInput{}
	if err := c.BodyParser(&input); err != nil {
		return handler.respondSettingsError(c, fiber.StatusBadRequest, "invalid settings input")
	}

	input.CurrentPassword = strings.TrimSpace(input.CurrentPassword)
	input.NewPassword = strings.TrimSpace(input.NewPassword)
	input.ConfirmPassword = strings.TrimSpace(input.ConfirmPassword)
	if input.CurrentPassword == "" || input.NewPassword == "" || input.ConfirmPassword == "" {
		return handler.respondSettingsError(c, fiber.StatusBadRequest, "invalid settings input")
	}
	if input.NewPassword != input.ConfirmPassword {
		return handler.respondSettingsError(c, fiber.StatusBadRequest, "password mismatch")
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.CurrentPassword)) != nil {
		return handler.respondSettingsError(c, fiber.StatusUnauthorized, "invalid current password")
	}
	if input.CurrentPassword == input.NewPassword {
		return handler.respondSettingsError(c, fiber.StatusBadRequest, "new password must differ")
	}
	if err := validatePasswordStrength(input.NewPassword); err != nil {
		return handler.respondSettingsError(c, fiber.StatusBadRequest, "weak password")
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

func (handler *Handler) UpdateCycleSettings(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return apiError(c, fiber.StatusUnauthorized, "unauthorized")
	}

	input := cycleSettingsInput{}
	if err := c.BodyParser(&input); err != nil {
		return handler.respondSettingsError(c, fiber.StatusBadRequest, "invalid settings input")
	}
	if !isValidOnboardingCycleLength(input.CycleLength) {
		return handler.respondSettingsError(c, fiber.StatusBadRequest, "cycle length must be between 21 and 50")
	}
	if !isValidOnboardingPeriodLength(input.PeriodLength) {
		return handler.respondSettingsError(c, fiber.StatusBadRequest, "period length must be between 2 and 7")
	}

	if err := handler.db.Model(&models.User{}).Where("id = ?", user.ID).Updates(map[string]any{
		"cycle_length":  input.CycleLength,
		"period_length": input.PeriodLength,
	}).Error; err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to update cycle settings")
	}

	user.CycleLength = input.CycleLength
	user.PeriodLength = input.PeriodLength

	if acceptsJSON(c) {
		return c.JSON(fiber.Map{"ok": true})
	}
	handler.setFlashCookie(c, FlashPayload{SettingsSuccess: "cycle_updated"})
	return redirectOrJSON(c, "/settings")
}

func (handler *Handler) RegenerateRecoveryCode(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return apiError(c, fiber.StatusUnauthorized, "unauthorized")
	}

	recoveryCode, recoveryHash, err := generateRecoveryCodeHash()
	if err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to create recovery code")
	}

	if err := handler.db.Model(&models.User{}).Where("id = ?", user.ID).Update("recovery_code_hash", recoveryHash).Error; err != nil {
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

func (handler *Handler) DeleteAccount(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return apiError(c, fiber.StatusUnauthorized, "unauthorized")
	}

	input := deleteAccountInput{}
	if err := c.BodyParser(&input); err != nil && acceptsJSON(c) {
		return handler.respondSettingsError(c, fiber.StatusBadRequest, "invalid password")
	}

	input.Password = strings.TrimSpace(input.Password)
	if input.Password == "" {
		input.Password = strings.TrimSpace(c.FormValue("password"))
	}
	if input.Password == "" {
		return handler.respondSettingsError(c, fiber.StatusBadRequest, "invalid password")
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)) != nil {
		return handler.respondSettingsError(c, fiber.StatusUnauthorized, "invalid password")
	}

	if err := handler.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", user.ID).Delete(&models.DailyLog{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", user.ID).Delete(&models.SymptomType{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&models.User{}, user.ID).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to delete account")
	}

	handler.clearAuthCookie(c)
	if acceptsJSON(c) {
		return c.JSON(fiber.Map{"ok": true})
	}
	return redirectOrJSON(c, "/login")
}
