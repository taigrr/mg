package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	git "github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"

	"github.com/taigrr/mg/parse"
)

// cloneCmd represents the clone command
var cloneCmd = &cobra.Command{
	Use:   "clone",
	Short: "ensure all repos defined in the config are cloned",
	Run: func(_ *cobra.Command, args []string) {
		repos := commandRepos(args)
		outcomes := runPool(repos, func(repo parse.Repo) repoOutcome {
			_, err := git.PlainOpenWithOptions(repo.Path, &git.PlainOpenOptions{DetectDotGit: true})
			switch err {
			case nil:
				log.Printf("already cloned: %s\n", repo.Path)
				return repoOutcome{repo: repo.Path, skip: true}
			case git.ErrRepositoryNotExists:
				log.Printf("attempting clone: %s\n", repo.Path)
				parentPath := filepath.Dir(repo.Path)
				if _, err := os.Stat(parentPath); err != nil {
					os.MkdirAll(parentPath, os.ModeDir|os.ModePerm)
				}
				if _, err := git.PlainClone(repo.Path, false, &git.CloneOptions{URL: repo.Remote}); err != nil {
					log.Printf("clone failed for %s: %v\n", repo.Path, err)
					return repoOutcome{repo: repo.Path, err: err}
				}
				fmt.Printf("successfully cloned %s\n", repo.Path)
				return repoOutcome{repo: repo.Path}
			default:
				log.Printf("clone failed for %s: %v\n", repo.Path, err)
				return repoOutcome{repo: repo.Path, err: err}
			}
		})
		reportOutcomes(outcomes, "cloned", "cloning", "clone", "already cloned", true)
	},
}

func init() {
	RootCmd.AddCommand(cloneCmd)
	cloneCmd.Flags().IntVarP(&jobs, "jobs", "j", 1, "number of jobs to run in parallel")
}
