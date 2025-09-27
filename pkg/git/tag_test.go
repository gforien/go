package git

import (
	"testing"

	"github.com/gforien/go/pkg/semver"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"
)

func TestGetVersion(t *testing.T) {
	tests := []struct {
		name     string
		given    []string
		expected semver.Version
	}{
		{
			given:    []string{"1.0.0", "2.0.0", "1.2.0"},
			expected: semver.Version{Major: 2, Minor: 0, Patch: 0},
		},
		{
			name:     "0.17.0 should be >= 0.9.0",
			given:    []string{"0.1.0", "0.9.0", "0.17.0"},
			expected: semver.Version{Major: 0, Minor: 17, Patch: 0},
		},
		{
			name:     "No Tags",
			given:    []string{},
			expected: semver.Version{Major: 0, Minor: 0, Patch: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Initialize an in-memory Git repository
			repo, err := git.Init(memory.NewStorage(), memfs.New())
			if err != nil {
				t.Fatalf("error in test setup: creating in-memory repository: %v", err)
			}
			wt, err := repo.Worktree()
			if err != nil {
				t.Fatalf("error in test setup: retrieving worktree: %v", err)
			}

			// If there are any tags to be added, create them
			for _, tag := range tt.given {
				h, err := wt.Commit(tag, &git.CommitOptions{AllowEmptyCommits: true})
				if err != nil {
					t.Fatalf("error in test setup: creating commit: %v", err)
				}
				_, err = repo.CreateTag(tag, h, nil)
				if err != nil {
					t.Fatalf("error in test setup: creating tag %v: %v", tag, err)
				}
			}

			_, v, err := GetVersion(repo)
			if err != nil {
				t.Errorf("expected no error, got %#v", err)
			}
			if !tt.expected.Equals(v) {
				t.Errorf("expected %#v, got %#v", tt.expected, v)
			}
		})
	}
}
