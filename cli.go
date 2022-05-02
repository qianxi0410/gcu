package main

import (
	"log"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
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
				Usage:   "Path to go.mod file",
				Value:   ".",
			},
			&cli.BoolFlag{
				Name:    "stable",
				Aliases: []string{"s"},
				Usage:   "Only fetch stable version",
				Value:   true,
			},
			&cli.BoolFlag{
				Name:    "cached",
				Aliases: []string{"c"},
				Usage:   "Use cached version if available",
				Value:   false,
			},
			&cli.BoolFlag{
				Name:    "all",
				Aliases: []string{"a"},
				Usage:   "Upgrade all dependencies without asking",
				Value:   false,
			},
			&cli.BoolFlag{
				Name:    "rewrite",
				Aliases: []string{"w"},
				Usage:   "Rewrite all dependencies to latest version in your project",
				Value:   true,
			},
		},
		Commands: []*cli.Command{
			{
				Name:   "list",
				Usage:  "List all direct dependencies available for update",
				Action: listCmd,
			},
		},
		Action: gcuCmd,
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
		return nil
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"lib", "current version", "latest version"})
	for _, v := range versions {
		t.AppendRow(table.Row{v.path, v.oldversion(), v.newVersion()})
	}

	t.Render()

	return nil
}

func gcuCmd(ctx *cli.Context) error {
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
		return nil
	}

	if ctx.Bool("all") {
		for _, v := range versions {
			if err := upgrade(v.path, v.new, ctx.String("modfiles"), ctx.Bool("rewrite")); err != nil {
				return err
			}
		}

		c := color.New(color.FgGreen, color.Bold)
		c.Println("🎉 All the latest dependencies!")
		return nil
	}

	options := make([]string, 0, len(versions))

	for _, v := range versions {
		options = append(options, v.String())
	}

	idxs := []int{}
	prompt := &survey.MultiSelect{
		Message:  "Select the dependencies you need to upgrade:",
		Options:  options,
		PageSize: 10,
	}
	if err := survey.AskOne(prompt, &idxs, survey.WithPageSize(10)); err != nil {
		return err
	}

	for _, idx := range idxs {
		if err := upgrade(versions[idx].path, versions[idx].new, ctx.String("modfile"), ctx.Bool("rewrite")); err != nil {
			return err
		}
	}
	c := color.New(color.FgGreen, color.Bold)
	c.Println("🎉 All the latest dependencies!")

	return nil
}
