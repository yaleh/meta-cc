package filter

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/yale/meta-cc/internal/parser"
)

// TimeFilter represents time-based filtering options
type TimeFilter struct {
	Since      string // "5 minutes ago", "1 hour ago"
	LastNTurns int    // Last N entries
	FromTs     int64  // Unix timestamp (start)
	ToTs       int64  // Unix timestamp (end)
}

// Apply applies the time filter to session entries
func (f *TimeFilter) Apply(entries []parser.SessionEntry) ([]parser.SessionEntry, error) {
	// No filter, return all
	if f.Since == "" && f.LastNTurns == 0 && f.FromTs == 0 && f.ToTs == 0 {
		return entries, nil
	}

	// Handle LastNTurns first (simple slice operation)
	if f.LastNTurns > 0 {
		if f.LastNTurns >= len(entries) {
			return entries, nil
		}
		return entries[len(entries)-f.LastNTurns:], nil
	}

	// Parse Since duration
	var cutoffTime time.Time
	if f.Since != "" {
		duration, err := ParseDuration(f.Since)
		if err != nil {
			return nil, fmt.Errorf("invalid --since format: %w", err)
		}
		cutoffTime = time.Now().Add(-duration)
	}

	// Filter by timestamp
	var result []parser.SessionEntry
	for _, entry := range entries {
		entryTime, err := time.Parse(time.RFC3339Nano, entry.Timestamp)
		if err != nil {
			// Skip entries with invalid timestamps
			continue
		}

		// Check Since filter
		if f.Since != "" && entryTime.Before(cutoffTime) {
			continue
		}

		// Check FromTs filter
		if f.FromTs > 0 && entryTime.Unix() < f.FromTs {
			continue
		}

		// Check ToTs filter
		if f.ToTs > 0 && entryTime.Unix() > f.ToTs {
			continue
		}

		result = append(result, entry)
	}

	return result, nil
}

// ParseDuration parses human-readable duration strings like "5 minutes ago"
func ParseDuration(s string) (time.Duration, error) {
	// Expected format: "<number> <unit> ago"
	// Examples: "5 minutes ago", "1 hour ago", "30 seconds ago"

	s = strings.ToLower(strings.TrimSpace(s))

	if !strings.HasSuffix(s, " ago") {
		return 0, fmt.Errorf("duration must end with ' ago'")
	}

	// Remove " ago" suffix
	s = strings.TrimSuffix(s, " ago")
	s = strings.TrimSpace(s)

	// Split into number and unit
	parts := strings.Fields(s)
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid duration format (expected '<number> <unit> ago')")
	}

	// Parse number
	num, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid number: %w", err)
	}

	// Parse unit
	unit := parts[1]

	// Handle plural forms
	if strings.HasSuffix(unit, "s") && len(unit) > 1 {
		unit = unit[:len(unit)-1]
	}

	var duration time.Duration
	switch unit {
	case "second":
		duration = time.Duration(num) * time.Second
	case "minute":
		duration = time.Duration(num) * time.Minute
	case "hour":
		duration = time.Duration(num) * time.Hour
	case "day":
		duration = time.Duration(num) * 24 * time.Hour
	default:
		return 0, fmt.Errorf("unknown time unit: %s", unit)
	}

	return duration, nil
}
