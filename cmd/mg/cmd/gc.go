package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"

	"github.com/spf13/cobra"
)

var gcAggressive bool

var gcCmd = &cobra.Command{
	Use:   "gc",
	Short: "run git gc on all git repos",
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
			mutex sync.Mutex
			wg    sync.WaitGroup
		)
		wg.Add(len(conf.Repos))
		for i := 0; i < jobs; i++ {
			go func() {
				for repo := range repoChan {
					log.Printf("attempting gc: %s\n", repo)
					gcArgs := []string{"gc"}
					if gcAggressive {
						gcArgs = append(gcArgs, "--aggressive")
					}
					cmd := exec.Command("git", gcArgs...)
					cmd.Dir = repo
					out, err := cmd.CombinedOutput()
					if err != nil {
						mutex.Lock()
						errs = append(errs, RepoError{Error: fmt.Errorf("%v: %s", err, out), Repo: repo})
						mutex.Unlock()
						log.Printf("gc failed for %s: %v\n", repo, err)
						wg.Done()
						continue
					}
					fmt.Printf("successfully ran gc on %s\n", repo)
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
			log.Printf("error running gc on %s: %s\n", err.Repo, err.Error)
		}
		lenErrs := len(errs)
		fmt.Println()
		fmt.Printf("successfully ran gc on %d/%d repos\n", len(conf.Repos)-lenErrs, len(conf.Repos))
		fmt.Printf("failed to run gc on %d/%d repos\n", lenErrs, len(conf.Repos))
	},
}

func init() {
	RootCmd.AddCommand(gcCmd)
	gcCmd.Flags().IntVarP(&jobs, "jobs", "j", 1, "number of jobs to run in parallel")
	gcCmd.Flags().BoolVar(&gcAggressive, "aggressive", false, "run git gc with --aggressive")
}
