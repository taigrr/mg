package cmd

import (
	"log"
	"os"

	git "github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
)

var registerCmd = &cobra.Command{
	Use:   "register",
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
		}
		r, err := git.PlainOpenWithOptions(path, &(git.PlainOpenOptions{DetectDotGit: true}))
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		remotes, err := r.Remotes()
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		if len(remotes) == 0 {
			log.Println("no remotes found")
			os.Exit(1)
		}
		remote := remotes[0]
		urls := remote.Config().URLs
		if len(urls) == 0 {
			log.Println("no urls found for remote")
			os.Exit(1)
		}
		url := urls[0]
		newPath, err := r.Worktree()
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
		path = newPath.Filesystem.Root()
		for _, v := range conf.Repos {
			if v.Path == path {
				log.Printf("repo %s already registered\n", path)
				os.Exit(1)
			}
		}
		conf.AddRepo(path, url)
		err = conf.Save()
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(registerCmd)
}
