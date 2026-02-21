package api

import "github.com/gofiber/fiber/v2"

func (handler *Handler) ShowLoginPage(c *fiber.Ctx) error {
	redirected, err := handler.redirectAuthenticatedUserIfPresent(c)
	if err != nil {
		return err
	}
	if redirected {
		return nil
	}

	needsSetup, err := handler.requiresInitialSetup()
	if err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to load setup state")
	}

	flash := handler.popFlashCookie(c)
	data := buildLoginPageData(c, currentMessages(c), flash, needsSetup)
	return handler.render(c, "login", data)
}

func (handler *Handler) ShowRegisterPage(c *fiber.Ctx) error {
	redirected, err := handler.redirectAuthenticatedUserIfPresent(c)
	if err != nil {
		return err
	}
	if redirected {
		return nil
	}

	needsSetup, err := handler.requiresInitialSetup()
	if err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to load setup state")
	}

	flash := handler.popFlashCookie(c)
	data := buildRegisterPageData(c, currentMessages(c), flash, needsSetup)
	return handler.render(c, "register", data)
}

func (handler *Handler) ShowForgotPasswordPage(c *fiber.Ctx) error {
	flash := handler.popFlashCookie(c)
	data := buildForgotPasswordPageData(c, currentMessages(c), flash)
	return handler.render(c, "forgot_password", data)
}

func (handler *Handler) ShowResetPasswordPage(c *fiber.Ctx) error {
	flash := handler.popFlashCookie(c)
	data := handler.buildResetPasswordPageData(c, currentMessages(c), flash)
	return handler.render(c, "reset_password", data)
}
