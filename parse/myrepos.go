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
// and returns a MRConfig struct
// TODO: load aliases into map instead of hardcoded Unregister prop
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
	text := string(f)
	lines := strings.Split(text, "\n")
	config := MRConfig{
		Aliases: make(map[string]string),
		Repos:   []Repo{},
	}

	length := -1
	mode := "default"
	for n, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// ignore comments in mrconfig
		if strings.HasPrefix(line, "#") {
			continue
		}
		if line == "[DEFAULT]" {
			mode = "default"
			continue
		} else if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			length++
			path := strings.Trim(line, "[]")
			if !strings.HasPrefix(path, "/") {
				path = filepath.Join(home, path)
			}
			mode = "repo"
			config.Repos = append(config.Repos, Repo{Path: path})
			continue
		}
		split := strings.SplitN(line, " = ", 2)
		if len(split) != 2 {
			return MRConfig{}, fmt.Errorf("unexpected argument on line %d: %s", n, line)
		}
		switch mode {
		case "repo":
			if split[0] != "checkout" {
				return MRConfig{}, fmt.Errorf("unexpected argument on line %d: %s", n, line)
			}

			config.Repos[length].Remote = split[1]

		case "default":

			// TODO load text into Aliases map instead of hardcoded Unregister prop
			switch split[0] {
			case "unregister":
				config.Aliases["unregister"] = split[1]
			case "git_gc":
				config.Aliases["gc"] = split[1]
			default:
				return MRConfig{}, fmt.Errorf("unexpected argument on line %d: %s", n, line)
			}
		}
	}
	return config, nil
}
