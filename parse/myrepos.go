package parse

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type MRConfig struct {
	Repos   []Repo
	Aliases map[string]string
}
type Repo struct {
	Path     string
	Checkout string
	Aliases  map[string]string
}

func (m MRConfig) GetRepoPaths() []string {
	paths := []string{}
	for _, r := range m.Repos {
		paths = append(paths, r.Path)
	}
	return paths
}

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
	f, err := ioutil.ReadFile(mrconfPath)
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
			config.Repos[length].Checkout = split[1]

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
