package cmd

import (
	"fmt"
	"log"
	"os"

	git "github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"

	"github.com/taigrr/mg/parse"
)

var commitMessage string

// commitCmd represents the commit command
var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "commit staged changes across all repos with the same message",
	Run: func(_ *cobra.Command, args []string) {
		if commitMessage == "" {
			log.Println("commit message is required (-m)")
			os.Exit(1)
		}
		repos := commandRepos(args)
		outcomes := runPool(repos, func(repo parse.Repo) repoOutcome {
			path := repo.Path
			r, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{DetectDotGit: true})
			if err != nil {
				return repoOutcome{repo: path, err: err}
			}
			w, err := r.Worktree()
			if err != nil {
				return repoOutcome{repo: path, err: err}
			}
			st, err := w.Status()
			if err != nil {
				return repoOutcome{repo: path, err: err}
			}
			if !hasStagedChanges(st) {
				fmt.Printf("repo %s: nothing staged to commit\n", path)
				return repoOutcome{repo: path, skip: true}
			}
			if _, err := w.Commit(commitMessage, &git.CommitOptions{}); err != nil {
				log.Printf("commit failed for %s: %v\n", path, err)
				return repoOutcome{repo: path, err: err}
			}
			fmt.Printf("successfully committed in %s\n", path)
			return repoOutcome{repo: path}
		})
		reportOutcomes(outcomes, "committed", "committing", "commit", "had nothing staged", false)
	},
}

// hasStagedChanges reports whether any file has changes staged in the index.
func hasStagedChanges(st git.Status) bool {
	for _, s := range st {
		if s.Staging != git.Unmodified && s.Staging != git.Untracked {
			return true
		}
	}
	return false
}

func init() {
	RootCmd.AddCommand(commitCmd)
	commitCmd.Flags().IntVarP(&jobs, "jobs", "j", 1, "number of jobs to run in parallel")
	commitCmd.Flags().StringVarP(&commitMessage, "message", "m", "", "commit message")
}
