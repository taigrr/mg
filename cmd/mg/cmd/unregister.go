package cmd

import (
	"log"
	"os"

	git "github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
)

// unregisterCmd represents the unregister command
var unregisterCmd = &cobra.Command{
	Use:   "unregister",
	Short: "add current path to list of repos",
	Run: func(_ *cobra.Command, args []string) {
		conf := GetConfig()
		path, err := os.Getwd()
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		if len(args) == 1 {
			path = args[0]
			err = conf.DelRepo(path)
			if err != nil {
				log.Println(err)
				os.Exit(1)
			}
			return
		} else if len(args) > 1 {
			log.Println("too many arguments")
			os.Exit(1)
		}
		r, err := git.PlainOpenWithOptions(path, &(git.PlainOpenOptions{DetectDotGit: true}))
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		newPath, err := r.Worktree()
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		path = newPath.Filesystem.Root()
		err = conf.DelRepo(path)
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		err = conf.Save()
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(unregisterCmd)
}
