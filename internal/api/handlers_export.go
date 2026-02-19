package api

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

func (handler *Handler) ExportCSV(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return apiError(c, fiber.StatusUnauthorized, "unauthorized")
	}

	logs, symptomNames, err := handler.fetchExportData(user.ID)
	if err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to fetch logs")
	}

	var output bytes.Buffer
	writer := csv.NewWriter(&output)
	if err := writer.Write([]string{
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
	}); err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to build export")
	}

	for _, logEntry := range logs {
		flags, other := buildCSVSymptomColumns(logEntry.SymptomIDs, symptomNames)
		if err := writer.Write([]string{
			dateAtLocation(logEntry.Date, handler.location).Format("2006-01-02"),
			csvYesNo(logEntry.IsPeriod),
			csvFlowLabel(logEntry.Flow),
			csvYesNo(flags.Cramps),
			csvYesNo(flags.Headache),
			csvYesNo(flags.Acne),
			csvYesNo(flags.Mood),
			csvYesNo(flags.Bloating),
			csvYesNo(flags.Fatigue),
			csvYesNo(flags.BreastTenderness),
			csvYesNo(flags.BackPain),
			csvYesNo(flags.Nausea),
			csvYesNo(flags.Spotting),
			csvYesNo(flags.Irritability),
			csvYesNo(flags.Insomnia),
			csvYesNo(flags.FoodCravings),
			csvYesNo(flags.Diarrhea),
			csvYesNo(flags.Constipation),
			strings.Join(other, "; "),
			logEntry.Notes,
		}); err != nil {
			return apiError(c, fiber.StatusInternalServerError, "failed to build export")
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to build export")
	}

	filename := fmt.Sprintf("lume-export-%s.csv", time.Now().In(handler.location).Format("2006-01-02"))
	c.Set(fiber.HeaderContentType, "text/csv")
	c.Set(fiber.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%s", filename))
	return c.Send(output.Bytes())
}

func (handler *Handler) ExportJSON(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return apiError(c, fiber.StatusUnauthorized, "unauthorized")
	}

	logs, symptomNames, err := handler.fetchExportData(user.ID)
	if err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to fetch logs")
	}

	entries := make([]exportJSONEntry, 0, len(logs))
	for _, logEntry := range logs {
		flags, other := buildCSVSymptomColumns(logEntry.SymptomIDs, symptomNames)
		entries = append(entries, exportJSONEntry{
			Date:   dateAtLocation(logEntry.Date, handler.location).Format("2006-01-02"),
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

	payload := fiber.Map{
		"exported_at": time.Now().In(handler.location).Format(time.RFC3339),
		"entries":     entries,
	}

	serialized, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to build export")
	}

	filename := fmt.Sprintf("lume-export-%s.json", time.Now().In(handler.location).Format("2006-01-02"))
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	c.Set(fiber.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%s", filename))
	return c.Send(serialized)
}
