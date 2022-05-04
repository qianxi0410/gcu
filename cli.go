package main

import (
	"log"
	"os"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/briandowns/spinner"
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
				Usage:   "Path to go.mod file.",
				Value:   ".",
			},
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
	versions, err := getVersions(*ctx)
	if err != nil {
		return err
	}

	if len(versions) == 0 {
		printAllDepLatest()
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

type MultiSelect struct {
	survey.MultiSelect
}

func (m MultiSelect) Cleanup(config *survey.PromptConfig, val interface{}) error {
	return m.Render("", nil)
}

func gcuCmd(ctx *cli.Context) error {
	versions, err := getVersions(*ctx)
	if err != nil {
		return err
	}

	if len(versions) == 0 {
		printAllDepLatest()
		return nil
	}

	if ctx.Bool("all") {
		s := spinner.New(spinner.CharSets[0], 100*time.Millisecond)
		s.Prefix = "Updating... Please wait."
		if err := s.Color("cyan"); err != nil {
			return err
		}

		s.Start()
		for _, v := range versions {
			if err := upgrade(v.path, v.new, ctx.String("modfile"), ctx.Bool("rewrite") && !ctx.Bool("safe")); err != nil {
				return err
			}
		}
		s.Stop()

		printAllDepLatest()
		return nil
	}

	options := make([]string, 0, len(versions))
	m1, m2, m3 := caculateMaxLenForEachItem(versions)

	for _, v := range versions {
		options = append(options, v.String(m1, m2, m3))
	}

	idxs := make([]int, 0, len(options))

	// disable MultiSelect answer output.
	prompt := &MultiSelect{
		survey.MultiSelect{
			Message:  "Select the dependencies you need to upgrade: ",
			Options:  options,
			PageSize: ctx.Int("size"),
		},
	}
	err = survey.AskOne(prompt, &idxs)
	if err == terminal.InterruptErr {
		printBye()
		os.Exit(0)
	} else if err != nil {
		return err
	}

	s := spinner.New(spinner.CharSets[0], 100*time.Millisecond)
	s.Prefix = "Updating... Please wait."
	if err := s.Color("cyan"); err != nil {
		return err
	}

	s.Start()
	for _, idx := range idxs {
		if err := upgrade(versions[idx].path, versions[idx].new, ctx.String("modfile"), ctx.Bool("rewrite") && !ctx.Bool("safe")); err != nil {
			return err
		}
	}
	s.Stop()
	printPartDepLatest()

	return nil
}
