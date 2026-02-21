package api

import "strings"

func normalizeLegacySymptomName(name string) string {
	if strings.EqualFold(strings.TrimSpace(name), "fatique") {
		return "Fatigue"
	}
	return name
}
