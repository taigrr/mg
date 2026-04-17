package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	git "github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"

	"github.com/taigrr/mg/parse"
)

// cloneCmd represents the clone command
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
			alreadyCloned := 0
			mutex := sync.Mutex{}
			wg := sync.WaitGroup{}
			wg.Add(len(conf.Repos))
			cloneFunc := func() {
				for repo := range repoChan {
					_, err := git.PlainOpenWithOptions(repo.Path, &(git.PlainOpenOptions{DetectDotGit: true}))
					if err == nil {
						log.Printf("already cloned: %s\n", repo.Path)
						mutex.Lock()
						alreadyCloned++
						mutex.Unlock()
						wg.Done()
						continue
					} else if err == git.ErrRepositoryNotExists {
						log.Printf("attempting clone: %s\n", repo.Path)
						parentPath := filepath.Dir(repo.Path)
						if _, err := os.Stat(parentPath); err != nil {
							os.MkdirAll(parentPath, os.ModeDir|os.ModePerm)
						}
						_, err = git.PlainClone(repo.Path, false, &git.CloneOptions{
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
						fmt.Printf("successfully cloned %s\n", repo.Path)
						wg.Done()
						continue
					} else {
						mutex.Lock()
						errs = append(errs, RepoError{Error: err, Repo: repo.Path})
						mutex.Unlock()
						log.Printf("clone failed for %s: %v\n", repo.Path, err)
						wg.Done()
						continue
					}
				}
			}
			for i := 0; i < jobs; i++ {
				go cloneFunc()
			}
			fmt.Println(len(conf.Repos))
			for _, repo := range conf.Repos {
				repoChan <- repo
			}
			close(repoChan)
			fmt.Println("waiting...")
			wg.Wait()
			for _, err := range errs {
				log.Printf("error pulling %s: %s\n", err.Repo, err.Error)
			}
			lenErrs := len(errs)
			fmt.Println()
			fmt.Printf("successfully cloned %d/%d repos\n", len(conf.Repos)-lenErrs, len(conf.Repos))
			fmt.Printf("%d repos already cloned\n", alreadyCloned)
			fmt.Printf("failed to clone %d/%d repos\n", lenErrs, len(conf.Repos))
		},
	}
)

func init() {
	rootCmd.AddCommand(cloneCmd)
	cloneCmd.Flags().IntVarP(&jobs, "jobs", "j", 1, "number of jobs to run in parallel")
}
