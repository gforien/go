package main

import (
	"testing"
	"time"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name string
		d    time.Duration
		want string
	}{
		{
			name: "nanoseconds",
			d:    500 * time.Nanosecond,
			want: "500 ns",
		},
		{
			name: "microseconds",
			d:    153*time.Microsecond + 343*time.Nanosecond,
			want: "153 µs",
		},
		{
			name: "milliseconds",
			d:    12*time.Millisecond + 345*time.Microsecond,
			want: "12 ms",
		},
		{
			name: "less than a second",
			d:    999 * time.Millisecond,
			want: "999 ms",
		},
		{
			name: "seconds",
			d:    3*time.Second + 123*time.Millisecond,
			want: "3.12 s",
		},
		{
			name: "exact second",
			d:    2 * time.Second,
			want: "2.00 s",
		},
		{
			name: "exact millisecond",
			d:    500 * time.Millisecond,
			want: "500 ms",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatDuration(tt.d)
			if got != tt.want {
				t.Errorf("FormatDuration(%v) = %q, want %q", tt.d, got, tt.want)
			}
		})
	}
}
