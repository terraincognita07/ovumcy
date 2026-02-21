package api

import "github.com/gofiber/fiber/v2"

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
