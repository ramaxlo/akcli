# CLAUDE.md — Project Notes for akcli

## Project overview

`akcli` is a Go CLI tool (`ak`) for automating Yocto build environment setup and image builds. It wraps Google `repo`, `oe-init-build-env`, and `bitbake` into a simple subcommand workflow.

## Key decisions

- **Module path**: `github.com/ramaxlo/akcli`
- **CLI framework**: `github.com/spf13/cobra`
- **TOML parser**: `github.com/BurntSushi/toml`
- **Config caching**: After `ak yocto init`, the parsed config is cached to `.ak/config.cache.toml`. All subsequent subcommands (sync, setup, build, pack) load from the cache rather than the original config file.
- **`template_conf` is required** for `ak yocto setup`. Return an error if it is missing — do not treat it as optional.
- **`DL_DIR` and `SSTATE_DIR`** default to `downloads/` and `sstate-cache/` at the project root (siblings to the build directory), and are appended to `local.conf` after `oe-init-build-env` creates it.
- **`MACHINE`** is set as an inline env variable when invoking `bitbake` (e.g. `MACHINE=qemux86-64 bitbake ...`), not only via `local.conf`.
- **Version**: `v1.0.0`, injected at build time via `-ldflags`.

## Patterns

- `dryRun` and `cfgFile` are global vars declared in `cmd/root.go` and used across all subcommands.
- `findPokyDir()` scans a list of common locations for `oe-init-build-env`, then falls back to a one-level-deep scan of the current directory.
- `findRepo()` checks `.ak/bin/repo` (local download) first, then falls back to `PATH`.
- Shell commands (sourcing `oe-init-build-env`, running `repo`, running `bitbake`) are all executed via `bash -c "<cmd>"`.

## Gotchas

### `ak yocto build` — `DisableFlagParsing`
The build command uses `buildCmd.DisableFlagParsing = true` so that unknown flags (bitbake flags like `--runall`, `-c`) are passed through to bitbake without cobra erroring. The consequence is that cobra is unaware of `--dryrun` and `--target`, so:
- These flags must be parsed manually in `runBuild`.
- `cmd.Help()` does not show them — a fully custom help block must be printed instead.

### `tar --transform` modifies symlink targets
By default, GNU tar's `--transform` applies to all names including symlink targets, which breaks relative symlinks in the Yocto deploy directory. Use `flags=rsh` to restrict the transform to regular files, symlink path entries, and hard links only:
```
--transform=flags=rsh;s,^\.,<machine>,
```

### `make` does not inherit shell `PATH`
`make` runs with a restricted `PATH` that may not include `/usr/local/go/bin`. Setting an inline environment variable (e.g. `GOOS=linux go build ...`) causes make to invoke the command through the shell, which resolves the full `PATH` correctly.

### Sourcing `oe-init-build-env` cannot modify the parent process environment
A child process cannot export environment variables back to its parent. `ak yocto setup` therefore only creates the build directory and config files — it does not attempt to set up the caller's shell environment.
