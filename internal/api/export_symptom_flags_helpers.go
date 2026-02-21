package api

import (
	"sort"
	"strings"
)

func buildCSVSymptomColumns(symptomIDs []uint, symptomNames map[uint]string) (exportSymptomFlags, []string) {
	flags := exportSymptomFlags{}
	otherSet := make(map[string]struct{})

	for _, symptomID := range symptomIDs {
		name, ok := symptomNames[symptomID]
		if !ok {
			continue
		}

		if setExportSymptomFlag(&flags, exportSymptomColumn(name)) {
			continue
		}

		trimmed := strings.TrimSpace(name)
		if trimmed != "" {
			otherSet[trimmed] = struct{}{}
		}
	}

	other := make([]string, 0, len(otherSet))
	for name := range otherSet {
		other = append(other, name)
	}
	sort.Strings(other)

	return flags, other
}

func setExportSymptomFlag(flags *exportSymptomFlags, column string) bool {
	switch column {
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
		return false
	}
	return true
}
