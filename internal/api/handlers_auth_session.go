package api

import (
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
	if validationError := validateRegistrationCredentials(credentials); validationError != "" {
		return handler.respondAuthError(c, fiber.StatusBadRequest, validationError)
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
		CycleLength:      models.DefaultCycleLength,
		PeriodLength:     models.DefaultPeriodLength,
		AutoPeriodFill:   true,
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

	return handler.renderRecoveryCodeResponse(c, &user, recoveryCode, fiber.StatusCreated)
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
		path := buildResetPasswordPath(token, true)
		if acceptsJSON(c) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":       "password change required",
				"reset_token": token,
			})
		}
		return redirectToPath(c, path)
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
	if err := handler.db.Save(user).Error; err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to reset password")
	}

	if err := handler.setAuthCookie(c, user, true); err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to create session")
	}

	return handler.renderRecoveryCodeResponse(c, user, recoveryCode, fiber.StatusOK)
}
