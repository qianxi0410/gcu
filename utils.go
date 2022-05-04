package main

import (
	"github.com/fatih/color"
)

func printAllDepLatest() {
	c := color.New(color.FgCyan, color.Bold)
	c.Println("🎉 All the latest dependencies!")
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
