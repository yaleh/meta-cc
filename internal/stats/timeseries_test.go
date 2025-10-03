package stats

import (
	"testing"
	"time"

	"github.com/yale/meta-cc/internal/parser"
)

func TestAnalyzeTimeSeries_ToolCallsMetric(t *testing.T) {
	// Test data: 4 tool calls across 3 hours
	baseTime := time.Date(2025, 10, 3, 10, 0, 0, 0, time.UTC)
	tools := []parser.ToolCall{
		{ToolName: "Bash", Status: "success", Timestamp: baseTime.Format(time.RFC3339)},
		{ToolName: "Read", Status: "success", Timestamp: baseTime.Add(30 * time.Minute).Format(time.RFC3339)},
		{ToolName: "Edit", Status: "error", Timestamp: baseTime.Add(90 * time.Minute).Format(time.RFC3339)},
		{ToolName: "Bash", Status: "success", Timestamp: baseTime.Add(150 * time.Minute).Format(time.RFC3339)},
	}

	config := TimeSeriesConfig{
		Metric:   "tool-calls",
		Interval: "hour",
	}

	points, err := AnalyzeTimeSeries(tools, config)
	if err != nil {
		t.Fatalf("AnalyzeTimeSeries failed: %v", err)
	}

	// Should have 3 time buckets (hour 0, hour 1, hour 2)
	if len(points) != 3 {
		t.Errorf("expected 3 time buckets, got %d", len(points))
	}

	// Hour 0: 2 tool calls (at 0 min and 30 min)
	if points[0].Value != 2.0 {
		t.Errorf("hour 0: expected 2 tool calls, got %.0f", points[0].Value)
	}

	// Hour 1: 1 tool call (at 90 min)
	if points[1].Value != 1.0 {
		t.Errorf("hour 1: expected 1 tool call, got %.0f", points[1].Value)
	}

	// Hour 2: 1 tool call (at 150 min)
	if points[2].Value != 1.0 {
		t.Errorf("hour 2: expected 1 tool call, got %.0f", points[2].Value)
	}
}

func TestAnalyzeTimeSeries_ErrorRateMetric(t *testing.T) {
	baseTime := time.Date(2025, 10, 3, 8, 0, 0, 0, time.UTC)
	tools := []parser.ToolCall{
		{Status: "success", Timestamp: baseTime.Format(time.RFC3339)},
		{Status: "success", Timestamp: baseTime.Add(10 * time.Minute).Format(time.RFC3339)},
		{Status: "error", Timestamp: baseTime.Add(20 * time.Minute).Format(time.RFC3339)},
		{Status: "success", Timestamp: baseTime.Add(70 * time.Minute).Format(time.RFC3339)},
	}

	config := TimeSeriesConfig{
		Metric:   "error-rate",
		Interval: "hour",
	}

	points, err := AnalyzeTimeSeries(tools, config)
	if err != nil {
		t.Fatalf("AnalyzeTimeSeries failed: %v", err)
	}

	// Hour 0: 3 calls, 1 error = 33.33% error rate
	expectedRate := 1.0 / 3.0
	if points[0].Value < expectedRate-0.01 || points[0].Value > expectedRate+0.01 {
		t.Errorf("hour 0: expected error rate ~%.2f, got %.2f", expectedRate, points[0].Value)
	}

	// Hour 1: 1 call, 0 errors = 0% error rate
	if points[1].Value != 0.0 {
		t.Errorf("hour 1: expected 0%% error rate, got %.2f", points[1].Value)
	}
}

func TestAnalyzeTimeSeries_DayInterval(t *testing.T) {
	baseTime := time.Date(2025, 10, 1, 10, 0, 0, 0, time.UTC)
	tools := []parser.ToolCall{
		{ToolName: "Bash", Timestamp: baseTime.Format(time.RFC3339)},
		{ToolName: "Read", Timestamp: baseTime.Add(6 * time.Hour).Format(time.RFC3339)},
		{ToolName: "Edit", Timestamp: baseTime.Add(25 * time.Hour).Format(time.RFC3339)},
		{ToolName: "Grep", Timestamp: baseTime.Add(26 * time.Hour).Format(time.RFC3339)},
	}

	config := TimeSeriesConfig{
		Metric:   "tool-calls",
		Interval: "day",
	}

	points, err := AnalyzeTimeSeries(tools, config)
	if err != nil {
		t.Fatalf("AnalyzeTimeSeries failed: %v", err)
	}

	// Should have 2 days
	if len(points) != 2 {
		t.Errorf("expected 2 day buckets, got %d", len(points))
	}

	// Day 0: 2 calls
	if points[0].Value != 2.0 {
		t.Errorf("day 0: expected 2 tool calls, got %.0f", points[0].Value)
	}

	// Day 1: 2 calls
	if points[1].Value != 2.0 {
		t.Errorf("day 1: expected 2 tool calls, got %.0f", points[1].Value)
	}
}

func TestAnalyzeTimeSeries_WeekInterval(t *testing.T) {
	// Start on Monday (2025-09-29 is a Monday)
	baseTime := time.Date(2025, 9, 29, 10, 0, 0, 0, time.UTC)
	tools := []parser.ToolCall{
		{ToolName: "Bash", Timestamp: baseTime.Format(time.RFC3339)},                        // Monday
		{ToolName: "Read", Timestamp: baseTime.Add(3 * 24 * time.Hour).Format(time.RFC3339)}, // Thursday
		{ToolName: "Edit", Timestamp: baseTime.Add(8 * 24 * time.Hour).Format(time.RFC3339)}, // Next Tuesday
	}

	config := TimeSeriesConfig{
		Metric:   "tool-calls",
		Interval: "week",
	}

	points, err := AnalyzeTimeSeries(tools, config)
	if err != nil {
		t.Fatalf("AnalyzeTimeSeries failed: %v", err)
	}

	// Should have 2 weeks
	if len(points) != 2 {
		t.Errorf("expected 2 week buckets, got %d", len(points))
	}

	// Week 0: 2 calls
	if points[0].Value != 2.0 {
		t.Errorf("week 0: expected 2 tool calls, got %.0f", points[0].Value)
	}

	// Week 1: 1 call
	if points[1].Value != 1.0 {
		t.Errorf("week 1: expected 1 tool call, got %.0f", points[1].Value)
	}
}

func TestAnalyzeTimeSeries_EmptyInput(t *testing.T) {
	config := TimeSeriesConfig{
		Metric:   "tool-calls",
		Interval: "hour",
	}

	points, err := AnalyzeTimeSeries([]parser.ToolCall{}, config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if points != nil {
		t.Errorf("expected nil for empty input, got %d points", len(points))
	}
}

func TestAnalyzeTimeSeries_SingleDataPoint(t *testing.T) {
	baseTime := time.Date(2025, 10, 3, 15, 30, 0, 0, time.UTC)
	tools := []parser.ToolCall{
		{ToolName: "Bash", Status: "success", Timestamp: baseTime.Format(time.RFC3339)},
	}

	config := TimeSeriesConfig{
		Metric:   "tool-calls",
		Interval: "hour",
	}

	points, err := AnalyzeTimeSeries(tools, config)
	if err != nil {
		t.Fatalf("AnalyzeTimeSeries failed: %v", err)
	}

	// Should have 1 bucket
	if len(points) != 1 {
		t.Errorf("expected 1 bucket, got %d", len(points))
	}

	if points[0].Value != 1.0 {
		t.Errorf("expected 1 tool call, got %.0f", points[0].Value)
	}
}

func TestAnalyzeTimeSeries_InvalidTimestamp(t *testing.T) {
	tools := []parser.ToolCall{
		{ToolName: "Bash", Timestamp: "invalid-timestamp"},
	}

	config := TimeSeriesConfig{
		Metric:   "tool-calls",
		Interval: "hour",
	}

	_, err := AnalyzeTimeSeries(tools, config)
	if err == nil {
		t.Error("expected error for invalid timestamp, got nil")
	}
}

func TestTruncateTime(t *testing.T) {
	tests := []struct {
		name     string
		time     time.Time
		interval string
		expected time.Time
	}{
		{
			name:     "hour truncation",
			time:     time.Date(2025, 10, 3, 14, 45, 30, 123456789, time.UTC),
			interval: "hour",
			expected: time.Date(2025, 10, 3, 14, 0, 0, 0, time.UTC),
		},
		{
			name:     "day truncation",
			time:     time.Date(2025, 10, 3, 14, 45, 30, 0, time.UTC),
			interval: "day",
			expected: time.Date(2025, 10, 3, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "week truncation - Monday",
			time:     time.Date(2025, 9, 29, 10, 0, 0, 0, time.UTC), // Monday
			interval: "week",
			expected: time.Date(2025, 9, 29, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "week truncation - Sunday",
			time:     time.Date(2025, 10, 5, 10, 0, 0, 0, time.UTC), // Sunday
			interval: "week",
			expected: time.Date(2025, 9, 29, 0, 0, 0, 0, time.UTC), // Previous Monday
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateTime(tt.time, tt.interval)
			if !result.Equal(tt.expected) {
				t.Errorf("truncateTime(%v, %s) = %v, want %v",
					tt.time, tt.interval, result, tt.expected)
			}
		})
	}
}

func TestCreateTimeBuckets(t *testing.T) {
	start := time.Date(2025, 10, 3, 10, 30, 0, 0, time.UTC)
	end := time.Date(2025, 10, 3, 12, 15, 0, 0, time.UTC)

	buckets := createTimeBuckets(start, end, "hour")

	// Should have 3 hour buckets: 10:00, 11:00, 12:00
	if len(buckets) != 3 {
		t.Errorf("expected 3 hour buckets, got %d", len(buckets))
	}

	expectedFirst := time.Date(2025, 10, 3, 10, 0, 0, 0, time.UTC)
	if !buckets[0].Equal(expectedFirst) {
		t.Errorf("first bucket: expected %v, got %v", expectedFirst, buckets[0])
	}

	expectedLast := time.Date(2025, 10, 3, 12, 0, 0, 0, time.UTC)
	if !buckets[2].Equal(expectedLast) {
		t.Errorf("last bucket: expected %v, got %v", expectedLast, buckets[2])
	}
}
