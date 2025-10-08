package stats

import (
	"fmt"
	"time"

	"github.com/yaleh/meta-cc/internal/parser"
)

// TimeSeriesConfig defines time series analysis parameters
type TimeSeriesConfig struct {
	Metric   string // Metric: "tool-calls", "error-rate", "avg-duration"
	Interval string // Interval: "hour", "day", "week"
}

// TimeSeriesPoint represents a single data point in time series
type TimeSeriesPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// AnalyzeTimeSeries generates time series data from tool calls
func AnalyzeTimeSeries(tools []parser.ToolCall, config TimeSeriesConfig) ([]TimeSeriesPoint, error) {
	if len(tools) == 0 {
		return nil, nil
	}

	// Parse timestamps
	var times []time.Time
	for _, tool := range tools {
		t, err := time.Parse(time.RFC3339, tool.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("invalid timestamp %s: %w", tool.Timestamp, err)
		}
		times = append(times, t)
	}

	// Find time range
	minTime := times[0]
	maxTime := times[0]
	for _, t := range times {
		if t.Before(minTime) {
			minTime = t
		}
		if t.After(maxTime) {
			maxTime = t
		}
	}

	// Create time buckets
	buckets := createTimeBuckets(minTime, maxTime, config.Interval)

	// Group tools into buckets
	bucketData := make(map[time.Time][]parser.ToolCall)
	for i, tool := range tools {
		bucket := findBucket(times[i], buckets, config.Interval)
		bucketData[bucket] = append(bucketData[bucket], tool)
	}

	// Calculate metrics for each bucket
	var points []TimeSeriesPoint
	for _, bucket := range buckets {
		data := bucketData[bucket]
		value := calculateTimeSeriesMetric(data, config.Metric)

		points = append(points, TimeSeriesPoint{
			Timestamp: bucket,
			Value:     value,
		})
	}

	return points, nil
}

// createTimeBuckets creates a list of time buckets from start to end
func createTimeBuckets(start, end time.Time, interval string) []time.Time {
	var buckets []time.Time

	current := truncateTime(start, interval)
	endTruncated := truncateTime(end, interval)

	for !current.After(endTruncated) {
		buckets = append(buckets, current)
		current = advanceTime(current, interval)
	}

	return buckets
}

// truncateTime truncates time to the start of the interval
func truncateTime(t time.Time, interval string) time.Time {
	switch interval {
	case "hour":
		return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
	case "day":
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	case "week":
		// ISO week starts on Monday
		weekday := int(t.Weekday())
		if weekday == 0 {
			weekday = 7 // Sunday
		}
		daysToSubtract := weekday - 1
		return time.Date(t.Year(), t.Month(), t.Day()-daysToSubtract, 0, 0, 0, 0, t.Location())
	default:
		return t
	}
}

// advanceTime advances time by one interval
func advanceTime(t time.Time, interval string) time.Time {
	switch interval {
	case "hour":
		return t.Add(time.Hour)
	case "day":
		return t.AddDate(0, 0, 1)
	case "week":
		return t.AddDate(0, 0, 7)
	default:
		return t
	}
}

// findBucket finds the appropriate bucket for a given time
func findBucket(t time.Time, buckets []time.Time, interval string) time.Time {
	truncated := truncateTime(t, interval)
	for _, bucket := range buckets {
		if bucket.Equal(truncated) {
			return bucket
		}
	}
	// If not found, return the closest bucket
	return buckets[len(buckets)-1]
}

// calculateTimeSeriesMetric calculates the metric value for a bucket
func calculateTimeSeriesMetric(tools []parser.ToolCall, metric string) float64 {
	if len(tools) == 0 {
		return 0.0
	}

	switch metric {
	case "tool-calls":
		return float64(len(tools))

	case "error-rate":
		errorCount := 0
		for _, tool := range tools {
			if tool.Status == "error" {
				errorCount++
			}
		}
		return float64(errorCount) / float64(len(tools))

	default:
		return 0.0
	}
}
