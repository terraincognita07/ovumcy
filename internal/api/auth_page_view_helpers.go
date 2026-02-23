package api

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

func authErrorKeyFromFlashOrQuery(c *fiber.Ctx, flashAuthError string) string {
	errorSource := strings.TrimSpace(flashAuthError)
	if errorSource == "" {
		errorSource = strings.TrimSpace(c.Query("error"))
	}
	return authErrorTranslationKey(errorSource)
}

func loginEmailFromFlashOrQuery(c *fiber.Ctx, flashEmail string) string {
	email := normalizeLoginEmail(flashEmail)
	if email == "" {
		email = normalizeLoginEmail(c.Query("email"))
	}
	return email
}

func buildLoginPageData(c *fiber.Ctx, messages map[string]string, flash FlashPayload, needsSetup bool) fiber.Map {
	return fiber.Map{
		"Title":         localizedPageTitle(messages, "meta.title.login", "Ovumcy | Login"),
		"ErrorKey":      authErrorKeyFromFlashOrQuery(c, flash.AuthError),
		"Email":         loginEmailFromFlashOrQuery(c, flash.LoginEmail),
		"IsFirstLaunch": needsSetup,
	}
}

func buildRegisterPageData(c *fiber.Ctx, messages map[string]string, flash FlashPayload, needsSetup bool) fiber.Map {
	return fiber.Map{
		"Title":         localizedPageTitle(messages, "meta.title.register", "Ovumcy | Sign Up"),
		"ErrorKey":      authErrorKeyFromFlashOrQuery(c, flash.AuthError),
		"Email":         loginEmailFromFlashOrQuery(c, flash.RegisterEmail),
		"IsFirstLaunch": needsSetup,
	}
}

func buildForgotPasswordPageData(c *fiber.Ctx, messages map[string]string, flash FlashPayload) fiber.Map {
	return fiber.Map{
		"Title":    localizedPageTitle(messages, "meta.title.forgot_password", "Ovumcy | Password Recovery"),
		"ErrorKey": authErrorKeyFromFlashOrQuery(c, flash.AuthError),
	}
}

func (handler *Handler) buildResetPasswordPageData(c *fiber.Ctx, messages map[string]string, flash FlashPayload) fiber.Map {
	token := strings.TrimSpace(c.Query("token"))
	invalidToken := false
	if _, err := handler.parsePasswordResetToken(token); err != nil {
		invalidToken = true
	}

	return fiber.Map{
		"Title":        localizedPageTitle(messages, "meta.title.reset_password", "Ovumcy | Reset Password"),
		"Token":        token,
		"InvalidToken": invalidToken,
		"ForcedReset":  parseBoolValue(c.Query("forced")),
		"ErrorKey":     authErrorKeyFromFlashOrQuery(c, flash.AuthError),
	}
}
