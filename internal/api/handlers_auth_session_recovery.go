package api

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
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

	token, err := handler.buildPasswordResetToken(user.ID, 30*time.Minute)
	if err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to create reset token")
	}
	handler.recoveryLimiter.reset(limiterKey)

	if acceptsJSON(c) {
		return c.JSON(fiber.Map{
			"ok":          true,
			"reset_token": token,
		})
	}

	return redirectToPath(c, buildResetPasswordPath(token, false))
}

func (handler *Handler) ResetPassword(c *fiber.Ctx) error {
	input, parseError := parseResetPasswordInput(c)
	if parseError != "" {
		return handler.respondAuthError(c, fiber.StatusBadRequest, parseError)
	}

	user, err := handler.lookupUserByResetToken(input.Token)
	if err != nil {
		return handler.respondAuthError(c, fiber.StatusBadRequest, "invalid reset token")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to secure password")
	}
	recoveryCode, recoveryHash, err := generateRecoveryCodeHash()
	if err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to create recovery code")
	}

	user.PasswordHash = string(passwordHash)
	user.RecoveryCodeHash = recoveryHash
	user.MustChangePassword = false
	handler.ensureDependencies()
	if err := handler.authService.SaveUser(user); err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to reset password")
	}

	if err := handler.setAuthCookie(c, user, true); err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to create session")
	}

	return handler.renderRecoveryCodeResponse(c, user, recoveryCode, fiber.StatusOK)
}
