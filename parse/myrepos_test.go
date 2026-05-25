package parse

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadMRConfig(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		setupDir  bool
		wantErr   string
		wantRepos []Repo
		wantAlias map[string]string
	}{
		{
			name: "loads aliases and repo checkouts",
			content: `[DEFAULT]
fetch = git fetch --all --prune
status = git status --short

[code/mg]
checkout = git clone 'git@github.com:taigrr/mg.git' 'mg'

[/opt/repos/grlx]
checkout = git clone 'git@github.com:gogrlx/grlx.git' 'grlx'
`,
			wantRepos: []Repo{
				{Path: "code/mg", Remote: "git clone 'git@github.com:taigrr/mg.git' 'mg'"},
				{Path: "/opt/repos/grlx", Remote: "git clone 'git@github.com:gogrlx/grlx.git' 'grlx'"},
			},
			wantAlias: map[string]string{
				"fetch":  "git fetch --all --prune",
				"status": "git status --short",
			},
		},
		{
			name: "rejects unexpected repo arguments",
			content: `[code/mg]
update = git pull --ff-only
`,
			wantErr: "unexpected argument",
		},
		{
			name:     "rejects directory mrconfig path",
			setupDir: true,
			wantErr:  "expected mrconfig file but got a directory",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			homeDir := t.TempDir()
			t.Setenv("HOME", homeDir)
			mrconfigPath := filepath.Join(homeDir, ".mrconfig")

			if test.setupDir {
				if err := os.Mkdir(mrconfigPath, 0o755); err != nil {
					t.Fatalf("failed to create mrconfig dir: %v", err)
				}
			} else {
				if err := os.WriteFile(mrconfigPath, []byte(test.content), 0o644); err != nil {
					t.Fatalf("failed to write mrconfig: %v", err)
				}
			}

			config, err := LoadMRConfig()
			if test.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", test.wantErr)
				}
				if !strings.Contains(err.Error(), test.wantErr) {
					t.Fatalf("expected error containing %q, got %q", test.wantErr, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("LoadMRConfig() error = %v", err)
			}

			if len(config.Repos) != len(test.wantRepos) {
				t.Fatalf("expected %d repos, got %d", len(test.wantRepos), len(config.Repos))
			}
			for index, repo := range config.Repos {
				wantPath := test.wantRepos[index].Path
				if !filepath.IsAbs(wantPath) {
					wantPath = filepath.Join(homeDir, wantPath)
				}
				if repo.Path != wantPath {
					t.Fatalf("repo %d path = %q, want %q", index, repo.Path, wantPath)
				}
				if repo.Remote != test.wantRepos[index].Remote {
					t.Fatalf("repo %d remote = %q, want %q", index, repo.Remote, test.wantRepos[index].Remote)
				}
			}

			if len(config.Aliases) != len(test.wantAlias) {
				t.Fatalf("expected %d aliases, got %d", len(test.wantAlias), len(config.Aliases))
			}
			for key, wantValue := range test.wantAlias {
				if gotValue := config.Aliases[key]; gotValue != wantValue {
					t.Fatalf("alias %q = %q, want %q", key, gotValue, wantValue)
				}
			}
		})
	}
}

func TestToMGConfig(t *testing.T) {
	mrConfig := MRConfig{
		Repos: []Repo{
			{Path: "/repos/mg", Remote: "git clone 'git@github.com:taigrr/mg.git' 'mg'"},
			{Path: "/repos/plain", Remote: "git@github.com:taigrr/plain.git"},
		},
		Aliases: map[string]string{"status": "git status --short"},
	}

	mgConfig := mrConfig.ToMGConfig()

	if len(mgConfig.Repos) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(mgConfig.Repos))
	}
	if mgConfig.Repos[0].Remote != "git@github.com:taigrr/mg.git" {
		t.Fatalf("first remote = %q, want %q", mgConfig.Repos[0].Remote, "git@github.com:taigrr/mg.git")
	}
	if mgConfig.Repos[1].Remote != "git@github.com:taigrr/plain.git" {
		t.Fatalf("second remote = %q, want unchanged remote", mgConfig.Repos[1].Remote)
	}
	if mgConfig.Aliases["status"] != "git status --short" {
		t.Fatalf("aliases were not preserved")
	}
}
