package main

import (
	"fmt"
	"time"
)

const (
	grafanaDashboardBaseURL = "http://localhost:3001/d/ddnw277huiv40ae/ftdc-dashboard"
	grafanaDateFormat       = "2006-01-02T15:04:05.000Z"
	defaultDashboardWindow  = 48 * time.Hour
)

// dashboardRangeFromBounds returns Grafana from/to timestamps covering at most
// window ending at latest. If the ingested span is shorter than window, from
// is clamped to earliest so the URL never starts before available data.
func dashboardRangeFromBounds(earliest, latest time.Time, window time.Duration) (from, to time.Time) {
	if latest.Before(earliest) {
		earliest, latest = latest, earliest
	}
	to = latest
	from = latest.Add(-window)
	if from.Before(earliest) {
		from = earliest
	}
	return from, to
}

func formatGrafanaDashboardURL(earliest, latest time.Time) string {
	from, to := dashboardRangeFromBounds(earliest, latest, defaultDashboardWindow)
	return fmt.Sprintf("%s?from=%s&to=%s&timezone=UTC",
		grafanaDashboardBaseURL,
		from.UTC().Format(grafanaDateFormat),
		to.UTC().Format(grafanaDateFormat),
	)
}

func parseGrafanaTimestamp(s string) (time.Time, error) {
	t, err := time.Parse(grafanaDateFormat, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse grafana timestamp %q: %w", s, err)
	}
	return t, nil
}
