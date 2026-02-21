package api

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

func (handler *Handler) ExportCSV(c *fiber.Ctx) error {
	user, from, to, status, message := handler.exportUserAndRange(c)
	if status != 0 {
		return apiError(c, status, message)
	}

	logs, symptomNames, err := handler.fetchExportData(user.ID, from, to)
	if err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to fetch logs")
	}
	now := time.Now().In(handler.location)

	var output bytes.Buffer
	writer := csv.NewWriter(&output)
	if err := writer.Write(exportCSVHeaders); err != nil {
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

	setExportAttachmentHeaders(c, "text/csv", buildExportFilename(now, "csv"))
	return c.Send(output.Bytes())
}

func (handler *Handler) ExportSummary(c *fiber.Ctx) error {
	user, from, to, status, message := handler.exportUserAndRange(c)
	if status != 0 {
		return apiError(c, status, message)
	}

	totalEntries, firstDate, lastDate, err := handler.fetchExportSummaryForRange(user.ID, from, to)
	if err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to fetch logs")
	}

	return c.JSON(fiber.Map{
		"total_entries": int(totalEntries),
		"has_data":      totalEntries > 0,
		"date_from":     firstDate,
		"date_to":       lastDate,
	})
}

func (handler *Handler) ExportJSON(c *fiber.Ctx) error {
	user, from, to, status, message := handler.exportUserAndRange(c)
	if status != 0 {
		return apiError(c, status, message)
	}

	logs, symptomNames, err := handler.fetchExportData(user.ID, from, to)
	if err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to fetch logs")
	}
	now := time.Now().In(handler.location)

	entries := buildExportJSONEntries(logs, symptomNames, handler.location)

	payload := fiber.Map{
		"exported_at": now.Format(time.RFC3339),
		"entries":     entries,
	}

	serialized, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to build export")
	}

	setExportAttachmentHeaders(c, fiber.MIMEApplicationJSON, buildExportFilename(now, "json"))
	return c.Send(serialized)
}
