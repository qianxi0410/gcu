package main

import (
	"fmt"
	"regexp"

	"github.com/fatih/color"
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
