package http

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kore/kore/pkg/kernel"
)

func TestParsePeriodQuery_Window60(t *testing.T) {
	req := httptest.NewRequest("GET", "/planning?window=60", nil)
	period, err := parsePeriodQuery(req)
	if err != nil {
		t.Fatalf("parsePeriodQuery: %v", err)
	}
	now := time.Now().UTC()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 59)
	if !period.Start.Equal(start) || !period.End.Equal(end) {
		t.Fatalf("unexpected 60d window: %+v", period)
	}
}

func TestParsePeriodQuery_ExplicitRange(t *testing.T) {
	req := httptest.NewRequest("GET", "/planning?start=2026-07-01&end=2026-07-31", nil)
	period, err := parsePeriodQuery(req)
	if err != nil {
		t.Fatalf("parsePeriodQuery: %v", err)
	}
	expectedStart := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	expectedEnd := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)
	if !period.Start.Equal(expectedStart) || !period.End.Equal(expectedEnd) {
		t.Fatalf("unexpected range: %+v", period)
	}
}

func TestParsePeriodQuery_DefaultMonth(t *testing.T) {
	req := httptest.NewRequest("GET", "/planning", nil)
	period, err := parsePeriodQuery(req)
	if err != nil {
		t.Fatalf("parsePeriodQuery: %v", err)
	}
	now := time.Now().UTC()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, -1)
	if !period.Start.Equal(start) || !period.End.Equal(end) {
		t.Fatalf("unexpected default month: %+v", period)
	}
}

var _ kernel.Period
