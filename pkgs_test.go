package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJoinPath(t *testing.T) {
	tests := []struct {
		modprefix string
		version   string
		pkgdir    string
		want      string
	}{
		{
			modprefix: "github.com/google/go-cmp",
			version:   "v0.1.0",
			pkgdir:    "cmp",
			want:      "github.com/google/go-cmp/cmp",
		},
		{
			modprefix: "github.com/google/go-cmp",
			version:   "v0.1.2",
			pkgdir:    "",
			want:      "github.com/google/go-cmp",
		},
		{
			modprefix: "github.com/google/go-cmp",
			version:   "v1.0.0",
			pkgdir:    "cmp",
			want:      "github.com/google/go-cmp/cmp",
		},
		{
			modprefix: "github.com/google/go-cmp",
			version:   "v2.0.0",
			pkgdir:    "cmp",
			want:      "github.com/google/go-cmp/v2/cmp",
		},
		{
			modprefix: "github.com/google/go-cmp",
			version:   "v0.1.0+incompatible",
			pkgdir:    "cmp",
			want:      "github.com/google/go-cmp/cmp",
		},
		{
			modprefix: "github.com/google/go-cmp",
			version:   "v2.0.0+incompatible",
			pkgdir:    "cmp",
			want:      "github.com/google/go-cmp/cmp",
		},
		{
			modprefix: "github.com/google/go-cmp",
			version:   "v0.0.0-20191010130408-0a8f2c9e6d3d",
			pkgdir:    "cmp",
			want:      "github.com/google/go-cmp/cmp",
		},
		{
			modprefix: "github.com/google/go-cmp",
			version:   "v2.0.0-20191010130408-0a8f2c9e6d3d",
			pkgdir:    "cmp",
			want:      "github.com/google/go-cmp/v2/cmp",
		},
		{
			modprefix: "gopkg.in/yaml",
			version:   "v2",
			pkgdir:    "",
			want:      "gopkg.in/yaml.v2",
		},
		{
			modprefix: "gopkg.in/yaml",
			version:   "v2",
			pkgdir:    "cc",
			want:      "gopkg.in/yaml.v2/cc",
		},
		{
			modprefix: "gopkg.in/yaml",
			version:   "v1",
			pkgdir:    "",
			want:      "gopkg.in/yaml.v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := joinPath(tt.modprefix, tt.version, tt.pkgdir)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestModMajor(t *testing.T) {
	tests := []struct {
		modp string
		want string
	}{
		{
			"gopkg.in/yaml.v2",
			"v2",
		},
		{
			"gopkg.in/yaml.v2.v2",
			"v2",
		},
		{
			"github.com/golang/protobuf/v2",
			"v2",
		},
		{
			"github.com/golang/protobuf/v0.0.0-20191010130408-0a8f2c9e6d3d",
			"",
		},
		{
			"github.com/golang/protobuf/v0.0.0-20191010130408-0a8f2c9e6d3d+incompatible",
			"",
		},
		{
			"github.com/golang/protobuf",
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got, ok := modMajor(tt.modp)
			assert.True(t, ok)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSplitPath(t *testing.T) {
	tests := []struct {
		modprefix string
		pkgpath   string
		modpath   string
		pkgdir    string
		ok        bool
	}{
		{
			modprefix: "github.com/google/go-cmp",
			pkgpath:   "github.com/google/go-cmp/cmp",
			modpath:   "github.com/google/go-cmp",
			pkgdir:    "cmp",
			ok:        true,
		},
		{
			modprefix: "github.com/google/go-cmp",
			pkgpath:   "github.com/google/go-cmp/v2/cmp",
			modpath:   "github.com/google/go-cmp/v2",
			pkgdir:    "cmp",
			ok:        true,
		},
		{
			modprefix: "github.com/google/go-cmp",
			pkgpath:   "github.com/google/go-cmp/v2/cmp/cmp",
			modpath:   "github.com/google/go-cmp/v2",
			pkgdir:    "cmp/cmp",
			ok:        true,
		},
		{
			modprefix: "github.com/google/go-cmp",
			pkgpath:   "github.com/google/go-cmp/",
			modpath:   "github.com/google/go-cmp",
			pkgdir:    "",
			ok:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.modpath, func(t *testing.T) {
			modpath, pkgdir, ok := splitPath(tt.modprefix, tt.pkgpath)
			assert.Equal(t, tt.ok, ok)
			assert.Equal(t, tt.modpath, modpath)
			assert.Equal(t, tt.pkgdir, pkgdir)
		})
	}
}

func TestDirect(t *testing.T) {
	mods, _ := direct(".")
	for _, m := range mods {
		t.Logf("%s", m)
	}
}
