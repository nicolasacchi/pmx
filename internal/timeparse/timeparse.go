// Package timeparse converts human time expressions to Postmark's date filters.
// Postmark message/bounce/stat filters (fromdate/todate) are date-granular
// (YYYY-MM-DD), so sub-day inputs collapse to the calendar day in UTC.
package timeparse

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ParseDate converts a time expression to "YYYY-MM-DD" (UTC). Accepts:
//   - "" → "" (no filter)
//   - "now" / "today" / "yesterday"
//   - relative "7d", "24h", "30m", "2w" (interpreted as "ago")
//   - "now-2h"
//   - absolute "2006-01-02"
//   - RFC3339, or a Unix epoch (seconds)
func ParseDate(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", nil
	}
	t, err := parseTime(s)
	if err != nil {
		return "", err
	}
	return t.UTC().Format("2006-01-02"), nil
}

func parseTime(s string) (time.Time, error) {
	now := time.Now().UTC()
	switch strings.ToLower(s) {
	case "now", "today":
		return now, nil
	case "yesterday":
		return now.AddDate(0, 0, -1), nil
	}
	if strings.HasPrefix(s, "now-") {
		d, err := parseDuration(s[4:])
		if err != nil {
			return time.Time{}, err
		}
		return now.Add(-d), nil
	}
	if d, err := parseDuration(s); err == nil {
		return now.Add(-d), nil
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	if epoch, err := strconv.ParseInt(s, 10, 64); err == nil {
		return time.Unix(epoch, 0).UTC(), nil
	}
	return time.Time{}, fmt.Errorf("cannot parse time %q (use 7d, 24h, 2006-01-02, RFC3339, or epoch)", s)
}

func parseDuration(s string) (time.Duration, error) {
	if s == "" {
		return 0, fmt.Errorf("empty duration")
	}
	last := s[len(s)-1]
	if last == 'd' || last == 'w' {
		n, err := strconv.Atoi(s[:len(s)-1])
		if err != nil {
			return 0, fmt.Errorf("invalid duration %q", s)
		}
		if last == 'd' {
			return time.Duration(n) * 24 * time.Hour, nil
		}
		return time.Duration(n) * 7 * 24 * time.Hour, nil
	}
	// h / m / s handled by the stdlib parser.
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("invalid duration %q", s)
	}
	return d, nil
}
