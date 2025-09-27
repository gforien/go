package git

import (
	"errors"
	"testing"

	"github.com/gforien/go/pkg/semver"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
)

func TestGetVersion(t *testing.T) {
	type testCase struct {
		name      string
		given     []string
		prefix    string
		expected  semver.Version
		expectRef bool // whether we expect a non-nil reference
	}

	tests := []testCase{
		{
			name:      "simple tags",
			given:     []string{"1.0.0", "2.0.0", "1.2.0"},
			expected:  semver.Version{Major: 2, Minor: 0, Patch: 0},
			expectRef: true,
		},
		{
			name:      "0.17.0 should be >= 0.9.0",
			given:     []string{"0.1.0", "0.9.0", "0.17.0"},
			expected:  semver.Version{Major: 0, Minor: 17, Patch: 0},
			expectRef: true,
		},
		{
			name:      "No Tags",
			given:     []string{},
			expected:  semver.Version{Major: 0, Minor: 0, Patch: 0},
			expectRef: false,
		},
		{
			name:      "simple monorepo tag",
			given:     []string{"module/v1.0.0", "module/v1.2.0", "other/v2.0.0"},
			prefix:    "module/",
			expected:  semver.Version{Major: 1, Minor: 2, Patch: 0},
			expectRef: true,
		},
		{
			name:      "many monorepo tags",
			given:     []string{"v0.1.0", "v0.2.0", "my/module/v1.0.0", "my/module/v1.2.0", "other/v2.0.0"},
			prefix:    "my/module/",
			expected:  semver.Version{Major: 1, Minor: 2, Patch: 0},
			expectRef: true,
		},
		{
			name:      "Prefix with no matching tags",
			given:     []string{"other/v1.0.0", "other/v1.2.0"},
			prefix:    "my/module/",
			expected:  semver.Version{Major: 0, Minor: 0, Patch: 0},
			expectRef: false,
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

			ref, v, err := GetVersion(repo, tt.prefix)
			if err != nil {
				t.Errorf("expected no error, got %#v", err)
			}

			if !tt.expected.Equals(v) {
				t.Errorf("expected %#v, got %#v", tt.expected, v)
			}
			if tt.expectRef && ref == nil {
				t.Errorf("expected a reference, got nil")
			}
			if !tt.expectRef && ref != nil {
				t.Errorf("expected no reference, got %#v", ref)
			}
		})
	}
}

func TestEnsureCommitSince(t *testing.T) {
	type CommitMsg struct {
		commitMsg string
		tagMsg    string // empty if no tag
	}
	tests := []struct {
		name               string
		givenTag           string
		commitTagMsgs      []CommitMsg
		expectErrRefIsHead bool // expect ErrRefIsHead
		expectAnyErr       bool // expect any error
	}{
		{
			name:     "HEAD equals tag",
			givenTag: "tag1",
			commitTagMsgs: []CommitMsg{
				{"commit1", "tag1"},
			},
			expectErrRefIsHead: true,
		},
		{
			name:     "one new commit after tag",
			givenTag: "tag1",
			commitTagMsgs: []CommitMsg{
				{"commit1", "tag1"},
				{"commit2", ""},
			},
		},
		{
			name:     "several commits and tags",
			givenTag: "tag2",
			commitTagMsgs: []CommitMsg{
				{"c1", "tag1"},
				{"c2", ""},
				{"c3", "tag2"},
				{"c4", ""},
				{"c5", ""},
			},
		},
		{
			name:     "nonexistent tag",
			givenTag: "nonexistent",
			commitTagMsgs: []CommitMsg{
				{"c1", "tag1"},
				{"c2", ""},
			},
			expectAnyErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo, err := git.Init(memory.NewStorage(), memfs.New())
			if err != nil {
				t.Fatalf("failed to init repo: %v", err)
			}
			wt, err := repo.Worktree()
			if err != nil {
				t.Fatalf("failed to get worktree: %v", err)
			}

			tagRefs := map[string]*plumbing.Reference{}
			for _, ct := range tt.commitTagMsgs {
				h, err := wt.Commit(ct.commitMsg, &git.CommitOptions{AllowEmptyCommits: true})
				if err != nil {
					t.Fatalf("failed to commit: %v", err)
				}
				if ct.tagMsg != "" {
					ref, err := repo.CreateTag(ct.tagMsg, h, nil)
					if err != nil {
						t.Fatalf("failed to create tag %s: %v", ct.tagMsg, err)
					}
					tagRefs[ct.tagMsg] = ref
				}
			}

			ref, ok := tagRefs[tt.givenTag]
			if !ok {
				// simulate missing tag reference
				ref = &plumbing.Reference{}
			}

			got := EnsureCommitSince(repo, ref)

			switch {
			case tt.expectErrRefIsHead:
				var noNew ErrRefIsHead
				if !errors.As(got, &noNew) {
					t.Errorf("expected ErrNoNewCommit, got %v", got)
				}
			case tt.expectAnyErr:
				if got == nil {
					t.Errorf("expected an error, got nil")
				}
			default:
				if got != nil {
					t.Errorf("unexpected error: %v", got)
				}
			}
		})
	}
}
