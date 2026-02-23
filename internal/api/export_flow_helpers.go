package api

import (
	"strings"

	"github.com/terraincognita07/ovumcy/internal/models"
)

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
