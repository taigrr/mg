package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// unregisterCmd represents the unregister command
var unregisterCmd = &cobra.Command{
	Use:   "unregister",
	Short: "remove the current repo from the config",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("unregister called")
	},
}

func init() {
	rootCmd.AddCommand(unregisterCmd)
}
