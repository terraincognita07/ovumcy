package api

import (
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/lume/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func (handler *Handler) Register(c *fiber.Ctx) error {
	credentials, err := parseCredentials(c)
	if err != nil {
		return handler.respondAuthError(c, fiber.StatusBadRequest, "invalid input")
	}
	if credentials.ConfirmPassword == "" {
		return handler.respondAuthError(c, fiber.StatusBadRequest, "invalid input")
	}
	if credentials.Password != credentials.ConfirmPassword {
		return handler.respondAuthError(c, fiber.StatusBadRequest, "password mismatch")
	}
	if err := validatePasswordStrength(credentials.Password); err != nil {
		return handler.respondAuthError(c, fiber.StatusBadRequest, "weak password")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(credentials.Password), bcrypt.DefaultCost)
	if err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to secure password")
	}

	recoveryCode, recoveryHash, err := generateRecoveryCodeHash()
	if err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to create recovery code")
	}

	user := models.User{
		Email:            credentials.Email,
		PasswordHash:     string(passwordHash),
		RecoveryCodeHash: recoveryHash,
		Role:             models.RoleOwner,
		CreatedAt:        time.Now().In(handler.location),
	}
	if err := handler.db.Create(&user).Error; err != nil {
		return handler.respondAuthError(c, fiber.StatusConflict, "email already exists")
	}

	if err := handler.seedBuiltinSymptoms(user.ID); err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to seed symptoms")
	}

	if err := handler.setAuthCookie(c, &user, true); err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to create session")
	}

	if acceptsJSON(c) {
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"ok":            true,
			"recovery_code": recoveryCode,
		})
	}

	return handler.render(c, "recovery_code", fiber.Map{
		"Title":        localizedPageTitle(currentMessages(c), "meta.title.recovery_code", "Lume | Recovery Code"),
		"RecoveryCode": recoveryCode,
		"ContinuePath": postLoginRedirectPath(&user),
	})
}

func (handler *Handler) Login(c *fiber.Ctx) error {
	credentials, err := parseCredentials(c)
	if err != nil {
		return handler.respondAuthError(c, fiber.StatusBadRequest, "invalid input")
	}

	var user models.User
	if err := handler.db.Where("email = ?", credentials.Email).First(&user).Error; err != nil {
		return handler.respondAuthError(c, fiber.StatusUnauthorized, "invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(credentials.Password)); err != nil {
		return handler.respondAuthError(c, fiber.StatusUnauthorized, "invalid credentials")
	}

	if user.MustChangePassword {
		token, err := handler.buildPasswordResetToken(user.ID, 30*time.Minute)
		if err != nil {
			return apiError(c, fiber.StatusInternalServerError, "failed to create reset token")
		}
		path := "/reset-password?token=" + url.QueryEscape(token) + "&forced=1"
		if acceptsJSON(c) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":       "password change required",
				"reset_token": token,
			})
		}
		if isHTMX(c) {
			c.Set("HX-Redirect", path)
			return c.SendStatus(fiber.StatusOK)
		}
		return c.Redirect(path, fiber.StatusSeeOther)
	}

	if err := handler.setAuthCookie(c, &user, credentials.RememberMe); err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to create session")
	}

	return redirectOrJSON(c, postLoginRedirectPath(&user))
}

func (handler *Handler) Logout(c *fiber.Ctx) error {
	handler.clearAuthCookie(c)
	if isHTMX(c) {
		c.Set("HX-Redirect", "/login")
		return c.SendStatus(fiber.StatusOK)
	}
	if acceptsJSON(c) {
		return c.JSON(fiber.Map{"ok": true})
	}
	return c.Redirect("/login", fiber.StatusSeeOther)
}

func (handler *Handler) ForgotPassword(c *fiber.Ctx) error {
	const recoveryAttemptsLimit = 8
	const recoveryAttemptsWindow = 15 * time.Minute

	now := time.Now().In(handler.location)
	limiterKey := requestLimiterKey(c)
	if handler.recoveryLimiter.tooManyRecent(limiterKey, now, recoveryAttemptsLimit, recoveryAttemptsWindow) {
		return handler.respondAuthError(c, fiber.StatusTooManyRequests, "too many recovery attempts")
	}

	input := forgotPasswordInput{}
	if err := c.BodyParser(&input); err != nil {
		handler.recoveryLimiter.addFailure(limiterKey, now, recoveryAttemptsWindow)
		return handler.respondAuthError(c, fiber.StatusBadRequest, "invalid input")
	}

	code := normalizeRecoveryCode(input.RecoveryCode)
	if !recoveryCodeRegex.MatchString(code) {
		handler.recoveryLimiter.addFailure(limiterKey, now, recoveryAttemptsWindow)
		return handler.respondAuthError(c, fiber.StatusBadRequest, "invalid recovery code")
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

	path := "/reset-password?token=" + url.QueryEscape(token)
	if isHTMX(c) {
		c.Set("HX-Redirect", path)
		return c.SendStatus(fiber.StatusOK)
	}
	return c.Redirect(path, fiber.StatusSeeOther)
}

func (handler *Handler) ResetPassword(c *fiber.Ctx) error {
	input := resetPasswordInput{}
	if err := c.BodyParser(&input); err != nil {
		return handler.respondAuthError(c, fiber.StatusBadRequest, "invalid input")
	}

	input.Token = strings.TrimSpace(input.Token)
	input.Password = strings.TrimSpace(input.Password)
	input.ConfirmPassword = strings.TrimSpace(input.ConfirmPassword)
	if input.Token == "" || input.Password == "" || input.ConfirmPassword == "" {
		return handler.respondAuthError(c, fiber.StatusBadRequest, "invalid input")
	}
	if input.Password != input.ConfirmPassword {
		return handler.respondAuthError(c, fiber.StatusBadRequest, "password mismatch")
	}
	if err := validatePasswordStrength(input.Password); err != nil {
		return handler.respondAuthError(c, fiber.StatusBadRequest, "weak password")
	}

	userID, err := handler.parsePasswordResetToken(input.Token)
	if err != nil {
		return handler.respondAuthError(c, fiber.StatusBadRequest, "invalid reset token")
	}

	var user models.User
	if err := handler.db.First(&user, userID).Error; err != nil {
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
	if err := handler.db.Save(&user).Error; err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to reset password")
	}

	if err := handler.setAuthCookie(c, &user, true); err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to create session")
	}

	if acceptsJSON(c) {
		return c.JSON(fiber.Map{
			"ok":            true,
			"recovery_code": recoveryCode,
		})
	}

	return handler.render(c, "recovery_code", fiber.Map{
		"Title":        localizedPageTitle(currentMessages(c), "meta.title.recovery_code", "Lume | Recovery Code"),
		"RecoveryCode": recoveryCode,
		"ContinuePath": postLoginRedirectPath(&user),
	})
}
