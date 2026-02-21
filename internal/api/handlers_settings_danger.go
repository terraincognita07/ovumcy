package api

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/lume/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func (handler *Handler) ClearAllData(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return apiError(c, fiber.StatusUnauthorized, "unauthorized")
	}

	if err := handler.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", user.ID).Delete(&models.DailyLog{}).Error; err != nil {
			return err
		}
		return tx.Model(&models.User{}).Where("id = ?", user.ID).Updates(map[string]any{
			"cycle_length":      models.DefaultCycleLength,
			"period_length":     models.DefaultPeriodLength,
			"auto_period_fill":  true,
			"last_period_start": nil,
		}).Error
	}); err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to clear data")
	}

	user.CycleLength = models.DefaultCycleLength
	user.PeriodLength = models.DefaultPeriodLength
	user.AutoPeriodFill = true
	user.LastPeriodStart = nil

	if acceptsJSON(c) {
		return c.JSON(fiber.Map{"ok": true})
	}
	handler.setFlashCookie(c, FlashPayload{SettingsSuccess: "data_cleared"})
	return redirectOrJSON(c, "/settings")
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
