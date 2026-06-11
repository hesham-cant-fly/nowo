package vm

import (
	"fmt"
	"strings"
)

type ValueType byte

const (
	ValNil ValueType = iota
	ValNumber
	ValFunction
	ValBuiltin
	ValArray
)

type FuncValue struct {
	Name      string
	Chunk     *Chunk
	HardArity int
	Params    []string
	Env       *Environment
}

type BuiltinFunc func(args []Value) Value

type BuiltinValue struct {
	Name string
	Fn   BuiltinFunc
}

type ArrayValue struct {
	Elements []Value
}

type Value struct {
	Type    ValueType
	Num     float64
	Fn      *FuncValue
	Builtin *BuiltinValue
	Arr     *ArrayValue
}

func NilVal() Value { return Value{Type: ValNil} }
func BoolVal(b bool) Value {
	if b {
		return NumVal(1)
	}
	return NumVal(0)
}
func NumVal(n float64) Value           { return Value{Type: ValNumber, Num: n} }
func FnVal(fn *FuncValue) Value        { return Value{Type: ValFunction, Fn: fn} }
func BuiltinVal(b *BuiltinValue) Value { return Value{Type: ValBuiltin, Builtin: b} }
func ArrVal(a *ArrayValue) Value       { return Value{Type: ValArray, Arr: a} }

func (v Value) String() string {
	switch v.Type {
	case ValNil:
		return "nil"
	case ValNumber:
		return fmt.Sprintf("%g", v.Num)
	case ValFunction:
		return fmt.Sprintf("<fn %s>", v.Fn.Name)
	case ValBuiltin:
		return fmt.Sprintf("<builtin %s>", v.Builtin.Name)
	case ValArray:
		var b strings.Builder
		b.WriteByte('[')
		for i, e := range v.Arr.Elements {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(e.String())
		}
		b.WriteByte(']')
		return b.String()
	default:
		return "?"
	}
}

func (v Value) Truthy() bool {
	switch v.Type {
	case ValNil:
		return false
	case ValNumber:
		return v.Num > 0
	default:
		return true
	}
}

type Environment struct {
	vars   map[string]Value
	parent *Environment
}

func NewEnv(parent *Environment) *Environment {
	return &Environment{vars: make(map[string]Value), parent: parent}
}

func (e *Environment) Get(name string) (Value, bool) {
	for s := e; s != nil; s = s.parent {
		if v, ok := s.vars[name]; ok {
			return v, true
		}
	}
	return NilVal(), false
}

func (e *Environment) Set(name string, val Value) {
	for s := e; s != nil; s = s.parent {
		if _, ok := s.vars[name]; ok {
			s.vars[name] = val
			return
		}
	}
	e.vars[name] = val
}

func (e *Environment) SetLocal(name string, val Value) {
	e.vars[name] = val
}

func StdEnv() *Environment {
	env := NewEnv(nil)
	env.SetLocal("nil", NilVal())
	env.SetLocal("print", BuiltinVal(&BuiltinValue{
		Name: "print",
		Fn: func(args []Value) Value {
			for i, arg := range args {
				if i > 0 {
					fmt.Print(" ")
				}
				fmt.Print(arg)
			}
			fmt.Println()
			return NilVal()
		},
	}))
	return env
}
