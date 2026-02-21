package api

import (
	"strings"

	"github.com/terraincognita07/lume/internal/models"
)

var authErrorKeys = map[string]string{
	"invalid input":                                 "auth.error.invalid_input",
	"registration disabled":                         "auth.error.registration_disabled",
	"invalid credentials":                           "auth.error.invalid_credentials",
	"email already exists":                          "auth.error.email_exists",
	"weak password":                                 "auth.error.weak_password",
	"password mismatch":                             "auth.error.password_mismatch",
	"invalid recovery code":                         "auth.error.invalid_recovery_code",
	"too many recovery attempts":                    "auth.error.too_many_recovery_attempts",
	"too_many_login_attempts":                       "auth.error.too_many_login_attempts",
	"too many login attempts":                       "auth.error.too_many_login_attempts",
	"too_many_forgot_password_attempts":             "auth.error.too_many_forgot_password_attempts",
	"too many forgot password attempts":             "auth.error.too_many_forgot_password_attempts",
	"invalid reset token":                           "auth.error.invalid_reset_token",
	"invalid current password":                      "settings.error.invalid_current_password",
	"new password must differ":                      "settings.error.password_unchanged",
	"invalid settings input":                        "settings.error.invalid_input",
	"invalid profile input":                         "settings.error.invalid_profile_input",
	"display name too long":                         "settings.error.display_name_too_long",
	"invalid password":                              "settings.error.invalid_password",
	"period flow is required":                       "calendar.error.period_flow_required",
	"date is required":                              "onboarding.error.date_required",
	"invalid last period start":                     "onboarding.error.invalid_last_period_start",
	"last period start must be within last 60 days": "onboarding.error.last_period_range",
	"period status is required":                     "onboarding.error.period_status_required",
	"invalid period status":                         "onboarding.error.invalid_period_status",
	"period end is required":                        "onboarding.error.period_end_required",
	"invalid period end":                            "onboarding.error.invalid_period_end",
	"period end must be between start and today":    "onboarding.error.period_end_range",
	"cycle length must be between 15 and 90":        "onboarding.error.cycle_length_range",
	"period length must be between 1 and 10":        "onboarding.error.period_length_range",
	"complete onboarding steps first":               "onboarding.error.incomplete",
	"failed to save onboarding step":                "onboarding.error.generic",
	"failed to finish onboarding":                   "onboarding.error.generic",
}

var builtinSymptomKeys = map[string]string{
	"acne":              "symptoms.acne",
	"back pain":         "symptom.back_pain",
	"bloating":          "symptoms.bloating",
	"breast tenderness": "symptoms.breast_tenderness",
	"constipation":      "symptom.constipation",
	"cramps":            "symptoms.cramps",
	"diarrhea":          "symptom.diarrhea",
	"fatique":           "symptoms.fatigue",
	"fatigue":           "symptoms.fatigue",
	"food cravings":     "symptom.food_cravings",
	"headache":          "symptoms.headache",
	"insomnia":          "symptom.insomnia",
	"irritability":      "symptom.irritability",
	"mood swings":       "symptoms.mood_swings",
	"nausea":            "symptom.nausea",
	"spotting":          "symptom.spotting",
	"swelling":          "symptom.swelling",
}

func authErrorTranslationKey(message string) string {
	key, ok := authErrorKeys[strings.ToLower(strings.TrimSpace(message))]
	if !ok {
		return ""
	}
	return key
}

func settingsStatusTranslationKey(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "password_changed":
		return "settings.success.password_changed"
	case "cycle_updated":
		return "settings.success.cycle_updated"
	case "profile_updated":
		return "settings.success.profile_updated"
	case "profile_name_cleared":
		return "settings.success.profile_name_cleared"
	case "data_cleared":
		return "settings.success.data_cleared"
	default:
		return ""
	}
}

func phaseTranslationKey(phase string) string {
	switch strings.ToLower(strings.TrimSpace(phase)) {
	case "menstrual":
		return "phases.menstrual"
	case "follicular":
		return "phases.follicular"
	case "ovulation":
		return "phases.ovulation"
	case "fertile":
		return "phases.fertile"
	case "luteal":
		return "phases.luteal"
	default:
		return "phases.unknown"
	}
}

func flowTranslationKey(flow string) string {
	switch strings.ToLower(strings.TrimSpace(flow)) {
	case models.FlowLight:
		return "dashboard.flow.light"
	case models.FlowMedium:
		return "dashboard.flow.medium"
	case models.FlowHeavy:
		return "dashboard.flow.heavy"
	default:
		return "dashboard.flow.none"
	}
}

func roleTranslationKey(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case models.RoleOwner:
		return "role.owner"
	case models.RolePartner:
		return "role.partner"
	default:
		return role
	}
}

func localizedSymptomName(messages map[string]string, name string) string {
	key, ok := builtinSymptomKeys[strings.ToLower(strings.TrimSpace(name))]
	if !ok {
		return name
	}
	return translateMessage(messages, key)
}
