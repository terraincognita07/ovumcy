package api

import (
	"encoding/json"
	"fmt"
	"html"
	"regexp"
)

var statsChartDataPattern = regexp.MustCompile(`data-chart='([^']+)'`)

type statsChartPayload struct {
	Labels []string `json:"labels"`
	Values []int    `json:"values"`
}

func extractStatsChartPayload(rendered string) (statsChartPayload, error) {
	matches := statsChartDataPattern.FindStringSubmatch(rendered)
	if len(matches) != 2 {
		return statsChartPayload{}, fmt.Errorf("data-chart attribute not found")
	}

	rawJSON := html.UnescapeString(matches[1])
	payload := statsChartPayload{}
	if err := json.Unmarshal([]byte(rawJSON), &payload); err != nil {
		return statsChartPayload{}, err
	}
	return payload, nil
}
