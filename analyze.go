package avro

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/actgardner/gogen-avro/schema"
	"github.com/actgardner/gogen-avro/vm"
)

type decodeProgram struct {
	vm.Program

	// enter holds an entry for each Enter instruction in the
	// program, indexed by pc, that gets a value that
	// can be assigned to for the given index.
	// It reports whether the returned value is a reference
	// directly into the target value (for example when
	// the target is a struct type).
	enter []func(target reflect.Value) (reflect.Value, bool)

	// makeDefault holds an entry for each SetDefault instruction
	// in the program, indexed by pc, that gets the default
	// value for a field.
	makeDefault []func() interface{}
}

type analyser struct {
	prog        *vm.Program
	pcInfo      []pcInfo
	enter       []func(reflect.Value) (reflect.Value, bool)
	makeDefault []func() interface{}
}

type pcInfo struct {
	// path holds the descent path into the type for an instruction
	// in the program. It has an entry for each Enter
	// (record field or union), AppendArray or AppendMap
	// instruction encountered when executing the VM up
	// until the instruction.
	path []pathElem
}

type pathElem struct {
	// index holds an index into a record or union value.
	// It is zero for map and array types.
	index int
	// ftype holds the type of the value at the given index.
	ftype reflect.Type
	// info holds the type info for the element.
	info azTypeInfo
}

// compileDecoder returns a decoder program to decode into values of the given type
// Avro values encoded with the given writer schema.
func compileDecoder(t reflect.Type, writerType schema.AvroType) (*decodeProgram, error) {

}

// analyzeProgramTypes analyses the given program with
// respect to the given type (the program must have been generated for that
// type) and returns a program with a populated "enter" field allowing
// the VM to correctly create union and field values for Enter instructions.
func analyzeProgramTypes(prog *vm.Program, t reflect.Type) (*decodeProgram, error) {
	a := &analyser{
		prog:        prog,
		pcInfo:      make([]pcInfo, len(prog.Instructions)),
		enter:       make([]func(reflect.Value) (reflect.Value, bool), len(prog.Instructions)),
		makeDefault: make([]func() interface{}, len(prog.Instructions)),
	}
	debugf("analyze %d instructions\n%s {", len(prog.Instructions), prog)
	defer debugf("}")
	info, err := newAzTypeInfo(t)
	if err != nil {
		return nil, err
	}
	if err := a.eval([]int{0}, []pathElem{{
		ftype: t,
		info:  info,
	}}); err != nil {
		return nil, fmt.Errorf("eval: %v", err)
	}
	prog1 := &decodeProgram{
		Program:     *prog,
		enter:       a.enter,
		makeDefault: a.makeDefault,
	}
	// Sanity check that all Enter and SetDefault
	// instructions have associated info.
	for i, inst := range prog.Instructions {
		switch inst.Op {
		case vm.Enter:
			if prog1.enter[i] == nil {
				return nil, fmt.Errorf("enter not set; pc %d; instruction %v", i, inst)
			}
		case vm.SetDefault:
			if prog1.makeDefault[i] == nil {
				return nil, fmt.Errorf("makeDefault not set; pc %d; instruction %v", i, inst)
			}
		}
	}
	return prog1, nil
}

func (a *analyser) eval(stack []int, path []pathElem) (retErr error) {
	debugf("eval %v; path %s{", stack, pathStr(path))
	defer func() {
		debugf("} -> %v", retErr)
	}()
	for {
		pc := stack[len(stack)-1]
		if pc >= len(a.prog.Instructions) {
			break
		}
		if a.pcInfo[pc].path == nil {
			// Update the type info for the current PC with a copy
			// of the current path.
			a.pcInfo[pc].path = append(a.pcInfo[pc].path, path...)
		} else {
			debugf("already evaluated instruction %d", pc)
			// We've already visited this instruction which
			// means we can stop analysing here.
			// Make sure that the path is consistent though,
			// to sanity-check our assumptions about the VM.
			if !equalPathRef(path, a.pcInfo[pc].path) {
				return fmt.Errorf("type mismatch (\n\tprevious %s\n\tnew %s\n)", pathStr(a.pcInfo[pc].path), pathStr(path))
			}
			return nil
		}
		debugf("exec %d: %v", pc, a.prog.Instructions[pc])

		elem := path[len(path)-1]
		switch inst := a.prog.Instructions[pc]; inst.Op {
		case vm.Set:
			// TODO: sanity-check that if it's Set(Bytes), the previous
			// instruction was Read(Bytes) (i.e. frame.Bytes hasn't been invalidated).
			if !canAssignVMType(inst.Operand, elem.ftype) {
				return fmt.Errorf("cannot assign %v to %s", inst.Operand, elem.ftype)
			}
		case vm.Enter:
			elem := &path[len(path)-1]
			index := inst.Operand

			if index >= len(elem.info.entries) {
				return fmt.Errorf("union index out of bounds; pc %d; type %s", pc, elem.ftype)
			}
			info := elem.info.entries[index]
			debugf("enter %d -> %v, %d entries", index, info.ftype, len(info.entries))
			if info.ftype == nil {
				// Special case for the nil value. Return
				// a zero value that will never be used.
				a.enter[pc] = func(v reflect.Value) (reflect.Value, bool) {
					return reflect.Value{}, true
				}
				path = append(path, pathElem{
					index: index,
				})
				break
			}
			var enter func(v reflect.Value) (reflect.Value, bool)
			switch elem.ftype.Kind() {
			case reflect.Struct:
				enter = func(v reflect.Value) (reflect.Value, bool) {
					return v.Field(index), true
				}
			case reflect.Interface:
				enter = func(v reflect.Value) (reflect.Value, bool) {
					return reflect.New(info.ftype).Elem(), false
				}
			case reflect.Ptr:
				if len(elem.info.entries) != 2 {
					return fmt.Errorf("pointer type without a two-member union")
				}
				enter = func(v reflect.Value) (reflect.Value, bool) {
					inner := reflect.New(info.ftype)
					v.Set(inner)
					return inner.Elem(), true
				}
			default:
				return fmt.Errorf("unexpected type in union %T", elem.ftype)
			}
			if len(info.entries) == 0 {
				// The type itself might contribute information.
				info1, err := newAzTypeInfo(info.ftype)
				if err != nil {
					return fmt.Errorf("cannot get info for %s: %v", info.ftype, err)
				}
				info = info1
			}
			path = append(path, pathElem{
				index: index,
				ftype: info.ftype,
				info:  info,
			})
			a.enter[pc] = enter
		case vm.AppendArray:
			if elem.ftype.Kind() != reflect.Slice {
				return fmt.Errorf("cannot append to %T", elem.ftype)
			}
			path = append(path, pathElem{
				ftype: elem.ftype.Elem(),
				info:  elem.info,
			})
		case vm.AppendMap:
			if elem.ftype.Kind() != reflect.Map {
				return fmt.Errorf("cannot append to %T", elem.ftype)
			}
			if elem.ftype.Key().Kind() != reflect.String {
				return fmt.Errorf("invalid key type for map %s", elem.ftype)
			}
			path = append(path, pathElem{
				ftype: elem.ftype.Elem(),
				info:  elem.info,
			})
		case vm.Exit:
			if len(path) == 0 {
				return fmt.Errorf("unbalanced exit")
			}
			path = path[:len(path)-1]
		case vm.SetDefault:
			index := inst.Operand
			if index >= len(elem.info.entries) {
				return fmt.Errorf("set-default index out of bounds; pc %d; type %s", pc, elem.ftype)
			}
			info := elem.info.entries[index]
			if info.makeDefault == nil {
				return fmt.Errorf("no default info found index %d at %v", index, pathStr(path))
			}
			a.makeDefault[pc] = info.makeDefault
		case vm.Call:
			stack = append(stack, inst.Operand-1)
		case vm.Return:
			if len(stack) == 0 {
				return fmt.Errorf("empty stack")
			}
			stack = stack[:len(stack)-1]
		case vm.CondJump:
			debugf("split {")
			// Execute one path of the condition with a forked
			// version of the state before carrying on with the
			// current execution flow.
			stack1 := make([]int, len(stack), cap(stack))
			copy(stack1, stack)
			stack1[len(stack1)-1] = inst.Operand
			path1 := make([]pathElem, len(path), cap(path))
			copy(path1, path)
			if err := a.eval(stack1, path1); err != nil {
				return err
			}
			debugf("}")
		case vm.Jump:
			stack[len(stack)-1] = inst.Operand - 1
		case vm.EvalGreater,
			vm.EvalEqual,
			vm.SetLong,
			vm.AddLong,
			vm.MultLong,
			vm.PushLoop,
			vm.PopLoop,
			vm.Read:
			// We don't care about any of these instructions because
			// they can't influence the types that we're traversing.
		case vm.Halt:
			return nil
		default:
			return fmt.Errorf("unknown instruction %v", inst.Op)
		}
		stack[len(stack)-1]++
	}
	return nil
}

func canAssignVMType(operand int, t reflect.Type) bool {
	// TODO check
	return true
}

func equalPathRef(p1, p2 []pathElem) bool {
	p1 = singleRefPath(p1)
	p2 = singleRefPath(p2)
	if len(p1) != len(p2) {
		return false
	}
	for i := range p1 {
		e1, e2 := &p1[i], &p2[i]
		if e1.ftype != e2.ftype || e1.index != e2.index {
			return false
		}
	}
	return true
}

func singleRefPath(p []pathElem) []pathElem {
	if len(p) == 0 {
		return nil
	}
	refType := p[len(p)-1].info.referenceType
	for i := len(p) - 1; i >= 0; i-- {
		if p[i].info.referenceType != refType {
			return p[i+1:]
		}
	}
	return p
}

func pathStr(ps []pathElem) string {
	var buf strings.Builder
	buf.WriteString("{")
	for i, p := range ps {
		if i > 0 {
			buf.WriteString(", ")
		}
		fmt.Fprintf(&buf, "%d: %s", p.index, p.ftype)
	}
	buf.WriteString("}")
	return buf.String()
}
