package rhsm

import (
	"path/filepath"
	"testing"
)

func TestRepoEnabled(t *testing.T) {
	tests := []struct {
		name     string
		repoFile string
		repoID   string
		expected bool
	}{
		{
			name:     "enabled_numeric_1",
			repoID:   "repo-test",
			expected: true,
		},
		{
			name:     "enabled_boolean_true",
			repoID:   "repo-test",
			expected: true,
		},
		{
			name:     "enabled_invalid_value",
			repoID:   "repo-test",
			expected: false,
		},
		{
			name:     "enabled_mixed_case_true",
			repoID:   "repo-test",
			expected: true,
		},
		{
			name:     "comment_in_line",
			repoID:   "repo-test",
			expected: true,
		},
		{
			name:     "repo_not_found",
			repoID:   "non-existent-repo",
			expected: false,
		},
		{
			name:     "no_enabled_flag",
			repoID:   "repo-test",
			expected: false,
		},
		{
			name:     "empty_file",
			repoID:   "repo-test",
			expected: false,
		},
		{
			name:     "empty_file",
			repoID:   "repo-test",
			expected: false,
		},
		{
			name:     "rhel10_baseos",
			repoFile: "rhel10",
			repoID:   "rhel-10-for-x86_64-baseos-rpms",
			expected: true,
		},
		{
			name:     "rhel10_appstream",
			repoFile: "rhel10",
			repoID:   "rhel-10-for-x86_64-appstream-rpms",
			expected: true,
		},
		{
			name:     "rhel10_extensions",
			repoFile: "rhel10",
			repoID:   "rhel-10-for-x86_64-extensions-rpms",
			expected: true,
		},
		{
			name:     "rhel10_supplementary",
			repoFile: "rhel10",
			repoID:   "rhel-10-for-x86_64-supplementary-rpms",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoFile := tt.repoFile
			if repoFile == "" {
				repoFile = tt.name
			}
			repoPath := filepath.Join("testdata/" + repoFile + ".repo")
			states, err := readRepoStates(repoPath)
			if err != nil {
				t.Fatal(err)
			}
			result := states[tt.repoID]
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
