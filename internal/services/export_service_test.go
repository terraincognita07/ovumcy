package services

import (
	"errors"
	"testing"
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
)

type stubExportDayReader struct {
	logs []models.DailyLog
	err  error
}

func (stub *stubExportDayReader) FetchLogsForOptionalRange(uint, *time.Time, *time.Time, *time.Location) ([]models.DailyLog, error) {
	if stub.err != nil {
		return nil, stub.err
	}
	result := make([]models.DailyLog, len(stub.logs))
	copy(result, stub.logs)
	return result, nil
}

type stubExportSymptomReader struct {
	symptoms []models.SymptomType
	err      error
}

func (stub *stubExportSymptomReader) FetchSymptoms(uint) ([]models.SymptomType, error) {
	if stub.err != nil {
		return nil, stub.err
	}
	result := make([]models.SymptomType, len(stub.symptoms))
	copy(result, stub.symptoms)
	return result, nil
}

func TestExportBuildSummaryUsesDateBounds(t *testing.T) {
	service := NewExportService(
		&stubExportDayReader{
			logs: []models.DailyLog{
				{Date: mustParseExportDay(t, "2026-02-20")},
				{Date: mustParseExportDay(t, "2026-02-07")},
				{Date: mustParseExportDay(t, "2026-02-12")},
			},
		},
		&stubExportSymptomReader{},
	)

	summary, err := service.BuildSummary(42, nil, nil, time.UTC)
	if err != nil {
		t.Fatalf("BuildSummary() unexpected error: %v", err)
	}
	if !summary.HasData {
		t.Fatalf("expected summary.HasData=true")
	}
	if summary.TotalEntries != 3 {
		t.Fatalf("expected TotalEntries=3, got %d", summary.TotalEntries)
	}
	if summary.DateFrom != "2026-02-07" {
		t.Fatalf("expected DateFrom=2026-02-07, got %q", summary.DateFrom)
	}
	if summary.DateTo != "2026-02-20" {
		t.Fatalf("expected DateTo=2026-02-20, got %q", summary.DateTo)
	}
}

func TestExportBuildSummaryReturnsEmptyForNoLogs(t *testing.T) {
	service := NewExportService(&stubExportDayReader{logs: []models.DailyLog{}}, &stubExportSymptomReader{})
	summary, err := service.BuildSummary(42, nil, nil, time.UTC)
	if err != nil {
		t.Fatalf("BuildSummary() unexpected error: %v", err)
	}
	if summary.HasData {
		t.Fatalf("expected summary.HasData=false")
	}
	if summary.TotalEntries != 0 {
		t.Fatalf("expected TotalEntries=0, got %d", summary.TotalEntries)
	}
	if summary.DateFrom != "" || summary.DateTo != "" {
		t.Fatalf("expected empty date range, got %q..%q", summary.DateFrom, summary.DateTo)
	}
}

func TestExportBuildJSONEntriesNormalizesFlowAndMapsSymptoms(t *testing.T) {
	service := NewExportService(
		&stubExportDayReader{
			logs: []models.DailyLog{
				{
					Date:       mustParseExportDay(t, "2026-02-19"),
					Flow:       "unexpected-flow",
					SymptomIDs: []uint{1, 2, 3, 3},
					Notes:      "json-note",
				},
			},
		},
		&stubExportSymptomReader{
			symptoms: []models.SymptomType{
				{ID: 1, Name: "Mood swings"},
				{ID: 2, Name: "My Custom"},
				{ID: 3, Name: "Another Custom"},
			},
		},
	)

	entries, err := service.BuildJSONEntries(42, nil, nil, time.UTC)
	if err != nil {
		t.Fatalf("BuildJSONEntries() unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected one entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.Date != "2026-02-19" {
		t.Fatalf("expected Date=2026-02-19, got %q", entry.Date)
	}
	if entry.Flow != models.FlowNone {
		t.Fatalf("expected normalized flow=%q, got %q", models.FlowNone, entry.Flow)
	}
	if !entry.Symptoms.Mood {
		t.Fatalf("expected mood flag=true")
	}
	if len(entry.OtherSymptoms) != 2 || entry.OtherSymptoms[0] != "Another Custom" || entry.OtherSymptoms[1] != "My Custom" {
		t.Fatalf("expected sorted deduped other symptoms, got %#v", entry.OtherSymptoms)
	}
	if entry.Notes != "json-note" {
		t.Fatalf("expected notes preserved, got %q", entry.Notes)
	}
}

func TestExportBuildCSVRowsBuildsExpectedColumns(t *testing.T) {
	service := NewExportService(
		&stubExportDayReader{
			logs: []models.DailyLog{
				{
					Date:       mustParseExportDay(t, "2026-02-18"),
					IsPeriod:   true,
					Flow:       models.FlowLight,
					SymptomIDs: []uint{1, 2},
					Notes:      "note",
				},
			},
		},
		&stubExportSymptomReader{
			symptoms: []models.SymptomType{
				{ID: 1, Name: "Cramps"},
				{ID: 2, Name: "Custom Symptom"},
			},
		},
	)

	rows, err := service.BuildCSVRows(42, nil, nil, time.UTC)
	if err != nil {
		t.Fatalf("BuildCSVRows() unexpected error: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected one row, got %d", len(rows))
	}

	columns := rows[0].Columns()
	if len(columns) != len(ExportCSVHeaders) {
		t.Fatalf("expected %d csv columns, got %d", len(ExportCSVHeaders), len(columns))
	}
	if columns[0] != "2026-02-18" || columns[1] != "Yes" || columns[2] != "Light" {
		t.Fatalf("unexpected fixed csv columns: %#v", columns[:3])
	}
	if columns[3] != "Yes" {
		t.Fatalf("expected cramps column Yes, got %q", columns[3])
	}
	if columns[18] != "Custom Symptom" {
		t.Fatalf("expected other symptom column, got %q", columns[18])
	}
	if columns[19] != "note" {
		t.Fatalf("expected notes column, got %q", columns[19])
	}
}

func TestExportServicePropagatesDependencyErrors(t *testing.T) {
	dayErrService := NewExportService(
		&stubExportDayReader{err: errors.New("load failed")},
		&stubExportSymptomReader{},
	)
	if _, err := dayErrService.BuildSummary(1, nil, nil, time.UTC); err == nil {
		t.Fatalf("expected summary error when day reader fails")
	}

	symptomErrService := NewExportService(
		&stubExportDayReader{logs: []models.DailyLog{{Date: mustParseExportDay(t, "2026-02-18")}}},
		&stubExportSymptomReader{err: errors.New("symptom load failed")},
	)
	if _, err := symptomErrService.BuildJSONEntries(1, nil, nil, time.UTC); err == nil {
		t.Fatalf("expected json entries error when symptom reader fails")
	}
}

func mustParseExportDay(t *testing.T, raw string) time.Time {
	t.Helper()
	parsed, err := time.ParseInLocation("2006-01-02", raw, time.UTC)
	if err != nil {
		t.Fatalf("parse day %q: %v", raw, err)
	}
	return parsed
}
