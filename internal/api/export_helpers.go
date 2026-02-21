package api

import (
	"sort"
	"strings"

	"github.com/terraincognita07/lume/internal/models"
)

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
