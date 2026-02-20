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

func (handler *Handler) parseExportRange(c *fiber.Ctx) (*time.Time, *time.Time, string) {
	rawFrom := strings.TrimSpace(c.Query("from"))
	rawTo := strings.TrimSpace(c.Query("to"))

	var from *time.Time
	if rawFrom != "" {
		parsedFrom, err := parseDayParam(rawFrom, handler.location)
		if err != nil {
			return nil, nil, "invalid from date"
		}
		from = &parsedFrom
	}

	var to *time.Time
	if rawTo != "" {
		parsedTo, err := parseDayParam(rawTo, handler.location)
		if err != nil {
			return nil, nil, "invalid to date"
		}
		to = &parsedTo
	}

	if from != nil && to != nil && to.Before(*from) {
		return nil, nil, "invalid range"
	}

	return from, to, ""
}

func (handler *Handler) ExportCSV(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return apiError(c, fiber.StatusUnauthorized, "unauthorized")
	}

	from, to, rangeError := handler.parseExportRange(c)
	if rangeError != "" {
		return apiError(c, fiber.StatusBadRequest, rangeError)
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

	filename := fmt.Sprintf("lume-export-%s.csv", now.Format("2006-01-02"))
	c.Set(fiber.HeaderContentType, "text/csv")
	c.Set(fiber.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%s", filename))
	return c.Send(output.Bytes())
}

func (handler *Handler) ExportSummary(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return apiError(c, fiber.StatusUnauthorized, "unauthorized")
	}

	from, to, rangeError := handler.parseExportRange(c)
	if rangeError != "" {
		return apiError(c, fiber.StatusBadRequest, rangeError)
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
	user, ok := currentUser(c)
	if !ok {
		return apiError(c, fiber.StatusUnauthorized, "unauthorized")
	}

	from, to, rangeError := handler.parseExportRange(c)
	if rangeError != "" {
		return apiError(c, fiber.StatusBadRequest, rangeError)
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

	filename := fmt.Sprintf("lume-export-%s.json", now.Format("2006-01-02"))
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	c.Set(fiber.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%s", filename))
	return c.Send(serialized)
}
