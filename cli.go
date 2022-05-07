package main

import (
	"bufio"
	"bytes"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
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

type MultiSelect struct {
	survey.MultiSelect
}

func (m MultiSelect) Cleanup(config *survey.PromptConfig, val interface{}) error {
	return m.Render("", nil)
}

func gcuCmd(ctx *cli.Context) error {
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

		func() {
			s.Stop()
			printAllLibLatest()
		}()

		for _, v := range versions {
			if err := upgrade(v.path, v.new, filePath, ctx.Bool("rewrite") && !ctx.Bool("safe"), ctx.Bool("tidy")); err != nil {
				return err
			}
		}

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

	func() {
		s.Stop()
		printPartDepLatest()
	}()

	for _, idx := range idxs {
		if err := upgrade(versions[idx].path, versions[idx].new, filePath, ctx.Bool("rewrite") && !ctx.Bool("safe"), ctx.Bool("tidy")); err != nil {
			return err
		}
	}

	return nil
}

func checkBinaries(fp string) error {
	s := spinner.New(spinner.CharSets[0], 100*time.Millisecond)
	s.Prefix = "Updating... Please wait. "
	if err := s.Color("cyan"); err != nil {
		return err
	}

	s.Start()

	defer func() {
		s.Stop()
		printAllLibLatest()
	}()

	output, err := exec.Command("go", "version", "-m", fp).Output()
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	paths := make([]string, 0)
	wg := sync.WaitGroup{}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "path") {
			path := strings.SplitN(line, "\t", 2)[1]
			paths = append(paths, path)

		}
	}

	wg.Add(len(paths))

	for _, path := range paths {
		go func(path string) {
			defer wg.Done()
			exec.Command("go", "install", path+"@latest").Run()
		}(path)
	}

	wg.Wait()

	return nil
}
