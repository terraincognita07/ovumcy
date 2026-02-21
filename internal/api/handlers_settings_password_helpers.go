package api

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/lume/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func parseChangePasswordInput(c *fiber.Ctx) (changePasswordInput, string) {
	input := changePasswordInput{}
	if err := c.BodyParser(&input); err != nil {
		return changePasswordInput{}, "invalid settings input"
	}

	input.CurrentPassword = strings.TrimSpace(input.CurrentPassword)
	input.NewPassword = strings.TrimSpace(input.NewPassword)
	input.ConfirmPassword = strings.TrimSpace(input.ConfirmPassword)
	if input.CurrentPassword == "" || input.NewPassword == "" || input.ConfirmPassword == "" {
		return changePasswordInput{}, "invalid settings input"
	}
	if input.NewPassword != input.ConfirmPassword {
		return changePasswordInput{}, "password mismatch"
	}
	return input, ""
}

func validateChangePasswordInput(input changePasswordInput, user *models.User) (int, string) {
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.CurrentPassword)) != nil {
		return fiber.StatusUnauthorized, "invalid current password"
	}
	if input.CurrentPassword == input.NewPassword {
		return fiber.StatusBadRequest, "new password must differ"
	}
	if err := validatePasswordStrength(input.NewPassword); err != nil {
		return fiber.StatusBadRequest, "weak password"
	}
	return 0, ""
}
