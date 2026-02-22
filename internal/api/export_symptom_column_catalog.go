package api

import "strings"

var exportSymptomColumnsByName = map[string]string{
	"cramps":            "cramps",
	"headache":          "headache",
	"acne":              "acne",
	"mood":              "mood",
	"mood swings":       "mood",
	"bloating":          "bloating",
	"fatigue":           "fatigue",
	"breast tenderness": "breast_tenderness",
	"back pain":         "back_pain",
	"nausea":            "nausea",
	"spotting":          "spotting",
	"irritability":      "irritability",
	"insomnia":          "insomnia",
	"food cravings":     "food_cravings",
	"diarrhea":          "diarrhea",
	"constipation":      "constipation",
}

func exportSymptomColumn(name string) string {
	normalized := strings.ToLower(strings.TrimSpace(name))
	if column, ok := exportSymptomColumnsByName[normalized]; ok {
		return column
	}
	return "other"
}
