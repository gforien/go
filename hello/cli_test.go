package main

import (
	"os"
	"testing"
	"time"
)

// Mock Footer() for testing
func mockFooter(_ time.Time) string { return "v1.2.3\n" }

func TestCliHandler(t *testing.T) {
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }() // restore after test

	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "no args",
			args: []string{"app"},
			want: "hello!\n" +
				"    $0: app\n" +
				"v1.2.3\n",
		},
		{
			name: "two args",
			args: []string{"app", "foo", "bar"},
			want: "hello!\n" +
				"    $0: app\n" +
				"    $1: foo\n" +
				"    $2: bar\n" +
				"v1.2.3\n",
		},
	}

	// Override Date() and Version() for deterministic output
	Footer = mockFooter

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args
			got := cliHandler()
			if got != tt.want {
				t.Errorf("cliHandler() = %q, want %q", got, tt.want)
			}
		})
	}
}
