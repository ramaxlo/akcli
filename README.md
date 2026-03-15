# akcli

A CLI tool for automating Yocto build environment setup and image builds.

## Requirements

- **Go** 1.25 or later
- **make**
- **Python 3** (for running Google repo tool)
- **GNU tar** (for `ak yocto pack`)
- **bash** (for sourcing `oe-init-build-env`)

## Building

```sh
make
```

The binary is written to `bin/ak`.

To remove the built binary:

```sh
make clean
```

## Installation

Copy the binary to a directory in your `PATH`, for example:

```sh
cp bin/ak ~/.local/bin/
```

Or system-wide:

```sh
sudo cp bin/ak /usr/local/bin/
```

## Configuration

`akcli` reads a TOML config file (default: `ak.toml` in the current directory).
Use the `-c` / `--config` flag to specify an alternative path.

### Config file format

```toml
[manifest]
url    = "https://github.com/example/yocto-manifest.git"
branch = "main"

[build]
distro   = "poky"
machine  = "qemux86-64"
target   = "core-image-minimal"

# Optional fields:

# Directory name for the Yocto build folder (default: "build")
# build_dir = "build"

# Path to a custom template directory for oe-init-build-env.
# Sets the TEMPLATECONF environment variable before sourcing the init script.
template_conf = "meta-custom/conf/templates/default"

# Directories for download cache and shared state cache.
# Defaults to "downloads/" and "sstate-cache/" at the project root.
# dl_dir     = "/shared/downloads"
# sstate_dir = "/shared/sstate-cache"
```

| Field                  | Required | Description                                      |
|------------------------|----------|--------------------------------------------------|
| `manifest.url`         | Yes      | URL of the repo manifest repository              |
| `manifest.branch`      | Yes      | Branch of the manifest repository                |
| `build.distro`         | Yes      | Yocto distro (e.g. `poky`)                       |
| `build.machine`        | Yes      | Target machine (e.g. `qemux86-64`)               |
| `build.target`         | Yes      | Default bitbake build target                     |
| `build.build_dir`      | No       | Build directory name (default: `build`)          |
| `build.template_conf`  | Yes      | Path to conf template dir for `TEMPLATECONF`     |
| `build.dl_dir`         | No       | Download cache directory (default: `downloads/`) |
| `build.sstate_dir`     | No       | Shared state cache dir (default: `sstate-cache/`)|

## Usage

### Global flags

| Flag              | Short | Description                                    |
|-------------------|-------|------------------------------------------------|
| `--config <path>` | `-c`  | Path to config file (default: `ak.toml`)       |
| `--dryrun`        |       | Print commands that would run without executing|

### Workflow

The typical workflow follows these steps in order:

```
ak yocto init → ak yocto sync → ak yocto setup → ak yocto build
```

---

### `ak yocto init`

Reads the config file, downloads the Google `repo` tool if not already available,
and initializes the manifest repository.

```sh
ak yocto init
ak -c custom.toml yocto init
```

The config is cached to `.ak/config.cache.toml` for use by subsequent commands.

The `repo` tool is looked up in `PATH` first. If not found, it is downloaded from
Google and stored at `.ak/bin/repo`.

---

### `ak yocto sync`

Fetches all Yocto meta layers defined in the manifest using `repo sync`.

```sh
ak yocto sync
ak yocto sync -j 8
```

| Flag          | Short | Description                          |
|---------------|-------|--------------------------------------|
| `--jobs <n>`  | `-j`  | Number of parallel sync jobs (default: 4) |

---

### `ak yocto setup`

Creates the Yocto build directory by sourcing `oe-init-build-env`. The script is
located automatically by scanning for it under the current directory.

`template_conf` must be set in the config. `TEMPLATECONF` is exported to that
value before sourcing so the init script uses the specified template directory
to generate `local.conf` and `bblayers.conf`.

After the build directory is created, `DL_DIR` and `SSTATE_DIR` are appended to
`local.conf` to place the download and shared state caches outside the build
directory.

```sh
ak yocto setup
```

---

### `ak yocto build`

Sources `oe-init-build-env` and runs `bitbake` to build the image. `MACHINE` is
set from the config file. Any unrecognized flags are passed directly to bitbake.

```sh
ak yocto build
ak yocto build --target core-image-sato
ak yocto build --runall fetch
ak yocto build --target core-image-sato -c populate_sdk
```

| Flag              | Description                                              |
|-------------------|----------------------------------------------------------|
| `--target <name>` | Override the default build target from the config file   |
| `--dryrun`        | Print the shell command without executing it             |

---

### `ak yocto pack`

Packs the built image artifacts from `<build_dir>/tmp/deploy/images/<machine>`
into a `<machine>.tar.gz` file. All files inside the tarball are prefixed with
the machine name as a directory. Symlink targets within the archive are preserved
as-is.

```sh
ak yocto pack
```

The output file is written to the current directory as `<machine>.tar.gz`.
