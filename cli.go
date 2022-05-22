package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

const gcuVersion = "0.1.0-dev"

func main() {
	app := &cli.App{
		Name:  "gcu (go-check-updates)",
		Usage: "check for updates in go.mod dependency and go's binary files",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "stable",
				Aliases: []string{"s"},
				Usage:   "Only fetch stable version.",
				Value:   true,
			},
			&cli.BoolFlag{
				Name:    "cached",
				Aliases: []string{"c"},
				Usage:   "Use cached version if available.",
				Value:   false,
			},
			&cli.BoolFlag{
				Name:    "all",
				Aliases: []string{"a"},
				Usage:   "Upgrade all dependencies without asking.",
				Value:   false,
			},
			&cli.BoolFlag{
				Name:    "rewrite",
				Aliases: []string{"w"},
				Usage:   "Rewrite all dependencies to latest version in your project.",
				Value:   true,
			},
			&cli.BoolFlag{
				Name:  "safe",
				Usage: "Only minor and patch releases are checked and updated.",
				Value: false,
			},
			&cli.IntFlag{
				Name:  "size",
				Usage: "Number of items to show in the select list.",
				Value: 10,
			},
			&cli.BoolFlag{
				Name:    "tidy",
				Aliases: []string{"t"},
				Usage:   "Tidy up your go.mod working file.",
				Value:   true,
			},
			&cli.BoolFlag{
				Name:    "binary",
				Aliases: []string{"b"},
				Usage:   "Check for updates in your binaries.",
				Value:   false,
			},
			&cli.BoolFlag{
				Name:    "global",
				Aliases: []string{"g"},
				Usage:   "Check for binaries updates in your global directory.",
				Value:   false,
			},
			&cli.BoolFlag{
				Name:    "version",
				Aliases: []string{"v"},
				Usage:   "Print the version and exit",
			},
		},
		Commands: []*cli.Command{
			{
				Name:   "list",
				Usage:  "List all direct dependencies available for update",
				Action: listCmd,
			},
			{
				Name:    "version",
				Usage:   "Print the version number of gcu",
				Action:  versionCmd,
				Aliases: []string{"v"},
			},
		},
		Action: gcuCmd,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalln(err)
	}
}
