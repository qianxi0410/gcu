# gcu (go check updates)

Due to following the [semantic version number](https://semver.org/), Go will not actively update the major version when checking and updating dependencies, and Go does not provide us with optional dependency update items. When updating, either all of them are updated or none of them are updated.

But this tool makes up for these deficiencies.

- It provides update checking for major versions
- Visual update selection
- Colored version number distinguishing hints
- Automatically rewrite import paths (default)

warning:

- You need to ensure your own compatibility after updating major versions
- If the major version of the library is discontinuous, the latest version may not be available (e.g. 1.0.0 -> 3.1.0 without v2)
- Still Work In Progress

```bash
> gcu help

NAME:
   gcu (go-check-updates) - check for updates in go mod dependency

USAGE:
   gcu (go-check-updates) [global options] command [command options] [arguments...]

COMMANDS:
   list     List all direct dependencies available for update
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --modfile value, -m value  Path to go.mod file (default: ".")
   --stable, -s               Only fetch stable version (default: true)
   --cached, -c               Use cached version if available (default: false)
   --all, -a                  Upgrade all dependencies without asking (default: false)
   --rewrite, -w              Rewrite all dependencies to latest version in your project (default: true)
   --help, -h                 show help (default: false)
```