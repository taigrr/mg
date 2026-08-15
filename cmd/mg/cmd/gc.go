package cmd

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/taigrr/mg/parse"
)

var gcAggressive bool

var gcCmd = &cobra.Command{
	Use:   "gc",
	Short: "run git gc on all git repos",
	Run: func(_ *cobra.Command, args []string) {
		repos := commandRepos(args)
		outcomes := runPool(repos, func(repo parse.Repo) repoOutcome {
			path := repo.Path
			log.Printf("attempting gc: %s\n", path)
			gcArgs := []string{"gc"}
			if gcAggressive {
				gcArgs = append(gcArgs, "--aggressive")
			}
			cmd := exec.Command("git", gcArgs...)
			cmd.Dir = path
			out, err := cmd.CombinedOutput()
			if err != nil {
				log.Printf("gc failed for %s: %v\n", path, err)
				return repoOutcome{repo: path, err: fmt.Errorf("%v: %s", err, out)}
			}
			fmt.Printf("successfully ran gc on %s\n", path)
			return repoOutcome{repo: path}
		})
		reportOutcomes(outcomes, "ran gc on", "running gc on", "run gc on", "", true)
	},
}

func init() {
	RootCmd.AddCommand(gcCmd)
	gcCmd.Flags().IntVarP(&jobs, "jobs", "j", 1, "number of jobs to run in parallel")
	gcCmd.Flags().BoolVar(&gcAggressive, "aggressive", false, "run git gc with --aggressive")
}
