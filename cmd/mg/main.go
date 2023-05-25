package main

import "github.com/taigrr/mg/cmd/mg/cmd"

func main() {
	// conf, err := parse.LoadMGConfig()
	// if err != nil {
	//	if os.IsNotExist(err) {
	//		// Try to load mr config instead
	//		mrconf, err := parse.LoadMRConfig()
	//		if err != nil {
	//			log.Println(err)
	//			os.Exit(1)
	//		}

	//		conf = mrconf.ToMGConfig()
	//		log.Println("migrated mrconfig to mgconfig")
	//		err = conf.Save()
	//		if err != nil {
	//			log.Println(err)
	//			os.Exit(1)
	//		}
	//	}
	//}

	// fmt.Println(conf)
	cmd.Execute()
}

func Register() {
}
