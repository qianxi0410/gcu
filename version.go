package main

import (
	"fmt"
	"os/exec"
	"regexp"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
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

func (v *version) String() string {
	return fmt.Sprintf("%s %s -> %s", v.path, v.oldversion(), v.newVersion())
}

func getVersions(ctx cli.Context) ([]version, error) {
	deps, err := direct(ctx.String("modfile"))
	if err != nil {
		return nil, err
	}

	versions := make([]version, 0, len(deps))

	pattern := regexp.MustCompile(`v[\d]+.0.0-[\d]{14}-[\d\s\S]{12}`)

	output, err := exec.Command("go", "list", "-m", "-u", "all").Output()
	if err != nil {
		return nil, err
	}

	for _, dep := range deps {
		if pattern.MatchString(dep.Version) {
			old := dep.Version
			extractPattern := regexp.MustCompile(dep.Path + ` v[\d]+.0.0-[\d]{14}-[\d\s\S]{12} \[(.*)\]`)
			result := extractPattern.FindStringSubmatch(string(output))
			if len(result) != 2 {
				continue
			}
			new := result[1]
			versions = append(versions, version{
				path: modPrefix(dep.Path),
				old:  old,
				new:  new,
			})
		} else {
			mod, err := latest(dep.Path, ctx.Bool("cached"))
			if err != nil {
				return nil, err
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
	}

	return versions, nil
}
