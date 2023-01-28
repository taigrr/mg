package main

import (
	"fmt"

	"github.com/taigrr/mg/parse"
)

func main() {
	mrconf, err := parse.LoadMRConfig()
	if err != nil {
		panic(err)
	}
	for _, path := range mrconf.GetRepoPaths() {
		fmt.Println(path)
	}
}
