package parse

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type MRConfig struct {
	Repos   []Repo
	Aliases map[string]string
}
type Repo struct {
	Path    string
	Remote  string
	Aliases map[string]string `json:"aliases,omitempty"`
}

// GetRepoPaths returns a slice of strings containing the paths of all repos
// in the MRConfig struct
func (m MRConfig) GetRepoPaths() []string {
	paths := []string{}
	for _, r := range m.Repos {
		paths = append(paths, r.Path)
	}
	return paths
}

func (m MRConfig) ToMGConfig() MGConfig {
	mgconf := MGConfig(m)
	for i, repo := range mgconf.Repos {
		checkout := repo.Remote
		if after, ok := strings.CutPrefix(checkout, "git clone '"); ok {
			// git clone 'git@bitbucket.org:taigrr/mg.git' 'mg'
			remote := after
			sp := strings.Split(remote, "' '")
			remote = sp[0]
			mgconf.Repos[i].Remote = remote
		}
	}
	return mgconf
}

// LoadMRConfig loads the mrconfig file from the user's home directory
// and returns a MRConfig struct with all repos and aliases from the [DEFAULT] section
func LoadMRConfig() (MRConfig, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return MRConfig{}, err
	}
	mrconfPath := filepath.Join(home, ".mrconfig")
	s, err := os.Stat(mrconfPath)
	if err != nil {
		return MRConfig{}, err
	}
	if s.IsDir() {
		return MRConfig{}, errors.New("expected mrconfig file but got a directory")
	}
	f, err := os.ReadFile(mrconfPath)
	if err != nil {
		return MRConfig{}, err
	}
	config := MRConfig{
		Aliases: make(map[string]string),
		Repos:   []Repo{},
	}

	inRepo := false
	for n, line := range strings.Split(string(f), "\n") {
		line = strings.TrimSpace(line)
		// skip blank lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if line == "[DEFAULT]" {
			inRepo = false
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			path := strings.Trim(line, "[]")
			if !strings.HasPrefix(path, "/") {
				path = filepath.Join(home, path)
			}
			config.Repos = append(config.Repos, Repo{Path: path})
			inRepo = true
			continue
		}
		key, value, ok := strings.Cut(line, " = ")
		if !ok {
			return MRConfig{}, fmt.Errorf("unexpected argument on line %d: %s", n, line)
		}
		if inRepo {
			if key != "checkout" {
				return MRConfig{}, fmt.Errorf("unexpected argument on line %d: %s", n, line)
			}
			config.Repos[len(config.Repos)-1].Remote = value
		} else {
			// Load all DEFAULT section aliases into the map
			config.Aliases[key] = value
		}
	}
	return config, nil
}
