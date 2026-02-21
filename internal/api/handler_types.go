package api

import (
	"html/template"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/terraincognita07/lume/internal/i18n"
	"gorm.io/gorm"
)

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

type FlashPayload struct {
	AuthError       string `json:"auth_error,omitempty"`
	SettingsError   string `json:"settings_error,omitempty"`
	SettingsSuccess string `json:"settings_success,omitempty"`
	LoginEmail      string `json:"login_email,omitempty"`
	RegisterEmail   string `json:"register_email,omitempty"`
}

const (
	defaultAuthTokenTTL  = 7 * 24 * time.Hour
	rememberAuthTokenTTL = 30 * 24 * time.Hour
)

type passwordResetClaims struct {
	UserID  uint   `json:"uid"`
	Purpose string `json:"purpose"`
	jwt.RegisteredClaims
}
