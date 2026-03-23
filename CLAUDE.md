# CLAUDE.md — Project Notes for akcli

## Project overview

`akcli` is a Go CLI tool (`ak`) for automating Yocto build environment setup and image builds, and Linux kernel builds. It wraps Google `repo`, `oe-init-build-env`, `bitbake`, `git`, and `make` into a simple subcommand workflow.

## Key decisions

- **Module path**: `github.com/ramaxlo/akcli`
- **CLI framework**: `github.com/spf13/cobra`
- **TOML parser**: `github.com/BurntSushi/toml`
- **Config caching**: After `ak yocto init`, the parsed config is cached to `.ak/config.cache.toml`. All subsequent subcommands (sync, setup, build, pack) load from the cache rather than the original config file.
- **`template_conf` is required** for `ak yocto setup`. Return an error if it is missing — do not treat it as optional.
- **`DL_DIR` and `SSTATE_DIR`** default to `downloads/` and `sstate-cache/` at the project root (siblings to the build directory), and are appended to `local.conf` after `oe-init-build-env` creates it.
- **`MACHINE`** is set as an inline env variable when invoking `bitbake` (e.g. `MACHINE=qemux86-64 bitbake ...`), not only via `local.conf`.
- **Version**: `v1.0.0`, injected at build time via `-ldflags`.
- **`branch` and `tag` are mutually exclusive** in `[manifest]`. `config.Load()` returns an error if both are set. When `tag` is set, it is passed to `repo init` as `refs/tags/<tag>` via the `-b` flag.
- **Kernel config is cached separately**: After `ak kernel init`, the config is cached to `.ak/kernel.cache.toml` (distinct from `.ak/config.cache.toml` used by yocto). `KernelCache()` and `LoadKernelCache()` in `config/config.go` mirror the yocto `Cache()` / `LoadCache()` pattern.
- **Kernel config sections are distinct from yocto**: `[kernel]`, `[[kernel.remote]]`, and `[[kernel.defconfig]]` are separate from `[manifest]` and `[build]`. Both can coexist in the same `ak.toml`.
- **Kernel defconfigs are full `.config` files**, not defconfig names. Each `[[kernel.defconfig]]` entry has a `name` (used as the output dir under `kbuild/`) and a `config` path to the kernel config file.
- **Multiple kernel remotes, single local repo**: `ak kernel init` creates one local git repo and adds each `[[kernel.remote]]` as a named remote. The remote with `checkout = true` (last one wins if multiple) is checked out into detached HEAD.
- **`ARCH`, `CROSS_COMPILE`, and `O` are make variable assignments**, not shell env vars. They are passed directly on the `make` command line (e.g. `make -C kernel ARCH=arm64 CROSS_COMPILE=... O=...`), which is the idiomatic approach — no `bash -c` needed for kernel builds.
- **Per-defconfig output dirs**: `ak kernel build` uses `kbuild/<name>/` as the `O=` output directory for each defconfig, keeping build artifacts separated.

## Patterns

- `dryRun` and `cfgFile` are global vars declared in `cmd/root.go` and used across all subcommands.
- `findPokyDir()` scans a list of common locations for `oe-init-build-env`, then falls back to a one-level-deep scan of the current directory.
- `findRepo()` checks `.ak/bin/repo` (local download) first, then falls back to `PATH`.
- Shell commands (sourcing `oe-init-build-env`, running `repo`, running `bitbake`) are all executed via `bash -c "<cmd>"`.
- `runCmd(name string, args ...string)` in `cmd/kernel.go` is a shared helper for the kernel subcommands: prints the command then runs it with `Stdout`/`Stderr` wired to `os.Stdout`/`os.Stderr`. Use this instead of inlining `exec.Command` when the same pattern repeats across multiple kernel files.
- Each subcommand group (`yocto`, `kernel`) has its own `startTime` variable (`startTime` vs `kernelStartTime`) to avoid collision at the package level.

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

### `ak kernel build` — `olddefconfig` before building
After copying a `.config` file into the output dir, always run `make olddefconfig` before the actual build targets. This normalizes the config (fills in missing symbols with defaults) and prevents build failures on mismatched kernel versions.

### `git checkout <remote>/<branch>` produces detached HEAD
This is intentional for a read-only kernel source tree. Do not attempt to create a local tracking branch — detached HEAD is correct here.

### `git remote add` fails if remote already exists
Re-running `ak kernel init` on an existing `src_dir` will fail at `git remote add`. The user must delete `src_dir` and retry. This mirrors the `ak yocto init` / `repo init` behavior.

### Kernel `O=` path must be absolute
`make` with `O=` requires an absolute path for the output directory. Always use `filepath.Abs("kbuild/" + name)` before passing it to `make`.
