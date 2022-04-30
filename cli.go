package main

import (
	"fmt"
	"log"
	"os"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "gcu (go-check-updates)",
		Usage: "check for updates in go mod dependency",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "modfile",
				Aliases: []string{"m"},
				Usage:   "path to go.mod file",
				Value:   ".",
			},
			&cli.BoolFlag{
				Name:    "stable",
				Aliases: []string{"s"},
				Usage:   "only fetch stable version",
				Value:   true,
			},
			&cli.BoolFlag{
				Name:    "cached",
				Aliases: []string{"c"},
				Usage:   "use cached version if available",
				Value:   false,
			},
		},
		Commands: []*cli.Command{
			{
				Name:   "list",
				Usage:  "list all direct dependencies available for update",
				Action: listCmd,
			},
		},
		// TODO: add main handler
		Action: nil,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalln(err)
	}
}

func listCmd(ctx *cli.Context) error {
	deps, err := direct(ctx.String("modfile"))
	if err != nil {
		return err
	}

	versions := make([]version, 0, len(deps))

	for _, dep := range deps {
		mod, err := latest(dep.Path, ctx.Bool("cached"))
		if err != nil {
			return err
		}
		old, new := dep.Version, mod.maxVersion("", ctx.Bool("stable"))
		if diff(old, new) {
			versions = append(versions, version{
				path: modPrefix(mod.Path),
				old:  old,
				new:  new,
			})
		}
	}

	if len(versions) == 0 {
		c := color.New(color.FgCyan, color.Bold)
		c.Println("🎉 All the latest dependencies!")
	}

	for _, v := range versions {
		fmt.Printf("%s\t %s\t %s\n", v.path, v.oldversion(), v.newVersion())
	}

	return nil
}
