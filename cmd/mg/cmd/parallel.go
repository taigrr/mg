package cmd

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/taigrr/mg/parse"
)

// commandRepos validates the flags and positional args shared by every
// repo-spanning command and returns the configured repos. It exits the
// process on invalid input, matching the previous per-command behavior.
func commandRepos(args []string) []parse.Repo {
	if jobs < 1 {
		log.Println("jobs must be greater than 0")
		os.Exit(1)
	}
	if len(args) > 0 {
		log.Println("too many arguments")
		os.Exit(1)
	}
	return GetConfig().Repos
}

// runPool applies fn to every repo across `jobs` workers and returns the
// results in the same order as repos. fn is invoked concurrently, so it must
// not touch shared state without synchronization; each result is written to a
// distinct slot, so no locking is required for the results themselves.
func runPool[T any](repos []parse.Repo, fn func(parse.Repo) T) []T {
	results := make([]T, len(repos))
	queue := make(chan int)
	var wg sync.WaitGroup
	for w := 0; w < jobs; w++ {
		wg.Go(func() {
			for i := range queue {
				results[i] = fn(repos[i])
			}
		})
	}
	for i := range repos {
		queue <- i
	}
	close(queue)
	wg.Wait()
	return results
}

// repoOutcome is the result of an action command (clone, pull, push, ...) for
// a single repo: a failure (err != nil), a skip (nothing to do), or a success.
type repoOutcome struct {
	repo string
	err  error
	skip bool
}

// reportOutcomes prints the standard "successfully / skipped / failed" summary
// shared by the action commands. verbs are the past ("pulled"), gerund
// ("pulling"), and base ("pull") forms used in the messages. skipLabel is the
// noun phrase for skipped repos (empty to omit the line); skipIsSuccess counts
// skipped repos toward the success total (true for "already up to date", false
// for "nothing staged").
func reportOutcomes(outcomes []repoOutcome, past, gerund, base, skipLabel string, skipIsSuccess bool) {
	var failed, skipped int
	for _, o := range outcomes {
		switch {
		case o.err != nil:
			failed++
			log.Printf("error %s %s: %s\n", gerund, o.repo, o.err)
		case o.skip:
			skipped++
		}
	}
	total := len(outcomes)
	success := total - failed
	if !skipIsSuccess {
		success -= skipped
	}
	fmt.Println()
	fmt.Printf("successfully %s %d/%d repos\n", past, success, total)
	if skipLabel != "" {
		fmt.Printf("%d repos %s\n", skipped, skipLabel)
	}
	fmt.Printf("failed to %s %d/%d repos\n", base, failed, total)
}
