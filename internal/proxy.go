package internal

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
	"strings"

	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"
)

const limit = 100

type Module struct {
	Path     string
	Versions []string
}

// MaxVersion returns the highest version of the module.
// if there is no version return a empty string
// if pre is false, prerelease version will also exclude.
func (m *Module) MaxVersion(prefix string, stable bool) (max string) {
	for _, v := range m.Versions {
		if !semver.IsValid(v) || !strings.HasPrefix(v, prefix) {
			continue
		}

		if !stable && semver.Prerelease(v) != "" {
			continue
		}

		if max == "" {
			max = v
		}

		if semver.Compare(max, v) == -1 {
			max = v
		}
	}

	return max
}

// nextMajorVersion returns the next major version of the module.
func nextMajorVersion(version string) (next string, err error) {
	major, err := strconv.Atoi(strings.TrimPrefix(semver.Major(version), "v"))
	if err != nil {
		return
	}

	next = fmt.Sprintf("v%d", major+1)
	return
}

func (m *Module) VersionPath(version string) string {
	prefix := ModPrefix(m.Path)
	return JoinPath(prefix, version, "")
}

func (m *Module) NextMajorPath() (string, bool) {
	latest := m.MaxVersion("", true)
	if latest == "" {
		return "", false
	}

	if semver.Major(latest) == "v0" {
		return "", false
	}

	next, err := nextMajorVersion(latest)
	if err != nil {
		return "", false
	}

	return m.VersionPath(next), true
}

// MakeModule will fetch versions from the proxy and return a Module.
func Query(modp string, cached bool) (*Module, bool, error) {
	escaped, err := module.EscapePath(modp)
	if err != nil {
		return nil, false, err
	}

	url := fmt.Sprintf("https://proxy.golang.org/%s/@v/list", escaped)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, false, err
	}

	if cached {
		req.Header.Set("Disable-Module-Fetch", "true")
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, false, err
	}

	defer res.Body.Close()

	if res.ContentLength == 0 {
		return nil, false, nil
	}

	if res.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(res.Body)

		if res.StatusCode == http.StatusGone && bytes.HasPrefix(body, []byte("not found: ")) {
			return nil, false, nil
		}

		msg := string(body)
		if msg == "" {
			msg = res.Status
		}

		return nil, false, fmt.Errorf("proxy: %s", msg)
	}

	mod := new(Module)
	mod.Path = modp

	sc := bufio.NewScanner(res.Body)
	for sc.Scan() {
		mod.Versions = append(mod.Versions, sc.Text())
	}

	if err := sc.Err(); err != nil {
		return nil, false, err
	}

	return mod, true, nil
}

func Latest(modp string, cached bool) (*Module, error) {
	latest, ok, err := Query(modp, cached)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, fmt.Errorf("module not found: %s", modp)
	}

	for i := 0; i < limit; i++ {
		nextp, ok := latest.NextMajorPath()
		if !ok {
			return latest, nil
		}

		next, ok, err := Query(nextp, cached)
		if err != nil {
			return nil, err
		}

		if !ok {
			version := latest.MaxVersion("", true)
			if semver.Build(version) == "+incompatible" {
				nextp = latest.VersionPath((semver.Major(version)))
				if nextp != latest.Path {
					next, ok, err = Query(nextp, cached)
					if err != nil {
						return nil, err
					}
				}
			}
		}
		if !ok {
			return latest, nil
		}
		latest = next
	}

	return nil, fmt.Errorf("request too many times")
}

func QueryPkg(pkgpath string, cached bool) (*Module, error) {
	prefix := pkgpath
	for prefix != "" {
		if module.CheckPath(prefix) == nil {
			mod, ok, err := Query(prefix, cached)
			if err != nil {
				return nil, err
			}

			if ok {
				modprefix := ModPrefix(mod.Path)
				if modpath, pkgdir, ok := SplitPath(modprefix, pkgpath); ok && modpath != mod.Path {
					if major, ok := ModMajor(modpath); ok {
						if v := mod.MaxVersion(major, false); v != "" {
							spec := JoinPath(modprefix, "", pkgdir) + "@" + v
							return nil, fmt.Errorf("%s is not in %s", pkgpath, spec)
						}
						return nil, fmt.Errorf("failed to find %s in %s", pkgpath, modprefix)
					}
				}
				return mod, nil
			}
		}
		remain, last := path.Split(prefix)
		if last == "" {
			break
		}
		prefix = strings.TrimSuffix(remain, "/")
	}

	return nil, fmt.Errorf("failed to find module for %s", pkgpath)
}
