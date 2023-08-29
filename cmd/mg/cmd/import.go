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
		if args[0] == "-" {
			f, err := io.ReadAll(os.Stdin)
			if err != nil {
				log.Println(err)
				os.Exit(1)
			}
			parsed, err := parse.ParseMGConfig(f)
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
		} else {
			f, err := os.ReadFile(args[0])
			if err != nil {
				log.Println(err)
				os.Exit(1)
			}
			parsed, err := parse.ParseMGConfig(f)
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
		}
		err := conf.Save()
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(importCmd)
}
