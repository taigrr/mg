package cmd

import (
	"fmt"
	"log"
	"os"
	"sync"

	git "github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"

	"github.com/taigrr/mg/parse"
)

// pullCmd represents the pull command
var (
	cloneCmd = &cobra.Command{
		Use:   "clone",
		Short: "ensure all repos defined in the config are cloned",
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
			repoChan := make(chan parse.Repo, len(conf.Repos))
			errs := []RepoError{}
			alreadyUpToDate := 0
			mutex := sync.Mutex{}
			wg := sync.WaitGroup{}
			wg.Add(len(conf.Repos))
			for i := 0; i < jobs; i++ {
				go func() {
					for repo := range repoChan {
						_, err := git.PlainOpenWithOptions(repo.Remote, &(git.PlainOpenOptions{DetectDotGit: true}))
						if err == nil {
							log.Printf("already cloned: %s\n", repo.Path)
						} else if err == git.ErrRepositoryNotExists {
							log.Printf("attempting clone: %s\n", repo)
							_, err := git.PlainClone(repo.Path, false, &git.CloneOptions{
								URL: repo.Remote,
							})
							if err != nil {
								mutex.Lock()
								errs = append(errs, RepoError{Error: err, Repo: repo.Path})
								mutex.Unlock()
								log.Printf("clone failed for %s: %v\n", repo.Path, err)
								wg.Done()
								continue
							}
							fmt.Printf("successfully cloned %s\n", repo)
							wg.Done()
							continue
						} else {
							mutex.Lock()
							errs = append(errs, RepoError{Error: err, Repo: repo.Path})
							mutex.Unlock()
							log.Printf("clone failed for %s: %v\n", repo, err)
							wg.Done()
							continue
						}
					}
				}()
			}
			for _, repo := range conf.Repos {
				repoChan <- repo
			}
			close(repoChan)
			wg.Wait()
			for _, err := range errs {
				log.Printf("error pulling %s: %s\n", err.Repo, err.Error)
			}
			lenErrs := len(errs)
			fmt.Println()
			fmt.Printf("successfully pulled %d/%d repos\n", len(conf.Repos)-lenErrs, len(conf.Repos))
			fmt.Printf("%d repos already up to date\n", alreadyUpToDate)
			fmt.Printf("failed to pull %d/%d repos\n", lenErrs, len(conf.Repos))
		},
	}
)

func init() {
	rootCmd.AddCommand(pullCmd)
	pullCmd.Flags().IntVarP(&jobs, "jobs", "j", 1, "number of jobs to run in parallel")
}
