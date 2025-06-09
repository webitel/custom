package data

import (
	"fmt"
	"testing"
	"time"
)

func TestCastNumberAsDateTime(t *testing.T) {
	type expected struct {
		precise time.Duration
		time    time.Time
	}

	tests := []struct {
		name      string
		timestamp []any
		expected  []expected
	}{
		{
			name:      "nano", // 1
			timestamp: []any{int64(1731951463_527302496), 1731951463.527302496},
			expected: []expected{
				// float64() is not precise enough to represent nanoseconds.
				// {
				//	 precise: time.Nanosecond,
				//	 time:    time.Date(2024, 11, 18, 17, 37, 43, 527302496, time.UTC),
				// },
				{
					precise: time.Microsecond,
					time:    time.Date(2024, 11, 18, 17, 37, 43, 527302000, time.UTC),
				},
				{
					precise: time.Millisecond,
					time:    time.Date(2024, 11, 18, 17, 37, 43, 527000000, time.UTC),
				},
				{
					precise: time.Second,
					time:    time.Date(2024, 11, 18, 17, 37, 43, 0, time.UTC),
				},
			},
		},
		{
			name:      "micro", // 1e3
			timestamp: []any{int64(1731951463_527302), 1731951463.527302},
			expected: []expected{
				{
					precise: time.Microsecond,
					time:    time.Date(2024, 11, 18, 17, 37, 43, 527302000, time.UTC),
				},
				{
					precise: time.Millisecond,
					time:    time.Date(2024, 11, 18, 17, 37, 43, 527000000, time.UTC),
				},
				{
					precise: time.Second,
					time:    time.Date(2024, 11, 18, 17, 37, 43, 0, time.UTC),
				},
			},
		},
		{
			name:      "milli", // 1e6
			timestamp: []any{int64(1731951463_527), 1731951463.527},
			expected: []expected{
				{
					precise: time.Microsecond,
					time:    time.Date(2024, 11, 18, 17, 37, 43, 527000000, time.UTC),
				},
				{
					precise: time.Millisecond,
					time:    time.Date(2024, 11, 18, 17, 37, 43, 527000000, time.UTC),
				},
				{
					precise: time.Second,
					time:    time.Date(2024, 11, 18, 17, 37, 43, 0, time.UTC),
				},
			},
		},
		{
			name:      "sec", // 1e9
			timestamp: []any{int64(1731951463), 1731951463.0},
			expected: []expected{
				{
					precise: time.Nanosecond,
					time:    time.Date(2024, 11, 18, 17, 37, 43, 0, time.UTC),
				},
				{
					precise: time.Microsecond,
					time:    time.Date(2024, 11, 18, 17, 37, 43, 0, time.UTC),
				},
				{
					precise: time.Millisecond,
					time:    time.Date(2024, 11, 18, 17, 37, 43, 0, time.UTC),
				},
				{
					precise: time.Second,
					time:    time.Date(2024, 11, 18, 17, 37, 43, 0, time.UTC),
				},
			},
		},
	}

	for _, tt := range tests {
		for _, ts := range tt.timestamp {
			for _, want := range tt.expected {
				t.Run(fmt.Sprintf("%T %s with precise of %s", ts, tt.name, want.precise.String()), func(t *testing.T) {
					var date time.Time
					switch ts := ts.(type) {
					case int64:
						date = CastNumberAsDateTime(ts, want.precise).UTC()
					case float64:
						date = CastNumberAsDateTime(ts, want.precise).UTC()
					default:
						t.Errorf("cannot cast time %T with precise of %s", ts, want.precise.String())
					}

					if !date.Equal(want.time) {
						t.Errorf("CastNumberAsDateTime() = %v, want %v", date, want.time)
					}
				})
			}
		}
	}
}

func TestCastDateTimeAsNumber(t *testing.T) {
	tests := []struct {
		name     string
		time     time.Time
		precise  time.Duration
		expected int64
	}{
		{
			name:     "nano",
			time:     time.Date(2024, 11, 18, 17, 37, 43, 527302496, time.UTC),
			precise:  time.Nanosecond,
			expected: 1731951463527302496,
		},
		{
			name:     "micro",
			time:     time.Date(2024, 11, 18, 17, 37, 43, 527302496, time.UTC),
			precise:  time.Microsecond,
			expected: 1731951463527302,
		},
		{
			name:     "milli",
			time:     time.Date(2024, 11, 18, 17, 37, 43, 527302496, time.UTC),
			precise:  time.Millisecond,
			expected: 1731951463527,
		},
		{
			name:     "milli without nsec",
			time:     time.Date(2024, 11, 18, 17, 37, 43, 0, time.UTC),
			precise:  time.Millisecond,
			expected: 1731951463000,
		},
		{
			name:     "sec",
			time:     time.Date(2024, 11, 18, 17, 37, 43, 527302496, time.UTC),
			precise:  time.Second,
			expected: 1731951463,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CastDateTimeAsNumber(tt.time, tt.precise)
			if result != tt.expected {
				t.Errorf("CastDateTimeAsNumber() = %v, want %v", result, tt.expected)
			}

			// Verify roundtrip conversion
			reconverted := CastNumberAsDateTime(result, tt.precise)
			if !reconverted.Equal(tt.time.Truncate(tt.precise)) {
				t.Errorf("Roundtrip conversion failed: got %v, want %v",
					reconverted, tt.time.Truncate(tt.precise))
			}
		})
	}
}
