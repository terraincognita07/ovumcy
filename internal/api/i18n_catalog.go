package api

import (
	"strings"

	"github.com/terraincognita07/lume/internal/models"
)

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
