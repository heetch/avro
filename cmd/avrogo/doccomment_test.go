package main

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

var docTests = []struct {
	testName string
	val      interface{}
	indent   string
	expect   string
}{{
	testName: "simple-doc-string",
	val:      docString("hello"),
	indent:   "// ",
	expect: `
// hello
`,
}, {
	testName: "not-a-doc-object",
	val:      245,
	indent:   "// ",
	expect:   "",
}, {
	testName: "multi-line",
	val: docString(`one line
and another`),
	indent: "// ",
	expect: `
// one line
// and another
`,
}, {
	testName: "empty-doc-string",
	val:      docString(""),
	indent:   "// ",
	expect:   "",
}, {
	testName: "extra-white-space",
	val:      docString(" \n  \thello\n \t "),
	indent:   "// ",
	expect: `
// hello
`,
}, {
	testName: "remove-stars",
	val:      docString("* one two three\n\t\t * four five."),
	indent:   "// ",
	expect: `
// one two three
// four five.
`,
}, {
	testName: "remove-stars-no-newlines",
	val:      docString("* one two three"),
	indent:   "// ",
	expect: `
// one two three
`,
}, {
	testName: "remove-stars-less-spaces",
	val:      docString("*one two three\n\t\t*four five\n\t\t*six seven\n\t\t*eight nine."),
	indent:   "// ",
	expect: `
// one two three
// four five
// six seven
// eight nine.
`,
}, {
	testName: "different-indent",
	val:      docString("hello"),
	indent:   "!!",
	expect: `
!!hello
`,
}}

func TestDoc(t *testing.T) {
	c := qt.New(t)
	for _, test := range docTests {
		c.Run(test.testName, func(c *qt.C) {
			c.Assert(doc(test.indent, test.val), qt.Equals, test.expect)
		})
	}
}

type docString string

func (d docString) Doc() string {
	return string(d)
}
