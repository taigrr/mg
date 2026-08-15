package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestUnregisterCommandWithPathSavesConfig(t *testing.T) {
	homeDir := t.TempDir()
	configPath := filepath.Join(homeDir, "mgconfig")
	t.Setenv("HOME", homeDir)
	t.Setenv("MGCONFIG", configPath)

	removePath := filepath.Join(homeDir, "code/remove")
	configJSON := `{
  "Repos": [
    {
      "Path": "$HOME/code/remove",
      "Remote": "git@github.com:user/remove.git"
    },
    {
      "Path": "$HOME/code/keep",
      "Remote": "git@github.com:user/keep.git"
    }
  ]
}`
	if err := os.WriteFile(configPath, []byte(configJSON), 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	executeRootCommand(t, "unregister", removePath)

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
	if saved.Repos[0].Path != "$HOME/code/keep" {
		t.Fatalf("remaining repo path = %q, want collapsed keep path", saved.Repos[0].Path)
	}
	if saved.Repos[0].Remote != "git@github.com:user/keep.git" {
		t.Fatalf("remaining repo remote = %q, want keep remote", saved.Repos[0].Remote)
	}
}
