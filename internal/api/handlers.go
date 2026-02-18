package api

import (
	"bytes"
	"crypto/rand"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/mail"
	"net/url"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/terraincognita07/lume/internal/i18n"
	"github.com/terraincognita07/lume/internal/models"
	"github.com/terraincognita07/lume/internal/services"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var hexColorRegex = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
var recoveryCodeRegex = regexp.MustCompile(`^LUME-[A-Z0-9]{4}-[A-Z0-9]{4}-[A-Z0-9]{4}$`)

type Handler struct {
	db              *gorm.DB
	secretKey       []byte
	location        *time.Location
	lutealPhaseDays int
	i18n            *i18n.Manager
	templates       map[string]*template.Template
	partials        map[string]*template.Template
	recoveryLimiter *attemptLimiter
}

type CalendarDay struct {
	Date         time.Time
	DateString   string
	Day          int
	InMonth      bool
	IsToday      bool
	IsPeriod     bool
	IsPredicted  bool
	IsFertility  bool
	IsOvulation  bool
	HasData      bool
	CellClass    string
	TextClass    string
	BadgeClass   string
	OvulationDot bool
}

type SymptomCount struct {
	Name             string
	Icon             string
	Count            int
	TotalDays        int
	FrequencySummary string
}

type credentialsInput struct {
	Email    string `json:"email" form:"email"`
	Password string `json:"password" form:"password"`
}

type dayPayload struct {
	IsPeriod   bool   `json:"is_period"`
	Flow       string `json:"flow"`
	SymptomIDs []uint `json:"symptom_ids"`
	Notes      string `json:"notes"`
}

type symptomPayload struct {
	Name  string `json:"name" form:"name"`
	Icon  string `json:"icon" form:"icon"`
	Color string `json:"color" form:"color"`
}

type forgotPasswordInput struct {
	RecoveryCode string `json:"recovery_code" form:"recovery_code"`
}

type resetPasswordInput struct {
	Token           string `json:"token" form:"token"`
	Password        string `json:"password" form:"password"`
	ConfirmPassword string `json:"confirm_password" form:"confirm_password"`
}

type changePasswordInput struct {
	CurrentPassword string `json:"current_password" form:"current_password"`
	NewPassword     string `json:"new_password" form:"new_password"`
}

type cycleSettingsInput struct {
	CycleLength  int `json:"cycle_length" form:"cycle_length"`
	PeriodLength int `json:"period_length" form:"period_length"`
}

type deleteAccountInput struct {
	Password string `json:"password" form:"password"`
}

type exportSymptomFlags struct {
	Cramps           bool
	Headache         bool
	Acne             bool
	Mood             bool
	Bloating         bool
	Fatigue          bool
	BreastTenderness bool
	BackPain         bool
	Nausea           bool
	Spotting         bool
	Irritability     bool
	Insomnia         bool
	FoodCravings     bool
	Diarrhea         bool
	Constipation     bool
}

type exportJSONSymptomFlags struct {
	Cramps           bool `json:"cramps"`
	Headache         bool `json:"headache"`
	Acne             bool `json:"acne"`
	Mood             bool `json:"mood"`
	Bloating         bool `json:"bloating"`
	Fatigue          bool `json:"fatigue"`
	BreastTenderness bool `json:"breast_tenderness"`
	BackPain         bool `json:"back_pain"`
	Nausea           bool `json:"nausea"`
	Spotting         bool `json:"spotting"`
	Irritability     bool `json:"irritability"`
	Insomnia         bool `json:"insomnia"`
	FoodCravings     bool `json:"food_cravings"`
	Diarrhea         bool `json:"diarrhea"`
	Constipation     bool `json:"constipation"`
}

type exportJSONEntry struct {
	Date          string                 `json:"date"`
	Period        bool                   `json:"period"`
	Flow          string                 `json:"flow"`
	Symptoms      exportJSONSymptomFlags `json:"symptoms"`
	OtherSymptoms []string               `json:"other_symptoms"`
	Notes         string                 `json:"notes"`
}

type attemptLimiter struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
}

type passwordResetClaims struct {
	UserID  uint   `json:"uid"`
	Purpose string `json:"purpose"`
	jwt.RegisteredClaims
}

func NewHandler(database *gorm.DB, secret string, templateDir string, location *time.Location, i18nManager *i18n.Manager) (*Handler, error) {
	if location == nil {
		location = time.Local
	}
	if i18nManager == nil {
		return nil, errors.New("i18n manager is required")
	}

	funcMap := template.FuncMap{
		"formatDate": func(value time.Time, layout string) string {
			if value.IsZero() {
				return ""
			}
			return value.Format(layout)
		},
		"formatFloat": func(value float64) string {
			return fmt.Sprintf("%.1f", value)
		},
		"t": func(messages map[string]string, key string) string {
			return translateMessage(messages, key)
		},
		"phaseLabel": func(messages map[string]string, phase string) string {
			return translateMessage(messages, phaseTranslationKey(phase))
		},
		"phaseIcon": func(phase string) string {
			switch strings.ToLower(strings.TrimSpace(phase)) {
			case "menstrual":
				return "🌙"
			case "follicular":
				return "🌸"
			case "ovulation":
				return "☀️"
			case "fertile":
				return "🌿"
			case "luteal":
				return "🍂"
			default:
				return "✨"
			}
		},
		"flowLabel": func(messages map[string]string, flow string) string {
			return translateMessage(messages, flowTranslationKey(flow))
		},
		"symptomLabel": func(messages map[string]string, name string) string {
			return localizedSymptomName(messages, name)
		},
		"roleLabel": func(messages map[string]string, role string) string {
			return translateMessage(messages, roleTranslationKey(role))
		},
		"isActiveRoute": func(currentPath string, route string) bool {
			path := strings.TrimSpace(currentPath)
			if path == "" {
				return route == "/"
			}
			if route == "/" {
				return path == "/" || strings.HasPrefix(path, "/?")
			}
			return path == route || strings.HasPrefix(path, route+"?") || strings.HasPrefix(path, route+"/")
		},
		"hasSymptom": func(set map[uint]bool, id uint) bool {
			return set[id]
		},
		"toJSON": func(value any) template.JS {
			serialized, _ := json.Marshal(value)
			return template.JS(serialized)
		},
	}

	templates := make(map[string]*template.Template)
	pages := []string{
		"login",
		"register",
		"recovery_code",
		"forgot_password",
		"reset_password",
		"onboarding",
		"dashboard",
		"calendar",
		"stats",
		"settings",
		"privacy",
	}
	for _, page := range pages {
		templatePath := filepath.Join(templateDir, page+".html")
		parsed, err := template.New("base").Funcs(funcMap).ParseFiles(
			filepath.Join(templateDir, "base.html"),
			templatePath,
		)
		if err != nil {
			return nil, fmt.Errorf("parse template %s: %w", page, err)
		}
		templates[page] = parsed
	}

	partials := make(map[string]*template.Template)
	partialFiles := []string{"day_editor_partial.html"}
	for _, partial := range partialFiles {
		name := strings.TrimSuffix(partial, ".html")
		parsed, err := template.New(name).Funcs(funcMap).ParseFiles(filepath.Join(templateDir, partial))
		if err != nil {
			return nil, fmt.Errorf("parse partial %s: %w", partial, err)
		}
		partials[name] = parsed
	}

	return &Handler{
		db:              database,
		secretKey:       []byte(secret),
		location:        location,
		lutealPhaseDays: 14,
		i18n:            i18nManager,
		templates:       templates,
		partials:        partials,
		recoveryLimiter: newAttemptLimiter(),
	}, nil
}

func (handler *Handler) Health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}

func (handler *Handler) SetupStatus(c *fiber.Ctx) error {
	var usersCount int64
	if err := handler.db.Model(&models.User{}).Count(&usersCount).Error; err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to load setup state")
	}
	return c.JSON(fiber.Map{"needs_setup": usersCount == 0})
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

	errorKey := authErrorTranslationKey(c.Query("error"))
	messages := currentMessages(c)
	title := translateMessage(messages, "meta.title.login")
	if title == "meta.title.login" {
		title = "Lume | Login"
	}
	data := fiber.Map{
		"Title":    title,
		"ErrorKey": errorKey,
	}
	return handler.render(c, "login", data)
}

func (handler *Handler) ShowRegisterPage(c *fiber.Ctx) error {
	if user, err := handler.authenticateRequest(c); err == nil {
		return c.Redirect(postLoginRedirectPath(user), fiber.StatusSeeOther)
	}

	errorKey := authErrorTranslationKey(c.Query("error"))
	messages := currentMessages(c)
	title := translateMessage(messages, "meta.title.register")
	if title == "meta.title.register" {
		title = "Lume | Sign Up"
	}
	data := fiber.Map{
		"Title":    title,
		"ErrorKey": errorKey,
	}
	return handler.render(c, "register", data)
}

func (handler *Handler) ShowForgotPasswordPage(c *fiber.Ctx) error {
	errorKey := authErrorTranslationKey(c.Query("error"))
	messages := currentMessages(c)
	title := translateMessage(messages, "meta.title.forgot_password")
	if title == "meta.title.forgot_password" {
		title = "Lume | Password Recovery"
	}
	data := fiber.Map{
		"Title":    title,
		"ErrorKey": errorKey,
	}
	return handler.render(c, "forgot_password", data)
}

func (handler *Handler) ShowResetPasswordPage(c *fiber.Ctx) error {
	token := strings.TrimSpace(c.Query("token"))
	messages := currentMessages(c)
	title := translateMessage(messages, "meta.title.reset_password")
	if title == "meta.title.reset_password" {
		title = "Lume | Reset Password"
	}

	invalidToken := false
	if _, err := handler.parsePasswordResetToken(token); err != nil {
		invalidToken = true
	}

	data := fiber.Map{
		"Title":        title,
		"Token":        token,
		"InvalidToken": invalidToken,
		"ForcedReset":  parseBoolValue(c.Query("forced")),
		"ErrorKey":     authErrorTranslationKey(c.Query("error")),
	}
	return handler.render(c, "reset_password", data)
}

func (handler *Handler) ShowOnboarding(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return c.Redirect("/login", fiber.StatusSeeOther)
	}
	if !requiresOnboarding(user) {
		return c.Redirect("/dashboard", fiber.StatusSeeOther)
	}

	now := dateAtLocation(time.Now().In(handler.location), handler.location)
	messages := currentMessages(c)

	lastPeriodStart := ""
	if user.LastPeriodStart != nil {
		lastPeriodStart = dateAtLocation(*user.LastPeriodStart, handler.location).Format("2006-01-02")
	}

	cycleLength := user.CycleLength
	if cycleLength < 21 || cycleLength > 35 {
		cycleLength = 28
	}
	periodLength := user.PeriodLength
	if periodLength < 2 || periodLength > 7 {
		periodLength = 5
	}

	data := fiber.Map{
		"Title":           localizedPageTitle(messages, "meta.title.onboarding", "Lume | Onboarding"),
		"CurrentUser":     user,
		"HideNavigation":  true,
		"MinDate":         now.AddDate(0, 0, -60).Format("2006-01-02"),
		"MaxDate":         now.Format("2006-01-02"),
		"LastPeriodStart": lastPeriodStart,
		"CycleLength":     cycleLength,
		"PeriodLength":    periodLength,
	}
	return handler.render(c, "onboarding", data)
}

func (handler *Handler) OnboardingStep1(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return apiError(c, fiber.StatusUnauthorized, "unauthorized")
	}
	if !requiresOnboarding(user) {
		return redirectOrJSON(c, "/dashboard")
	}

	input := struct {
		LastPeriodStart string `json:"last_period_start" form:"last_period_start"`
	}{}
	if err := c.BodyParser(&input); err != nil {
		return apiError(c, fiber.StatusBadRequest, "invalid input")
	}

	parsedDay, err := parseDayParam(strings.TrimSpace(input.LastPeriodStart), handler.location)
	if err != nil {
		return apiError(c, fiber.StatusBadRequest, "invalid last period start")
	}

	today := dateAtLocation(time.Now().In(handler.location), handler.location)
	minDate := today.AddDate(0, 0, -60)
	if parsedDay.After(today) || parsedDay.Before(minDate) {
		return apiError(c, fiber.StatusBadRequest, "last period start must be within last 60 days")
	}

	if err := handler.db.Model(&models.User{}).Where("id = ?", user.ID).Update("last_period_start", parsedDay).Error; err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to save onboarding step")
	}
	user.LastPeriodStart = &parsedDay

	if acceptsJSON(c) {
		return c.JSON(fiber.Map{"ok": true})
	}
	if isHTMX(c) {
		return c.SendStatus(fiber.StatusNoContent)
	}
	return c.Redirect("/onboarding", fiber.StatusSeeOther)
}

func (handler *Handler) OnboardingStep2(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return apiError(c, fiber.StatusUnauthorized, "unauthorized")
	}
	if !requiresOnboarding(user) {
		return redirectOrJSON(c, "/dashboard")
	}

	input := struct {
		CycleLength  int `json:"cycle_length" form:"cycle_length"`
		PeriodLength int `json:"period_length" form:"period_length"`
	}{}
	if err := c.BodyParser(&input); err != nil {
		return apiError(c, fiber.StatusBadRequest, "invalid input")
	}
	if input.CycleLength < 21 || input.CycleLength > 35 {
		return apiError(c, fiber.StatusBadRequest, "cycle length must be between 21 and 35")
	}
	if input.PeriodLength < 2 || input.PeriodLength > 7 {
		return apiError(c, fiber.StatusBadRequest, "period length must be between 2 and 7")
	}

	if err := handler.db.Model(&models.User{}).Where("id = ?", user.ID).Updates(map[string]any{
		"cycle_length":  input.CycleLength,
		"period_length": input.PeriodLength,
	}).Error; err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to save onboarding step")
	}
	user.CycleLength = input.CycleLength
	user.PeriodLength = input.PeriodLength

	if acceptsJSON(c) {
		return c.JSON(fiber.Map{"ok": true})
	}
	if isHTMX(c) {
		return c.SendStatus(fiber.StatusNoContent)
	}
	return c.Redirect("/onboarding", fiber.StatusSeeOther)
}

func (handler *Handler) OnboardingComplete(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return apiError(c, fiber.StatusUnauthorized, "unauthorized")
	}
	if !requiresOnboarding(user) {
		return redirectOrJSON(c, "/dashboard")
	}
	if user.LastPeriodStart == nil {
		return apiError(c, fiber.StatusBadRequest, "complete onboarding steps first")
	}

	startDay := dateAtLocation(*user.LastPeriodStart, handler.location)
	if err := handler.db.Transaction(func(tx *gorm.DB) error {
		var current models.User
		if err := tx.First(&current, user.ID).Error; err != nil {
			return err
		}
		if current.LastPeriodStart == nil {
			return errors.New("complete onboarding steps first")
		}
		startDay = dateAtLocation(*current.LastPeriodStart, handler.location)

		var entry models.DailyLog
		result := tx.Where("user_id = ? AND date = ?", current.ID, startDay).First(&entry)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			entry = models.DailyLog{
				UserID:     current.ID,
				Date:       startDay,
				IsPeriod:   true,
				Flow:       models.FlowMedium,
				SymptomIDs: []uint{},
			}
			if err := tx.Create(&entry).Error; err != nil {
				return err
			}
		} else if result.Error != nil {
			return result.Error
		} else {
			updates := map[string]any{"is_period": true}
			if strings.TrimSpace(entry.Flow) == "" || entry.Flow == models.FlowNone {
				updates["flow"] = models.FlowMedium
			}
			if err := tx.Model(&entry).Updates(updates).Error; err != nil {
				return err
			}
		}

		if err := tx.Model(&models.User{}).Where("id = ?", current.ID).Update("onboarding_completed", true).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "complete onboarding steps first") {
			return apiError(c, fiber.StatusBadRequest, "complete onboarding steps first")
		}
		return apiError(c, fiber.StatusInternalServerError, "failed to finish onboarding")
	}

	user.OnboardingCompleted = true
	user.LastPeriodStart = &startDay
	return redirectOrJSON(c, "/dashboard")
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

	allLogs, err := handler.fetchLogsForUser(user.ID, today.AddDate(-2, 0, 0), today)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("failed to load logs")
	}

	stats := services.BuildCycleStats(allLogs, now, handler.lutealPhaseDays)
	stats = handler.applyUserCycleBaseline(user, allLogs, stats, now)
	todayLog, err := handler.fetchLogByDate(user.ID, today)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("failed to load today log")
	}

	symptoms := make([]models.SymptomType, 0)
	if user.Role == models.RoleOwner {
		symptoms, err = handler.fetchSymptoms(user.ID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("failed to load symptoms")
		}
	}

	if user.Role == models.RolePartner {
		todayLog = sanitizeLogForPartner(todayLog)
	}

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
		"IsOwner":           user.Role == models.RoleOwner,
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
	activeMonth, err := parseMonthQuery(c.Query("month"), now, handler.location)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("invalid month")
	}

	monthStart := activeMonth
	monthEnd := monthStart.AddDate(0, 1, -1)

	logRangeStart := monthStart.AddDate(0, 0, -70)
	logRangeEnd := monthEnd.AddDate(0, 0, 70)
	logs, err := handler.fetchLogsForUser(user.ID, logRangeStart, logRangeEnd)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("failed to load calendar")
	}

	statsLogs, err := handler.fetchLogsForUser(user.ID, now.AddDate(-2, 0, 0), now)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("failed to load stats")
	}
	stats := services.BuildCycleStats(statsLogs, now, handler.lutealPhaseDays)
	stats = handler.applyUserCycleBaseline(user, statsLogs, stats, now)

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
		"CalendarDays": days,
		"Today":        dateAtLocation(now, handler.location).Format("2006-01-02"),
		"Stats":        stats,
		"IsOwner":      user.Role == models.RoleOwner,
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

	isOwner := user.Role == models.RoleOwner
	symptoms := make([]models.SymptomType, 0)
	if isOwner {
		symptoms, err = handler.fetchSymptoms(user.ID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("failed to load symptoms")
		}
	} else {
		logEntry = sanitizeLogForPartner(logEntry)
	}

	payload := fiber.Map{
		"Date":              day,
		"DateString":        day.Format("2006-01-02"),
		"DateLabel":         localizedDateLabel(currentLanguage(c), day),
		"IsFutureDate":      day.After(dateAtLocation(time.Now().In(handler.location), handler.location)),
		"NoDataLabel":       translateMessage(messages, "common.not_available"),
		"Log":               logEntry,
		"Symptoms":          symptoms,
		"SelectedSymptomID": symptomIDSet(logEntry.SymptomIDs),
		"HasDayData":        dayHasData(logEntry),
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
	logs, err := handler.fetchLogsForUser(user.ID, now.AddDate(-2, 0, 0), now)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("failed to load stats")
	}

	stats := services.BuildCycleStats(logs, now, handler.lutealPhaseDays)
	stats = handler.applyUserCycleBaseline(user, logs, stats, now)
	lengths := services.CycleLengths(logs)
	if len(lengths) > 12 {
		lengths = lengths[len(lengths)-12:]
	}

	labels := make([]string, 0, len(lengths))
	cycleLabelPattern := translateMessage(messages, "stats.cycle_label")
	if cycleLabelPattern == "stats.cycle_label" {
		cycleLabelPattern = "Cycle %d"
	}
	for index := range lengths {
		labels = append(labels, fmt.Sprintf(cycleLabelPattern, index+1))
	}

	baselineCycleLength := 0
	if user.Role == models.RoleOwner && isValidOnboardingCycleLength(user.CycleLength) {
		baselineCycleLength = user.CycleLength
	}

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
		for index := range symptomCounts {
			symptomCounts[index].FrequencySummary = localizedSymptomFrequencySummary(language, symptomCounts[index].Count, symptomCounts[index].TotalDays)
		}
	}

	data := fiber.Map{
		"Title":           localizedPageTitle(messages, "meta.title.stats", "Lume | Stats"),
		"CurrentUser":     user,
		"Stats":           stats,
		"ChartData":       chartPayload,
		"ChartBaseline":   baselineCycleLength,
		"TrendPointCount": len(lengths),
		"SymptomCounts":   symptomCounts,
		"IsOwner":         user.Role == models.RoleOwner,
	}

	return handler.render(c, "stats", data)
}

func (handler *Handler) ShowSettings(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return c.Redirect("/login", fiber.StatusSeeOther)
	}
	data, err := handler.buildSettingsViewData(c, user)
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

	data := fiber.Map{
		"Title":           localizedPageTitle(messages, "meta.title.privacy", "Lume | Privacy Policy"),
		"MetaDescription": metaDescription,
	}

	if user, err := handler.authenticateRequest(c); err == nil {
		data["CurrentUser"] = user
	}

	return handler.render(c, "privacy", data)
}

func (handler *Handler) Register(c *fiber.Ctx) error {
	credentials, err := parseCredentials(c)
	if err != nil {
		return handler.respondAuthError(c, fiber.StatusBadRequest, "invalid input")
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

	if err := handler.setAuthCookie(c, &user); err != nil {
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

	if err := handler.setAuthCookie(c, &user); err != nil {
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

	if err := handler.setAuthCookie(c, &user); err != nil {
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

func (handler *Handler) ChangePassword(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return apiError(c, fiber.StatusUnauthorized, "unauthorized")
	}

	input := changePasswordInput{}
	if err := c.BodyParser(&input); err != nil {
		return handler.respondSettingsError(c, fiber.StatusBadRequest, "invalid settings input")
	}

	input.CurrentPassword = strings.TrimSpace(input.CurrentPassword)
	input.NewPassword = strings.TrimSpace(input.NewPassword)
	if input.CurrentPassword == "" || input.NewPassword == "" {
		return handler.respondSettingsError(c, fiber.StatusBadRequest, "invalid settings input")
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.CurrentPassword)) != nil {
		return handler.respondSettingsError(c, fiber.StatusUnauthorized, "invalid current password")
	}
	if input.CurrentPassword == input.NewPassword {
		return handler.respondSettingsError(c, fiber.StatusBadRequest, "new password must differ")
	}
	if err := validatePasswordStrength(input.NewPassword); err != nil {
		return handler.respondSettingsError(c, fiber.StatusBadRequest, "weak password")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to secure password")
	}

	if err := handler.db.Model(&models.User{}).Where("id = ?", user.ID).Updates(map[string]any{
		"password_hash":        string(passwordHash),
		"must_change_password": false,
	}).Error; err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to update password")
	}

	if acceptsJSON(c) {
		return c.JSON(fiber.Map{"ok": true})
	}
	return redirectOrJSON(c, "/settings?status=password_changed")
}

func (handler *Handler) UpdateCycleSettings(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return apiError(c, fiber.StatusUnauthorized, "unauthorized")
	}

	input := cycleSettingsInput{}
	if err := c.BodyParser(&input); err != nil {
		return handler.respondSettingsError(c, fiber.StatusBadRequest, "invalid settings input")
	}
	if !isValidOnboardingCycleLength(input.CycleLength) {
		return handler.respondSettingsError(c, fiber.StatusBadRequest, "cycle length must be between 21 and 35")
	}
	if !isValidOnboardingPeriodLength(input.PeriodLength) {
		return handler.respondSettingsError(c, fiber.StatusBadRequest, "period length must be between 2 and 7")
	}

	if err := handler.db.Model(&models.User{}).Where("id = ?", user.ID).Updates(map[string]any{
		"cycle_length":  input.CycleLength,
		"period_length": input.PeriodLength,
	}).Error; err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to update cycle settings")
	}

	user.CycleLength = input.CycleLength
	user.PeriodLength = input.PeriodLength

	if acceptsJSON(c) {
		return c.JSON(fiber.Map{"ok": true})
	}
	return redirectOrJSON(c, "/settings?success=cycle_updated")
}

func (handler *Handler) RegenerateRecoveryCode(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return apiError(c, fiber.StatusUnauthorized, "unauthorized")
	}

	recoveryCode, recoveryHash, err := generateRecoveryCodeHash()
	if err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to create recovery code")
	}

	if err := handler.db.Model(&models.User{}).Where("id = ?", user.ID).Update("recovery_code_hash", recoveryHash).Error; err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to update recovery code")
	}

	if acceptsJSON(c) {
		return c.JSON(fiber.Map{
			"ok":            true,
			"recovery_code": recoveryCode,
		})
	}

	data, err := handler.buildSettingsViewData(c, user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("failed to load settings")
	}
	data["SuccessKey"] = "settings.success.recovery_code_regenerated"
	data["GeneratedRecoveryCode"] = recoveryCode
	return handler.render(c, "settings", data)
}

func (handler *Handler) DeleteAccount(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return apiError(c, fiber.StatusUnauthorized, "unauthorized")
	}

	input := deleteAccountInput{}
	if err := c.BodyParser(&input); err != nil && acceptsJSON(c) {
		return handler.respondSettingsError(c, fiber.StatusBadRequest, "invalid password")
	}

	input.Password = strings.TrimSpace(input.Password)
	if input.Password == "" {
		input.Password = strings.TrimSpace(c.FormValue("password"))
	}
	if input.Password == "" {
		return handler.respondSettingsError(c, fiber.StatusBadRequest, "invalid password")
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)) != nil {
		return handler.respondSettingsError(c, fiber.StatusUnauthorized, "invalid password")
	}

	if err := handler.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", user.ID).Delete(&models.DailyLog{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", user.ID).Delete(&models.SymptomType{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&models.User{}, user.ID).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to delete account")
	}

	handler.clearAuthCookie(c)
	if acceptsJSON(c) {
		return c.JSON(fiber.Map{"ok": true})
	}
	return redirectOrJSON(c, "/login")
}

func (handler *Handler) GetDays(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return apiError(c, fiber.StatusUnauthorized, "unauthorized")
	}

	from, err := parseDayParam(c.Query("from"), handler.location)
	if err != nil {
		return apiError(c, fiber.StatusBadRequest, "invalid from date")
	}
	to, err := parseDayParam(c.Query("to"), handler.location)
	if err != nil {
		return apiError(c, fiber.StatusBadRequest, "invalid to date")
	}
	if to.Before(from) {
		return apiError(c, fiber.StatusBadRequest, "invalid range")
	}

	logs, err := handler.fetchLogsForUser(user.ID, from, to)
	if err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to fetch logs")
	}

	if user.Role == models.RolePartner {
		for index := range logs {
			logs[index] = sanitizeLogForPartner(logs[index])
		}
	}

	return c.JSON(logs)
}

func (handler *Handler) GetDay(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return apiError(c, fiber.StatusUnauthorized, "unauthorized")
	}

	day, err := parseDayParam(c.Params("date"), handler.location)
	if err != nil {
		return apiError(c, fiber.StatusBadRequest, "invalid date")
	}

	logEntry, err := handler.fetchLogByDate(user.ID, day)
	if err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to fetch day")
	}

	if user.Role == models.RolePartner {
		logEntry = sanitizeLogForPartner(logEntry)
	}

	return c.JSON(logEntry)
}

func (handler *Handler) CheckDayExists(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return apiError(c, fiber.StatusUnauthorized, "unauthorized")
	}

	day, err := parseDayParam(c.Params("date"), handler.location)
	if err != nil {
		return apiError(c, fiber.StatusBadRequest, "invalid date")
	}

	logEntry, err := handler.fetchLogByDate(user.ID, day)
	if err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to fetch day")
	}

	return c.JSON(fiber.Map{"exists": dayHasData(logEntry)})
}

func (handler *Handler) UpsertDay(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return apiError(c, fiber.StatusUnauthorized, "unauthorized")
	}

	day, err := parseDayParam(c.Params("date"), handler.location)
	if err != nil {
		return apiError(c, fiber.StatusBadRequest, "invalid date")
	}

	payload, err := parseDayPayload(c)
	if err != nil {
		return apiError(c, fiber.StatusBadRequest, "invalid payload")
	}

	if !isValidFlow(payload.Flow) {
		return apiError(c, fiber.StatusBadRequest, "invalid flow value")
	}
	if payload.IsPeriod && payload.Flow == models.FlowNone {
		return apiError(c, fiber.StatusBadRequest, "period flow is required")
	}

	if !payload.IsPeriod {
		payload.Flow = models.FlowNone
	}

	if len(payload.Notes) > 2000 {
		payload.Notes = payload.Notes[:2000]
	}

	cleanIDs, err := handler.validateSymptomIDs(user.ID, payload.SymptomIDs)
	if err != nil {
		return apiError(c, fiber.StatusBadRequest, "invalid symptom ids")
	}

	var entry models.DailyLog
	result := handler.db.Where("user_id = ? AND date = ?", user.ID, day).First(&entry)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		entry = models.DailyLog{
			UserID:   user.ID,
			Date:     day,
			IsPeriod: payload.IsPeriod,
			Flow:     payload.Flow,
			Notes:    payload.Notes,
		}
		entry.SymptomIDs = cleanIDs
		if err := handler.db.Create(&entry).Error; err != nil {
			return apiError(c, fiber.StatusInternalServerError, "failed to create day")
		}
	} else if result.Error != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to load day")
	} else {
		entry.IsPeriod = payload.IsPeriod
		entry.Flow = payload.Flow
		entry.SymptomIDs = cleanIDs
		entry.Notes = payload.Notes
		if err := handler.db.Save(&entry).Error; err != nil {
			return apiError(c, fiber.StatusInternalServerError, "failed to update day")
		}
	}

	if err := handler.refreshUserLastPeriodStart(user.ID); err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to sync last period start")
	}

	if isHTMX(c) {
		c.Set("HX-Trigger", "calendar-day-updated")
		timestamp := time.Now().In(handler.location).Format("15:04")
		pattern := translateMessage(currentMessages(c), "common.saved_at")
		if pattern == "common.saved_at" {
			pattern = "Saved at %s"
		}
		message := fmt.Sprintf(pattern, timestamp)
		return c.SendString(fmt.Sprintf("<div class=\"status-ok status-transient\">%s</div>", template.HTMLEscapeString(message)))
	}

	return c.JSON(entry)
}

func (handler *Handler) DeleteDailyLog(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return apiError(c, fiber.StatusUnauthorized, "unauthorized")
	}

	day, err := parseDayParam(c.Query("date"), handler.location)
	if err != nil {
		return apiError(c, fiber.StatusBadRequest, "invalid date")
	}

	if err := handler.deleteDailyLogByDate(user.ID, day); err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to delete day")
	}
	if err := handler.refreshUserLastPeriodStart(user.ID); err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to sync last period start")
	}

	source := strings.ToLower(strings.TrimSpace(c.Query("source")))
	if isHTMX(c) {
		c.Set("HX-Trigger", "calendar-day-updated")
		switch source {
		case "calendar":
			return handler.renderDayEditorPartial(c, user, day)
		case "dashboard":
			c.Set("HX-Redirect", "/dashboard")
			return c.SendStatus(fiber.StatusOK)
		default:
			return c.SendStatus(fiber.StatusNoContent)
		}
	}

	if source == "dashboard" {
		return redirectOrJSON(c, "/dashboard")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (handler *Handler) DeleteDay(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return apiError(c, fiber.StatusUnauthorized, "unauthorized")
	}

	day, err := parseDayParam(c.Params("date"), handler.location)
	if err != nil {
		return apiError(c, fiber.StatusBadRequest, "invalid date")
	}

	if err := handler.deleteDailyLogByDate(user.ID, day); err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to delete day")
	}
	if err := handler.refreshUserLastPeriodStart(user.ID); err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to sync last period start")
	}

	if isHTMX(c) {
		c.Set("HX-Trigger", "calendar-day-updated")
		return handler.renderDayEditorPartial(c, user, day)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (handler *Handler) GetSymptoms(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return apiError(c, fiber.StatusUnauthorized, "unauthorized")
	}
	if user.Role != models.RoleOwner {
		return apiError(c, fiber.StatusForbidden, "owner access required")
	}

	symptoms, err := handler.fetchSymptoms(user.ID)
	if err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to fetch symptoms")
	}
	return c.JSON(symptoms)
}

func (handler *Handler) CreateSymptom(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return apiError(c, fiber.StatusUnauthorized, "unauthorized")
	}

	payload := symptomPayload{}
	if err := c.BodyParser(&payload); err != nil {
		return apiError(c, fiber.StatusBadRequest, "invalid payload")
	}

	payload.Name = strings.TrimSpace(payload.Name)
	payload.Icon = strings.TrimSpace(payload.Icon)
	payload.Color = strings.TrimSpace(payload.Color)
	payload.Name = normalizeLegacySymptomName(payload.Name)

	if payload.Name == "" || len(payload.Name) > 80 {
		return apiError(c, fiber.StatusBadRequest, "invalid symptom name")
	}
	if payload.Icon == "" {
		payload.Icon = "✨"
	}
	if !hexColorRegex.MatchString(payload.Color) {
		return apiError(c, fiber.StatusBadRequest, "invalid symptom color")
	}

	symptom := models.SymptomType{
		UserID:    user.ID,
		Name:      payload.Name,
		Icon:      payload.Icon,
		Color:     payload.Color,
		IsBuiltin: false,
	}

	if err := handler.db.Create(&symptom).Error; err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to create symptom")
	}
	return c.Status(fiber.StatusCreated).JSON(symptom)
}

func (handler *Handler) DeleteSymptom(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return apiError(c, fiber.StatusUnauthorized, "unauthorized")
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return apiError(c, fiber.StatusBadRequest, "invalid symptom id")
	}

	var symptom models.SymptomType
	if err := handler.db.Where("id = ? AND user_id = ?", id, user.ID).First(&symptom).Error; err != nil {
		return apiError(c, fiber.StatusNotFound, "symptom not found")
	}
	if symptom.IsBuiltin {
		return apiError(c, fiber.StatusBadRequest, "built-in symptom cannot be deleted")
	}

	if err := handler.db.Delete(&symptom).Error; err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to delete symptom")
	}

	if err := handler.removeSymptomFromLogs(user.ID, symptom.ID); err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to clean symptom logs")
	}

	return c.JSON(fiber.Map{"ok": true})
}

func (handler *Handler) GetStatsOverview(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return apiError(c, fiber.StatusUnauthorized, "unauthorized")
	}

	now := time.Now().In(handler.location)
	logs, err := handler.fetchLogsForUser(user.ID, now.AddDate(-2, 0, 0), now)
	if err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to fetch stats")
	}

	stats := services.BuildCycleStats(logs, now, handler.lutealPhaseDays)
	stats = handler.applyUserCycleBaseline(user, logs, stats, now)
	return c.JSON(stats)
}

func (handler *Handler) ExportCSV(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return apiError(c, fiber.StatusUnauthorized, "unauthorized")
	}

	logs, symptomNames, err := handler.fetchExportData(user.ID)
	if err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to fetch logs")
	}

	var output bytes.Buffer
	writer := csv.NewWriter(&output)
	if err := writer.Write([]string{
		"Date",
		"Period",
		"Flow",
		"Cramps",
		"Headache",
		"Acne",
		"Mood",
		"Bloating",
		"Fatigue",
		"Breast tenderness",
		"Back pain",
		"Nausea",
		"Spotting",
		"Irritability",
		"Insomnia",
		"Food cravings",
		"Diarrhea",
		"Constipation",
		"Other",
		"Notes",
	}); err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to build export")
	}

	for _, logEntry := range logs {
		flags, other := buildCSVSymptomColumns(logEntry.SymptomIDs, symptomNames)
		if err := writer.Write([]string{
			dateAtLocation(logEntry.Date, handler.location).Format("2006-01-02"),
			csvYesNo(logEntry.IsPeriod),
			csvFlowLabel(logEntry.Flow),
			csvYesNo(flags.Cramps),
			csvYesNo(flags.Headache),
			csvYesNo(flags.Acne),
			csvYesNo(flags.Mood),
			csvYesNo(flags.Bloating),
			csvYesNo(flags.Fatigue),
			csvYesNo(flags.BreastTenderness),
			csvYesNo(flags.BackPain),
			csvYesNo(flags.Nausea),
			csvYesNo(flags.Spotting),
			csvYesNo(flags.Irritability),
			csvYesNo(flags.Insomnia),
			csvYesNo(flags.FoodCravings),
			csvYesNo(flags.Diarrhea),
			csvYesNo(flags.Constipation),
			strings.Join(other, "; "),
			logEntry.Notes,
		}); err != nil {
			return apiError(c, fiber.StatusInternalServerError, "failed to build export")
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to build export")
	}

	filename := fmt.Sprintf("lume-export-%s.csv", time.Now().In(handler.location).Format("2006-01-02"))
	c.Set(fiber.HeaderContentType, "text/csv")
	c.Set(fiber.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%s", filename))
	return c.Send(output.Bytes())
}

func (handler *Handler) ExportJSON(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return apiError(c, fiber.StatusUnauthorized, "unauthorized")
	}

	logs, symptomNames, err := handler.fetchExportData(user.ID)
	if err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to fetch logs")
	}

	entries := make([]exportJSONEntry, 0, len(logs))
	for _, logEntry := range logs {
		flags, other := buildCSVSymptomColumns(logEntry.SymptomIDs, symptomNames)
		entries = append(entries, exportJSONEntry{
			Date:   dateAtLocation(logEntry.Date, handler.location).Format("2006-01-02"),
			Period: logEntry.IsPeriod,
			Flow:   normalizeExportFlow(logEntry.Flow),
			Symptoms: exportJSONSymptomFlags{
				Cramps:           flags.Cramps,
				Headache:         flags.Headache,
				Acne:             flags.Acne,
				Mood:             flags.Mood,
				Bloating:         flags.Bloating,
				Fatigue:          flags.Fatigue,
				BreastTenderness: flags.BreastTenderness,
				BackPain:         flags.BackPain,
				Nausea:           flags.Nausea,
				Spotting:         flags.Spotting,
				Irritability:     flags.Irritability,
				Insomnia:         flags.Insomnia,
				FoodCravings:     flags.FoodCravings,
				Diarrhea:         flags.Diarrhea,
				Constipation:     flags.Constipation,
			},
			OtherSymptoms: other,
			Notes:         logEntry.Notes,
		})
	}

	payload := fiber.Map{
		"exported_at": time.Now().In(handler.location).Format(time.RFC3339),
		"entries":     entries,
	}

	serialized, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to build export")
	}

	filename := fmt.Sprintf("lume-export-%s.json", time.Now().In(handler.location).Format("2006-01-02"))
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	c.Set(fiber.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%s", filename))
	return c.Send(serialized)
}

func (handler *Handler) render(c *fiber.Ctx, name string, data fiber.Map) error {
	tmpl, ok := handler.templates[name]
	if !ok {
		return c.Status(fiber.StatusInternalServerError).SendString("template not found")
	}
	payload := handler.withTemplateDefaults(c, data)
	var output bytes.Buffer
	if err := tmpl.ExecuteTemplate(&output, "base", payload); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("failed to render template")
	}
	c.Type("html", "utf-8")
	return c.Send(output.Bytes())
}

func (handler *Handler) renderPartial(c *fiber.Ctx, name string, data fiber.Map) error {
	tmpl, ok := handler.partials[name]
	if !ok {
		return c.Status(fiber.StatusInternalServerError).SendString("partial not found")
	}
	payload := handler.withTemplateDefaults(c, data)
	var output bytes.Buffer
	if err := tmpl.ExecuteTemplate(&output, name, payload); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("failed to render partial")
	}
	c.Type("html", "utf-8")
	return c.Send(output.Bytes())
}

func (handler *Handler) setAuthCookie(c *fiber.Ctx, user *models.User) error {
	token, err := handler.buildToken(user)
	if err != nil {
		return err
	}

	cookie := &fiber.Cookie{
		Name:     authCookieName,
		Value:    token,
		Path:     "/",
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
		Expires:  time.Now().Add(7 * 24 * time.Hour),
	}
	c.Cookie(cookie)
	return nil
}

func (handler *Handler) clearAuthCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     authCookieName,
		Value:    "",
		Path:     "/",
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
		Expires:  time.Now().Add(-1 * time.Hour),
	})
}

func (handler *Handler) buildToken(user *models.User) (string, error) {
	claims := authClaims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatUint(uint64(user.ID), 10),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(handler.secretKey)
}

func (handler *Handler) buildPasswordResetToken(userID uint, ttl time.Duration) (string, error) {
	if ttl <= 0 {
		ttl = 30 * time.Minute
	}

	now := time.Now()
	claims := passwordResetClaims{
		UserID:  userID,
		Purpose: "password_reset",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatUint(uint64(userID), 10),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(handler.secretKey)
}

func (handler *Handler) parsePasswordResetToken(rawToken string) (uint, error) {
	if strings.TrimSpace(rawToken) == "" {
		return 0, errors.New("missing token")
	}

	claims := &passwordResetClaims{}
	token, err := jwt.ParseWithClaims(rawToken, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return handler.secretKey, nil
	})
	if err != nil || !token.Valid {
		return 0, errors.New("invalid token")
	}
	if claims.Purpose != "password_reset" {
		return 0, errors.New("invalid token purpose")
	}
	if claims.ExpiresAt == nil || claims.ExpiresAt.Time.Before(time.Now()) {
		return 0, errors.New("token expired")
	}
	if claims.UserID == 0 {
		return 0, errors.New("invalid user id")
	}
	return claims.UserID, nil
}

func (handler *Handler) seedBuiltinSymptoms(userID uint) error {
	var count int64
	if err := handler.db.Model(&models.SymptomType{}).
		Where("user_id = ? AND is_builtin = ?", userID, true).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	builtin := models.DefaultBuiltinSymptoms()
	records := make([]models.SymptomType, 0, len(builtin))
	for _, symptom := range builtin {
		records = append(records, models.SymptomType{
			UserID:    userID,
			Name:      symptom.Name,
			Icon:      symptom.Icon,
			Color:     symptom.Color,
			IsBuiltin: true,
		})
	}

	return handler.db.Create(&records).Error
}

func (handler *Handler) fetchSymptoms(userID uint) ([]models.SymptomType, error) {
	if err := handler.ensureBuiltinSymptoms(userID); err != nil {
		return nil, err
	}

	symptoms := make([]models.SymptomType, 0)
	if err := handler.db.Where("user_id = ?", userID).Find(&symptoms).Error; err != nil {
		return nil, err
	}
	for index := range symptoms {
		symptoms[index].Name = normalizeLegacySymptomName(symptoms[index].Name)
	}

	builtinOrder := make(map[string]int)
	for index, symptom := range models.DefaultBuiltinSymptoms() {
		builtinOrder[strings.ToLower(strings.TrimSpace(symptom.Name))] = index
	}

	sort.Slice(symptoms, func(i, j int) bool {
		left := symptoms[i]
		right := symptoms[j]
		if left.IsBuiltin != right.IsBuiltin {
			return left.IsBuiltin
		}
		if left.IsBuiltin && right.IsBuiltin {
			leftIndex, leftHas := builtinOrder[strings.ToLower(strings.TrimSpace(left.Name))]
			rightIndex, rightHas := builtinOrder[strings.ToLower(strings.TrimSpace(right.Name))]
			switch {
			case leftHas && rightHas && leftIndex != rightIndex:
				return leftIndex < rightIndex
			case leftHas != rightHas:
				return leftHas
			}
		}
		return strings.ToLower(strings.TrimSpace(left.Name)) < strings.ToLower(strings.TrimSpace(right.Name))
	})

	return symptoms, nil
}

func (handler *Handler) ensureBuiltinSymptoms(userID uint) error {
	if err := handler.db.
		Model(&models.SymptomType{}).
		Where("user_id = ? AND lower(trim(name)) = ?", userID, "fatique").
		Update("name", "Fatigue").Error; err != nil {
		return err
	}

	existing := make([]models.SymptomType, 0)
	if err := handler.db.Where("user_id = ?", userID).Find(&existing).Error; err != nil {
		return err
	}

	existingByName := make(map[string]struct{}, len(existing))
	for _, symptom := range existing {
		key := strings.ToLower(strings.TrimSpace(symptom.Name))
		if key != "" {
			existingByName[key] = struct{}{}
		}
	}

	missing := make([]models.SymptomType, 0)
	for _, symptom := range models.DefaultBuiltinSymptoms() {
		key := strings.ToLower(strings.TrimSpace(symptom.Name))
		if _, ok := existingByName[key]; ok {
			continue
		}
		missing = append(missing, models.SymptomType{
			UserID:    userID,
			Name:      symptom.Name,
			Icon:      symptom.Icon,
			Color:     symptom.Color,
			IsBuiltin: true,
		})
	}

	if len(missing) == 0 {
		return nil
	}
	return handler.db.Create(&missing).Error
}

func (handler *Handler) buildSettingsViewData(c *fiber.Ctx, user *models.User) (fiber.Map, error) {
	messages := currentMessages(c)
	status := strings.TrimSpace(c.Query("success"))
	if status == "" {
		status = strings.TrimSpace(c.Query("status"))
	}

	cycleLength := user.CycleLength
	if !isValidOnboardingCycleLength(cycleLength) {
		cycleLength = 28
	}
	periodLength := user.PeriodLength
	if !isValidOnboardingPeriodLength(periodLength) {
		periodLength = 5
	}

	data := fiber.Map{
		"Title":        localizedPageTitle(messages, "meta.title.settings", "Lume | Settings"),
		"CurrentUser":  user,
		"ErrorKey":     authErrorTranslationKey(c.Query("error")),
		"SuccessKey":   settingsStatusTranslationKey(status),
		"CycleLength":  cycleLength,
		"PeriodLength": periodLength,
	}

	if user.Role == models.RoleOwner {
		totalEntries, firstDate, lastDate, err := handler.fetchExportSummary(user.ID)
		if err != nil {
			return nil, err
		}
		data["ExportTotalEntries"] = int(totalEntries)
		data["HasExportData"] = totalEntries > 0
		data["ExportDateFrom"] = firstDate
		data["ExportDateTo"] = lastDate
	}

	return data, nil
}

func (handler *Handler) fetchExportData(userID uint) ([]models.DailyLog, map[uint]string, error) {
	logs := make([]models.DailyLog, 0)
	if err := handler.db.Where("user_id = ?", userID).Order("date ASC").Find(&logs).Error; err != nil {
		return nil, nil, err
	}

	symptoms, err := handler.fetchSymptoms(userID)
	if err != nil {
		return nil, nil, err
	}

	symptomNames := make(map[uint]string, len(symptoms))
	for _, symptom := range symptoms {
		symptomNames[symptom.ID] = symptom.Name
	}

	return logs, symptomNames, nil
}

func (handler *Handler) fetchExportSummary(userID uint) (int64, string, string, error) {
	var total int64
	if err := handler.db.Model(&models.DailyLog{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return 0, "", "", err
	}
	if total == 0 {
		return 0, "", "", nil
	}

	var first models.DailyLog
	if err := handler.db.Select("date").Where("user_id = ?", userID).Order("date ASC").First(&first).Error; err != nil {
		return 0, "", "", err
	}

	var last models.DailyLog
	if err := handler.db.Select("date").Where("user_id = ?", userID).Order("date DESC").First(&last).Error; err != nil {
		return 0, "", "", err
	}

	return total,
		dateAtLocation(first.Date, handler.location).Format("2006-01-02"),
		dateAtLocation(last.Date, handler.location).Format("2006-01-02"),
		nil
}

func (handler *Handler) fetchLogsForUser(userID uint, from time.Time, to time.Time) ([]models.DailyLog, error) {
	logs := make([]models.DailyLog, 0)
	err := handler.db.
		Where("user_id = ? AND date >= ? AND date <= ?", userID, dateAtLocation(from, handler.location), dateAtLocation(to, handler.location)).
		Order("date ASC").
		Find(&logs).Error
	return logs, err
}

func (handler *Handler) fetchAllLogsForUser(userID uint) ([]models.DailyLog, error) {
	logs := make([]models.DailyLog, 0)
	err := handler.db.Where("user_id = ?", userID).Order("date ASC").Find(&logs).Error
	return logs, err
}

func (handler *Handler) fetchLogByDate(userID uint, day time.Time) (models.DailyLog, error) {
	entry := models.DailyLog{}
	result := handler.db.Where("user_id = ? AND date = ?", userID, dateAtLocation(day, handler.location)).Limit(1).Find(&entry)
	if result.Error != nil {
		return models.DailyLog{}, result.Error
	}
	if result.RowsAffected == 0 {
		return models.DailyLog{
			UserID:     userID,
			Date:       day,
			Flow:       models.FlowNone,
			SymptomIDs: []uint{},
		}, nil
	}
	return entry, nil
}

func (handler *Handler) deleteDailyLogByDate(userID uint, day time.Time) error {
	return handler.db.Where("user_id = ? AND date = ?", userID, dateAtLocation(day, handler.location)).Delete(&models.DailyLog{}).Error
}

func (handler *Handler) refreshUserLastPeriodStart(userID uint) error {
	periodLogs := make([]models.DailyLog, 0)
	if err := handler.db.
		Select("date", "is_period").
		Where("user_id = ? AND is_period = ?", userID, true).
		Order("date ASC").
		Find(&periodLogs).Error; err != nil {
		return err
	}

	starts := services.DetectCycleStarts(periodLogs)
	if len(starts) == 0 {
		return handler.db.Model(&models.User{}).Where("id = ?", userID).Update("last_period_start", nil).Error
	}

	latest := dateAtLocation(starts[len(starts)-1], handler.location)
	return handler.db.Model(&models.User{}).Where("id = ?", userID).Update("last_period_start", latest).Error
}

func (handler *Handler) validateSymptomIDs(userID uint, ids []uint) ([]uint, error) {
	if len(ids) == 0 {
		return []uint{}, nil
	}

	unique := make(map[uint]struct{}, len(ids))
	for _, id := range ids {
		unique[id] = struct{}{}
	}
	filtered := make([]uint, 0, len(unique))
	for id := range unique {
		filtered = append(filtered, id)
	}

	var matched int64
	if err := handler.db.Model(&models.SymptomType{}).
		Where("user_id = ? AND id IN ?", userID, filtered).
		Count(&matched).Error; err != nil {
		return nil, err
	}
	if int(matched) != len(filtered) {
		return nil, errors.New("invalid symptom id")
	}
	sort.Slice(filtered, func(i, j int) bool { return filtered[i] < filtered[j] })
	return filtered, nil
}

func (handler *Handler) removeSymptomFromLogs(userID uint, symptomID uint) error {
	logs := make([]models.DailyLog, 0)
	if err := handler.db.Where("user_id = ?", userID).Find(&logs).Error; err != nil {
		return err
	}

	for _, logEntry := range logs {
		updated := removeUint(logEntry.SymptomIDs, symptomID)
		if len(updated) != len(logEntry.SymptomIDs) {
			if err := handler.db.Model(&logEntry).Update("symptom_ids", updated).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func (handler *Handler) buildCalendarDays(monthStart time.Time, logs []models.DailyLog, stats services.CycleStats, now time.Time) []CalendarDay {
	monthEnd := monthStart.AddDate(0, 1, -1)
	gridStart := monthStart.AddDate(0, 0, -int(monthStart.Weekday()))
	gridEnd := monthEnd.AddDate(0, 0, 6-int(monthEnd.Weekday()))

	periodMap := make(map[string]bool)
	hasDataMap := make(map[string]bool)
	for _, logEntry := range logs {
		key := dateAtLocation(logEntry.Date, handler.location).Format("2006-01-02")
		periodMap[key] = logEntry.IsPeriod
		hasDataMap[key] = dayHasData(logEntry)
	}

	predictedPeriodMap := make(map[string]bool)
	predictedPeriodLength := int(stats.AveragePeriodLength + 0.5)
	if predictedPeriodLength <= 0 {
		predictedPeriodLength = 5
	}
	if !stats.NextPeriodStart.IsZero() {
		for offset := 0; offset < predictedPeriodLength; offset++ {
			day := stats.NextPeriodStart.AddDate(0, 0, offset)
			predictedPeriodMap[day.Format("2006-01-02")] = true
		}
	}

	fertilityMap := make(map[string]bool)
	if !stats.FertilityWindowStart.IsZero() && !stats.FertilityWindowEnd.IsZero() {
		for day := stats.FertilityWindowStart; !day.After(stats.FertilityWindowEnd); day = day.AddDate(0, 0, 1) {
			fertilityMap[day.Format("2006-01-02")] = true
		}
	}

	todayKey := dateAtLocation(now, handler.location).Format("2006-01-02")
	ovulationKey := stats.OvulationDate.Format("2006-01-02")

	days := make([]CalendarDay, 0, 42)
	for day := gridStart; !day.After(gridEnd); day = day.AddDate(0, 0, 1) {
		key := day.Format("2006-01-02")
		inMonth := day.Month() == monthStart.Month()
		isPeriod := periodMap[key]
		isPredicted := predictedPeriodMap[key]
		isFertility := fertilityMap[key]
		isToday := key == todayKey
		isOvulation := key == ovulationKey
		hasData := hasDataMap[key]

		cellClass := "calendar-cell"
		textClass := "calendar-day-number"
		badgeClass := "calendar-tag"
		if isPeriod {
			cellClass += " calendar-cell-period"
			badgeClass += " calendar-tag-period"
		} else if isPredicted {
			cellClass += " calendar-cell-predicted"
			badgeClass += " calendar-tag-predicted"
		} else if isFertility {
			cellClass += " calendar-cell-fertile"
			badgeClass += " calendar-tag-fertile"
		}
		if !inMonth {
			cellClass += " calendar-cell-out"
			textClass += " calendar-day-out"
		}
		if isToday {
			cellClass += " calendar-cell-today"
		}

		days = append(days, CalendarDay{
			Date:         day,
			DateString:   key,
			Day:          day.Day(),
			InMonth:      inMonth,
			IsToday:      isToday,
			IsPeriod:     isPeriod,
			IsPredicted:  isPredicted,
			IsFertility:  isFertility,
			IsOvulation:  isOvulation,
			HasData:      hasData,
			CellClass:    cellClass,
			TextClass:    textClass,
			BadgeClass:   badgeClass,
			OvulationDot: isOvulation,
		})
	}
	return days
}

func (handler *Handler) calculateSymptomFrequencies(userID uint, logs []models.DailyLog) ([]SymptomCount, error) {
	if len(logs) == 0 {
		return []SymptomCount{}, nil
	}
	totalDays := len(logs)

	counts := make(map[uint]int)
	for _, logEntry := range logs {
		for _, id := range logEntry.SymptomIDs {
			counts[id]++
		}
	}
	if len(counts) == 0 {
		return []SymptomCount{}, nil
	}

	symptoms, err := handler.fetchSymptoms(userID)
	if err != nil {
		return nil, err
	}

	symptomByID := make(map[uint]models.SymptomType, len(symptoms))
	for _, symptom := range symptoms {
		symptomByID[symptom.ID] = symptom
	}

	result := make([]SymptomCount, 0, len(counts))
	for id, count := range counts {
		if symptom, ok := symptomByID[id]; ok {
			result = append(result, SymptomCount{Name: symptom.Name, Icon: symptom.Icon, Count: count, TotalDays: totalDays})
		}
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Count == result[j].Count {
			return result[i].Name < result[j].Name
		}
		return result[i].Count > result[j].Count
	})

	return result, nil
}

func (handler *Handler) applyUserCycleBaseline(user *models.User, logs []models.DailyLog, stats services.CycleStats, now time.Time) services.CycleStats {
	if user == nil || user.Role != models.RoleOwner {
		return stats
	}

	latestLoggedPeriodStart := time.Time{}
	detectedStarts := services.DetectCycleStarts(logs)
	if len(detectedStarts) > 0 {
		latestLoggedPeriodStart = dateAtLocation(detectedStarts[len(detectedStarts)-1], handler.location)
	}

	cycleLength := 0
	if isValidOnboardingCycleLength(user.CycleLength) {
		cycleLength = user.CycleLength
	}

	periodLength := 0
	if isValidOnboardingPeriodLength(user.PeriodLength) {
		periodLength = user.PeriodLength
	}

	reliableCycleData := len(services.CycleLengths(logs)) >= 2
	if !reliableCycleData {
		if cycleLength > 0 {
			stats.AverageCycleLength = float64(cycleLength)
			stats.MedianCycleLength = cycleLength
		}
		if periodLength > 0 {
			stats.AveragePeriodLength = float64(periodLength)
		}
		switch {
		case !latestLoggedPeriodStart.IsZero():
			stats.LastPeriodStart = latestLoggedPeriodStart
		case user.LastPeriodStart != nil:
			stats.LastPeriodStart = dateAtLocation(*user.LastPeriodStart, handler.location)
		}
	} else if !latestLoggedPeriodStart.IsZero() {
		stats.LastPeriodStart = latestLoggedPeriodStart
	}

	if !stats.LastPeriodStart.IsZero() && cycleLength > 0 && (!reliableCycleData || stats.NextPeriodStart.IsZero()) {
		stats.NextPeriodStart = dateAtLocation(stats.LastPeriodStart.AddDate(0, 0, cycleLength), handler.location)
		stats.OvulationDate = dateAtLocation(stats.NextPeriodStart.AddDate(0, 0, -handler.lutealPhaseDays), handler.location)
		stats.FertilityWindowStart = dateAtLocation(stats.OvulationDate.AddDate(0, 0, -5), handler.location)
		stats.FertilityWindowEnd = dateAtLocation(stats.OvulationDate.AddDate(0, 0, 1), handler.location)
	}

	today := dateAtLocation(now.In(handler.location), handler.location)
	if !stats.LastPeriodStart.IsZero() && !today.Before(stats.LastPeriodStart) {
		stats.CurrentCycleDay = int(today.Sub(stats.LastPeriodStart).Hours()/24) + 1
	} else {
		stats.CurrentCycleDay = 0
	}

	stats.CurrentPhase = handler.detectCurrentPhase(stats, logs, today)

	return stats
}

func (handler *Handler) detectCurrentPhase(stats services.CycleStats, logs []models.DailyLog, today time.Time) string {
	periodByDate := make(map[string]bool, len(logs))
	for _, logEntry := range logs {
		if logEntry.IsPeriod {
			periodByDate[dateAtLocation(logEntry.Date, handler.location).Format("2006-01-02")] = true
		}
	}
	if periodByDate[today.Format("2006-01-02")] {
		return "menstrual"
	}

	periodLength := int(stats.AveragePeriodLength + 0.5)
	if periodLength <= 0 {
		periodLength = 5
	}
	if !stats.LastPeriodStart.IsZero() {
		periodEnd := dateAtLocation(stats.LastPeriodStart.AddDate(0, 0, periodLength-1), handler.location)
		if betweenCalendarDaysInclusive(today, stats.LastPeriodStart, periodEnd) {
			return "menstrual"
		}
	}

	if !stats.OvulationDate.IsZero() {
		switch {
		case sameCalendarDay(today, stats.OvulationDate):
			return "ovulation"
		case betweenCalendarDaysInclusive(today, stats.FertilityWindowStart, stats.FertilityWindowEnd):
			return "fertile"
		case today.Before(stats.OvulationDate):
			return "follicular"
		default:
			return "luteal"
		}
	}

	return "unknown"
}

func parseCredentials(c *fiber.Ctx) (credentialsInput, error) {
	credentials := credentialsInput{}
	if err := c.BodyParser(&credentials); err != nil {
		return credentialsInput{}, err
	}

	credentials.Email = strings.ToLower(strings.TrimSpace(credentials.Email))
	credentials.Password = strings.TrimSpace(credentials.Password)

	if credentials.Email == "" || credentials.Password == "" {
		return credentialsInput{}, errors.New("missing credentials")
	}
	if _, err := mail.ParseAddress(credentials.Email); err != nil {
		return credentialsInput{}, errors.New("invalid email")
	}

	return credentials, nil
}

func isValidOnboardingCycleLength(value int) bool {
	return value >= 21 && value <= 35
}

func isValidOnboardingPeriodLength(value int) bool {
	return value >= 2 && value <= 7
}

func validatePasswordStrength(password string) error {
	if len(password) < 8 {
		return errors.New("password too short")
	}

	var hasUpper bool
	var hasLower bool
	var hasDigit bool
	for _, character := range password {
		if unicode.IsUpper(character) {
			hasUpper = true
			continue
		}
		if unicode.IsLower(character) {
			hasLower = true
			continue
		}
		if unicode.IsDigit(character) {
			hasDigit = true
		}
	}

	if hasUpper && hasLower && hasDigit {
		return nil
	}
	return errors.New("weak password")
}

func parseDayPayload(c *fiber.Ctx) (dayPayload, error) {
	payload := dayPayload{Flow: models.FlowNone, SymptomIDs: []uint{}}
	contentType := strings.ToLower(c.Get("Content-Type"))

	if strings.Contains(contentType, "application/json") {
		if err := c.BodyParser(&payload); err != nil {
			return payload, err
		}
	} else {
		payload.IsPeriod = parseBoolValue(c.FormValue("is_period"))
		payload.Flow = strings.ToLower(strings.TrimSpace(c.FormValue("flow")))
		payload.Notes = strings.TrimSpace(c.FormValue("notes"))

		symptomRaw := c.Context().PostArgs().PeekMulti("symptom_ids")
		for _, value := range symptomRaw {
			parsed, err := strconv.ParseUint(string(value), 10, 64)
			if err == nil {
				payload.SymptomIDs = append(payload.SymptomIDs, uint(parsed))
			}
		}
	}

	payload.Flow = strings.ToLower(strings.TrimSpace(payload.Flow))
	if payload.Flow == "" {
		payload.Flow = models.FlowNone
	}
	payload.Notes = strings.TrimSpace(payload.Notes)

	return payload, nil
}

func isValidFlow(flow string) bool {
	switch flow {
	case models.FlowNone, models.FlowLight, models.FlowMedium, models.FlowHeavy:
		return true
	default:
		return false
	}
}

func dayHasData(entry models.DailyLog) bool {
	if entry.IsPeriod {
		return true
	}
	if len(entry.SymptomIDs) > 0 {
		return true
	}
	if strings.TrimSpace(entry.Notes) != "" {
		return true
	}
	return strings.TrimSpace(entry.Flow) != "" && entry.Flow != models.FlowNone
}

func normalizeLegacySymptomName(name string) string {
	if strings.EqualFold(strings.TrimSpace(name), "fatique") {
		return "Fatigue"
	}
	return name
}

func buildCSVSymptomColumns(symptomIDs []uint, symptomNames map[uint]string) (exportSymptomFlags, []string) {
	flags := exportSymptomFlags{}
	otherSet := make(map[string]struct{})

	for _, symptomID := range symptomIDs {
		name, ok := symptomNames[symptomID]
		if !ok {
			continue
		}

		switch exportSymptomColumn(name) {
		case "cramps":
			flags.Cramps = true
		case "headache":
			flags.Headache = true
		case "acne":
			flags.Acne = true
		case "mood":
			flags.Mood = true
		case "bloating":
			flags.Bloating = true
		case "fatigue":
			flags.Fatigue = true
		case "breast_tenderness":
			flags.BreastTenderness = true
		case "back_pain":
			flags.BackPain = true
		case "nausea":
			flags.Nausea = true
		case "spotting":
			flags.Spotting = true
		case "irritability":
			flags.Irritability = true
		case "insomnia":
			flags.Insomnia = true
		case "food_cravings":
			flags.FoodCravings = true
		case "diarrhea":
			flags.Diarrhea = true
		case "constipation":
			flags.Constipation = true
		default:
			trimmed := strings.TrimSpace(name)
			if trimmed != "" {
				otherSet[trimmed] = struct{}{}
			}
		}
	}

	other := make([]string, 0, len(otherSet))
	for name := range otherSet {
		other = append(other, name)
	}
	sort.Strings(other)

	return flags, other
}

func exportSymptomColumn(name string) string {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "cramps":
		return "cramps"
	case "headache":
		return "headache"
	case "acne":
		return "acne"
	case "mood", "mood swings":
		return "mood"
	case "bloating":
		return "bloating"
	case "fatigue", "fatique":
		return "fatigue"
	case "breast tenderness":
		return "breast_tenderness"
	case "back pain":
		return "back_pain"
	case "nausea":
		return "nausea"
	case "spotting":
		return "spotting"
	case "irritability":
		return "irritability"
	case "insomnia":
		return "insomnia"
	case "food cravings":
		return "food_cravings"
	case "diarrhea":
		return "diarrhea"
	case "constipation":
		return "constipation"
	default:
		return "other"
	}
}

func csvYesNo(value bool) string {
	if value {
		return "Yes"
	}
	return "No"
}

func csvFlowLabel(flow string) string {
	switch strings.ToLower(strings.TrimSpace(flow)) {
	case models.FlowLight:
		return "Light"
	case models.FlowMedium:
		return "Medium"
	case models.FlowHeavy:
		return "Heavy"
	default:
		return "None"
	}
}

func normalizeExportFlow(flow string) string {
	switch strings.ToLower(strings.TrimSpace(flow)) {
	case models.FlowLight:
		return models.FlowLight
	case models.FlowMedium:
		return models.FlowMedium
	case models.FlowHeavy:
		return models.FlowHeavy
	default:
		return models.FlowNone
	}
}

func parseBoolValue(value string) bool {
	normalized := strings.ToLower(strings.TrimSpace(value))
	return normalized == "1" || normalized == "true" || normalized == "on" || normalized == "yes"
}

func parseDayParam(raw string, location *time.Location) (time.Time, error) {
	if raw == "" {
		return time.Time{}, errors.New("date is required")
	}
	parsed, err := time.ParseInLocation("2006-01-02", raw, location)
	if err != nil {
		return time.Time{}, err
	}
	return dateAtLocation(parsed, location), nil
}

func parseMonthQuery(raw string, now time.Time, location *time.Location) (time.Time, error) {
	if raw == "" {
		current := dateAtLocation(now, location)
		return time.Date(current.Year(), current.Month(), 1, 0, 0, 0, 0, location), nil
	}
	parsed, err := time.ParseInLocation("2006-01", raw, location)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(parsed.Year(), parsed.Month(), 1, 0, 0, 0, 0, location), nil
}

func sameCalendarDay(a time.Time, b time.Time) bool {
	return a.Format("2006-01-02") == b.Format("2006-01-02")
}

func betweenCalendarDaysInclusive(day time.Time, start time.Time, end time.Time) bool {
	if start.IsZero() || end.IsZero() {
		return false
	}
	return (day.Equal(start) || day.After(start)) && (day.Equal(end) || day.Before(end))
}

func sanitizeLogForPartner(entry models.DailyLog) models.DailyLog {
	entry.Notes = ""
	entry.SymptomIDs = []uint{}
	return entry
}

func dateAtLocation(value time.Time, location *time.Location) time.Time {
	localized := value.In(location)
	year, month, day := localized.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, location)
}

func symptomIDSet(ids []uint) map[uint]bool {
	set := make(map[uint]bool, len(ids))
	for _, id := range ids {
		set[id] = true
	}
	return set
}

func removeUint(values []uint, needle uint) []uint {
	filtered := make([]uint, 0, len(values))
	for _, value := range values {
		if value != needle {
			filtered = append(filtered, value)
		}
	}
	return filtered
}

func postLoginRedirectPath(user *models.User) string {
	if requiresOnboarding(user) {
		return "/onboarding"
	}
	return "/dashboard"
}

func requiresOnboarding(user *models.User) bool {
	if user == nil {
		return false
	}
	return user.Role == models.RoleOwner && !user.OnboardingCompleted
}

func redirectOrJSON(c *fiber.Ctx, path string) error {
	if isHTMX(c) {
		c.Set("HX-Redirect", path)
		return c.SendStatus(fiber.StatusOK)
	}
	if acceptsJSON(c) {
		return c.JSON(fiber.Map{"ok": true})
	}
	return c.Redirect(path, fiber.StatusSeeOther)
}

func apiError(c *fiber.Ctx, status int, message string) error {
	if isHTMX(c) {
		rendered := message
		if key := authErrorTranslationKey(message); key != "" {
			if localized := translateMessage(currentMessages(c), key); localized != key {
				rendered = localized
			}
		}
		return c.Status(status).SendString(fmt.Sprintf("<div class=\"status-error\">%s</div>", template.HTMLEscapeString(rendered)))
	}
	return c.Status(status).JSON(fiber.Map{"error": message})
}

func (handler *Handler) respondAuthError(c *fiber.Ctx, status int, message string) error {
	if strings.HasPrefix(c.Path(), "/api/auth/") && !acceptsJSON(c) && !isHTMX(c) {
		errorParam := "error=" + url.QueryEscape(message)
		switch c.Path() {
		case "/api/auth/register":
			return c.Redirect("/register?"+errorParam, fiber.StatusSeeOther)
		case "/api/auth/forgot-password":
			return c.Redirect("/forgot-password?"+errorParam, fiber.StatusSeeOther)
		case "/api/auth/reset-password":
			token := strings.TrimSpace(c.FormValue("token"))
			if token == "" {
				token = strings.TrimSpace(c.Query("token"))
			}
			if token == "" {
				return c.Redirect("/reset-password?"+errorParam, fiber.StatusSeeOther)
			}
			return c.Redirect("/reset-password?token="+url.QueryEscape(token)+"&"+errorParam, fiber.StatusSeeOther)
		default:
			return c.Redirect("/login?"+errorParam, fiber.StatusSeeOther)
		}
	}
	return apiError(c, status, message)
}

func (handler *Handler) respondSettingsError(c *fiber.Ctx, status int, message string) error {
	if isHTMX(c) {
		rendered := message
		if key := authErrorTranslationKey(message); key != "" {
			if localized := translateMessage(currentMessages(c), key); localized != key {
				rendered = localized
			}
		}
		return c.Status(fiber.StatusOK).SendString(fmt.Sprintf("<div class=\"status-error\">%s</div>", template.HTMLEscapeString(rendered)))
	}
	if (strings.HasPrefix(c.Path(), "/api/settings/") || strings.HasPrefix(c.Path(), "/settings/")) && !acceptsJSON(c) {
		errorParam := "error=" + url.QueryEscape(message)
		return c.Redirect("/settings?"+errorParam, fiber.StatusSeeOther)
	}
	return apiError(c, status, message)
}

func acceptsJSON(c *fiber.Ctx) bool {
	return strings.Contains(strings.ToLower(c.Get("Accept")), "application/json")
}

func isHTMX(c *fiber.Ctx) bool {
	return strings.EqualFold(c.Get("HX-Request"), "true")
}

func csrfToken(c *fiber.Ctx) string {
	token, _ := c.Locals("csrf").(string)
	return token
}

func localizedPageTitle(messages map[string]string, key string, fallback string) string {
	title := translateMessage(messages, key)
	if title == key || strings.TrimSpace(title) == "" {
		return fallback
	}
	return title
}

func sanitizeRedirectPath(raw string, fallback string) string {
	candidate := strings.TrimSpace(raw)
	if candidate == "" {
		return fallback
	}
	if strings.HasPrefix(candidate, "//") || !strings.HasPrefix(candidate, "/") {
		return fallback
	}
	parsed, err := url.Parse(candidate)
	if err != nil || parsed.IsAbs() {
		return fallback
	}
	return candidate
}

func newAttemptLimiter() *attemptLimiter {
	return &attemptLimiter{
		attempts: make(map[string][]time.Time),
	}
}

func (limiter *attemptLimiter) tooManyRecent(key string, now time.Time, limit int, window time.Duration) bool {
	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	pruned := limiter.pruneLocked(key, now, window)
	return len(pruned) >= limit
}

func (limiter *attemptLimiter) addFailure(key string, now time.Time, window time.Duration) {
	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	pruned := limiter.pruneLocked(key, now, window)
	pruned = append(pruned, now)
	limiter.attempts[key] = pruned
}

func (limiter *attemptLimiter) reset(key string) {
	limiter.mu.Lock()
	defer limiter.mu.Unlock()
	delete(limiter.attempts, key)
}

func (limiter *attemptLimiter) pruneLocked(key string, now time.Time, window time.Duration) []time.Time {
	values := limiter.attempts[key]
	if len(values) == 0 {
		return []time.Time{}
	}

	threshold := now.Add(-window)
	pruned := make([]time.Time, 0, len(values))
	for _, value := range values {
		if value.After(threshold) {
			pruned = append(pruned, value)
		}
	}

	if len(pruned) == 0 {
		delete(limiter.attempts, key)
		return []time.Time{}
	}

	limiter.attempts[key] = pruned
	return pruned
}

func requestLimiterKey(c *fiber.Ctx) string {
	key := strings.TrimSpace(c.IP())
	if key == "" {
		return "unknown"
	}
	return key
}

func normalizeRecoveryCode(raw string) string {
	normalized := strings.ToUpper(strings.TrimSpace(raw))
	normalized = strings.ReplaceAll(normalized, " ", "")
	normalized = strings.ReplaceAll(normalized, "-", "")
	if strings.HasPrefix(normalized, "LUME") {
		normalized = strings.TrimPrefix(normalized, "LUME")
	}
	if len(normalized) != 12 {
		return strings.ToUpper(strings.TrimSpace(raw))
	}
	return fmt.Sprintf("LUME-%s-%s-%s", normalized[:4], normalized[4:8], normalized[8:12])
}

func generateRecoveryCodeHash() (string, string, error) {
	code, err := generateRecoveryCode()
	if err != nil {
		return "", "", err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}
	return code, string(hash), nil
}

func generateRecoveryCode() (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	randomBytes := make([]byte, 12)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	chars := make([]byte, 12)
	for index, value := range randomBytes {
		chars[index] = alphabet[int(value)%len(alphabet)]
	}

	return fmt.Sprintf("LUME-%s-%s-%s", string(chars[:4]), string(chars[4:8]), string(chars[8:12])), nil
}

func (handler *Handler) findUserByRecoveryCode(code string) (*models.User, error) {
	users := make([]models.User, 0)
	if err := handler.db.Where("recovery_code_hash <> ''").Find(&users).Error; err != nil {
		return nil, err
	}

	for index := range users {
		hash := strings.TrimSpace(users[index].RecoveryCodeHash)
		if hash == "" {
			continue
		}
		if bcrypt.CompareHashAndPassword([]byte(hash), []byte(code)) == nil {
			return &users[index], nil
		}
	}
	return nil, errors.New("recovery code not found")
}
