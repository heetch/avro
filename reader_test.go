package avro

import (
	"io"
	"strings"
	"testing"
	"testing/iotest"

	qt "github.com/frankban/quicktest"
)

func TestReadFixed(t *testing.T) {
	c := qt.New(t)
	d := &decoder{
		buf: make([]byte, 0, 10),
		r:   iotest.OneByteReader(strings.NewReader("abcdefghijklmnopqrstuvwxyz")),
	}
	b := d.readFixed(5)
	c.Assert(string(b), qt.Equals, "abcde")
	b = d.readFixed(3)
	c.Assert(string(b), qt.Equals, "fgh")
	b = d.readFixed(5)
	c.Assert(string(b), qt.Equals, "ijklm")
	p := catch(func() {
		d.readFixed(30)
	})
	c.Assert(p, qt.Not(qt.IsNil))
	c.Assert(p.(*decodeError).err, qt.Equals, io.ErrUnexpectedEOF)
}

func catch(f func()) (v interface{}) {
	defer func() {
		v = recover()
	}()
	f()
	return
}
