package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
)

func TestRegisterCommandWithPathSavesConfig(t *testing.T) {
	homeDir := t.TempDir()
	configPath := filepath.Join(homeDir, "mgconfig")
	t.Setenv("HOME", homeDir)
	t.Setenv("MGCONFIG", configPath)

	configJSON := `{"Repos":[]}`
	if err := os.WriteFile(configPath, []byte(configJSON), 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	repoPath := filepath.Join(homeDir, "code/project")
	repo, err := git.PlainInit(repoPath, false)
	if err != nil {
		t.Fatalf("failed to initialize repo: %v", err)
	}

	remoteURL := "git@github.com:user/project.git"
	if _, err := repo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{remoteURL},
	}); err != nil {
		t.Fatalf("failed to create remote: %v", err)
	}

	executeRootCommand(t, "register", repoPath)

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	var saved struct {
		Repos []struct {
			Path   string
			Remote string
		}
	}
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("failed to parse config: %v", err)
	}
	if len(saved.Repos) != 1 {
		t.Fatalf("got %d repos, want 1", len(saved.Repos))
	}
	if saved.Repos[0].Path != "$HOME/code/project" {
		t.Fatalf("repo path = %q, want collapsed project path", saved.Repos[0].Path)
	}
	if saved.Repos[0].Remote != remoteURL {
		t.Fatalf("repo remote = %q, want %q", saved.Repos[0].Remote, remoteURL)
	}
}
