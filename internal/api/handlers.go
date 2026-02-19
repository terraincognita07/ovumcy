package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/terraincognita07/lume/internal/i18n"
	"github.com/terraincognita07/lume/internal/models"
	"github.com/terraincognita07/lume/internal/services"
	"gorm.io/gorm"
)

var hexColorRegex = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
var recoveryCodeRegex = regexp.MustCompile(`^LUME-[A-Z0-9]{4}-[A-Z0-9]{4}-[A-Z0-9]{4}$`)
var passwordLengthRegex = regexp.MustCompile(`^.{8,}$`)
var passwordUpperRegex = regexp.MustCompile(`\p{Lu}`)
var passwordLowerRegex = regexp.MustCompile(`\p{Ll}`)
var passwordDigitRegex = regexp.MustCompile(`\d`)

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
	Email           string `json:"email" form:"email"`
	Password        string `json:"password" form:"password"`
	ConfirmPassword string `json:"confirm_password" form:"confirm_password"`
	RememberMe      bool   `json:"remember_me" form:"remember_me"`
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
	ConfirmPassword string `json:"confirm_password" form:"confirm_password"`
}

type FlashPayload struct {
	AuthError       string `json:"auth_error,omitempty"`
	SettingsError   string `json:"settings_error,omitempty"`
	SettingsSuccess string `json:"settings_success,omitempty"`
	LoginEmail      string `json:"login_email,omitempty"`
}

const (
	defaultAuthTokenTTL  = 7 * 24 * time.Hour
	rememberAuthTokenTTL = 30 * 24 * time.Hour
)

type cycleSettingsInput struct {
	CycleLength  int `json:"cycle_length" form:"cycle_length"`
	PeriodLength int `json:"period_length" form:"period_length"`
}

type deleteAccountInput struct {
	Password string `json:"password" form:"password"`
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

func (handler *Handler) setAuthCookie(c *fiber.Ctx, user *models.User, rememberMe bool) error {
	tokenTTL := defaultAuthTokenTTL
	if rememberMe {
		tokenTTL = rememberAuthTokenTTL
	}

	token, err := handler.buildToken(user, tokenTTL)
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
	}
	if rememberMe {
		cookie.Expires = time.Now().Add(tokenTTL)
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

func (handler *Handler) buildToken(user *models.User, ttl time.Duration) (string, error) {
	if ttl <= 0 {
		ttl = defaultAuthTokenTTL
	}
	now := time.Now()

	claims := authClaims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatUint(uint64(user.ID), 10),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
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

func (handler *Handler) buildCalendarDays(monthStart time.Time, logs []models.DailyLog, stats services.CycleStats, now time.Time) []CalendarDay {
	monthEnd := monthStart.AddDate(0, 1, -1)
	gridStart := monthStart.AddDate(0, 0, -int(monthStart.Weekday()))
	gridEnd := monthEnd.AddDate(0, 0, 6-int(monthEnd.Weekday()))

	latestLogByDate := make(map[string]models.DailyLog)
	hasDataMap := make(map[string]bool)
	for _, logEntry := range logs {
		key := dateAtLocation(logEntry.Date, handler.location).Format("2006-01-02")
		existing, exists := latestLogByDate[key]
		if !exists || logEntry.Date.After(existing.Date) || (logEntry.Date.Equal(existing.Date) && logEntry.ID > existing.ID) {
			latestLogByDate[key] = logEntry
		}
		hasDataMap[key] = hasDataMap[key] || dayHasData(logEntry)
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
		entry, hasEntry := latestLogByDate[key]
		isPeriod := hasEntry && entry.IsPeriod
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
