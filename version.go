package main

import (
	"fmt"
	"os/exec"
	"regexp"
	"sync"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"
)

type version struct {
	path string
	old  string
	new  string
}

// if v1 != v2 diff will returns true else false.
func diff(v1, v2 string) bool {
	return semver.Compare(v1, v2) != 0
}

// oldversion
func (v *version) oldversion() string {
	return v.old[:len(v.old)-len(semver.Build(v.old))]
}

// colorful print new version's diffent part.
func (v *version) newVersion() string {
	major, minor, patch, pre := color.New(color.FgWhite).SprintFunc(), color.New(color.FgWhite).SprintFunc(), color.New(color.FgWhite).SprintFunc(), color.New(color.FgWhite).SprintFunc()

	pattern := regexp.MustCompile(`(v[\d]+).([\d]+).([\d]+)([-\w]*)([+\w]*)`)
	olds := pattern.FindStringSubmatch(v.old)
	news := pattern.FindStringSubmatch(v.new)

	if olds[1] != news[1] {
		major = color.New(color.FgRed).SprintFunc()
	}
	if olds[2] != news[2] {
		minor = color.New(color.FgBlue).SprintFunc()
	}
	if olds[3] != news[3] {
		patch = color.New(color.FgGreen).SprintFunc()
	}
	if olds[4] != news[4] {
		pre = color.New(color.FgYellow).SprintFunc()
	}

	return fmt.Sprintf("%s.%s.%s%s", major(news[1]), minor(news[2]), patch(news[3]), pre(news[4]))
}

func (v *version) String(m1, m2, m3 int) string {
	format := fmt.Sprintf("%%-%ds %%-%ds %%-%ds", m1, m2, m3)
	return fmt.Sprintf(format, v.path, v.oldversion(), v.newVersion())
}

func getVersions(ctx cli.Context) ([]version, error) {
	s := spinner.New(spinner.CharSets[36], 100*time.Millisecond)
	s.Prefix = "Checking... Please wait."
	if err := s.Color("cyan"); err != nil {
		return nil, err
	}

	s.Start()
	defer s.Stop()

	// accelerate the process.
	wg := &sync.WaitGroup{}
	wgCmd := &sync.WaitGroup{}
	mu := &sync.Mutex{}

	deps, err := direct(ctx.String("modfile"))
	if err != nil {
		return nil, err
	}

	versions := make([]version, 0, len(deps))

	pattern := regexp.MustCompile(`v0.0.0-[\d]{14}-[\d\s\S]{12}`)
	output := []byte{}

	wgCmd.Add(1)
	go func() {
		defer wgCmd.Done()
		output, _ = exec.Command("go", "list", "-u",
			"-f", "'{{if (and (not (or .Main .Indirect)) .Update)}}{{.Path}}: [{{.Version}}] [{{.Update.Version}}]{{end}}'",
			"-m", "all").Output()
	}()
	if err != nil {
		return nil, err
	}

	wg.Add(len(deps))

	for _, dep := range deps {
		go func(dep module.Version) {
			defer wg.Done()

			if ctx.Bool("safe") {
				wgCmd.Wait()

				old := dep.Version
				extractPattern := regexp.MustCompile(dep.Path + `: \[.*]\ \[(.*)\]`)
				result := extractPattern.FindStringSubmatch(string(output))
				if len(result) != 2 {
					return
				}
				new := result[1]
				mu.Lock()
				versions = append(versions, version{
					path: modPrefix(dep.Path),
					old:  old,
					new:  new,
				})
				mu.Unlock()

				return
			}

			if pattern.MatchString(dep.Version) {
				wgCmd.Wait()

				old := dep.Version
				extractPattern := regexp.MustCompile(dep.Path + `: \[v0.0.0-[\s\S\d\.]*[\d]{14}-[\d\s\S]{12}\] \[(.*)\]`)
				result := extractPattern.FindStringSubmatch(string(output))
				if len(result) != 2 {
					return
				}
				new := result[1]
				mu.Lock()
				versions = append(versions, version{
					path: modPrefix(dep.Path),
					old:  old,
					new:  new,
				})
				mu.Unlock()
			} else {
				mod, err := latest(dep.Path, ctx.Bool("cached"))
				if err != nil {
					return
				}
				old, new := dep.Version, mod.maxVersion("", ctx.Bool("stable"))
				if diff(old, new) {
					mu.Lock()
					versions = append(versions, version{
						path: modPrefix(mod.Path),
						old:  old,
						new:  new,
					})
					mu.Unlock()
				}
			}
		}(dep)
	}

	wg.Wait()

	return versions, nil
}
