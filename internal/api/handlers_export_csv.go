package api

import (
	"bytes"
	"encoding/csv"
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
