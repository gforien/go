package git

import (
	"fmt"

	"github.com/gforien/go/pkg/semver"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// GetVersion returns the latest semver tag in a repo
func GetVersion(repo *git.Repository) (*semver.Version, error) {
	tags, err := repo.Tags()
	if err != nil {
		return nil, fmt.Errorf("error fetching tags: %w", err)
	}

	var latestTag, t semver.Version

	err = tags.ForEach(func(ref *plumbing.Reference) error {
		t, err = semver.FromString(ref.Name().Short())
		if err != nil {
			return nil
		}
		if t.GreaterThan(&latestTag) {
			latestTag = t
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error iterating tags: %w", err)
	}

	return &latestTag, nil
}
