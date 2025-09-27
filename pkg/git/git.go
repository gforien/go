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

// EnsureCommitSince ensures that there is at least one commit since the given ref
// Returns ErrRefIsHead if the given ref is the current HEAD
func EnsureCommitSince(repo *git.Repository, ref *plumbing.Reference) error {
	head, err := repo.Head()
	if err != nil {
		return fmt.Errorf("error getting HEAD: %w", err)
	}

	if ref.Hash() == head.Hash() {
		return ErrRefIsHead{Ref: ref}
	}

	refCommit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return fmt.Errorf("error resolving ref commit: %w", err)
	}

	headCommit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return fmt.Errorf("error resolving head commit: %w", err)
	}

	isAncestor, err := refCommit.IsAncestor(headCommit)
	if err != nil {
		return fmt.Errorf("error checking ancestry: %w", err)
	}
	if !isAncestor {
		return fmt.Errorf("HEAD is not a descendant of ref %s", ref.Name().Short())
	}

	return nil
}

type ErrRefIsHead struct {
	Ref *plumbing.Reference
}

func (e ErrRefIsHead) Error() string {
	return fmt.Sprintf("no new commits since tag %v (%v)", e.Ref.Name().Short(), e.Ref.Hash())
}
