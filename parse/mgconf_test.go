package parse

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestExpandPaths(t *testing.T) {
	// Set up test environment variables
	t.Setenv("HOME", "/home/testuser")
	t.Setenv("GOPATH", "/home/testuser/go")
	t.Setenv("CUSTOM_VAR", "/custom/path")

	tests := []struct {
		name     string
		input    []Repo
		expected []string
	}{
		{
			name: "expand $HOME",
			input: []Repo{
				{Path: "$HOME/code/project", Remote: "git@github.com:user/project.git"},
			},
			expected: []string{"/home/testuser/code/project"},
		},
		{
			name: "expand multiple variables",
			input: []Repo{
				{Path: "$HOME/code/project", Remote: "git@github.com:user/project.git"},
				{Path: "$GOPATH/src/github.com/user/repo", Remote: "git@github.com:user/repo.git"},
			},
			expected: []string{
				"/home/testuser/code/project",
				"/home/testuser/go/src/github.com/user/repo",
			},
		},
		{
			name: "expand custom variable",
			input: []Repo{
				{Path: "$CUSTOM_VAR/subdir", Remote: "git@github.com:user/project.git"},
			},
			expected: []string{"/custom/path/subdir"},
		},
		{
			name: "no expansion needed",
			input: []Repo{
				{Path: "/absolute/path/to/repo", Remote: "git@github.com:user/project.git"},
			},
			expected: []string{"/absolute/path/to/repo"},
		},
		{
			name: "empty repos",
			input: []Repo{},
			expected: []string{},
		},
		{
			name: "undefined variable stays as-is",
			input: []Repo{
				{Path: "$UNDEFINED_VAR/code", Remote: "git@github.com:user/project.git"},
			},
			expected: []string{"/code"}, // undefined vars expand to empty string
		},
		{
			name: "braced variable syntax",
			input: []Repo{
				{Path: "${HOME}/code/project", Remote: "git@github.com:user/project.git"},
			},
			expected: []string{"/home/testuser/code/project"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := MGConfig{Repos: tt.input}
			conf.ExpandPaths()

			if len(conf.Repos) != len(tt.expected) {
				t.Fatalf("expected %d repos, got %d", len(tt.expected), len(conf.Repos))
			}

			for i, repo := range conf.Repos {
				if repo.Path != tt.expected[i] {
					t.Errorf("repo %d: expected path %q, got %q", i, tt.expected[i], repo.Path)
				}
			}
		})
	}
}

func TestCollapsePaths(t *testing.T) {
	// Get the actual home directory for this test
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home directory: %v", err)
	}

	tests := []struct {
		name     string
		input    []Repo
		expected []string
	}{
		{
			name: "collapse home directory",
			input: []Repo{
				{Path: filepath.Join(home, "code/project"), Remote: "git@github.com:user/project.git"},
			},
			expected: []string{"$HOME/code/project"},
		},
		{
			name: "collapse multiple paths",
			input: []Repo{
				{Path: filepath.Join(home, "code/project1"), Remote: "git@github.com:user/project1.git"},
				{Path: filepath.Join(home, "code/project2"), Remote: "git@github.com:user/project2.git"},
			},
			expected: []string{
				"$HOME/code/project1",
				"$HOME/code/project2",
			},
		},
		{
			name: "path not under home",
			input: []Repo{
				{Path: "/opt/repos/project", Remote: "git@github.com:user/project.git"},
			},
			expected: []string{"/opt/repos/project"},
		},
		{
			name: "mixed paths",
			input: []Repo{
				{Path: filepath.Join(home, "code/project"), Remote: "git@github.com:user/project.git"},
				{Path: "/opt/repos/other", Remote: "git@github.com:user/other.git"},
			},
			expected: []string{
				"$HOME/code/project",
				"/opt/repos/other",
			},
		},
		{
			name: "empty repos",
			input: []Repo{},
			expected: []string{},
		},
		{
			name: "home directory itself",
			input: []Repo{
				{Path: home, Remote: "git@github.com:user/home.git"},
			},
			expected: []string{"$HOME"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := MGConfig{Repos: tt.input}
			conf.CollapsePaths()

			if len(conf.Repos) != len(tt.expected) {
				t.Fatalf("expected %d repos, got %d", len(tt.expected), len(conf.Repos))
			}

			for i, repo := range conf.Repos {
				if repo.Path != tt.expected[i] {
					t.Errorf("repo %d: expected path %q, got %q", i, tt.expected[i], repo.Path)
				}
			}
		})
	}
}

func TestExpandAndCollapse_Roundtrip(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home directory: %v", err)
	}

	// Start with $HOME-based paths (as stored in config)
	original := MGConfig{
		Repos: []Repo{
			{Path: "$HOME/code/project1", Remote: "git@github.com:user/project1.git"},
			{Path: "$HOME/go/src/github.com/user/repo", Remote: "git@github.com:user/repo.git"},
			{Path: "/opt/external/repo", Remote: "git@github.com:user/external.git"},
		},
	}

	// Expand paths (as done when loading)
	conf := MGConfig{
		Repos: make([]Repo, len(original.Repos)),
	}
	copy(conf.Repos, original.Repos)
	conf.ExpandPaths()

	// Verify expansion worked
	expectedExpanded := []string{
		filepath.Join(home, "code/project1"),
		filepath.Join(home, "go/src/github.com/user/repo"),
		"/opt/external/repo",
	}
	for i, repo := range conf.Repos {
		if repo.Path != expectedExpanded[i] {
			t.Errorf("after expand, repo %d: expected %q, got %q", i, expectedExpanded[i], repo.Path)
		}
	}

	// Collapse paths (as done when saving)
	conf.CollapsePaths()

	// Verify we're back to the original
	for i, repo := range conf.Repos {
		if repo.Path != original.Repos[i].Path {
			t.Errorf("after roundtrip, repo %d: expected %q, got %q", i, original.Repos[i].Path, repo.Path)
		}
	}
}

func TestSave_CollapsesPaths(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home directory: %v", err)
	}

	// Create a temp file for the config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "mgconfig")
	t.Setenv("MGCONFIG", configPath)

	// Create config with expanded (absolute) paths
	conf := MGConfig{
		Repos: []Repo{
			{Path: filepath.Join(home, "code/project"), Remote: "git@github.com:user/project.git"},
			{Path: "/opt/external/repo", Remote: "git@github.com:user/external.git"},
		},
		Aliases: map[string]string{"test": "echo test"},
	}

	// Save should collapse paths
	err = conf.Save()
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Read back and verify paths are collapsed
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read saved config: %v", err)
	}

	var saved MGConfig
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("failed to parse saved config: %v", err)
	}

	expectedPaths := []string{
		"$HOME/code/project",
		"/opt/external/repo",
	}

	for i, repo := range saved.Repos {
		if repo.Path != expectedPaths[i] {
			t.Errorf("saved repo %d: expected path %q, got %q", i, expectedPaths[i], repo.Path)
		}
	}

	// Verify original config wasn't modified
	if conf.Repos[0].Path != filepath.Join(home, "code/project") {
		t.Errorf("original config was modified: expected %q, got %q",
			filepath.Join(home, "code/project"), conf.Repos[0].Path)
	}
}

func TestParseMGConfig(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantRepos int
		wantErr   bool
	}{
		{
			name: "valid config",
			input: `{
				"Repos": [
					{"Path": "$HOME/code/project", "Remote": "git@github.com:user/project.git"}
				],
				"Aliases": {"gc": "git gc"}
			}`,
			wantRepos: 1,
			wantErr:   false,
		},
		{
			name:      "empty config",
			input:     `{"Repos": [], "Aliases": {}}`,
			wantRepos: 0,
			wantErr:   false,
		},
		{
			name:      "invalid json",
			input:     `{invalid}`,
			wantRepos: 0,
			wantErr:   true,
		},
		{
			name: "multiple repos",
			input: `{
				"Repos": [
					{"Path": "$HOME/code/project1", "Remote": "git@github.com:user/project1.git"},
					{"Path": "$HOME/code/project2", "Remote": "git@github.com:user/project2.git"}
				]
			}`,
			wantRepos: 2,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf, err := ParseMGConfig([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMGConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(conf.Repos) != tt.wantRepos {
				t.Errorf("ParseMGConfig() got %d repos, want %d", len(conf.Repos), tt.wantRepos)
			}
		})
	}
}

func TestAddRepo(t *testing.T) {
	tests := []struct {
		name       string
		initial    []Repo
		addPath    string
		addRemote  string
		wantErr    bool
		wantCount  int
	}{
		{
			name:      "add to empty",
			initial:   []Repo{},
			addPath:   "$HOME/code/new",
			addRemote: "git@github.com:user/new.git",
			wantErr:   false,
			wantCount: 1,
		},
		{
			name: "add to existing",
			initial: []Repo{
				{Path: "$HOME/code/existing", Remote: "git@github.com:user/existing.git"},
			},
			addPath:   "$HOME/code/new",
			addRemote: "git@github.com:user/new.git",
			wantErr:   false,
			wantCount: 2,
		},
		{
			name: "add duplicate",
			initial: []Repo{
				{Path: "$HOME/code/existing", Remote: "git@github.com:user/existing.git"},
			},
			addPath:   "$HOME/code/existing",
			addRemote: "git@github.com:user/existing.git",
			wantErr:   true,
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := MGConfig{Repos: tt.initial}
			err := conf.AddRepo(tt.addPath, tt.addRemote)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddRepo() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(conf.Repos) != tt.wantCount {
				t.Errorf("AddRepo() repo count = %d, want %d", len(conf.Repos), tt.wantCount)
			}
		})
	}
}

func TestDelRepo(t *testing.T) {
	tests := []struct {
		name      string
		initial   []Repo
		delPath   string
		wantErr   bool
		wantCount int
	}{
		{
			name: "delete existing",
			initial: []Repo{
				{Path: "$HOME/code/project", Remote: "git@github.com:user/project.git"},
			},
			delPath:   "$HOME/code/project",
			wantErr:   false,
			wantCount: 0,
		},
		{
			name: "delete from multiple",
			initial: []Repo{
				{Path: "$HOME/code/project1", Remote: "git@github.com:user/project1.git"},
				{Path: "$HOME/code/project2", Remote: "git@github.com:user/project2.git"},
			},
			delPath:   "$HOME/code/project1",
			wantErr:   false,
			wantCount: 1,
		},
		{
			name: "delete non-existent",
			initial: []Repo{
				{Path: "$HOME/code/project", Remote: "git@github.com:user/project.git"},
			},
			delPath:   "$HOME/code/other",
			wantErr:   true,
			wantCount: 1,
		},
		{
			name:      "delete from empty",
			initial:   []Repo{},
			delPath:   "$HOME/code/project",
			wantErr:   true,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := MGConfig{Repos: tt.initial}
			err := conf.DelRepo(tt.delPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("DelRepo() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(conf.Repos) != tt.wantCount {
				t.Errorf("DelRepo() repo count = %d, want %d", len(conf.Repos), tt.wantCount)
			}
		})
	}
}

func TestMerge(t *testing.T) {
	tests := []struct {
		name           string
		initial        []Repo
		merge          []Repo
		wantCount      int
		wantDuplicates int
		wantNewPaths   int
	}{
		{
			name:    "merge into empty",
			initial: []Repo{},
			merge: []Repo{
				{Path: "$HOME/code/new", Remote: "git@github.com:user/new.git"},
			},
			wantCount:      1,
			wantDuplicates: 0,
			wantNewPaths:   1,
		},
		{
			name: "merge with duplicates",
			initial: []Repo{
				{Path: "$HOME/code/existing", Remote: "git@github.com:user/existing.git"},
			},
			merge: []Repo{
				{Path: "$HOME/code/existing", Remote: "git@github.com:user/existing.git"},
				{Path: "$HOME/code/new", Remote: "git@github.com:user/new.git"},
			},
			wantCount:      2,
			wantDuplicates: 1,
			wantNewPaths:   1,
		},
		{
			name: "merge all duplicates",
			initial: []Repo{
				{Path: "$HOME/code/project", Remote: "git@github.com:user/project.git"},
			},
			merge: []Repo{
				{Path: "$HOME/code/project", Remote: "git@github.com:user/project.git"},
			},
			wantCount:      1,
			wantDuplicates: 1,
			wantNewPaths:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := MGConfig{Repos: tt.initial}
			mergeConf := MGConfig{Repos: tt.merge}

			stats, err := conf.Merge(mergeConf)
			if err != nil {
				t.Fatalf("Merge() unexpected error: %v", err)
			}

			if len(conf.Repos) != tt.wantCount {
				t.Errorf("Merge() repo count = %d, want %d", len(conf.Repos), tt.wantCount)
			}
			if stats.Duplicates != tt.wantDuplicates {
				t.Errorf("Merge() duplicates = %d, want %d", stats.Duplicates, tt.wantDuplicates)
			}
			if len(stats.NewPaths) != tt.wantNewPaths {
				t.Errorf("Merge() new paths = %d, want %d", len(stats.NewPaths), tt.wantNewPaths)
			}
		})
	}
}

func TestGetRepoPaths(t *testing.T) {
	conf := MGConfig{
		Repos: []Repo{
			{Path: "$HOME/code/project1", Remote: "git@github.com:user/project1.git"},
			{Path: "$HOME/code/project2", Remote: "git@github.com:user/project2.git"},
		},
	}

	paths := conf.GetRepoPaths()

	if len(paths) != 2 {
		t.Fatalf("GetRepoPaths() returned %d paths, want 2", len(paths))
	}

	expected := []string{"$HOME/code/project1", "$HOME/code/project2"}
	for i, path := range paths {
		if path != expected[i] {
			t.Errorf("GetRepoPaths()[%d] = %q, want %q", i, path, expected[i])
		}
	}
}
