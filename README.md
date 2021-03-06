# gcu (go check updates)

Due to following the [semantic version number](https://semver.org/), Go will not actively update the major version when checking and updating dependencies, and Go does not provide us with optional dependency update items. When updating, either all of them are updated or none of them are updated.

But this tool makes up for these deficiencies.

- It provides update checking for major versions (you can add the `--safe` flag to ignore major version checks)
- Visual update selection
- Colored version number distinguishing hints
- Automatically rewrite import paths (default)
- Support binary file upgrade written in go language (list display is currently not supported)

warning:

- Will only check directly dependent libraries
- If there is a mutual dependency between two directly dependent libraries, unless both libraries depend on each other's latest library, there will be strange behavior
- You need to ensure your own compatibility after updating major versions
- If the major version of the library is discontinuous, the latest version may not be available (e.g. 1.0.0 -> 3.1.0 without v2)
- Still Work In Progress

install:

```bash
go install github.com/qianxi0410/gcu@latest
```

usage:

```txt
> gcu help

NAME:
   gcu (go-check-updates) - check for updates in go.mod dependency and go's binary files

USAGE:
   gcu (go-check-updates) [global options] command [command options] [arguments...]

COMMANDS:
   list        List all direct dependencies available for update
   version, v  Print the version number of gcu
   help, h     Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --stable, -s   Only fetch stable version (default: true)
   --cached, -c   Use cached version if available (default: false)
   --all, -a      Upgrade all dependencies without asking (default: false)
   --rewrite, -w  Rewrite all dependencies to latest version in your project (default: true)
   --safe         Only minor and patch releases are checked and updated (default: false)
   --size value   Number of items to show in the select list (default: 10)
   --tidy, -t     Tidy up your go.mod working file (default: true)
   --binary, -b   Check for updates in your binaries (default: false)
   --global, -g   Check for binaries updates in your global directory (default: false)
   --version, -v  Print the version and exit (default: false)
   --help, -h     show help (default: false)
```
