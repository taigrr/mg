# mg

Golang replacement for [myrepos](https://myrepos.branchable.com/) which only supports git repos.

This app will support the following subcommands:

- mg commit
- mg push
- mg status
- mg diff
- mg pull
- mg fetch
- mg register
- mg unregister

Passing the `-jX` argument will spin up X jobs simultaneously

mg supports loading an existing ~/.mrconfig and migrating it to ~/.config/mg.conf, provided no mg.conf file exists.


## Improvements over mr:
1. No external dependencies (even git!*) 
1. More output options (summary of failures)
1. More deterministic behavior (global vs local run, register from git project subdir)
1. Exports public functions and can be embedded into other Go programs idiomatically


## Why to stick with mr:
1. If you need support for non-git VCS tooling
1. If you want to use the [mr plugin ecosystem](https://myrepos.branchable.com/#:~:text=repos%20to%20myrepos-,related%20software,-garden%3A%20manage%20git)

*: custom-registered commands may rely on external applications.
