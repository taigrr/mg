package cmd

import (
	"fmt"
	"log"
	"sort"

	git "github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"

	"github.com/taigrr/mg/parse"
)

type repoStatus struct {
	Path     string
	Modified int
	Added    int
	Deleted  int
	Renamed  int
	Copied   int
	Untrack  int
	Clean    bool
}

type statusResult struct {
	status repoStatus
	err    error
	repo   string
}

// changeCode returns the effective change code for a file, preferring the
// worktree state and falling back to the staging state when unmodified.
func changeCode(s *git.FileStatus) git.StatusCode {
	if s.Worktree == git.Unmodified {
		return s.Staging
	}
	return s.Worktree
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "get the combined git status for all git repos",
	Run: func(_ *cobra.Command, args []string) {
		repos := commandRepos(args)
		results := runPool(repos, func(repo parse.Repo) statusResult {
			path := repo.Path
			r, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{DetectDotGit: true})
			if err != nil {
				return statusResult{repo: path, err: err}
			}
			w, err := r.Worktree()
			if err != nil {
				return statusResult{repo: path, err: err}
			}
			st, err := w.Status()
			if err != nil {
				return statusResult{repo: path, err: err}
			}
			rs := repoStatus{Path: path, Clean: st.IsClean()}
			for _, s := range st {
				switch changeCode(s) {
				case git.Modified:
					rs.Modified++
				case git.Added:
					rs.Added++
				case git.Deleted:
					rs.Deleted++
				case git.Renamed:
					rs.Renamed++
				case git.Copied:
					rs.Copied++
				case git.Untracked:
					rs.Untrack++
				}
			}
			return statusResult{repo: path, status: rs}
		})

		var (
			errs     []statusResult
			statuses []repoStatus
		)
		for _, res := range results {
			if res.err != nil {
				errs = append(errs, res)
				continue
			}
			statuses = append(statuses, res.status)
		}

		sort.Slice(statuses, func(i, j int) bool {
			return statuses[i].Path < statuses[j].Path
		})

		dirtyCount := 0
		for _, rs := range statuses {
			if rs.Clean {
				continue
			}
			dirtyCount++
			fmt.Printf("%s:\n", rs.Path)
			if rs.Modified > 0 {
				fmt.Printf("  modified: %d\n", rs.Modified)
			}
			if rs.Added > 0 {
				fmt.Printf("  added:    %d\n", rs.Added)
			}
			if rs.Deleted > 0 {
				fmt.Printf("  deleted:  %d\n", rs.Deleted)
			}
			if rs.Renamed > 0 {
				fmt.Printf("  renamed:  %d\n", rs.Renamed)
			}
			if rs.Copied > 0 {
				fmt.Printf("  copied:   %d\n", rs.Copied)
			}
			if rs.Untrack > 0 {
				fmt.Printf("  untracked: %d\n", rs.Untrack)
			}
		}

		for _, res := range errs {
			log.Printf("error reading %s: %s\n", res.repo, res.err)
		}

		fmt.Println()
		fmt.Printf("%d/%d repos have uncommitted changes\n", dirtyCount, len(repos))
		if len(errs) > 0 {
			fmt.Printf("failed to read %d/%d repos\n", len(errs), len(repos))
		}
	},
}

func init() {
	RootCmd.AddCommand(statusCmd)
	statusCmd.Flags().IntVarP(&jobs, "jobs", "j", 1, "number of jobs to run in parallel")
}
