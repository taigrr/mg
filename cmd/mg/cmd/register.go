package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "add current path to list of repos",
	Run: func(cmd *cobra.Command, args []string) {
		path, err := os.Getwd()
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		if len(args) == 1 {
			path = args[0]
		}
		fmt.Printf("register called for %s\n", path)
	},
}

func init() {
	rootCmd.AddCommand(registerCmd)
}
