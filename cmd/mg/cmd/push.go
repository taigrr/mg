package cmd

import (
	"fmt"
	"log"
	"os"
	"sync"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/spf13/cobra"
)

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "push all git repos",
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
			errs            []RepoError
			alreadyUpToDate int
			mutex           sync.Mutex
			wg              sync.WaitGroup
		)
		wg.Add(len(conf.Repos))
		for i := 0; i < jobs; i++ {
			go func() {
				for repo := range repoChan {
					log.Printf("attempting push: %s\n", repo)
					r, err := git.PlainOpenWithOptions(repo, &git.PlainOpenOptions{DetectDotGit: true})
					if err != nil {
						mutex.Lock()
						errs = append(errs, RepoError{Error: err, Repo: repo})
						mutex.Unlock()
						log.Printf("push failed for %s: %v\n", repo, err)
						wg.Done()
						continue
					}
					err = r.Push(&git.PushOptions{
						RefSpecs: []config.RefSpec{"refs/heads/*:refs/heads/*"},
					})
					if err == git.NoErrAlreadyUpToDate {
						mutex.Lock()
						alreadyUpToDate++
						mutex.Unlock()
						fmt.Printf("repo %s: already up to date\n", repo)
						wg.Done()
						continue
					} else if err != nil {
						mutex.Lock()
						errs = append(errs, RepoError{Error: err, Repo: repo})
						mutex.Unlock()
						log.Printf("push failed for %s: %v\n", repo, err)
						wg.Done()
						continue
					}
					fmt.Printf("successfully pushed %s\n", repo)
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
			log.Printf("error pushing %s: %s\n", err.Repo, err.Error)
		}
		lenErrs := len(errs)
		fmt.Println()
		fmt.Printf("successfully pushed %d/%d repos\n", len(conf.Repos)-lenErrs, len(conf.Repos))
		fmt.Printf("%d repos already up to date\n", alreadyUpToDate)
		fmt.Printf("failed to push %d/%d repos\n", lenErrs, len(conf.Repos))
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)
	pushCmd.Flags().IntVarP(&jobs, "jobs", "j", 1, "number of jobs to run in parallel")
}
