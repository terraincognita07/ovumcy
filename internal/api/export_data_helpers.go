package api

import (
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
)

func (handler *Handler) fetchExportData(userID uint, from *time.Time, to *time.Time) ([]models.DailyLog, map[uint]string, error) {
	handler.ensureDependencies()
	return handler.exportService.LoadDataForRange(userID, from, to, handler.location)
}

func (handler *Handler) fetchExportSummaryForRange(userID uint, from *time.Time, to *time.Time) (int64, string, string, error) {
	handler.ensureDependencies()
	summary, err := handler.exportService.BuildSummary(userID, from, to, handler.location)
	if err != nil {
		return 0, "", "", err
	}
	return int64(summary.TotalEntries),
		summary.DateFrom,
		summary.DateTo,
		nil
}
