package cmd

import (
	"fmt"
	"log"
	"sort"

	git "github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"

	"github.com/taigrr/mg/parse"
)

type repoDiff struct {
	Path    string
	Changes []string
}

type diffResult struct {
	diff    repoDiff
	err     error
	repo    string
	changed bool
}

// diffCmd represents the diff command
var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "show uncommitted changes across all repos",
	Run: func(_ *cobra.Command, args []string) {
		repos := commandRepos(args)
		results := runPool(repos, func(repo parse.Repo) diffResult {
			path := repo.Path
			r, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{DetectDotGit: true})
			if err != nil {
				return diffResult{repo: path, err: err}
			}
			w, err := r.Worktree()
			if err != nil {
				return diffResult{repo: path, err: err}
			}
			st, err := w.Status()
			if err != nil {
				return diffResult{repo: path, err: err}
			}
			if st.IsClean() {
				return diffResult{repo: path}
			}
			rd := repoDiff{Path: path}
			for file, status := range st {
				var prefix string
				switch changeCode(status) {
				case git.Modified:
					prefix = "M"
				case git.Added:
					prefix = "A"
				case git.Deleted:
					prefix = "D"
				case git.Renamed:
					prefix = "R"
				case git.Copied:
					prefix = "C"
				case git.Untracked:
					prefix = "?"
				default:
					continue
				}
				rd.Changes = append(rd.Changes, fmt.Sprintf("  %s %s", prefix, file))
			}
			sort.Strings(rd.Changes)
			return diffResult{repo: path, diff: rd, changed: true}
		})

		var (
			errs  []diffResult
			diffs []repoDiff
		)
		for _, res := range results {
			if res.err != nil {
				errs = append(errs, res)
				continue
			}
			if res.changed {
				diffs = append(diffs, res.diff)
			}
		}

		sort.Slice(diffs, func(i, j int) bool {
			return diffs[i].Path < diffs[j].Path
		})

		for _, rd := range diffs {
			fmt.Printf("%s:\n", rd.Path)
			for _, change := range rd.Changes {
				fmt.Println(change)
			}
			fmt.Println()
		}

		for _, res := range errs {
			log.Printf("error reading %s: %s\n", res.repo, res.err)
		}

		fmt.Printf("%d/%d repos have changes\n", len(diffs), len(repos))
		if len(errs) > 0 {
			fmt.Printf("failed to read %d/%d repos\n", len(errs), len(repos))
		}
	},
}

func init() {
	RootCmd.AddCommand(diffCmd)
	diffCmd.Flags().IntVarP(&jobs, "jobs", "j", 1, "number of jobs to run in parallel")
}
