package cmd

import (
	"log"
	"os"

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
	return conf
}
