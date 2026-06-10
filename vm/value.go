package vm

import "fmt"

type ValueType byte

const (
	ValNil ValueType = iota
	ValNumber
	ValFunction
)

type FuncValue struct {
	Name   string
	Chunk  *Chunk
	Params []string
	Env    *Environment
}

type Value struct {
	Type ValueType
	Num  float64
	Fn   *FuncValue
}

func NilVal() Value             { return Value{Type: ValNil} }
func BoolVal(b bool) Value {
	if b {
		return NumVal(1)
	}
	return NumVal(0)
}
func NumVal(n float64) Value    { return Value{Type: ValNumber, Num: n} }
func FnVal(fn *FuncValue) Value { return Value{Type: ValFunction, Fn: fn} }

func (v Value) String() string {
	switch v.Type {
	case ValNil:
		return "nil"
	case ValNumber:
		return fmt.Sprintf("%g", v.Num)
	case ValFunction:
		return fmt.Sprintf("<fn %s>", v.Fn.Name)
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
