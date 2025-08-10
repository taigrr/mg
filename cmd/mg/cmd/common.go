package cmd

import (
	"log"
	"os"
	"strings"

	"github.com/taigrr/mg/parse"
)

func GetConfig() parse.MGConfig {
	conf, err := parse.LoadMGConfig()
	if err != nil {
		if os.IsNotExist(err) {
			// Try to load mr config instead
			mrconf, err := parse.LoadMRConfig()
			if err != nil {
				log.Println(err)
				os.Exit(1)
			}

			conf = mrconf.ToMGConfig()
			log.Println("migrated mrconfig to mgconfig")
			err = conf.Save()
			if err != nil {
				log.Println(err)
				os.Exit(1)
			}
		}
	}
	homeDir, _ := os.UserHomeDir()
	for i, repo := range conf.Repos {
		if strings.HasPrefix(repo.Path, "$HOME") {
			conf.Repos[i].Path = strings.Replace(repo.Path, "$HOME", homeDir, 1)
		}
	}
	return conf
}
