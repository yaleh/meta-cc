package filter

import (
	"fmt"
	"testing"
	"time"

	"github.com/yale/meta-cc/internal/parser"
)

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Duration
		wantErr bool
	}{
		{
			name:    "5 minutes ago",
			input:   "5 minutes ago",
			want:    5 * time.Minute,
			wantErr: false,
		},
		{
			name:    "1 hour ago",
			input:   "1 hour ago",
			want:    1 * time.Hour,
			wantErr: false,
		},
		{
			name:    "30 seconds ago",
			input:   "30 seconds ago",
			want:    30 * time.Second,
			wantErr: false,
		},
		{
			name:    "2 hours ago",
			input:   "2 hours ago",
			want:    2 * time.Hour,
			wantErr: false,
		},
		{
			name:    "invalid format",
			input:   "invalid",
			wantErr: true,
		},
		{
			name:    "missing ago",
			input:   "5 minutes",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDuration(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDuration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTimeFilter_Apply(t *testing.T) {
	now := time.Now()
	entries := []parser.SessionEntry{
		{
			UUID:      "uuid1",
			Timestamp: now.Add(-10 * time.Minute).Format(time.RFC3339Nano),
		},
		{
			UUID:      "uuid2",
			Timestamp: now.Add(-5 * time.Minute).Format(time.RFC3339Nano),
		},
		{
			UUID:      "uuid3",
			Timestamp: now.Add(-2 * time.Minute).Format(time.RFC3339Nano),
		},
		{
			UUID:      "uuid4",
			Timestamp: now.Add(-1 * time.Minute).Format(time.RFC3339Nano),
		},
	}

	tests := []struct {
		name      string
		filter    TimeFilter
		wantCount int
		wantFirst string
		wantLast  string
	}{
		{
			name: "since 3 minutes ago",
			filter: TimeFilter{
				Since: "3 minutes ago",
			},
			wantCount: 2,
			wantFirst: "uuid3",
			wantLast:  "uuid4",
		},
		{
			name: "last 2 entries",
			filter: TimeFilter{
				LastNTurns: 2,
			},
			wantCount: 2,
			wantFirst: "uuid3",
			wantLast:  "uuid4",
		},
		{
			name: "from/to timestamps",
			filter: TimeFilter{
				FromTs: now.Add(-6 * time.Minute).Unix(),
				ToTs:   now.Add(-1 * time.Minute).Unix(),
			},
			wantCount: 3,
			wantFirst: "uuid2",
			wantLast:  "uuid4",
		},
		{
			name:      "no filter",
			filter:    TimeFilter{},
			wantCount: 4,
			wantFirst: "uuid1",
			wantLast:  "uuid4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.filter.Apply(entries)
			if err != nil {
				t.Fatalf("TimeFilter.Apply() error = %v", err)
			}
			if len(got) != tt.wantCount {
				t.Errorf("TimeFilter.Apply() returned %d entries, want %d", len(got), tt.wantCount)
			}
			if tt.wantCount > 0 {
				if got[0].UUID != tt.wantFirst {
					t.Errorf("First entry UUID = %s, want %s", got[0].UUID, tt.wantFirst)
				}
				if got[len(got)-1].UUID != tt.wantLast {
					t.Errorf("Last entry UUID = %s, want %s", got[len(got)-1].UUID, tt.wantLast)
				}
			}
		})
	}
}

func TestTimeFilter_LastNTurns(t *testing.T) {
	entries := make([]parser.SessionEntry, 10)
	for i := 0; i < 10; i++ {
		entries[i] = parser.SessionEntry{
			UUID:      fmt.Sprintf("uuid-%d", i),
			Timestamp: time.Now().Add(time.Duration(i) * time.Minute).Format(time.RFC3339Nano),
		}
	}

	filter := TimeFilter{LastNTurns: 3}
	got, err := filter.Apply(entries)
	if err != nil {
		t.Fatalf("TimeFilter.Apply() error = %v", err)
	}

	if len(got) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(got))
	}

	// Should get the last 3 entries (indices 7, 8, 9)
	expected := []string{"uuid-7", "uuid-8", "uuid-9"}
	for i, entry := range got {
		if entry.UUID != expected[i] {
			t.Errorf("Entry %d: UUID = %s, want %s", i, entry.UUID, expected[i])
		}
	}
}
