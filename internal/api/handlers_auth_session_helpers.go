package api

import (
	"errors"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/lume/internal/models"
)

func validateRegistrationCredentials(credentials credentialsInput) string {
	if credentials.ConfirmPassword == "" {
		return "invalid input"
	}
	if credentials.Password != credentials.ConfirmPassword {
		return "password mismatch"
	}
	if err := validatePasswordStrength(credentials.Password); err != nil {
		return "weak password"
	}
	return ""
}

func parseForgotPasswordCode(c *fiber.Ctx) (string, string) {
	input := forgotPasswordInput{}
	if err := c.BodyParser(&input); err != nil {
		return "", "invalid input"
	}

	code := normalizeRecoveryCode(input.RecoveryCode)
	if !recoveryCodeRegex.MatchString(code) {
		return "", "invalid recovery code"
	}
	return code, ""
}

func parseResetPasswordInput(c *fiber.Ctx) (resetPasswordInput, string) {
	input := resetPasswordInput{}
	if err := c.BodyParser(&input); err != nil {
		return resetPasswordInput{}, "invalid input"
	}

	input.Token = strings.TrimSpace(input.Token)
	input.Password = strings.TrimSpace(input.Password)
	input.ConfirmPassword = strings.TrimSpace(input.ConfirmPassword)
	if input.Token == "" || input.Password == "" || input.ConfirmPassword == "" {
		return resetPasswordInput{}, "invalid input"
	}
	if input.Password != input.ConfirmPassword {
		return resetPasswordInput{}, "password mismatch"
	}
	if err := validatePasswordStrength(input.Password); err != nil {
		return resetPasswordInput{}, "weak password"
	}

	return input, ""
}

func buildResetPasswordPath(token string, forced bool) string {
	path := "/reset-password?token=" + url.QueryEscape(token)
	if forced {
		return path + "&forced=1"
	}
	return path
}

func redirectToPath(c *fiber.Ctx, path string) error {
	if isHTMX(c) {
		c.Set("HX-Redirect", path)
		return c.SendStatus(fiber.StatusOK)
	}
	return c.Redirect(path, fiber.StatusSeeOther)
}

func (handler *Handler) lookupUserByResetToken(token string) (*models.User, error) {
	userID, err := handler.parsePasswordResetToken(token)
	if err != nil {
		return nil, errors.New("invalid reset token")
	}

	var user models.User
	if err := handler.db.First(&user, userID).Error; err != nil {
		return nil, errors.New("invalid reset token")
	}
	return &user, nil
}

func (handler *Handler) renderRecoveryCodeResponse(c *fiber.Ctx, user *models.User, recoveryCode string, status int) error {
	if acceptsJSON(c) {
		return c.Status(status).JSON(fiber.Map{
			"ok":            true,
			"recovery_code": recoveryCode,
		})
	}

	continuePath := "/dashboard"
	userID := uint(0)
	if user != nil {
		userID = user.ID
		continuePath = postLoginRedirectPath(user)
	}
	handler.setRecoveryCodePageCookie(c, userID, recoveryCode, continuePath)

	return redirectToPath(c, "/recovery-code")
}
