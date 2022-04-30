package main

import (
	"testing"
)

func TestDiff(t *testing.T) {
	tests := []struct {
		old  string
		new  string
		diff bool
	}{
		{
			"v1.0.0",
			"v1.0.0",
			false,
		},
		{
			"v1.0.0",
			"v1.0.1",
			true,
		},
		{
			"v2.1.1",
			"v2.1.1+incompatible",
			false,
		},
		{
			"v2.2.2",
			"v2.2.2-dev",
			true,
		},
		{
			"v0.0.0-20190101120000-abcdef",
			"v0.0.0-20190101120000-abcdef",
			false,
		},
		{
			"v0.0.0-20190101120000-abcdef",
			"v0.0.0-20190101120000-abcdefg",
			true,
		},
		{
			"v1.1.1+meta",
			"v1.1.1+meta2",
			false,
		},
	}

	for _, test := range tests {
		if diff(test.old, test.new) != test.diff {
			t.Errorf("diff(%q, %q) != %v", test.old, test.new, test.diff)
		}
	}
}
