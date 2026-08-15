package cmd

import (
	"fmt"
	"log"

	git "github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"

	"github.com/taigrr/mg/parse"
)

// pullCmd represents the pull command
var (
	jobs    int
	pullCmd = &cobra.Command{
		Use:   "pull",
		Short: "update all git repos specified in config",
		Run: func(_ *cobra.Command, args []string) {
			repos := commandRepos(args)
			outcomes := runPool(repos, func(repo parse.Repo) repoOutcome {
				path := repo.Path
				log.Printf("attempting pull: %s\n", path)
				r, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{DetectDotGit: true})
				if err != nil {
					log.Printf("pull failed for %s: %v\n", path, err)
					return repoOutcome{repo: path, err: err}
				}
				w, err := r.Worktree()
				if err != nil {
					log.Printf("pull failed for %s: %v\n", path, err)
					return repoOutcome{repo: path, err: err}
				}
				err = w.Pull(&git.PullOptions{})
				switch {
				case err == git.NoErrAlreadyUpToDate:
					fmt.Printf("repo %s: already up to date\n", path)
					return repoOutcome{repo: path, skip: true}
				case err != nil:
					log.Printf("pull failed for %s: %v\n", path, err)
					return repoOutcome{repo: path, err: err}
				default:
					fmt.Printf("successfully pulled %s\n", w.Filesystem.Root())
					return repoOutcome{repo: path}
				}
			})
			reportOutcomes(outcomes, "pulled", "pulling", "pull", "already up to date", true)
		},
	}
)

func init() {
	RootCmd.AddCommand(pullCmd)
	pullCmd.Flags().IntVarP(&jobs, "jobs", "j", 1, "number of jobs to run in parallel")
}
