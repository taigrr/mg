package main

import (
	"flag"
	"log"
	"os"

	"github.com/taigrr/mg/parse"
)

var jobs = flag.Int("j", 1, "number of jobs to run in parallel")

func main() {
	flag.Parse()
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

	// fmt.Println(conf)
}

func Register() {
}
