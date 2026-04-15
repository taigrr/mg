package cmd

import (
	"fmt"
	"log"
	"os"
	"sort"
	"sync"

	git "github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
)

type repoDiff struct {
	Path    string
	Changes []string
}

// diffCmd represents the diff command
var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "show uncommitted changes across all repos",
	Run: func(_ *cobra.Command, args []string) {
		type RepoError struct {
			Error error
			Repo  string
		}
		if jobs < 1 {
			log.Println("jobs must be greater than 0")
			os.Exit(1)
		}
		conf := GetConfig()
		if len(args) > 0 {
			log.Println("too many arguments")
			os.Exit(1)
		}
		repoChan := make(chan string, len(conf.Repos))
		var (
			errs  []RepoError
			diffs []repoDiff
			mutex sync.Mutex
			wg    sync.WaitGroup
		)
		wg.Add(len(conf.Repos))
		for i := 0; i < jobs; i++ {
			go func() {
				for repo := range repoChan {
					r, err := git.PlainOpenWithOptions(repo, &git.PlainOpenOptions{DetectDotGit: true})
					if err != nil {
						mutex.Lock()
						errs = append(errs, RepoError{Error: err, Repo: repo})
						mutex.Unlock()
						wg.Done()
						continue
					}
					w, err := r.Worktree()
					if err != nil {
						mutex.Lock()
						errs = append(errs, RepoError{Error: err, Repo: repo})
						mutex.Unlock()
						wg.Done()
						continue
					}
					st, err := w.Status()
					if err != nil {
						mutex.Lock()
						errs = append(errs, RepoError{Error: err, Repo: repo})
						mutex.Unlock()
						wg.Done()
						continue
					}
					if st.IsClean() {
						wg.Done()
						continue
					}
					rd := repoDiff{Path: repo}
					for file, status := range st {
						code := status.Worktree
						if code == git.Unmodified {
							code = status.Staging
						}
						var prefix string
						switch code {
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
					mutex.Lock()
					diffs = append(diffs, rd)
					mutex.Unlock()
					wg.Done()
				}
			}()
		}
		for _, repo := range conf.Repos {
			repoChan <- repo.Path
		}
		close(repoChan)
		wg.Wait()

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

		for _, err := range errs {
			log.Printf("error reading %s: %s\n", err.Repo, err.Error)
		}

		fmt.Printf("%d/%d repos have changes\n", len(diffs), len(conf.Repos))
		if len(errs) > 0 {
			fmt.Printf("failed to read %d/%d repos\n", len(errs), len(conf.Repos))
		}
	},
}

func init() {
	rootCmd.AddCommand(diffCmd)
	diffCmd.Flags().IntVarP(&jobs, "jobs", "j", 1, "number of jobs to run in parallel")
}
