package api

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/lume/internal/models"
	"github.com/terraincognita07/lume/internal/services"
)

func (handler *Handler) requiresInitialSetup() (bool, error) {
	var usersCount int64
	if err := handler.db.Model(&models.User{}).Count(&usersCount).Error; err != nil {
		return false, err
	}
	return usersCount == 0, nil
}

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
	if user, err := handler.authenticateRequest(c); err == nil {
		return c.Redirect(postLoginRedirectPath(user), fiber.StatusSeeOther)
	}

	needsSetup, err := handler.requiresInitialSetup()
	if err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to load setup state")
	}

	flash := handler.popFlashCookie(c)
	data := fiber.Map{
		"Title":         localizedPageTitle(currentMessages(c), "meta.title.login", "Lume | Login"),
		"ErrorKey":      authErrorKeyFromFlashOrQuery(c, flash.AuthError),
		"Email":         loginEmailFromFlashOrQuery(c, flash.LoginEmail),
		"IsFirstLaunch": needsSetup,
	}
	return handler.render(c, "login", data)
}

func (handler *Handler) ShowRegisterPage(c *fiber.Ctx) error {
	if user, err := handler.authenticateRequest(c); err == nil {
		return c.Redirect(postLoginRedirectPath(user), fiber.StatusSeeOther)
	}

	needsSetup, err := handler.requiresInitialSetup()
	if err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to load setup state")
	}

	flash := handler.popFlashCookie(c)
	data := fiber.Map{
		"Title":         localizedPageTitle(currentMessages(c), "meta.title.register", "Lume | Sign Up"),
		"ErrorKey":      authErrorKeyFromFlashOrQuery(c, flash.AuthError),
		"Email":         loginEmailFromFlashOrQuery(c, flash.RegisterEmail),
		"IsFirstLaunch": needsSetup,
	}
	return handler.render(c, "register", data)
}

func (handler *Handler) ShowForgotPasswordPage(c *fiber.Ctx) error {
	flash := handler.popFlashCookie(c)
	data := fiber.Map{
		"Title":    localizedPageTitle(currentMessages(c), "meta.title.forgot_password", "Lume | Password Recovery"),
		"ErrorKey": authErrorKeyFromFlashOrQuery(c, flash.AuthError),
	}
	return handler.render(c, "forgot_password", data)
}

func (handler *Handler) ShowResetPasswordPage(c *fiber.Ctx) error {
	token := strings.TrimSpace(c.Query("token"))
	flash := handler.popFlashCookie(c)

	invalidToken := false
	if _, err := handler.parsePasswordResetToken(token); err != nil {
		invalidToken = true
	}

	data := fiber.Map{
		"Title":        localizedPageTitle(currentMessages(c), "meta.title.reset_password", "Lume | Reset Password"),
		"Token":        token,
		"InvalidToken": invalidToken,
		"ForcedReset":  parseBoolValue(c.Query("forced")),
		"ErrorKey":     authErrorKeyFromFlashOrQuery(c, flash.AuthError),
	}
	return handler.render(c, "reset_password", data)
}

func (handler *Handler) ShowDashboard(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return c.Redirect("/login", fiber.StatusSeeOther)
	}
	language := currentLanguage(c)
	messages := currentMessages(c)

	now := time.Now().In(handler.location)
	today := dateAtLocation(now, handler.location)
	isOwner := isOwnerUser(user)

	stats, _, err := handler.buildCycleStatsForRange(user, today.AddDate(-2, 0, 0), today, now)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("failed to load logs")
	}
	todayLog, err := handler.fetchLogByDate(user.ID, today)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("failed to load today log")
	}

	symptoms, err := handler.fetchSymptomsForViewer(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("failed to load symptoms")
	}

	todayLog = sanitizeLogForViewer(user, todayLog)

	data := fiber.Map{
		"Title":             localizedPageTitle(messages, "meta.title.dashboard", "Lume | Dashboard"),
		"CurrentUser":       user,
		"Stats":             stats,
		"Today":             today.Format("2006-01-02"),
		"FormattedDate":     localizedDashboardDate(language, today),
		"TodayLog":          todayLog,
		"TodayHasData":      dayHasData(todayLog),
		"Symptoms":          symptoms,
		"SelectedSymptomID": symptomIDSet(todayLog.SymptomIDs),
		"IsOwner":           isOwner,
	}

	return handler.render(c, "dashboard", data)
}

func (handler *Handler) ShowCalendar(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return c.Redirect("/login", fiber.StatusSeeOther)
	}
	messages := currentMessages(c)

	now := time.Now().In(handler.location)
	monthQuery := strings.TrimSpace(c.Query("month"))
	selectedDate := ""
	selectedDayRaw := strings.TrimSpace(c.Query("day"))

	activeMonth, err := parseMonthQuery(monthQuery, now, handler.location)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("invalid month")
	}
	if selectedDayRaw != "" {
		if selectedDay, parseErr := parseDayParam(selectedDayRaw, handler.location); parseErr == nil {
			selectedDate = selectedDay.Format("2006-01-02")
			if monthQuery == "" {
				activeMonth = time.Date(selectedDay.Year(), selectedDay.Month(), 1, 0, 0, 0, 0, handler.location)
			}
		}
	}

	monthStart := activeMonth
	monthEnd := monthStart.AddDate(0, 1, -1)

	logRangeStart := monthStart.AddDate(0, 0, -70)
	logRangeEnd := monthEnd.AddDate(0, 0, 70)
	logs, err := handler.fetchLogsForUser(user.ID, logRangeStart, logRangeEnd)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("failed to load calendar")
	}

	stats, _, err := handler.buildCycleStatsForRange(user, now.AddDate(-2, 0, 0), now, now)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("failed to load stats")
	}

	days := handler.buildCalendarDays(monthStart, logs, stats, now)

	prevMonth := monthStart.AddDate(0, -1, 0).Format("2006-01")
	nextMonth := monthStart.AddDate(0, 1, 0).Format("2006-01")

	data := fiber.Map{
		"Title":        localizedPageTitle(messages, "meta.title.calendar", "Lume | Calendar"),
		"CurrentUser":  user,
		"MonthLabel":   localizedMonthYear(currentLanguage(c), monthStart),
		"MonthValue":   monthStart.Format("2006-01"),
		"PrevMonth":    prevMonth,
		"NextMonth":    nextMonth,
		"SelectedDate": selectedDate,
		"CalendarDays": days,
		"Today":        dateAtLocation(now, handler.location).Format("2006-01-02"),
		"Stats":        stats,
		"IsOwner":      isOwnerUser(user),
	}

	return handler.render(c, "calendar", data)
}

func (handler *Handler) CalendarDayPanel(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).SendString("unauthorized")
	}

	day, err := parseDayParam(c.Params("date"), handler.location)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("invalid date")
	}

	return handler.renderDayEditorPartial(c, user, day)
}

func (handler *Handler) renderDayEditorPartial(c *fiber.Ctx, user *models.User, day time.Time) error {
	messages := currentMessages(c)

	logEntry, err := handler.fetchLogByDate(user.ID, day)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("failed to load day")
	}
	hasDayData, err := handler.dayHasDataForDate(user.ID, day)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("failed to load day state")
	}

	isOwner := isOwnerUser(user)
	symptoms, err := handler.fetchSymptomsForViewer(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("failed to load symptoms")
	}
	logEntry = sanitizeLogForViewer(user, logEntry)

	payload := fiber.Map{
		"Date":              day,
		"DateString":        day.Format("2006-01-02"),
		"DateLabel":         localizedDateLabel(currentLanguage(c), day),
		"IsFutureDate":      day.After(dateAtLocation(time.Now().In(handler.location), handler.location)),
		"NoDataLabel":       translateMessage(messages, "common.not_available"),
		"Log":               logEntry,
		"Symptoms":          symptoms,
		"SelectedSymptomID": symptomIDSet(logEntry.SymptomIDs),
		"HasDayData":        hasDayData,
		"IsOwner":           isOwner,
	}
	return handler.renderPartial(c, "day_editor_partial", payload)
}

func (handler *Handler) ShowStats(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return c.Redirect("/login", fiber.StatusSeeOther)
	}
	language := currentLanguage(c)
	messages := currentMessages(c)

	now := time.Now().In(handler.location)
	stats, logs, err := handler.buildCycleStatsForRange(user, now.AddDate(-2, 0, 0), now, now)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("failed to load stats")
	}
	lengths := handler.completedCycleTrendLengths(logs, now)
	if len(lengths) > 12 {
		lengths = lengths[len(lengths)-12:]
	}

	labels := buildCycleTrendLabels(messages, len(lengths))
	baselineCycleLength := ownerBaselineCycleLength(user)

	chartPayload := fiber.Map{
		"labels": labels,
		"values": lengths,
	}
	if baselineCycleLength > 0 {
		chartPayload["baseline"] = baselineCycleLength
	}

	symptomCounts := make([]SymptomCount, 0)
	if user.Role == models.RoleOwner {
		symptomLogs, loadErr := handler.fetchAllLogsForUser(user.ID)
		if loadErr != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("failed to load symptom logs")
		}
		symptomCounts, err = handler.calculateSymptomFrequencies(user.ID, symptomLogs)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("failed to load symptom stats")
		}
		localizeSymptomFrequencySummaries(language, symptomCounts)
	}

	data := fiber.Map{
		"Title":           localizedPageTitle(messages, "meta.title.stats", "Lume | Stats"),
		"CurrentUser":     user,
		"Stats":           stats,
		"ChartData":       chartPayload,
		"ChartBaseline":   baselineCycleLength,
		"TrendPointCount": len(lengths),
		"SymptomCounts":   symptomCounts,
		"IsOwner":         isOwnerUser(user),
	}

	return handler.render(c, "stats", data)
}

func (handler *Handler) completedCycleTrendLengths(logs []models.DailyLog, now time.Time) []int {
	starts := services.DetectCycleStarts(logs)
	if len(starts) < 2 {
		return nil
	}

	today := dateAtLocation(now, handler.location)
	lengths := make([]int, 0, len(starts)-1)
	for index := 1; index < len(starts); index++ {
		previousStart := dateAtLocation(starts[index-1], handler.location)
		currentStart := dateAtLocation(starts[index], handler.location)
		if !currentStart.Before(today) {
			break
		}
		lengths = append(lengths, int(currentStart.Sub(previousStart).Hours()/24))
	}

	return lengths
}

func (handler *Handler) ShowSettings(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return c.Redirect("/login", fiber.StatusSeeOther)
	}

	flash := handler.popFlashCookie(c)

	data, err := handler.buildSettingsViewData(c, user, flash)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("failed to load settings")
	}
	return handler.render(c, "settings", data)
}

func (handler *Handler) ShowPrivacyPage(c *fiber.Ctx) error {
	messages := currentMessages(c)

	metaDescription := translateMessage(messages, "meta.description.privacy")
	if metaDescription == "meta.description.privacy" {
		metaDescription = "Lume Privacy Policy - Zero data collection, self-hosted period tracker."
	}
	backFallback := "/login"

	data := fiber.Map{
		"Title":           localizedPageTitle(messages, "meta.title.privacy", "Lume | Privacy Policy"),
		"MetaDescription": metaDescription,
	}

	if user, err := handler.authenticateRequest(c); err == nil {
		data["CurrentUser"] = user
		backFallback = "/dashboard"
	}
	data["BackPath"] = sanitizeRedirectPath(c.Query("back"), backFallback)

	return handler.render(c, "privacy", data)
}
