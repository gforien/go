package semver

import "fmt"

type Bump string

const (
	Major Bump = "major"
	Minor Bump = "minor"
	Patch Bump = "patch"
)

func ParseBump(s string) (Bump, error) {
	switch s {
	case string(Major):
		return Major, nil
	case string(Minor):
		return Minor, nil
	case string(Patch):
		return Patch, nil
	default:
		return "", &InvalidBumpError{s}
	}
}

type InvalidBumpError struct {
	Text string // the value that caused the error
}

func (e *InvalidBumpError) Error() string {
	return fmt.Sprintf("invalid bump %q (must be major, minor, or patch)", e.Text)
}
