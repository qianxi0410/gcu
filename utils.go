package main

import (
	"bufio"
	"bytes"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
)

func printAllDepLatest() {
	c := color.New(color.FgCyan, color.Bold)
	c.Println("🎉 All the latest dependencies!")
}

func printAllLibLatest() {
	c := color.New(color.FgCyan, color.Bold)
	c.Println("🎉 All the latest libs!")
}

func printPartDepLatest() {
	c := color.New(color.FgCyan, color.Bold)
	c.Println("🎉 The dependencies you selected have been updated to the latest!")
}

func printBye() {
	c := color.New(color.FgGreen, color.Bold)
	c.Println("👋 Bye!")
}

func max(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}

func caculateMaxLenForEachItem(versions []version) (m1, m2, m3 int) {
	for _, v := range versions {
		m1 = max(m1, len(v.path))
		m2 = max(m2, len(v.oldversion()))
		m3 = max(m3, len(v.newVersion()))
	}

	return
}

type MultiSelect struct {
	survey.MultiSelect
}

func (m MultiSelect) Cleanup(config *survey.PromptConfig, val interface{}) error {
	return m.Render("", nil)
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
			_ = exec.Command("go", "install", path+"@latest").Run()
		}(path)
	}

	wg.Wait()

	return nil
}
