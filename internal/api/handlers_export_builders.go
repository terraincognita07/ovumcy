package api

import (
	"time"

	"github.com/terraincognita07/lume/internal/models"
)

var exportCSVHeaders = []string{
	"Date",
	"Period",
	"Flow",
	"Cramps",
	"Headache",
	"Acne",
	"Mood",
	"Bloating",
	"Fatigue",
	"Breast tenderness",
	"Back pain",
	"Nausea",
	"Spotting",
	"Irritability",
	"Insomnia",
	"Food cravings",
	"Diarrhea",
	"Constipation",
	"Other",
	"Notes",
}

func buildExportJSONEntries(logs []models.DailyLog, symptomNames map[uint]string, location *time.Location) []exportJSONEntry {
	entries := make([]exportJSONEntry, 0, len(logs))
	for _, logEntry := range logs {
		flags, other := buildCSVSymptomColumns(logEntry.SymptomIDs, symptomNames)
		entries = append(entries, exportJSONEntry{
			Date:   dateAtLocation(logEntry.Date, location).Format("2006-01-02"),
			Period: logEntry.IsPeriod,
			Flow:   normalizeExportFlow(logEntry.Flow),
			Symptoms: exportJSONSymptomFlags{
				Cramps:           flags.Cramps,
				Headache:         flags.Headache,
				Acne:             flags.Acne,
				Mood:             flags.Mood,
				Bloating:         flags.Bloating,
				Fatigue:          flags.Fatigue,
				BreastTenderness: flags.BreastTenderness,
				BackPain:         flags.BackPain,
				Nausea:           flags.Nausea,
				Spotting:         flags.Spotting,
				Irritability:     flags.Irritability,
				Insomnia:         flags.Insomnia,
				FoodCravings:     flags.FoodCravings,
				Diarrhea:         flags.Diarrhea,
				Constipation:     flags.Constipation,
			},
			OtherSymptoms: other,
			Notes:         logEntry.Notes,
		})
	}
	return entries
}
