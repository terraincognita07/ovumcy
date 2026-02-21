package api

import (
	"encoding/json"
	"fmt"
	"html/template"
	"time"
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

func templateToJSON(value any) template.JS {
	serialized, err := json.Marshal(value)
	if err != nil {
		return template.JS("null")
	}
	return template.JS(serialized)
}
