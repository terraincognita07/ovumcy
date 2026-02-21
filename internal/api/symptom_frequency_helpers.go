package api

import (
	"sort"

	"github.com/terraincognita07/lume/internal/models"
)

func (handler *Handler) calculateSymptomFrequencies(userID uint, logs []models.DailyLog) ([]SymptomCount, error) {
	if len(logs) == 0 {
		return []SymptomCount{}, nil
	}
	totalDays := len(logs)

	counts := make(map[uint]int)
	for _, logEntry := range logs {
		for _, id := range logEntry.SymptomIDs {
			counts[id]++
		}
	}
	if len(counts) == 0 {
		return []SymptomCount{}, nil
	}

	symptoms, err := handler.fetchSymptoms(userID)
	if err != nil {
		return nil, err
	}

	symptomByID := make(map[uint]models.SymptomType, len(symptoms))
	for _, symptom := range symptoms {
		symptomByID[symptom.ID] = symptom
	}

	result := make([]SymptomCount, 0, len(counts))
	for id, count := range counts {
		if symptom, ok := symptomByID[id]; ok {
			result = append(result, SymptomCount{Name: symptom.Name, Icon: symptom.Icon, Count: count, TotalDays: totalDays})
		}
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Count == result[j].Count {
			return result[i].Name < result[j].Name
		}
		return result[i].Count > result[j].Count
	})

	return result, nil
}
