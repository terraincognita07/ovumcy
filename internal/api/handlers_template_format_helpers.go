package api

import (
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"time"
)

func formatTemplateDate(value time.Time, layout string) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(layout)
}

func formatTemplateFloat(value float64) string {
	rounded := math.Round(value*10) / 10
	if math.Abs(rounded-math.Round(rounded)) < 1e-9 {
		return fmt.Sprintf("%.0f", rounded)
	}
	return fmt.Sprintf("%.1f", rounded)
}

func templateToJSON(value any) template.JS {
	serialized, err := json.Marshal(value)
	if err != nil {
		return template.JS("null")
	}
	return template.JS(serialized)
}
