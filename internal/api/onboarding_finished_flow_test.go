package api

import (
	"net/url"
	"testing"
	"time"
)

func TestOnboardingFlowLegacyEndDateOverridesSliderAsExclusiveRange(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "flow-finished@example.com", "StrongPass1", false)

	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")
	startDay := dateAtLocation(time.Now().In(time.UTC), time.UTC).AddDate(0, 0, -14)
	endDayExclusive := startDay.AddDate(0, 0, 5)

	submitOnboardingStep1(t, app, authCookie, url.Values{
		"last_period_start": {startDay.Format("2006-01-02")},
		"period_end":        {endDayExclusive.Format("2006-01-02")},
	})

	submitOnboardingStep2(t, app, authCookie, url.Values{
		"cycle_length":      {"28"},
		"period_length":     {"8"},
		"auto_period_fill":  {"true"},
		"last_period_start": {startDay.Format("2006-01-02")},
		"period_end":        {endDayExclusive.Format("2006-01-02")},
	})

	submitOnboardingComplete(t, app, authCookie)

	for day := startDay; day.Before(endDayExclusive); day = day.AddDate(0, 0, 1) {
		assertOnboardingPeriodLogForDay(t, database, user.ID, day)
	}

	assertNoOnboardingLogForDay(t, database, user.ID, endDayExclusive)
}
