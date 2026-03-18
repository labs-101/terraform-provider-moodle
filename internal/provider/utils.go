package provider

import (
	"fmt"
	"time"
)

// parseDateToUnix converts a date string (YYYY-MM-DD) to a Unix timestamp.
// An empty string returns 0.
func parseDateToUnix(dateStr string) (int64, error) {
	if dateStr == "" {
		return 0, nil
	}
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return 0, fmt.Errorf("invalid date format %q, expected YYYY-MM-DD", dateStr)
	}
	return t.Unix(), nil
}

// unixToDate converts a Unix timestamp to a date string (YYYY-MM-DD).
// 0 returns an empty string.
func unixToDate(unix int64) string {
	if unix == 0 {
		return ""
	}
	return time.Unix(unix, 0).UTC().Format("2006-01-02")
}
