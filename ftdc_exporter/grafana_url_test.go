package main

import (
	"strings"
	"testing"
	"time"
)

func TestDashboardRangeFromBounds_clampsToEarliest(t *testing.T) {
	earliest := time.Date(2026, 7, 16, 18, 0, 0, 0, time.UTC)
	latest := time.Date(2026, 7, 16, 19, 41, 4, 0, time.UTC)

	from, to := dashboardRangeFromBounds(earliest, latest, 48*time.Hour)
	if !from.Equal(earliest) {
		t.Fatalf("from = %v, want earliest %v", from, earliest)
	}
	if !to.Equal(latest) {
		t.Fatalf("to = %v, want latest %v", to, latest)
	}
}

func TestDashboardRangeFromBounds_last48h(t *testing.T) {
	earliest := time.Date(2026, 6, 10, 15, 11, 5, 0, time.UTC)
	latest := time.Date(2026, 7, 16, 19, 41, 4, 0, time.UTC)

	from, to := dashboardRangeFromBounds(earliest, latest, 48*time.Hour)
	wantFrom := latest.Add(-48 * time.Hour)
	if !from.Equal(wantFrom) {
		t.Fatalf("from = %v, want %v", from, wantFrom)
	}
	if !to.Equal(latest) {
		t.Fatalf("to = %v, want %v", to, latest)
	}
}

func TestFormatGrafanaDashboardURL(t *testing.T) {
	earliest := time.Date(2026, 6, 10, 15, 11, 5, 0, time.UTC)
	latest := time.Date(2026, 7, 16, 19, 41, 4, 0, time.UTC)

	url := formatGrafanaDashboardURL(earliest, latest)
	if !strings.Contains(url, "from=2026-07-14T19:41:04.000Z") {
		t.Fatalf("url missing expected from: %s", url)
	}
	if !strings.Contains(url, "to=2026-07-16T19:41:04.000Z") {
		t.Fatalf("url missing expected to: %s", url)
	}
	if !strings.Contains(url, "timezone=UTC") {
		t.Fatalf("url missing timezone: %s", url)
	}
	if strings.Contains(url, "from=2026-06-10") {
		t.Fatalf("url should not use full earliest span: %s", url)
	}
}
