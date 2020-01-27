package avro_test

import (
	"context"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/heetch/avro"
)

func BenchmarkMarshal(b *testing.B) {
	type R struct {
		A *string
		B *string
		C []int
	}
	type T struct {
		R R
	}
	x := T{
		R: R{
			A: newString("hello"),
			B: newString("goodbye"),
			C: []int{1, 3, 1 << 20},
		},
	}
	for i := 0; i < b.N; i++ {
		_, _, err := avro.Marshal(x)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSingleDecoderUnmarshal(b *testing.B) {
	c := qt.New(b)
	type R struct {
		A *string
		B *string
		C []int
	}
	type T struct {
		R R
	}
	at, err := avro.TypeOf(T{})
	c.Assert(err, qt.Equals, nil)
	r := memRegistry{
		1: at.String(),
	}
	enc := avro.NewSingleEncoder(r, nil)
	ctx := context.Background()
	data, err := enc.Marshal(ctx, T{
		R: R{
			A: newString("hello"),
			B: newString("goodbye"),
			C: []int{1, 3, 1 << 20},
		},
	})
	c.Assert(err, qt.Equals, nil)

	dec := avro.NewSingleDecoder(r, nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var x T
		_, err := dec.Unmarshal(ctx, data, &x)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func newString(s string) *string {
	return &s
}
