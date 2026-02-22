package api

import "strings"

var authErrorKeys = map[string]string{
	"invalid input":                                   "auth.error.invalid_input",
	"registration disabled":                           "auth.error.registration_disabled",
	"invalid credentials":                             "auth.error.invalid_credentials",
	"email already exists":                            "auth.error.email_exists",
	"weak password":                                   "auth.error.weak_password",
	"password mismatch":                               "auth.error.password_mismatch",
	"invalid recovery code":                           "auth.error.invalid_recovery_code",
	"too many recovery attempts":                      "auth.error.too_many_recovery_attempts",
	"too_many_login_attempts":                         "auth.error.too_many_login_attempts",
	"too many login attempts":                         "auth.error.too_many_login_attempts",
	"too_many_forgot_password_attempts":               "auth.error.too_many_forgot_password_attempts",
	"too many forgot password attempts":               "auth.error.too_many_forgot_password_attempts",
	"invalid reset token":                             "auth.error.invalid_reset_token",
	"invalid current password":                        "settings.error.invalid_current_password",
	"new password must differ":                        "settings.error.password_unchanged",
	"invalid settings input":                          "settings.error.invalid_input",
	"invalid profile input":                           "settings.error.invalid_profile_input",
	"display name too long":                           "settings.error.display_name_too_long",
	"invalid password":                                "settings.error.invalid_password",
	"period flow is required":                         "calendar.error.period_flow_required",
	"date is required":                                "onboarding.error.date_required",
	"invalid last period start":                       "onboarding.error.invalid_last_period_start",
	"last period start must be within last 60 days":   "onboarding.error.last_period_range",
	"period status is required":                       "onboarding.error.period_status_required",
	"invalid period status":                           "onboarding.error.invalid_period_status",
	"period end is required":                          "onboarding.error.period_end_required",
	"invalid period end":                              "onboarding.error.invalid_period_end",
	"period end must be between start and today":      "onboarding.error.period_end_range",
	"cycle length must be between 15 and 90":          "onboarding.error.cycle_length_range",
	"period length must be between 1 and 10":          "onboarding.error.period_length_range",
	"period length must be between 1 and 14":          "onboarding.error.period_length_range",
	"period length must not exceed cycle length":      "onboarding.error.period_length_exceeds_cycle",
	"period length is incompatible with cycle length": "onboarding.error.period_cycle_incompatible",
	"complete onboarding steps first":                 "onboarding.error.incomplete",
	"failed to save onboarding step":                  "onboarding.error.generic",
	"failed to finish onboarding":                     "onboarding.error.generic",
}

func authErrorTranslationKey(message string) string {
	key, ok := authErrorKeys[strings.ToLower(strings.TrimSpace(message))]
	if !ok {
		return ""
	}
	return key
}
