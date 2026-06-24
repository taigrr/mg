package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/taigrr/mg/parse"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "inspect mg configuration",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		configPath, err := parse.MGConfigPath()
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(cmd.OutOrStdout(), configPath)
		return err
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "print the active mg configuration",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		conf := GetConfig()
		configJSON, err := json.MarshalIndent(conf, "", "  ")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(cmd.OutOrStdout(), string(configJSON))
		return err
	},
}

func init() {
	RootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
}
