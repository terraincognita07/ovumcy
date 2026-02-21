package api

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/lume/internal/models"
)

func (handler *Handler) SetupStatus(c *fiber.Ctx) error {
	needsSetup, err := handler.requiresInitialSetup()
	if err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to load setup state")
	}
	return c.JSON(fiber.Map{"needs_setup": needsSetup})
}

func (handler *Handler) SetLanguage(c *fiber.Ctx) error {
	language := handler.i18n.NormalizeLanguage(c.Params("lang"))
	handler.setLanguageCookie(c, language)

	nextPath := sanitizeRedirectPath(c.Query("next"), "/")
	if isHTMX(c) {
		c.Set("HX-Redirect", nextPath)
		return c.SendStatus(fiber.StatusOK)
	}
	return c.Redirect(nextPath, fiber.StatusSeeOther)
}

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

func (handler *Handler) ShowDashboard(c *fiber.Ctx) error {
	user, handled, err := handler.currentUserOrRedirectToLogin(c)
	if err != nil {
		return err
	}
	if handled {
		return nil
	}
	language, messages, now := handler.currentPageViewContext(c)
	data, errorMessage, err := handler.buildDashboardViewData(user, language, messages, now)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(errorMessage)
	}

	return handler.render(c, "dashboard", data)
}

func (handler *Handler) ShowCalendar(c *fiber.Ctx) error {
	user, handled, err := handler.currentUserOrRedirectToLogin(c)
	if err != nil {
		return err
	}
	if handled {
		return nil
	}
	language, messages, now := handler.currentPageViewContext(c)
	activeMonth, selectedDate, err := resolveCalendarMonthAndSelectedDate(c.Query("month"), c.Query("day"), now, handler.location)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("invalid month")
	}
	data, errorMessage, err := handler.buildCalendarViewData(user, language, messages, now, activeMonth, selectedDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(errorMessage)
	}

	return handler.render(c, "calendar", data)
}

func (handler *Handler) CalendarDayPanel(c *fiber.Ctx) error {
	user, handled, err := currentUserOrUnauthorized(c)
	if err != nil {
		return err
	}
	if handled {
		return nil
	}

	day, err := parseDayParam(c.Params("date"), handler.location)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("invalid date")
	}

	return handler.renderDayEditorPartial(c, user, day)
}

func (handler *Handler) renderDayEditorPartial(c *fiber.Ctx, user *models.User, day time.Time) error {
	language, messages, now := handler.currentPageViewContext(c)
	payload, errorMessage, err := handler.buildDayEditorPartialData(user, language, messages, day, now)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(errorMessage)
	}
	return handler.renderPartial(c, "day_editor_partial", payload)
}

func (handler *Handler) ShowStats(c *fiber.Ctx) error {
	user, handled, err := handler.currentUserOrRedirectToLogin(c)
	if err != nil {
		return err
	}
	if handled {
		return nil
	}
	language, messages, now := handler.currentPageViewContext(c)
	data, errorMessage, err := handler.buildStatsPageData(user, language, messages, now)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(errorMessage)
	}

	return handler.render(c, "stats", data)
}

func (handler *Handler) ShowSettings(c *fiber.Ctx) error {
	user, handled, err := handler.currentUserOrRedirectToLogin(c)
	if err != nil {
		return err
	}
	if handled {
		return nil
	}

	data, errorMessage, err := handler.buildSettingsPageData(c, user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(errorMessage)
	}
	return handler.render(c, "settings", data)
}

func (handler *Handler) ShowPrivacyPage(c *fiber.Ctx) error {
	messages := currentMessages(c)
	authenticatedUser := handler.optionalAuthenticatedUser(c)
	data := buildPrivacyPageData(messages, c.Query("back"), authenticatedUser)
	return handler.render(c, "privacy", data)
}
