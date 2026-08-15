package cmd

import (
	"fmt"
	"log"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/spf13/cobra"

	"github.com/taigrr/mg/parse"
)

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "push all git repos",
	Run: func(_ *cobra.Command, args []string) {
		repos := commandRepos(args)
		outcomes := runPool(repos, func(repo parse.Repo) repoOutcome {
			path := repo.Path
			log.Printf("attempting push: %s\n", path)
			r, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{DetectDotGit: true})
			if err != nil {
				log.Printf("push failed for %s: %v\n", path, err)
				return repoOutcome{repo: path, err: err}
			}
			err = r.Push(&git.PushOptions{
				RefSpecs: []config.RefSpec{"refs/heads/*:refs/heads/*"},
			})
			switch {
			case err == git.NoErrAlreadyUpToDate:
				fmt.Printf("repo %s: already up to date\n", path)
				return repoOutcome{repo: path, skip: true}
			case err != nil:
				log.Printf("push failed for %s: %v\n", path, err)
				return repoOutcome{repo: path, err: err}
			default:
				fmt.Printf("successfully pushed %s\n", path)
				return repoOutcome{repo: path}
			}
		})
		reportOutcomes(outcomes, "pushed", "pushing", "push", "already up to date", true)
	},
}

func init() {
	RootCmd.AddCommand(pushCmd)
	pushCmd.Flags().IntVarP(&jobs, "jobs", "j", 1, "number of jobs to run in parallel")
}
