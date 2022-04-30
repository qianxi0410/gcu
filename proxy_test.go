package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLatest(t *testing.T) {
	tests := []string{
		"github.com/go-redis/redis",
		"github.com/labstack/echo",
	}

	for _, test := range tests {
		t.Run(test, func(t *testing.T) {
			mod, err := latest(test, true)
			assert.Nil(t, err)
			t.Logf("Latest: %s, %v", mod.Path, mod.maxVersion("", true))
		})
	}
}

func TestQuery(t *testing.T) {
	mod, ok, err := query("github.com/go-redis/redis", true)
	assert.Nil(t, err)
	assert.True(t, ok)
	t.Logf("query: %s, %v", mod.Path, mod.maxVersion("", true))

	mod, ok, err = query("github.com/labstack/echo/v5", true)
	assert.Nil(t, err)
	assert.False(t, ok)
}

func TestQueryPkg(t *testing.T) {
	tests := []struct {
		pkgpath string
		modpath string
	}{
		{
			"github.com/go-redis/redis",
			"github.com/go-redis/redis",
		},
		{
			"github.com/go-redis/redis/suffix",
			"github.com/go-redis/redis",
		},
	}
	for _, tt := range tests {
		t.Run(tt.pkgpath, func(t *testing.T) {
			mod, err := queryPkg(tt.pkgpath, true)
			assert.Nil(t, err)
			assert.Equal(t, tt.modpath, mod.Path)
		})
	}
}
