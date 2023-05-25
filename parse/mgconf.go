package parse

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

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

func (m *MGConfig) AddRepo(path, remote string) {
	m.Repos = append(m.Repos, Repo{Path: path, Remote: remote})
}

// LoadMGConfig loads the mgconfig file from the XDG_CONFIG_HOME directory
// or from the default location of $HOME/.config/mgconfig
// If the file is not found, an error is returned
func LoadMGConfig() (MGConfig, error) {
	var config MGConfig
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
	err = json.Unmarshal(file, &config)

	return config, err
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
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(mgConf, b, 0o644)
}
