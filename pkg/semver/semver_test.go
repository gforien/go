package semver

import (
	"testing"
)

func TestFromString(t *testing.T) {
	cases := []struct {
		given    string
		expected Version
	}{
		{
			given:    "1.2.3",
			expected: Version{Major: 1, Minor: 2, Patch: 3},
		},
		{
			given:    "v1.2.3",
			expected: Version{Major: 1, Minor: 2, Patch: 3},
		},
		{
			given:    "0.0.1",
			expected: Version{Major: 0, Minor: 0, Patch: 1},
		},
		{
			given:    "10.0.0",
			expected: Version{Major: 10, Minor: 0, Patch: 0},
		},
	}

	for _, tc := range cases {
		t.Run(tc.given, func(t *testing.T) {
			t.Parallel()

			res, err := FromString(tc.given)
			if err != nil {
				t.Errorf("expected no error, but got %#v", err)
			}
			if res != tc.expected {
				t.Errorf("expected %#v, but got %#v", tc.expected, res)
			}
		})
	}
}

func TestFromStringFailing(t *testing.T) {
	cases := []struct {
		given    string
		expected Version
	}{
		{given: "", expected: Version{}},
		{given: "1", expected: Version{}},
		{given: "1.2", expected: Version{}},
		{given: "a.b.c", expected: Version{}},
	}

	for _, tc := range cases {
		t.Run(tc.given, func(t *testing.T) {
			t.Parallel()

			res, err := FromString(tc.given)
			if err == nil {
				t.Errorf("expected an error %#v but got none", tc.expected)
			}
			if res != tc.expected {
				t.Errorf("expected %#v, but got %#v", tc.expected, res)
			}
		})
	}
}

func TestString(t *testing.T) {
	t.Parallel()

	v := &Version{Major: 1, Minor: 2, Patch: 3}
	expected := "1.2.3"
	if v.String() != expected {
		t.Errorf("Version.String(): expected %s, got %s", expected, v.String())
	}
}

func TestNextMajor(t *testing.T) {
	t.Parallel()

	v := &Version{Major: 1, Minor: 2, Patch: 3}
	expected := &Version{Major: 2, Minor: 0, Patch: 0}
	result := v.NextMajor()
	if !expected.Equals(result) {
		t.Errorf("Version.ReleaseMajor(): expected %v, got %v", expected, result)
	}
}

func TestNextMinor(t *testing.T) {
	t.Parallel()

	v := &Version{Major: 1, Minor: 2, Patch: 3}
	expected := &Version{Major: 1, Minor: 3, Patch: 0}
	result := v.NextMinor()
	if !expected.Equals(result) {
		t.Errorf("Version.ReleaseMinor(): expected %v, got %v", expected, result)
	}
}

func TestNextPatch(t *testing.T) {
	t.Parallel()

	v := &Version{Major: 1, Minor: 2, Patch: 3}
	expected := &Version{Major: 1, Minor: 2, Patch: 4}
	result := v.NextPatch()
	if !expected.Equals(result) {
		t.Errorf("Version.ReleasePatch(): expected %v, got %v", expected, result)
	}
}

func TestErrCannotParse(t *testing.T) {
	t.Parallel()

	input := "1.2"
	_, err := FromString(input)
	if err == nil {
		t.Fatalf("FromString(%s): expected error, got nil", input)
	}

	if _, ok := err.(*ErrCannotParse); !ok {
		t.Fatalf("FromString(%s): expected *ErrCannotParse, got %T", input, err)
	}
}

func TestVersion_Bump(t *testing.T) {
	cases := []struct {
		name          string
		version       *Version
		bump          Bump
		expected      *Version
		expectedPanic bool
	}{
		{
			name:     "patch bump",
			version:  &Version{Major: 1, Minor: 2, Patch: 3},
			bump:     Patch,
			expected: &Version{Major: 1, Minor: 2, Patch: 4},
		},
		{
			name:     "minor bump",
			version:  &Version{Major: 1, Minor: 2, Patch: 3},
			bump:     Minor,
			expected: &Version{Major: 1, Minor: 3, Patch: 0},
		},
		{
			name:     "major bump",
			version:  &Version{Major: 1, Minor: 2, Patch: 3},
			bump:     Major,
			expected: &Version{Major: 2, Minor: 0, Patch: 0},
		},
		{
			name:     "nil version",
			version:  nil,
			bump:     Patch,
			expected: &One, // should return &One if s == nil
		},
		{
			name:          "invalid bump",
			version:       &Version{Major: 1, Minor: 2, Patch: 3},
			bump:          Bump("foo"),
			expectedPanic: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.expectedPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("expected panic, but did not panic")
					}
				}()
			}

			got := tc.version.Bump(tc.bump)
			if !tc.expectedPanic && !tc.expected.Equals(got) {
				t.Errorf("Version.Bump(%v): expected %v, got %v", tc.bump, tc.expected, got)
			}
		})
	}
}
