package api

import (
	"net/url"
	"testing"
	"time"
)

func TestOnboardingFlowFinishedPeriodUsesExactEndDateWithoutExtension(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "flow-finished@example.com", "StrongPass1", false)

	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")
	startDay := dateAtLocation(time.Now().In(time.UTC), time.UTC).AddDate(0, 0, -7)
	endDay := startDay.AddDate(0, 0, 3)

	submitOnboardingStep1(t, app, authCookie, url.Values{
		"last_period_start": {startDay.Format("2006-01-02")},
		"period_status":     {onboardingPeriodStatusFinished},
		"period_end":        {endDay.Format("2006-01-02")},
	})

	submitOnboardingStep2(t, app, authCookie, url.Values{
		"cycle_length":     {"30"},
		"period_length":    {"8"},
		"auto_period_fill": {"true"},
	})

	submitOnboardingComplete(t, app, authCookie)

	for day := startDay; !day.After(endDay); day = day.AddDate(0, 0, 1) {
		assertOnboardingPeriodLogForDay(t, database, user.ID, day)
	}

	afterEnd := endDay.AddDate(0, 0, 1)
	assertNoOnboardingLogForDay(t, database, user.ID, afterEnd)
}
