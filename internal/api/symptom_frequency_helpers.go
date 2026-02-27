package api

import "github.com/terraincognita07/ovumcy/internal/models"

func (handler *Handler) calculateSymptomFrequencies(userID uint, logs []models.DailyLog) ([]SymptomCount, error) {
	handler.ensureDependencies()
	frequencies, err := handler.symptomService.CalculateFrequencies(userID, logs)
	if err != nil {
		return nil, err
	}

	result := make([]SymptomCount, 0, len(frequencies))
	for _, item := range frequencies {
		result = append(result, SymptomCount{
			Name:      item.Name,
			Icon:      item.Icon,
			Count:     item.Count,
			TotalDays: item.TotalDays,
		})
	}
	return result, nil
}
