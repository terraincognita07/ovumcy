package api

import "strings"

func templateTranslate(messages map[string]string, key string) string {
	return translateMessage(messages, key)
}

func templatePhaseLabel(messages map[string]string, phase string) string {
	return translateMessage(messages, phaseTranslationKey(phase))
}

func templatePhaseIcon(phase string) string {
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
}

func templateFlowLabel(messages map[string]string, flow string) string {
	return translateMessage(messages, flowTranslationKey(flow))
}

func templateSymptomLabel(messages map[string]string, name string) string {
	return localizedSymptomName(messages, name)
}

func templateSymptomGroup(name string) string {
	normalized := strings.ToLower(strings.TrimSpace(name))
	switch normalized {
	case "cramps", "headache", "breast tenderness", "back pain":
		return "pain"
	case "mood swings", "fatigue", "irritability", "insomnia":
		return "mood"
	case "bloating", "nausea", "diarrhea", "constipation", "swelling", "food cravings":
		return "digestion"
	case "acne", "spotting":
		return "skin"
	default:
		return "other"
	}
}

func templateRoleLabel(messages map[string]string, role string) string {
	return translateMessage(messages, roleTranslationKey(role))
}
