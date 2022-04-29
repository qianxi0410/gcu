package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"
)

func modPrefix(modp string) string {
	prefix, _, ok := module.SplitPathVersion(modp)
	if !ok {
		prefix = modp
	}

	return prefix
}

// modMajaro return the major mod of the given mod
// like: vX.Y.Z -> vX; X >= 2
// will return empty string if the given mod is not a valid mod.
func modMajor(modp string) (string, bool) {
	_, major, ok := module.SplitPathVersion(modp)
	if ok {
		major = strings.TrimPrefix(major, "/")
		major = strings.TrimPrefix(major, ".")
	}

	return major, ok
}

// joinPath create a full pkg path.
func joinPath(modprefix, version, pkgdir string) string {
	version = strings.TrimPrefix(version, ".")
	version = strings.TrimPrefix(version, "/")

	major := semver.Major(version)
	pkgpath := modprefix

	switch {
	case strings.HasPrefix(pkgpath, "gopkg.in"):
		pkgpath += "." + major
	case major != "" && major != "v0" && major != "v1" && !strings.Contains(version, "+incompatible"):
		if !strings.HasPrefix(pkgpath, "/") {
			pkgpath += "/"
		}
		pkgpath += major
	}
	if pkgdir != "" {
		pkgpath += "/" + pkgdir
	}

	return pkgpath
}

// SplitPath split the pkgpath to the modpath and pkgdir.
func splitPath(modprefix, pkgpath string) (modpath, pkgdir string, ok bool) {
	if !strings.HasPrefix(pkgpath, modprefix) {
		return
	}

	modpathLen := len(modprefix)
	if strings.HasPrefix(pkgpath[modpathLen:], "/") {
		modpathLen++
	}

	if idx := strings.Index(pkgpath[modpathLen:], "/"); idx >= 0 {
		modpathLen += idx
	} else {
		modpathLen = len(pkgpath)
	}

	modpath = modprefix
	if major, ok := modMajor(pkgpath[:modpathLen]); ok {
		modpath = joinPath(modprefix, major, "")
	}
	pkgdir = strings.TrimPrefix(pkgpath[len(modpath):], "/")

	return modpath, pkgdir, true
}

// findModFile recursively search the given path for a go.mod file.
// only search up.
func findModFile(dir string) (path string, err error) {
	if dir, err = filepath.Abs(dir); err != nil {
		return "", err
	}

	for {
		path = filepath.Join(dir, "go.mod")
		if _, err = os.Stat(path); err == nil {
			return
		}

		if !os.IsNotExist(err) {
			return "", err
		}

		if dir == "" || dir == "." || dir == "/" {
			break
		}

		dir = filepath.Dir(dir)
	}

	return "", os.ErrNotExist
}

// direct returns the direct module deps.
func direct(dir string) ([]module.Version, error) {
	name, err := findModFile(dir)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}

	f, err := modfile.ParseLax(name, data, nil)
	if err != nil {
		return nil, err
	}

	mods := make([]module.Version, 0, len(f.Require))
	for _, req := range f.Require {
		if !req.Indirect {
			mods = append(mods, req.Mod)
		}
	}

	return mods, nil
}
