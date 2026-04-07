package cmd

import (
	"fmt"
	"log"
	"os"
	"sync"

	git "github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
)

var commitMessage string

// commitCmd represents the commit command
var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "commit staged changes across all repos with the same message",
	Run: func(_ *cobra.Command, args []string) {
		type RepoError struct {
			Error error
			Repo  string
		}
		if jobs < 1 {
			log.Println("jobs must be greater than 0")
			os.Exit(1)
		}
		if commitMessage == "" {
			log.Println("commit message is required (-m)")
			os.Exit(1)
		}
		conf := GetConfig()
		if len(args) > 0 {
			log.Println("too many arguments")
			os.Exit(1)
		}
		repoChan := make(chan string, len(conf.Repos))
		var (
			errs    []RepoError
			skipped int
			mutex   sync.Mutex
			wg      sync.WaitGroup
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
					// Check if there are any staged changes
					hasStagedChanges := false
					for _, s := range st {
						if s.Staging != git.Unmodified && s.Staging != git.Untracked {
							hasStagedChanges = true
							break
						}
					}
					if !hasStagedChanges {
						mutex.Lock()
						skipped++
						mutex.Unlock()
						fmt.Printf("repo %s: nothing staged to commit\n", repo)
						wg.Done()
						continue
					}
					_, err = w.Commit(commitMessage, &git.CommitOptions{})
					if err != nil {
						mutex.Lock()
						errs = append(errs, RepoError{Error: err, Repo: repo})
						mutex.Unlock()
						log.Printf("commit failed for %s: %v\n", repo, err)
						wg.Done()
						continue
					}
					fmt.Printf("successfully committed in %s\n", repo)
					wg.Done()
				}
			}()
		}
		for _, repo := range conf.Repos {
			repoChan <- repo.Path
		}
		close(repoChan)
		wg.Wait()
		for _, err := range errs {
			log.Printf("error committing %s: %s\n", err.Repo, err.Error)
		}
		lenErrs := len(errs)
		committed := len(conf.Repos) - lenErrs - skipped
		fmt.Println()
		fmt.Printf("successfully committed %d/%d repos\n", committed, len(conf.Repos))
		fmt.Printf("%d repos had nothing staged\n", skipped)
		fmt.Printf("failed to commit %d/%d repos\n", lenErrs, len(conf.Repos))
	},
}

func init() {
	rootCmd.AddCommand(commitCmd)
	commitCmd.Flags().IntVarP(&jobs, "jobs", "j", 1, "number of jobs to run in parallel")
	commitCmd.Flags().StringVarP(&commitMessage, "message", "m", "", "commit message")
}
