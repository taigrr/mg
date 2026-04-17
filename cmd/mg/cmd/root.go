package cmd

import (
	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "mg",
	Short: "go replacement for myrepos which only supports git repos",
	Long:  "mg is a go replacement for myrepos which only supports git repos.",
}
