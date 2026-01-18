package parse

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var errAlreadyRegistered = os.ErrExist

// MGConfig is the struct that represents the mgconfig file
// It contains a slice of Repo structs and a map of aliases
// The aliases map is a map of strings to strings, where the key is the alias
// and the value is a command to be run
type MGConfig struct {
	Repos   []Repo
	Aliases map[string]string
}

// GetRepoPaths returns a slice of strings containing the paths of the repos
// in the mgconfig file

func (m MGConfig) GetRepoPaths() []string {
	paths := []string{}
	for _, r := range m.Repos {
		paths = append(paths, r.Path)
	}
	return paths
}

func (m *MGConfig) DelRepo(path string) error {
	for i, v := range m.Repos {
		if v.Path == path {
			m.Repos = append(m.Repos[:i], m.Repos[i+1:]...)
			return nil
		}
	}
	return os.ErrNotExist
}

func (m *MGConfig) AddRepo(path, remote string) error {
	for _, v := range m.Repos {
		if v.Path == path {
			return errAlreadyRegistered
		}
	}

	m.Repos = append(m.Repos, Repo{Path: path, Remote: remote})
	return nil
}

type Stats struct {
	Duplicates int
	NewPaths   []string
}

func (s Stats) String() string {
	str := ""
	for _, v := range s.NewPaths {
		str += "Added repo " + v + "\n"
	}
	str += "\nAdded " + fmt.Sprintf("%d", len(s.NewPaths)) + " new repos\n"
	str += "Skipped " + fmt.Sprintf("%d", s.Duplicates) + " duplicate repos"
	return str
}

func (m *MGConfig) Merge(m2 MGConfig) (Stats, error) {
	stats := Stats{}
	for _, v := range m2.Repos {
		err := m.AddRepo(v.Path, v.Remote)
		switch err {
		case errAlreadyRegistered:
			stats.Duplicates++
			continue
		case nil:
			stats.NewPaths = append(stats.NewPaths, v.Path)
			continue
		default:

			return stats, err
		}
	}
	return stats, nil
}

// LoadMGConfig loads the mgconfig file from the XDG_CONFIG_HOME directory
// or from the default location of $HOME/.config/mgconfig
// If the file is not found, an error is returned
func LoadMGConfig() (MGConfig, error) {
	mgConf := os.Getenv("MGCONFIG")
	if mgConf == "" {
		confDir := os.Getenv("XDG_CONFIG_HOME")
		if confDir == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return MGConfig{}, err
			}
			confDir = filepath.Join(home, ".config")
			if _, err := os.Stat(confDir); err != nil {
				return MGConfig{}, err
			}
		}
		mgConf = filepath.Join(confDir, "mgconfig")
	}
	file, err := os.ReadFile(mgConf)
	if err != nil {
		return MGConfig{}, err
	}
	return ParseMGConfig(file)
}

// ParseMGConfig parses the mgconfig file from a byte slice
func ParseMGConfig(b []byte) (MGConfig, error) {
	var config MGConfig
	err := json.Unmarshal(b, &config)
	return config, err
}

// ExpandPaths expands shell variables in all repo paths using os.ExpandEnv.
// This allows paths like $HOME/code or $GOPATH/src to work cross-platform.
func (m *MGConfig) ExpandPaths() {
	for i := range m.Repos {
		m.Repos[i].Path = os.ExpandEnv(m.Repos[i].Path)
	}
}

// CollapsePaths replaces the user's home directory with $HOME in all repo paths.
// This allows config files to be shared across machines with different home paths.
func (m *MGConfig) CollapsePaths() {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return
	}
	for i := range m.Repos {
		if strings.HasPrefix(m.Repos[i].Path, home) {
			m.Repos[i].Path = "$HOME" + m.Repos[i].Path[len(home):]
		}
	}
}

func (m MGConfig) Save() error {
	mgConf := os.Getenv("MGCONFIG")
	if mgConf == "" {
		confDir := os.Getenv("XDG_CONFIG_HOME")
		if confDir == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			confDir = filepath.Join(home, ".config")
			if _, err := os.Stat(confDir); err != nil {
				return err
			}
		}
		mgConf = filepath.Join(confDir, "mgconfig")
	}
	// Collapse paths before saving so config is portable
	toSave := MGConfig{
		Repos:   make([]Repo, len(m.Repos)),
		Aliases: m.Aliases,
	}
	copy(toSave.Repos, m.Repos)
	toSave.CollapsePaths()

	b, err := json.MarshalIndent(toSave, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(mgConf, b, 0o644)
}
