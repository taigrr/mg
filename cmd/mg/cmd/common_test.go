package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetConfigRejectsInvalidConfig(t *testing.T) {
	if os.Getenv("MG_TEST_CONFIG_CHILD") == "1" {
		GetConfig()
		return
	}
	for _, test := range []struct{ name, contents, diagnostic string }{
		{"malformed JSON", `{"Repos":`, "unexpected end of JSON input"},
		{"partial config", `{"Repos":[{"Path":"/tmp/repo"}],"Aliases":42}`, "cannot unmarshal"},
	} {
		t.Run(test.name, func(t *testing.T) {
			configPath := filepath.Join(t.TempDir(), "mgconfig")
			if err := os.WriteFile(configPath, []byte(test.contents), 0o600); err != nil {
				t.Fatal(err)
			}
			// A valid migration source must not mask a broken existing mgconfig.
			homeDir := t.TempDir()
			if err := os.WriteFile(filepath.Join(homeDir, ".mrconfig"), []byte("[fallback]\ncheckout = git clone 'example' 'fallback'\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			t.Setenv("MGCONFIG", configPath)
			t.Setenv("HOME", homeDir)
			t.Setenv("MG_TEST_CONFIG_CHILD", "1")
			command := exec.Command(os.Args[0], "-test.run=^TestGetConfigRejectsInvalidConfig$")
			output, err := command.CombinedOutput()
			if err == nil {
				t.Fatalf("GetConfig accepted invalid config: %s", output)
			}
			if !strings.Contains(string(output), test.diagnostic) {
				t.Fatalf("missing diagnostic %q: %s", test.diagnostic, output)
			}
			contents, err := os.ReadFile(configPath)
			if err != nil {
				t.Fatal(err)
			}
			if string(contents) != test.contents {
				t.Fatalf("invalid config overwritten: %s", contents)
			}
		})
	}
}
