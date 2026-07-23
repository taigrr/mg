package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func executeRootCommand(t *testing.T, args ...string) string {
	t.Helper()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	RootCmd.SetOut(&stdout)
	RootCmd.SetErr(&stderr)
	RootCmd.SetArgs(args)

	t.Cleanup(func() {
		RootCmd.SetOut(nil)
		RootCmd.SetErr(nil)
		RootCmd.SetArgs(nil)
	})

	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("RootCmd.Execute() error = %v, stderr = %q", err, stderr.String())
	}
	return stdout.String()
}

func TestConfigCommandPrintsActivePath(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "custom-mgconfig")
	t.Setenv("MGCONFIG", configPath)

	output := executeRootCommand(t, "config")
	if strings.TrimSpace(output) != configPath {
		t.Fatalf("config output = %q, want %q", output, configPath)
	}
}

func TestConfigShowCommandPrintsActiveConfig(t *testing.T) {
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "mgconfig")
	t.Setenv("MGCONFIG", configPath)
	t.Setenv("HOME", "/home/testuser")

	configJSON := `{
  "Repos": [
    {
      "Path": "$HOME/code/project",
      "Remote": "git@github.com:user/project.git"
    }
  ],
  "Aliases": {
    "status": "git status"
  }
}`
	if err := os.WriteFile(configPath, []byte(configJSON), 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	output := executeRootCommand(t, "config", "show")

	var parsed struct {
		Repos []struct {
			Path   string
			Remote string
		}
		Aliases map[string]string
	}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("config show output is not JSON: %v\n%s", err, output)
	}
	if len(parsed.Repos) != 1 {
		t.Fatalf("got %d repos, want 1", len(parsed.Repos))
	}
	if parsed.Repos[0].Path != "/home/testuser/code/project" {
		t.Fatalf("repo path = %q, want expanded home path", parsed.Repos[0].Path)
	}
	if parsed.Repos[0].Remote != "git@github.com:user/project.git" {
		t.Fatalf("repo remote = %q, want git remote", parsed.Repos[0].Remote)
	}
	if parsed.Aliases["status"] != "git status" {
		t.Fatalf("alias status = %q, want git status", parsed.Aliases["status"])
	}
}
