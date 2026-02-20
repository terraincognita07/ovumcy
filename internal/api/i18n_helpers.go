package api

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
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
	"invalid password":                              "settings.error.invalid_password",
	"period flow is required":                       "calendar.error.period_flow_required",
	"invalid last period start":                     "onboarding.error.invalid_last_period_start",
	"last period start must be within last 60 days": "onboarding.error.last_period_range",
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

var monthNames = map[string][]string{
	"en": {"January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"},
	"ru": {"Январь", "Февраль", "Март", "Апрель", "Май", "Июнь", "Июль", "Август", "Сентябрь", "Октябрь", "Ноябрь", "Декабрь"},
}

var monthLongNames = map[string][]string{
	"en": {"January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"},
	"ru": {"января", "февраля", "марта", "апреля", "мая", "июня", "июля", "августа", "сентября", "октября", "ноября", "декабря"},
}

var weekdayShortNames = map[string][]string{
	"en": {"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"},
	"ru": {"Вс", "Пн", "Вт", "Ср", "Чт", "Пт", "Сб"},
}

var weekdayLongNames = map[string][]string{
	"en": {"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"},
	"ru": {"воскресенье", "понедельник", "вторник", "среда", "четверг", "пятница", "суббота"},
}

var monthShortNames = map[string][]string{
	"en": {"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"},
	"ru": {"Янв", "Фев", "Мар", "Апр", "Май", "Июн", "Июл", "Авг", "Сен", "Окт", "Ноя", "Дек"},
}

func translateMessage(messages map[string]string, key string) string {
	if key == "" {
		return ""
	}
	if messages != nil {
		if value, ok := messages[key]; ok && strings.TrimSpace(value) != "" {
			return value
		}
	}
	return key
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

func currentLanguage(c *fiber.Ctx) string {
	language, ok := c.Locals(contextLanguageKey).(string)
	if !ok || strings.TrimSpace(language) == "" {
		return ""
	}
	return language
}

func currentMessages(c *fiber.Ctx) map[string]string {
	messages, ok := c.Locals(contextMessagesKey).(map[string]string)
	if !ok || messages == nil {
		return map[string]string{}
	}
	return messages
}

func (handler *Handler) withTemplateDefaults(c *fiber.Ctx, data fiber.Map) fiber.Map {
	if data == nil {
		data = fiber.Map{}
	}

	messages := currentMessages(c)
	if _, ok := data["Messages"]; !ok {
		data["Messages"] = messages
	}

	if _, ok := data["Lang"]; !ok {
		language := currentLanguage(c)
		if language == "" {
			language = handler.i18n.DefaultLanguage()
		}
		data["Lang"] = language
	}

	if _, ok := data["CurrentPath"]; !ok {
		data["CurrentPath"] = currentPathWithQuery(c)
	}

	if _, ok := data["CSRFToken"]; !ok {
		data["CSRFToken"] = csrfToken(c)
	}

	if _, ok := data["NoDataLabel"]; !ok {
		noData := translateMessage(messages, "common.not_available")
		if noData == "common.not_available" {
			noData = "-"
		}
		data["NoDataLabel"] = noData
	}

	return data
}

func currentPathWithQuery(c *fiber.Ctx) string {
	path := string(c.Request().URI().RequestURI())
	if path == "" {
		return c.Path()
	}
	return path
}

func localizedMonthYear(language string, value time.Time) string {
	names, ok := monthNames[strings.ToLower(language)]
	if !ok || len(names) < 12 {
		return value.Format("January 2006")
	}
	monthIndex := int(value.Month()) - 1
	if monthIndex < 0 || monthIndex >= len(names) {
		return value.Format("January 2006")
	}
	return fmt.Sprintf("%s %d", names[monthIndex], value.Year())
}

func localizedDateLabel(language string, value time.Time) string {
	lang := strings.ToLower(strings.TrimSpace(language))
	weekdays, weekdaysOK := weekdayShortNames[lang]
	months, monthsOK := monthShortNames[lang]
	if !weekdaysOK || !monthsOK {
		return value.Format("Mon, Jan 2")
	}
	monthIndex := int(value.Month()) - 1
	if monthIndex < 0 || monthIndex >= len(months) {
		return value.Format("Mon, Jan 2")
	}

	weekday := weekdays[int(value.Weekday())]
	month := months[monthIndex]
	return fmt.Sprintf("%s, %s %d", weekday, month, value.Day())
}

func localizedDashboardDate(language string, value time.Time) string {
	lang := strings.ToLower(strings.TrimSpace(language))
	weekdays, weekdaysOK := weekdayLongNames[lang]
	months, monthsOK := monthLongNames[lang]
	if !weekdaysOK || !monthsOK {
		return value.Format("January 2, 2006, Monday")
	}
	monthIndex := int(value.Month()) - 1
	if monthIndex < 0 || monthIndex >= len(months) {
		return value.Format("January 2, 2006, Monday")
	}

	weekday := weekdays[int(value.Weekday())]
	month := months[monthIndex]
	if lang == "ru" {
		return fmt.Sprintf("%d %s %d, %s", value.Day(), month, value.Year(), weekday)
	}
	return fmt.Sprintf("%s %d, %d, %s", month, value.Day(), value.Year(), weekday)
}

func localizedSymptomFrequencySummary(language string, count int, days int) string {
	lang := strings.ToLower(strings.TrimSpace(language))
	if lang == "ru" {
		return fmt.Sprintf("%d %s (за %d %s)",
			count,
			russianPluralForm(count, "раз", "раза", "раз"),
			days,
			russianPluralForm(days, "день", "дня", "дней"),
		)
	}

	countWord := "times"
	if count == 1 {
		countWord = "time"
	}
	dayWord := "days"
	if days == 1 {
		dayWord = "day"
	}
	return fmt.Sprintf("%d %s (in %d %s)", count, countWord, days, dayWord)
}

func russianPluralForm(value int, one string, few string, many string) string {
	absolute := value
	if absolute < 0 {
		absolute = -absolute
	}
	lastTwoDigits := absolute % 100
	if lastTwoDigits >= 11 && lastTwoDigits <= 14 {
		return many
	}

	lastDigit := absolute % 10
	switch {
	case lastDigit == 1:
		return one
	case lastDigit >= 2 && lastDigit <= 4:
		return few
	default:
		return many
	}
}
