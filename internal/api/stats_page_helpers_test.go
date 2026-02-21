package api

import (
	"reflect"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/lume/internal/models"
)

func TestTrimTrailingCycleTrendLengths(t *testing.T) {
	t.Parallel()

	source := []int{1, 2, 3, 4, 5}
	if got := trimTrailingCycleTrendLengths(source, 10); !reflect.DeepEqual(got, source) {
		t.Fatalf("expected unchanged lengths, got %#v", got)
	}

	expected := []int{3, 4, 5}
	if got := trimTrailingCycleTrendLengths(source, 3); !reflect.DeepEqual(got, expected) {
		t.Fatalf("expected last 3 lengths %#v, got %#v", expected, got)
	}
}

func TestBuildStatsChartDataIncludesBaselineWhenPositive(t *testing.T) {
	t.Parallel()

	messages := map[string]string{"stats.cycle_label": "Cycle %d"}
	lengths := []int{28, 30}

	withBaseline := buildStatsChartData(messages, lengths, 29)
	if _, ok := withBaseline["baseline"]; !ok {
		t.Fatal("expected baseline in chart payload when baseline > 0")
	}
	if got, ok := withBaseline["labels"].([]string); !ok || len(got) != 2 {
		t.Fatalf("expected 2 labels, got %#v", withBaseline["labels"])
	}

	withoutBaseline := buildStatsChartData(messages, lengths, 0)
	if _, ok := withoutBaseline["baseline"]; ok {
		t.Fatal("did not expect baseline in chart payload when baseline == 0")
	}
}

func TestBuildStatsTrendViewTrimsPointsAndAppliesOwnerBaseline(t *testing.T) {
	t.Parallel()

	handler := &Handler{location: time.UTC}
	now := time.Date(2026, time.February, 21, 12, 0, 0, 0, time.UTC)
	today := dateAtLocation(now, time.UTC)

	start := today.AddDate(0, 0, -(14*28 + 2))
	logs := make([]models.DailyLog, 0, 15)
	for index := 0; index < 15; index++ {
		logs = append(logs, models.DailyLog{
			ID:       uint(index + 1),
			Date:     start.AddDate(0, 0, index*28),
			IsPeriod: true,
			Flow:     models.FlowMedium,
		})
	}

	user := &models.User{Role: models.RoleOwner, CycleLength: 28}
	chartPayload, baseline, trendCount := handler.buildStatsTrendView(user, logs, now, map[string]string{"stats.cycle_label": "Cycle %d"})

	if baseline != 28 {
		t.Fatalf("expected baseline 28, got %d", baseline)
	}
	if trendCount != maxStatsTrendPoints {
		t.Fatalf("expected trimmed trend point count %d, got %d", maxStatsTrendPoints, trendCount)
	}

	labels, labelsOK := chartPayload["labels"].([]string)
	values, valuesOK := chartPayload["values"].([]int)
	if !labelsOK || !valuesOK {
		t.Fatalf("expected labels/values slices in chart payload, got labels=%T values=%T", chartPayload["labels"], chartPayload["values"])
	}
	if len(labels) != maxStatsTrendPoints || len(values) != maxStatsTrendPoints {
		t.Fatalf("expected %d labels/values, got %d/%d", maxStatsTrendPoints, len(labels), len(values))
	}
	if _, ok := chartPayload["baseline"]; !ok {
		t.Fatal("expected baseline field in owner chart payload")
	}
}

func TestBuildStatsTrendViewPartnerHasNoBaseline(t *testing.T) {
	t.Parallel()

	handler := &Handler{location: time.UTC}
	now := time.Date(2026, time.February, 21, 12, 0, 0, 0, time.UTC)
	user := &models.User{Role: models.RolePartner, CycleLength: 28}

	chartPayload, baseline, trendCount := handler.buildStatsTrendView(user, []models.DailyLog{}, now, map[string]string{})
	if baseline != 0 {
		t.Fatalf("expected partner baseline 0, got %d", baseline)
	}
	if trendCount != 0 {
		t.Fatalf("expected zero trend count for empty logs, got %d", trendCount)
	}
	if _, ok := chartPayload["baseline"]; ok {
		t.Fatal("did not expect baseline field for partner chart payload")
	}
	if _, ok := chartPayload["values"].([]int); !ok {
		t.Fatalf("expected values slice in chart payload, got %T", chartPayload["values"])
	}
}

func TestBuildStatsChartDataTypeCompatibility(t *testing.T) {
	t.Parallel()

	payload := buildStatsChartData(map[string]string{}, []int{}, 0)
	if _, ok := any(payload).(fiber.Map); !ok {
		t.Fatalf("expected fiber.Map payload type, got %T", payload)
	}
}

func TestBuildStatsSymptomCountsNonOwnerSkipsDataAccess(t *testing.T) {
	t.Parallel()

	handler := &Handler{}
	user := &models.User{Role: models.RolePartner}

	counts, errorMessage, err := handler.buildStatsSymptomCounts(user, "en")
	if err != nil {
		t.Fatalf("buildStatsSymptomCounts returned error: %v", err)
	}
	if errorMessage != "" {
		t.Fatalf("expected empty error message, got %q", errorMessage)
	}
	if len(counts) != 0 {
		t.Fatalf("expected empty counts for non-owner, got %#v", counts)
	}
}

func TestBuildStatsSymptomCountsOwnerReturnsLocalizedCounts(t *testing.T) {
	t.Parallel()

	handler, database := newDataAccessTestHandler(t)
	user := createDataAccessTestUser(t, database, "stats-symptom-owner@example.com")

	symptom := models.SymptomType{
		UserID: user.ID,
		Name:   "Headache",
		Icon:   "🤕",
		Color:  "#CC8844",
	}
	if err := database.Create(&symptom).Error; err != nil {
		t.Fatalf("create symptom: %v", err)
	}

	logEntry := models.DailyLog{
		UserID:     user.ID,
		Date:       time.Date(2026, time.February, 21, 0, 0, 0, 0, time.UTC),
		IsPeriod:   false,
		Flow:       models.FlowNone,
		SymptomIDs: []uint{symptom.ID},
	}
	if err := database.Create(&logEntry).Error; err != nil {
		t.Fatalf("create daily log: %v", err)
	}

	counts, errorMessage, err := handler.buildStatsSymptomCounts(&user, "en")
	if err != nil {
		t.Fatalf("buildStatsSymptomCounts returned error: %v", err)
	}
	if errorMessage != "" {
		t.Fatalf("expected empty error message, got %q", errorMessage)
	}
	if len(counts) != 1 {
		t.Fatalf("expected one symptom count entry, got %d", len(counts))
	}
	if counts[0].Count != 1 || counts[0].TotalDays != 1 {
		t.Fatalf("unexpected count payload: %#v", counts[0])
	}
	if counts[0].FrequencySummary == "" {
		t.Fatalf("expected localized frequency summary, got empty value")
	}
}

func TestBuildStatsPageDataOwnerBaselineAndFlags(t *testing.T) {
	t.Parallel()

	handler, database := newDataAccessTestHandler(t)
	user := createDataAccessTestUser(t, database, "stats-page-owner@example.com")
	user.Role = models.RoleOwner

	now := time.Date(2026, time.February, 21, 0, 0, 0, 0, time.UTC)
	messages := map[string]string{"meta.title.stats": "Stats"}

	data, errorMessage, err := handler.buildStatsPageData(&user, "en", messages, now)
	if err != nil {
		t.Fatalf("buildStatsPageData returned error: %v", err)
	}
	if errorMessage != "" {
		t.Fatalf("expected empty error message, got %q", errorMessage)
	}
	if baseline, ok := data["ChartBaseline"].(int); !ok || baseline != 28 {
		t.Fatalf("expected ChartBaseline=28, got %#v", data["ChartBaseline"])
	}
	if isOwner, ok := data["IsOwner"].(bool); !ok || !isOwner {
		t.Fatalf("expected IsOwner=true, got %#v", data["IsOwner"])
	}
	if _, ok := data["ChartData"].(fiber.Map); !ok {
		t.Fatalf("expected ChartData fiber.Map, got %T", data["ChartData"])
	}
}

func TestBuildStatsPageDataPartnerNoBaseline(t *testing.T) {
	t.Parallel()

	handler, database := newDataAccessTestHandler(t)
	partner := models.User{
		Email:               "stats-page-partner@example.com",
		PasswordHash:        "test-hash",
		Role:                models.RolePartner,
		OnboardingCompleted: true,
		CycleLength:         28,
		PeriodLength:        5,
		CreatedAt:           time.Now().UTC(),
	}
	if err := database.Create(&partner).Error; err != nil {
		t.Fatalf("create partner user: %v", err)
	}

	now := time.Date(2026, time.February, 21, 0, 0, 0, 0, time.UTC)
	data, errorMessage, err := handler.buildStatsPageData(&partner, "en", map[string]string{}, now)
	if err != nil {
		t.Fatalf("buildStatsPageData returned error: %v", err)
	}
	if errorMessage != "" {
		t.Fatalf("expected empty error message, got %q", errorMessage)
	}
	if baseline, ok := data["ChartBaseline"].(int); !ok || baseline != 0 {
		t.Fatalf("expected ChartBaseline=0 for partner, got %#v", data["ChartBaseline"])
	}
	if isOwner, ok := data["IsOwner"].(bool); !ok || isOwner {
		t.Fatalf("expected IsOwner=false for partner, got %#v", data["IsOwner"])
	}
}
