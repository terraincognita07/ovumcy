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

	var existingUsers int64
	if err := handler.db.Model(&models.User{}).
		Where("lower(trim(email)) = ?", credentials.Email).
		Count(&existingUsers).Error; err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to create account")
	}
	if existingUsers > 0 {
		return handler.respondAuthError(c, fiber.StatusConflict, "email already exists")
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
	if err := handler.db.Where("lower(trim(email)) = ?", credentials.Email).First(&user).Error; err != nil {
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
