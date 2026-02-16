package api

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/lume/internal/models"
)

var authErrorKeys = map[string]string{
	"invalid input":         "auth.error.invalid_input",
	"registration disabled": "auth.error.registration_disabled",
	"invalid credentials":   "auth.error.invalid_credentials",
	"email already exists":  "auth.error.email_exists",
}

var builtinSymptomKeys = map[string]string{
	"acne":              "symptoms.acne",
	"bloating":          "symptoms.bloating",
	"breast tenderness": "symptoms.breast_tenderness",
	"cramps":            "symptoms.cramps",
	"fatigue":           "symptoms.fatigue",
	"headache":          "symptoms.headache",
	"mood swings":       "symptoms.mood_swings",
}

var monthNames = map[string][]string{
	"en": {"January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"},
	"ru": {"Январь", "Февраль", "Март", "Апрель", "Май", "Июнь", "Июль", "Август", "Сентябрь", "Октябрь", "Ноябрь", "Декабрь"},
}

var weekdayShortNames = map[string][]string{
	"en": {"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"},
	"ru": {"Вс", "Пн", "Вт", "Ср", "Чт", "Пт", "Сб"},
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
