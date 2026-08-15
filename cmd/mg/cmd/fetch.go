package cmd

import (
	"fmt"
	"log"

	git "github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"

	"github.com/taigrr/mg/parse"
)

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "fetch all git repos without merging",
	Run: func(_ *cobra.Command, args []string) {
		repos := commandRepos(args)
		outcomes := runPool(repos, func(repo parse.Repo) repoOutcome {
			path := repo.Path
			log.Printf("attempting fetch: %s\n", path)
			r, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{DetectDotGit: true})
			if err != nil {
				log.Printf("fetch failed for %s: %v\n", path, err)
				return repoOutcome{repo: path, err: err}
			}
			err = r.Fetch(&git.FetchOptions{})
			switch {
			case err == git.NoErrAlreadyUpToDate:
				fmt.Printf("repo %s: already up to date\n", path)
				return repoOutcome{repo: path, skip: true}
			case err != nil:
				log.Printf("fetch failed for %s: %v\n", path, err)
				return repoOutcome{repo: path, err: err}
			default:
				fmt.Printf("successfully fetched %s\n", path)
				return repoOutcome{repo: path}
			}
		})
		reportOutcomes(outcomes, "fetched", "fetching", "fetch", "already up to date", true)
	},
}

func init() {
	RootCmd.AddCommand(fetchCmd)
	fetchCmd.Flags().IntVarP(&jobs, "jobs", "j", 1, "number of jobs to run in parallel")
}
