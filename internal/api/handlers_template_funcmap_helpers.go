package api

import (
	"encoding/json"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/terraincognita07/lume/internal/models"
)

func formatTemplateDate(value time.Time, layout string) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(layout)
}

func formatTemplateFloat(value float64) string {
	return fmt.Sprintf("%.1f", value)
}

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

func templateRoleLabel(messages map[string]string, role string) string {
	return translateMessage(messages, roleTranslationKey(role))
}

func templateUserIdentity(user *models.User) string {
	if user == nil {
		return ""
	}
	if displayName := strings.TrimSpace(user.DisplayName); displayName != "" {
		return displayName
	}
	return strings.TrimSpace(user.Email)
}

func isActiveTemplateRoute(currentPath string, route string) bool {
	path := strings.TrimSpace(currentPath)
	if path == "" {
		return route == "/"
	}
	if route == "/" {
		return path == "/" || strings.HasPrefix(path, "/?")
	}
	return path == route || strings.HasPrefix(path, route+"?") || strings.HasPrefix(path, route+"/")
}

func hasTemplateSymptom(set map[uint]bool, id uint) bool {
	return set[id]
}

func templateToJSON(value any) template.JS {
	serialized, err := json.Marshal(value)
	if err != nil {
		return template.JS("null")
	}
	return template.JS(serialized)
}

func templateDict(values ...any) (map[string]any, error) {
	if len(values)%2 != 0 {
		return nil, fmt.Errorf("dict requires key-value pairs")
	}
	result := make(map[string]any, len(values)/2)
	for index := 0; index < len(values); index += 2 {
		key, ok := values[index].(string)
		if !ok {
			return nil, fmt.Errorf("dict key at index %d is not a string", index)
		}
		result[key] = values[index+1]
	}
	return result, nil
}
