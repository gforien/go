package semver

import (
	"errors"
	"strings"
	"testing"
)

func TestParseBump(t *testing.T) {
	cases := []struct {
		name        string
		given       string
		expected    Bump
		expectError bool
	}{
		{
			name:        "major",
			given:       "major",
			expected:    Major,
			expectError: false,
		},
		{
			name:        "minor",
			given:       "minor",
			expected:    Minor,
			expectError: false,
		},

		{
			name:        "patch",
			given:       "patch",
			expected:    Patch,
			expectError: false,
		},

		{
			name:        "invalid",
			given:       "invalid",
			expected:    "",
			expectError: true,
		},

		{
			name:        "empty",
			given:       "",
			expected:    "",
			expectError: true,
		},

		{
			name:        "MAJOR",
			given:       "MAJOR",
			expected:    "",
			expectError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseBump(tc.given)

			if tc.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				} else {
					var invalidErr *InvalidBumpError
					if !errors.As(err, &invalidErr) {
						t.Errorf("expected InvalidBumpError, got %T", err)
					}
					if invalidErr.Text != tc.given {
						t.Errorf("expected error text %q, got %q", tc.given, invalidErr.Text)
					}
					if !strings.Contains(err.Error(), tc.given) {
						t.Errorf("error string %q does not contain input %q", err.Error(), tc.given)
					}
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if got != tc.expected {
					t.Errorf("expected %q, got %q", tc.expected, got)
				}
			}
		})
	}
}
