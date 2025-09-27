package semver

import (
	"fmt"
	"regexp"
	"strconv"
)

// Version is a semver version with format Major.minor.patch
// An zero-valued Version behaves likes 0.0.0
type Version struct {
	Major int
	Minor int
	Patch int
}

var (
	zero = &Version{0, 0, 0}
	One  = Version{0, 1, 0} // v0.1.0
)

var semver = regexp.MustCompile(`v?(\d+)\.(\d+)\.(\d+)`)

// We expect exactly 3 matches to be found when parsing
// a provided string with the semver regexp
// If the provided string does not meet this criteria, ErrCannotParse is returned.
type ErrCannotParse struct {
	message string
}

func (e *ErrCannotParse) Error() string {
	return e.message
}

func FromString(s string) (Version, error) {
	// Regular expression to capture version numbers
	matches := semver.FindStringSubmatch(s)
	if len(matches) != 4 {
		err := &ErrCannotParse{
			message: fmt.Sprintf("Cannot parse '%s' into semver (found %d matches)", s, len(matches)),
		}
		return Version{}, err
	}

	// Convert matches to integers
	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	patch, _ := strconv.Atoi(matches[3])
	return Version{Major: major, Minor: minor, Patch: patch}, nil
}

func (s *Version) String() string {
	if s == nil {
		return zero.String()
	}

	return fmt.Sprintf("%d.%d.%d", s.Major, s.Minor, s.Patch)
}

func (s *Version) NextMajor() *Version {
	if s == nil {
		return zero.NextMajor()
	}

	return &Version{
		Major: s.Major + 1,
		Minor: 0,
		Patch: 0,
	}
}

func (s *Version) NextMinor() *Version {
	if s == nil {
		return zero.NextMinor()
	}

	return &Version{
		Major: s.Major,
		Minor: s.Minor + 1,
		Patch: 0,
	}
}

func (s *Version) NextPatch() *Version {
	if s == nil {
		return zero.NextPatch()
	}

	return &Version{
		Major: s.Major,
		Minor: s.Minor,
		Patch: s.Patch + 1,
	}
}

func (s *Version) Bump(b Bump) *Version {
	if s == nil {
		return &One
	}

	switch b {
	case Major:
		return s.NextMajor()
	case Minor:
		return s.NextMinor()
	case Patch:
		return s.NextPatch()
	default:
		panic(&InvalidBumpError{string(b)})
	}
}

// Compare returns
// -1 if the version is less than the other version,
// 0 if they are equal,
// +1 if the version is greater than the other version.
func (s *Version) Compare(other *Version) int {
	if s == nil {
		return zero.Compare(other)
	}
	if other == nil {
		return s.Compare(zero)
	}

	if (s.Major > other.Major) ||
		(s.Major == other.Major && s.Minor > other.Minor) ||
		(s.Major == other.Major && s.Minor == other.Minor && s.Patch > other.Patch) {
		return 1
	}

	if (s.Major < other.Major) ||
		(s.Major == other.Major && s.Minor < other.Minor) ||
		(s.Major == other.Major && s.Minor == other.Minor && s.Patch < other.Patch) {
		return -1
	}

	return 0
}

func (s *Version) GreaterThan(other *Version) bool {
	return s.Compare(other) > 0
}

func (s *Version) GreaterOrEqual(other *Version) bool {
	return s.Compare(other) >= 0
}

func (s *Version) LessThan(other *Version) bool {
	return s.Compare(other) < 0
}

func (s *Version) LessOrEqual(other *Version) bool {
	return s.Compare(other) <= 0
}

func (s *Version) Equals(other *Version) bool {
	return s.Compare(other) == 0
}
