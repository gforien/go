package git

import (
	"fmt"

	"github.com/gforien/go/pkg/semver"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// GetVersion returns the latest semver tag in a repo
func GetVersion(repo *git.Repository) (*plumbing.Reference, *semver.Version, error) {
	tags, err := repo.Tags()
	if err != nil {
		return nil, nil, fmt.Errorf("error fetching tags: %w", err)
	}

	var latestRef *plumbing.Reference
	var latestTag, t semver.Version

	if err = tags.ForEach(func(ref *plumbing.Reference) error {
		if t, err = semver.FromString(ref.Name().Short()); err != nil {
			return nil
		}
		if t.GreaterThan(&latestTag) {
			latestTag = t
			latestRef = ref
		}
		return nil
	}); err != nil {
		return nil, nil, fmt.Errorf("error iterating tags: %w", err)
	}

	return latestRef, &latestTag, nil
}
