package cmd

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/taigrr/mg/parse"
)

var importCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "merge a new mgconfig into the current one",
	Args:  cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		conf := GetConfig()
		var (
			data []byte
			err  error
		)
		if args[0] == "-" {
			data, err = io.ReadAll(os.Stdin)
		} else {
			data, err = os.ReadFile(args[0])
		}
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		parsed, err := parse.ParseMGConfig(data)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		stats, err := conf.Merge(parsed)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		fmt.Println(stats)
		if err := conf.Save(); err != nil {
			log.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(importCmd)
}
