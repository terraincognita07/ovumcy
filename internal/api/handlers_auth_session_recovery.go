package api

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/ovumcy/internal/services"
)

func (handler *Handler) ForgotPassword(c *fiber.Ctx) error {
	const recoveryAttemptsLimit = 8
	const recoveryAttemptsWindow = 15 * time.Minute

	now := time.Now().In(handler.location)
	limiterKey := requestLimiterKey(c)
	if handler.recoveryLimiter.tooManyRecent(limiterKey, now, recoveryAttemptsLimit, recoveryAttemptsWindow) {
		return handler.respondAuthError(c, fiber.StatusTooManyRequests, "too many recovery attempts")
	}

	code, parseError := parseForgotPasswordCode(c)
	if parseError != "" {
		handler.recoveryLimiter.addFailure(limiterKey, now, recoveryAttemptsWindow)
		return handler.respondAuthError(c, fiber.StatusBadRequest, parseError)
	}

	user, err := handler.findUserByRecoveryCode(code)
	if err != nil {
		handler.recoveryLimiter.addFailure(limiterKey, now, recoveryAttemptsWindow)
		return handler.respondAuthError(c, fiber.StatusBadRequest, "invalid recovery code")
	}

	token, err := handler.buildPasswordResetToken(user.ID, user.PasswordHash, 30*time.Minute)
	if err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to create reset token")
	}
	handler.setResetPasswordCookie(c, token, false)
	handler.recoveryLimiter.reset(limiterKey)

	if acceptsJSON(c) {
		return c.JSON(fiber.Map{
			"ok": true,
		})
	}

	return redirectToPath(c, buildResetPasswordPath())
}

func (handler *Handler) ResetPassword(c *fiber.Ctx) error {
	input, parseError := parseResetPasswordInput(c)
	if parseError != "" {
		return handler.respondAuthError(c, fiber.StatusBadRequest, parseError)
	}

	handler.ensureDependencies()
	if err := handler.authService.ValidateResetPasswordInput(input.Password, input.ConfirmPassword); err != nil {
		switch {
		case errors.Is(err, services.ErrAuthPasswordMismatch):
			return handler.respondAuthError(c, fiber.StatusBadRequest, "password mismatch")
		case errors.Is(err, services.ErrAuthWeakPassword):
			return handler.respondAuthError(c, fiber.StatusBadRequest, "weak password")
		default:
			return handler.respondAuthError(c, fiber.StatusBadRequest, "invalid input")
		}
	}

	token, _ := handler.readResetPasswordCookie(c)
	if token == "" {
		handler.clearResetPasswordCookie(c)
		return handler.respondAuthError(c, fiber.StatusBadRequest, "invalid reset token")
	}

	user, err := handler.lookupUserByResetToken(token)
	if err != nil {
		handler.clearResetPasswordCookie(c)
		return handler.respondAuthError(c, fiber.StatusBadRequest, "invalid reset token")
	}

	recoveryCode, err := handler.authService.ResetPasswordAndRotateRecoveryCode(user, input.Password)
	if err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to reset password")
	}

	if err := handler.setAuthCookie(c, user, true); err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to create session")
	}
	handler.clearResetPasswordCookie(c)

	return handler.renderRecoveryCodeResponse(c, user, recoveryCode, fiber.StatusOK)
}
