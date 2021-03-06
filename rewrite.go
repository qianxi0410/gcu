package main

import (
	"fmt"
	"go/parser"
	"go/printer"
	"go/token"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type replaceFunc func(pos token.Position, path string) (string, error)

func rewrite(dir string, replace replaceFunc) error {
	return filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			log.Println("import rewrite: ", err)
			return nil
		}

		if info.IsDir() {
			if info.Name() == "vendor" || info.Name() == ".git" || info.Name() == ".vscode" {
				return filepath.SkipDir
			}
			if path != dir {
				_, err := os.Lstat(filepath.Join(path, "go.mod"))
				if err == nil {
					return filepath.SkipDir
				}
				if !os.IsNotExist(err) {
					log.Panicln("import rewrite: ", err)
					return nil
				}
			}
			return nil
		}

		// only do rewrite in on go file.
		if strings.HasSuffix(path, ".go") {
			return rewriteFile(path, replace)
		}

		return nil
	})
}

func rewriteFile(name string, replace replaceFunc) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, name, nil, parser.ParseComments)
	if err != nil {
		e := err.Error()
		msg := "expected 'package'. found EOF"
		if e[len(e)-len(msg):] == msg {
			return nil
		}

		return err
	}

	change := false
	for _, i := range f.Imports {
		path, err := strconv.Unquote(i.Path.Value)
		if err != nil {
			return err
		}

		path, err = replace(fset.Position(i.Pos()), path)
		if err != nil {
			if err == filepath.SkipDir {
				continue
			}
			return err
		}

		i.Path.Value = strconv.Quote(path)
		change = true
	}

	if !change {
		return nil
	}

	tmp := name + ".temp"
	w, err := os.Create(tmp)
	if err != nil {
		return err
	}

	defer w.Close()
	info, err := os.Lstat(name)
	if err != nil {
		return err
	}

	if err := w.Chmod(info.Mode()); err != nil {
		return err
	}

	cfg := &printer.Config{
		Mode:     printer.TabIndent | printer.UseSpaces,
		Tabwidth: 4,
	}
	if err := cfg.Fprint(w, fset, f); err != nil {
		return err
	}

	return os.Rename(tmp, name)
}

func upgrade(modp, v, dir string, r, tidy bool) error {
	newp := joinPath(modp, v, "")

	// use go mod edit to update go.mod
	if err := exec.Command("go", "get", "-u", fmt.Sprintf("%s@%s", newp, v)).Run(); err != nil {
		return err
	}

	major, ok := modMajor(newp)
	if !ok {
		return fmt.Errorf("%s: failed to get major version", newp)
	}

	if !r || major == "v0" || major == "v1" || strings.Contains(v, "+incompatible") {
		return nil
	}
	// rewrite import path
	err := rewrite(dir, func(_ token.Position, path string) (string, error) {
		_, pkgdir, ok := splitPath(modp, path)
		if !ok {
			return "", filepath.SkipDir
		}
		newp := joinPath(modp, v, pkgdir)

		return newp, nil
	})
	if err != nil {
		return err
	}

	// after rewrite, we need to run go mod tidy to make sure go.mod is valid. ?
	if tidy {
		if err := exec.Command("go", "mod", "tidy").Run(); err != nil {
			return err
		}
	}
	

	return nil
}
