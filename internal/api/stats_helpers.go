package api

import (
	"time"

	"github.com/terraincognita07/lume/internal/models"
	"github.com/terraincognita07/lume/internal/services"
)

func (handler *Handler) buildCycleStatsForRange(user *models.User, from time.Time, to time.Time, now time.Time) (services.CycleStats, []models.DailyLog, error) {
	logs, err := handler.fetchLogsForUser(user.ID, from, to)
	if err != nil {
		return services.CycleStats{}, nil, err
	}

	stats := services.BuildCycleStats(logs, now, handler.lutealPhaseDays)
	stats = handler.applyUserCycleBaseline(user, logs, stats, now)
	return stats, logs, nil
}
