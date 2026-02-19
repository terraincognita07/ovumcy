package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/terraincognita07/lume/internal/i18n"
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
	cookieSecure    bool
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

func NewHandler(database *gorm.DB, secret string, templateDir string, location *time.Location, i18nManager *i18n.Manager, cookieSecure bool) (*Handler, error) {
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
		cookieSecure:    cookieSecure,
		lutealPhaseDays: 14,
		i18n:            i18nManager,
		templates:       templates,
		partials:        partials,
		recoveryLimiter: newAttemptLimiter(),
	}, nil
}
