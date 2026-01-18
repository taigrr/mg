# AGENTS.md

Agent guide for the `mg` codebase - a Go replacement for [myrepos](https://myrepos.branchable.com/) that only supports git repos.

## Project Overview

`mg` is a CLI tool for managing multiple git repositories simultaneously. It uses `go-git/go-git` for pure Go git operations (no external git dependency required) and `spf13/cobra` for CLI structure.

### Key Features

- Parallel operations via `-j` flag
- Compatible with existing `~/.mrconfig` files (auto-migrates to `mgconfig`)
- Pure Go implementation - no external git binary needed
- Embeddable as a library

## Commands

```bash
# Build
go build ./...

# Run tests
go test ./...

# Install the binary
go install ./cmd/mg

# Run directly
go run ./cmd/mg <command>
```

## Project Structure

```
mg/
├── cmd/
│   ├── mg/
│   │   ├── main.go          # Entry point
│   │   └── cmd/             # Cobra commands
│   │       ├── root.go      # Root command setup
│   │       ├── common.go    # Shared utilities (GetConfig)
│   │       ├── clone.go     # Clone all repos (implemented)
│   │       ├── pull.go      # Pull all repos (implemented)
│   │       ├── register.go  # Register repo (implemented)
│   │       ├── unregister.go# Unregister repo (implemented)
│   │       ├── import.go    # Merge configs (implemented)
│   │       ├── push.go      # Stub
│   │       ├── fetch.go     # Stub
│   │       ├── status.go    # Stub
│   │       ├── diff.go      # Stub
│   │       ├── commit.go    # Stub
│   │       └── config.go    # Stub
│   └── paths/
│       └── mrpaths.go       # Utility to list repo paths from mrconfig
└── parse/
    ├── mgconf.go            # MGConfig: JSON-based config format
    ├── myrepos.go           # MRConfig: Parse ~/.mrconfig (INI-style)
    └── myrepos_test.go      # Tests (skeleton)
```

## Implementation Status

| Command      | Status      | Notes                              |
|--------------|-------------|------------------------------------|
| `clone`      | Implemented | Parallel via `-j`, creates dirs   |
| `pull`       | Implemented | Parallel via `-j`                  |
| `register`   | Implemented | Detects git root, stores `$HOME`  |
| `unregister` | Implemented | By path or current dir             |
| `import`     | Implemented | Merge configs, supports stdin `-` |
| `push`       | Stub        | Prints "push called"               |
| `fetch`      | Stub        | Prints "fetch called"              |
| `status`     | Stub        | Prints "status called"             |
| `diff`       | Stub        | Prints "diff called"               |
| `commit`     | Stub        | Prints "commit called"             |
| `config`     | Stub        | Prints "config called"             |

## Configuration

### Config File Location

1. `$MGCONFIG` environment variable (if set)
2. `$XDG_CONFIG_HOME/mgconfig`
3. `~/.config/mgconfig`

### Config Format (JSON)

```json
{
  "Repos": [
    {
      "Path": "$HOME/code/project",
      "Remote": "git@github.com:user/project.git"
    }
  ],
  "Aliases": {}
}
```

### Migration from myrepos

If no `mgconfig` exists but `~/.mrconfig` does, `mg` auto-migrates on first run. The `MRConfig.ToMGConfig()` method handles conversion.

## Code Patterns

### Adding a New Command

1. Create `cmd/mg/cmd/<name>.go`
2. Define a `cobra.Command` variable
3. Register in `init()` via `rootCmd.AddCommand()`
4. For parallel operations, follow the pattern in `clone.go` or `pull.go`:

```go
var myCmd = &cobra.Command{
    Use:   "mycommand",
    Short: "description",
    Run: func(_ *cobra.Command, args []string) {
        conf := GetConfig()  // Load config with fallback to mrconfig
        // Implementation...
    },
}

func init() {
    rootCmd.AddCommand(myCmd)
    myCmd.Flags().IntVarP(&jobs, "jobs", "j", 1, "number of parallel jobs")
}
```

### Parallel Execution Pattern

Used in `clone.go` and `pull.go`:

```go
repoChan := make(chan RepoType, len(repos))
wg := sync.WaitGroup{}
mutex := sync.Mutex{}
errs := []Error{}

wg.Add(len(repos))
for i := 0; i < jobs; i++ {
    go func() {
        for repo := range repoChan {
            // Do work
            // Use mutex for shared state (errs, counters)
            wg.Done()
        }
    }()
}

for _, repo := range repos {
    repoChan <- repo
}
close(repoChan)
wg.Wait()
```

### Git Operations

Use `go-git/go-git/v5`:

```go
import git "github.com/go-git/go-git/v5"

// Open repo (detects .git in parent dirs)
r, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{DetectDotGit: true})

// Clone
_, err = git.PlainClone(path, false, &git.CloneOptions{URL: remote})

// Pull
w, _ := r.Worktree()
err = w.Pull(&git.PullOptions{})
// Check: err == git.NoErrAlreadyUpToDate
```

### Path Handling

- Paths in config use `$HOME` prefix for portability
- `GetConfig()` in `common.go` expands `$HOME` to actual home directory at runtime
- `register` command stores paths with `$HOME` prefix

## Key Types

### `parse.MGConfig`

```go
type MGConfig struct {
    Repos   []Repo
    Aliases map[string]string
}

// Methods
func LoadMGConfig() (MGConfig, error)
func (m *MGConfig) AddRepo(path, remote string) error
func (m *MGConfig) DelRepo(path string) error
func (m *MGConfig) Merge(m2 MGConfig) (Stats, error)
func (m MGConfig) Save() error
```

### `parse.Repo`

```go
type Repo struct {
    Path    string
    Remote  string
    Aliases map[string]string `json:"aliases,omitempty"`
}
```

## Dependencies

- `github.com/go-git/go-git/v5` - Pure Go git implementation
- `github.com/spf13/cobra` - CLI framework

## Testing

Tests are minimal. Only `parse/myrepos_test.go` exists with a skeleton structure:

```bash
go test ./...
```

## Known Issues / TODOs

1. Several commands are stubs (push, fetch, status, diff, commit, config)
2. `parse/mgconf.go:61` has a hint about inefficient string concatenation in a loop
3. Test coverage is minimal
4. `unregister` command short description incorrectly says "add current path" (copy-paste error)

## Error Handling Pattern

Commands typically:
1. Log errors via `log.Println(err)`
2. Exit with `os.Exit(1)` on fatal errors
3. Collect errors during parallel operations and report summary at end
