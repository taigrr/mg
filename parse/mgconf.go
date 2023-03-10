package parse

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type MGConfig struct {
	Repos   []Repo
	Aliases map[string]string
}

func (m MGConfig) GetRepoPaths() []string {
	paths := []string{}
	for _, r := range m.Repos {
		paths = append(paths, r.Path)
	}
	return paths
}

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
