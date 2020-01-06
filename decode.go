package avro

import (
	"fmt"
	"io"
	"reflect"

	"github.com/actgardner/gogen-avro/parser"
	"github.com/actgardner/gogen-avro/resolver"
	"github.com/actgardner/gogen-avro/schema"
	"github.com/actgardner/gogen-avro/vm"
)

// Unmarshal unmarshals the given Avro-encoded binary data, which must
// have been written with the given schema, wSchemaStr, into x, which must be a
// pointer to a struct type T.
//
// The schema of T must be compatible with the writer schema
// described by wSchemaStr according to the rules described
// here: https://avro.apache.org/docs/current/spec.html#Schema+Resolution
//
// The reader schema used is Schema(*x) except that interface{}-typed values
// are allowed in non-generated types. When an interface{}-typed value is encountered, it is
// as if that part of the type was filled in with the writer schema.
//
// When decoding Avro values into a Go interface{} type, the decoded Go type for Avro type
// T, decode(T), is derived according to the following rules (TODO implement this):
//
//	- "int" decodes as int32
//	- "long" decodes as int64
//	- "float" decodes as float32
//	- "double" decodes as float64
//	- "string" decodes as string
//	- "null" decodes as nil
//	- {"type": "fixed", "size": N} decodes as [N]byte
//	- {"type": "array", "items": T} decodes as []decode(T)
//	- {"type": "map", "values": T} decodes as map[string]decode(T)
//	- {"type": "record", "fields": [....]} decodes as map[string]interface{}
//	- ["null", T] decodes as *decode(T).
//	- [T1, T2, ...] when {decode(T1), decode(T2), ...} are all distinct Go types, decodes as the value itself.
//	- [T1, T2, ...] otherwise decodes as map[string] interface{} where
//	the map contains a single key holding the actual union type name as
//	represented in the Avro JSON encoding and the value holds a value of
//	type decode(Tn).
//
// TODO it might be better to decode in encoding/json-compatible format (e.g. all slices are []interface{} etc).
//
// TODO return the actual reader schema value used? This will be Schema(*x)
// unless x contains interface{} fields. That would allow any unmarshaled value
// to be marshaled again, even if it contained interface{}-typed values.
func Unmarshal(data []byte, x interface{}, wSchemaStr string) error {
	wSchema, err := parseSchema([]byte(wSchemaStr))
	if err != nil {
		return fmt.Errorf("cannot parse writer schema: %v", err)
	}
	v := reflect.ValueOf(x)
	t := v.Type()
	if t.Kind() != reflect.Ptr {
		return fmt.Errorf("destination is not a pointer %s", t)
	}
	prog, err := compileDecoder(t.Elem(), wSchema)
	if err != nil {
		return err
	}
	v = v.Elem()
	return unmarshal(nil, data, prog, v)
}

func parseSchema(s []byte) (schema.AvroType, error) {
	ns := parser.NewNamespace(false)
	avroType, err := ns.TypeForSchema(s)
	if err != nil {
		return nil, err
	}
	for _, def := range ns.Roots {
		if err := resolver.ResolveDefinition(def, ns.Definitions); err != nil {
			return nil, fmt.Errorf("cannot resolve references in schema: %v", err)
		}
	}
	return avroType, nil
}

type stackFrame struct {
	Boolean   bool
	Int       int64
	Float     float64
	Bytes     []byte
	String    string
	Condition bool
}

type decoder struct {
	pc      int
	program *decodeProgram

	// buf holds bytes read from r to be consumed
	// by the decoder. The unconsumed bytes are
	// in d.buf[d.scan:].
	buf     []byte
	scan    int
	r       io.Reader
	readErr error
}

type decodeError struct {
	err error
}

// unmarshal unmarshals Avro binary data from r and writes it to target
// following the given program.
func unmarshal(r io.Reader, buf []byte, prog *decodeProgram, target reflect.Value) (err error) {
	defer func() {
		switch panicErr := recover().(type) {
		case *decodeError:
			err = panicErr.err
		case nil:
		default:
			panic(panicErr)
		}
	}()
	d := decoder{
		r:       r,
		program: prog,
	}
	if r == nil {
		d.buf = buf
		d.readErr = io.EOF
	} else {
		d.buf = make([]byte, 0, bufSize)
	}
	d.eval(target)
	return nil
}

func (d *decoder) eval(target reflect.Value) {
	if target.IsValid() {
		debugf("eval %s", target.Type())
	} else {
		debugf("eval nil")
	}
	defer debugf("}")
	var frame stackFrame
	for ; d.pc < len(d.program.Instructions); d.pc++ {
		debugf("x %d: %v", d.pc, d.program.Instructions[d.pc])
		switch inst := d.program.Instructions[d.pc]; inst.Op {
		case vm.Read:
			switch inst.Operand {
			case vm.Null:
			case vm.Boolean:
				frame.Boolean = d.readBool()
			case vm.Int:
				// TODO bounds check
				frame.Int = d.readLong()
			case vm.Long:
				frame.Int = d.readLong()
			case vm.UnusedLong:
				d.readLong()
			case vm.Float:
				frame.Float = d.readFloat()
			case vm.Double:
				frame.Float = d.readDouble()
			case vm.Bytes:
				frame.Bytes = d.readBytes()
			case vm.String:
				frame.String = d.readString()
			default:
				frame.Bytes = d.readFixed(inst.Operand - 11)
			}
		case vm.Set:
			debugf("%v on %s", inst, target.Type())
			switch inst.Operand {
			case vm.Null:
			case vm.Boolean:
				target.SetBool(frame.Boolean)
			case vm.Int, vm.Long:
				// This is called on union types to set
				// the kind of union. TODO remove this hack!
				func() {
					defer func() {
						recover()
					}()
					target.SetInt(int64(frame.Int))
				}()
			case vm.Float, vm.Double:
				target.SetFloat(float64(frame.Float))
			case vm.Bytes:
				if target.Kind() == reflect.Array {
					n := reflect.Copy(target, reflect.ValueOf(frame.Bytes))
					if n != len(frame.Bytes) {
						d.error(fmt.Errorf("copied too little"))
					}
				} else {
					data := make([]byte, len(frame.Bytes))
					copy(data, frame.Bytes)
					target.SetBytes(data)
				}
			case vm.String:
				target.SetString(frame.String)
			}
		case vm.SetDefault:
			if d.program.makeDefault[d.pc] == nil {
				panic(fmt.Errorf("no makeDefault at PC %d; prog %p", d.pc, &d.program.makeDefault[0]))
			}
			target.Field(inst.Operand).Set(reflect.ValueOf(d.program.makeDefault[d.pc]()))
		case vm.Enter:
			val, isRef := d.program.enter[d.pc](target)
			debugf("enter %d -> %#v (isRef %v) {", inst.Operand, val, isRef)
			d.pc++
			d.eval(val)
			if !isRef {
				target.Set(val)
			}
		case vm.Exit:
			debugf("}")
			return
		case vm.AppendArray:
			target.Set(reflect.Append(target, reflect.Zero(target.Type().Elem())))
			d.pc++
			d.eval(target.Index(target.Len() - 1))
		case vm.AppendMap:
			d.pc++
			elem := reflect.New(target.Type().Elem()).Elem()
			d.eval(elem)
			if target.IsNil() {
				// TODO we'd like to encode (null | map) by using a nil
				// map value, but because we're only making the map
				// when we append the first element, all empty maps
				// will also be nil. Perhaps when SetLong is called on the
				// union type, we should create the map.
				target.Set(reflect.MakeMap(target.Type()))
			}
			target.SetMapIndex(reflect.ValueOf(frame.String), elem)
		case vm.Call:
			curr := d.pc
			d.pc = inst.Operand
			d.eval(target)
			d.pc = curr
		case vm.Return:
			return
		case vm.Jump:
			d.pc = inst.Operand - 1
		case vm.EvalGreater:
			frame.Condition = frame.Int > int64(inst.Operand)
		case vm.EvalEqual:
			frame.Condition = frame.Int == int64(inst.Operand)
		case vm.CondJump:
			if frame.Condition {
				d.pc = inst.Operand - 1
			}
		case vm.AddLong:
			frame.Int += int64(inst.Operand)
		case vm.SetLong:
			frame.Int = int64(inst.Operand)
		case vm.MultLong:
			frame.Int *= int64(inst.Operand)
		case vm.PushLoop:
			loop := frame.Int
			d.pc++
			d.eval(target)
			frame.Int = loop
		case vm.PopLoop:
			return
		case vm.Halt:
			if inst.Operand == 0 {
				// TODO this doesn't actually halt.
				return
			}
			d.error(fmt.Errorf("Runtime error: %v, frame: %v, pc: %v", d.program.Errors[inst.Operand-1], frame, d.pc))
		default:
			d.error(fmt.Errorf("Unknown instruction %v", d.program.Instructions[d.pc]))
		}
	}
}

func (d *decoder) check(err error, what string) {
	if err != nil {
		d.error(fmt.Errorf("%s: %v", what, err))
	}
}

func (d *decoder) error(err error) {
	panic(&decodeError{
		err: err,
	})
}
