package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/briandowns/spinner"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/urfave/cli/v2"
)

func gcuCmd(ctx *cli.Context) error {
	if ctx.Bool("version") {
		return versionCmd(ctx)
	}

	filePath := ctx.Args().First()
	if filePath == "" {
		filePath = "."
	}

	if ctx.Bool("binary") {
		if ctx.Bool("global") {
			filePath = filepath.Join(os.Getenv("GOPATH"), "bin")
		}

		if err := checkBinaries(filePath); err != nil {
			return err
		}

		return nil
	}

	versions, err := getVersions(*ctx, filePath)
	if err != nil {
		return err
	}

	if len(versions) == 0 {
		printAllDepLatest()
		return nil
	}

	if ctx.Bool("all") {
		s := spinner.New(spinner.CharSets[0], 100*time.Millisecond)
		s.Prefix = "Updating... Please wait. "
		if err := s.Color("cyan"); err != nil {
			return err
		}

		s.Start()

		for _, v := range versions {
			if err := upgrade(v.path, v.new, filePath, ctx.Bool("rewrite") && !ctx.Bool("safe"), ctx.Bool("tidy")); err != nil {
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
	s.Prefix = "Updating... Please wait. "
	if err := s.Color("cyan"); err != nil {
		return err
	}

	s.Start()

	for _, idx := range idxs {
		if err := upgrade(versions[idx].path, versions[idx].new, filePath, ctx.Bool("rewrite") && !ctx.Bool("safe"), ctx.Bool("tidy")); err != nil {
			return err
		}
	}

	s.Stop()
	printPartDepLatest()

	return nil
}

func listCmd(ctx *cli.Context) error {
	filePath := ctx.Args().First()
	if filePath == "" {
		filePath = "."
	}

	versions, err := getVersions(*ctx, filePath)
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

func versionCmd(_ *cli.Context) error {
	fmt.Printf("gcu(go check updates): %s\n", gcuVersion)
	return nil
}
