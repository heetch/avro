package main

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

var outputPathsTests = []struct {
	testName    string
	paths       []string
	testFile    bool
	expectError string
	expect      map[string]string
}{{
	testName: "single-path",
	paths:    []string{"foo.avsc"},
	expect: map[string]string{
		"foo.avsc": "foo_gen.go",
	},
}, {
	testName: "several-paths",
	paths:    []string{"foo.avsc", "bar.avsc"},
	expect: map[string]string{
		"foo.avsc": "foo_gen.go",
		"bar.avsc": "bar_gen.go",
	},
}, {
	testName: "ambiguous-paths",
	paths:    []string{"alpha/bravo/x.avsc", "alpha/charlie/x.avsc"},
	expect: map[string]string{
		"alpha/bravo/x.avsc":   "bravo_x_gen.go",
		"alpha/charlie/x.avsc": "charlie_x_gen.go",
	},
}, {
	testName: "multiple-levels",
	paths:    []string{"foo.avsc", "a/b/versions/v0.avsc", "a/c/versions/v0.avsc", "alpha/bravo/x.avsc", "alpha/charlie/x.avsc"},
	expect: map[string]string{
		"foo.avsc":             "foo_gen.go",
		"a/b/versions/v0.avsc": "b_versions_v0_gen.go",
		"a/c/versions/v0.avsc": "c_versions_v0_gen.go",
		"alpha/bravo/x.avsc":   "bravo_x_gen.go",
		"alpha/charlie/x.avsc": "charlie_x_gen.go",
	},
}, {
	testName: "ambiguous-until-root",
	paths:    []string{"/foo/bar.avsc", "foo/bar.avsc"},
	expect: map[string]string{
		"/foo/bar.avsc": "slash_foo_bar_gen.go",
		"foo/bar.avsc":  "foo_bar_gen.go",
	},
}, {
	testName: "ambiguous-until-start-of-relative",
	paths:    []string{"foo/bar.avsc", "bar.avsc"},
	expect: map[string]string{
		"foo/bar.avsc": "foo_bar_gen.go",
		"bar.avsc":     "bar_gen.go",
	},
}, {
	testName: "duplicate paths",
	paths:    []string{"x.avsc", "x.avsc"},
	expect: map[string]string{
		"x.avsc": "x_gen.go",
	},
}, {
	testName:    "ambiguous-underscores",
	paths:       []string{"x/y.avsc", "x_y.avsc", "z/y.avsc"},
	expectError: "could not make unambiguous output files from input files",
}, {
	testName:    "ambiguous-ext",
	paths:       []string{"x.json", "x.avsc"},
	expectError: "could not make unambiguous output files from input files",
}}

func TestOutputPaths(t *testing.T) {
	c := qt.New(t)
	for _, test := range outputPathsTests {
		c.Run(test.testName, func(c *qt.C) {
			result, err := outputPaths(test.paths, test.testFile)
			if test.expectError != "" {
				c.Check(result, qt.IsNil)
				c.Assert(err, qt.ErrorMatches, test.expectError)
				return
			}
			c.Assert(err, qt.IsNil)
			c.Assert(result, qt.DeepEquals, test.expect)
		})
	}
}
