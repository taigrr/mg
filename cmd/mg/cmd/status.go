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

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "get the combined git status for all git repos",
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
			errs     []RepoError
			statuses []repoStatus
			mutex    sync.Mutex
			wg       sync.WaitGroup
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
					rs := repoStatus{Path: repo, Clean: st.IsClean()}
					for _, s := range st {
						code := s.Worktree
						if code == git.Unmodified {
							code = s.Staging
						}
						switch code {
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
					mutex.Lock()
					statuses = append(statuses, rs)
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

		for _, err := range errs {
			log.Printf("error reading %s: %s\n", err.Repo, err.Error)
		}

		fmt.Println()
		fmt.Printf("%d/%d repos have uncommitted changes\n", dirtyCount, len(conf.Repos))
		if len(errs) > 0 {
			fmt.Printf("failed to read %d/%d repos\n", len(errs), len(conf.Repos))
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.Flags().IntVarP(&jobs, "jobs", "j", 1, "number of jobs to run in parallel")
}
