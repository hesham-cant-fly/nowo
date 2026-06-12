package vm

import (
	"fmt"
	"strings"
)

type runtimeErr struct {
	msg  string
	fn   string
	ip   int
	line int
}

type Frame struct {
	fn    *FuncValue
	ip    int
	stack []Value
	env   *Environment
}

func (f *Frame) push(v Value) { f.stack = append(f.stack, v) }
func (f *Frame) pop() Value {
	v := f.stack[len(f.stack)-1]
	f.stack = f.stack[:len(f.stack)-1]
	return v
}
func (f *Frame) peek() Value { return f.stack[len(f.stack)-1] }

type VM struct {
	frames []*Frame
}

func New() *VM { return &VM{} }

func (vm *VM) errorf(f *Frame, msg string, args ...any) {
	ip := f.ip - 1
	line := 0
	if ip >= 0 && ip < len(f.fn.Chunk.Lines) {
		line = f.fn.Chunk.Lines[ip]
	}
	panic(runtimeErr{
		msg:  fmt.Sprintf(msg, args...),
		fn:   f.fn.Name,
		ip:   ip,
		line: line,
	})
}

func (vm *VM) callStack() string {
	var b strings.Builder
	for i := len(vm.frames) - 1; i >= 0; i-- {
		f := vm.frames[i]
		line := 0
		if f.ip > 0 && f.ip-1 < len(f.fn.Chunk.Lines) {
			line = f.fn.Chunk.Lines[f.ip-1]
		}
		if line > 0 {
			fmt.Fprintf(&b, "\n  at %s() [ip=%d, line=%d]", f.fn.Name, f.ip, line)
		} else {
			fmt.Fprintf(&b, "\n  at %s() [ip=%d]", f.fn.Name, f.ip)
		}
	}
	return b.String()
}

func (vm *VM) Run(main *Chunk) (result Value, err error) {
	defer func() {
		if r := recover(); r != nil {
			result = NilVal()
			switch e := r.(type) {
			case runtimeErr:
				loc := ""
				if e.line > 0 {
					loc = fmt.Sprintf("line %d: ", e.line)
				}
				err = fmt.Errorf("runtime error: %s%s%s", loc, e.msg, vm.callStack())
			default:
				err = fmt.Errorf("runtime error: %v", r)
			}
		}
	}()

	global := StdEnv()
	proto := &FuncValue{Name: "main", Chunk: main, Env: global}
	vm.frames = append(vm.frames, &Frame{fn: proto, stack: make([]Value, 0, 256), env: global})

	for {
		f := vm.frames[len(vm.frames)-1]
		if f.ip >= len(f.fn.Chunk.Code) {
			break
		}

		inst := f.fn.Chunk.Code[f.ip]
		op := inst.Opcode()
		operand := inst.Operand()
		f.ip++

		switch op {
		case OP_NIL:
			f.push(NilVal())

		case OP_TRUE:
			f.push(NumVal(1))

		case OP_FALSE:
			f.push(NumVal(0))

		case OP_CONST:
			f.push(f.fn.Chunk.Constants[operand])

		case OP_POP:
			f.pop()

		case OP_ADD:
			b := f.pop()
			a := f.pop()
			if a.Type != ValNumber || b.Type != ValNumber {
				vm.errorf(f, "+ expects numbers, got %s and %s", a, b)
			}
			f.push(NumVal(a.Num + b.Num))

		case OP_SUB:
			b := f.pop()
			a := f.pop()
			if a.Type != ValNumber || b.Type != ValNumber {
				vm.errorf(f, "- expects numbers, got %s and %s", a, b)
			}
			f.push(NumVal(a.Num - b.Num))

		case OP_MUL:
			b := f.pop()
			a := f.pop()
			if a.Type != ValNumber || b.Type != ValNumber {
				vm.errorf(f, "* expects numbers, got %s and %s", a, b)
			}
			f.push(NumVal(a.Num * b.Num))

		case OP_DIV:
			b := f.pop()
			a := f.pop()
			if a.Type != ValNumber || b.Type != ValNumber {
				vm.errorf(f, "/ expects numbers, got %s and %s", a, b)
			}
			if b.Num == 0 {
				vm.errorf(f, "division by zero")
			}
			f.push(NumVal(a.Num / b.Num))

		case OP_NEG:
			a := f.pop()
			if a.Type != ValNumber {
				vm.errorf(f, "- expects a number, got %s", a)
			}
			f.push(NumVal(-a.Num))

		case OP_EQ:
			b := f.pop()
			a := f.pop()
			f.push(BoolVal(a.Num == b.Num))

		case OP_LT:
			b := f.pop()
			a := f.pop()
			f.push(BoolVal(a.Num < b.Num))

		case OP_GT:
			b := f.pop()
			a := f.pop()
			f.push(BoolVal(a.Num > b.Num))

		case OP_LE:
			b := f.pop()
			a := f.pop()
			f.push(BoolVal(a.Num <= b.Num))

		case OP_GE:
			b := f.pop()
			a := f.pop()
			f.push(BoolVal(a.Num >= b.Num))

		case OP_NE:
			b := f.pop()
			a := f.pop()
			f.push(BoolVal(a.Num != b.Num))

		case OP_LEN:
			arr := f.pop()
			if arr.Type != ValArray {
				vm.errorf(f, "# expects an array, got %s", arr)
			}
			f.push(NumVal(float64(len(arr.Arr.Elements))))

		case OP_ISNIL:
			v := f.pop()
			f.push(BoolVal(v.Type == ValNil))

		case OP_NOT:
			a := f.pop()
			f.push(BoolVal(!a.Truthy()))

		case OP_LOAD:
			name := f.fn.Chunk.Names[operand]
			v, ok := f.env.Get(name)
			if !ok {
				vm.errorf(f, "undefined variable %q", name)
			}
			f.push(v)

		case OP_STORE:
			name := f.fn.Chunk.Names[operand]
			v := f.pop()
			f.env.Set(name, v)
			f.push(v)

		case OP_JMP:
			f.ip += inst.Operand()

		case OP_JIF:
			if !f.pop().Truthy() {
				f.ip += inst.Operand()
			}

		case OP_CALL:
			nargs := operand
			args := make([]Value, nargs)
			for i := nargs - 1; i >= 0; i-- {
				args[i] = f.pop()
			}
			callee := f.pop()
			switch callee.Type {
			case ValBuiltin:
				result := callee.Builtin.Fn(args)
				f.push(result)
			case ValFunction:
				fn := callee.Fn
				if nargs > len(fn.Params) {
					vm.errorf(f, "expected at most %d args to %s(), got %d", len(fn.Params), fn.Name, nargs)
				}
				if nargs < fn.HardArity {
					vm.errorf(f, "expected at least %d args to %s(), got %d", fn.HardArity, fn.Name, nargs)
				}
				for i := 0; i < len(fn.Params)-fn.HardArity; i++ {
					args = append(args, NilVal())
				}
				env := NewEnv(fn.Env)
				for i, p := range fn.Params {
					env.SetLocal(p, args[i])
				}
				nf := &Frame{fn: fn, stack: make([]Value, 0, 64), env: env}
				vm.frames = append(vm.frames, nf)
			default:
				vm.errorf(f, "not callable: %s", callee)
			}

		case OP_RET:
			ret := f.pop()
			vm.frames = vm.frames[:len(vm.frames)-1]
			if len(vm.frames) == 0 {
				return ret, nil
			}
			caller := vm.frames[len(vm.frames)-1]
			caller.push(ret)

		case OP_MKFN:
			tmpl := f.fn.Chunk.Constants[operand].Fn
			// fn := &FuncValue{
			// 	Name:      tmpl.Name,
			// 	Chunk:     tmpl.Chunk,
			// 	Params:    tmpl.Params,
			// 	HardArity: tmpl.HardArity,
			// 	Env:       f.env,
			// }
			tmp := *tmpl
			tmp.Env = f.env
			f.push(FnVal(&tmp))

		case OP_ARRAY:
			elems := make([]Value, operand)
			for i := operand - 1; i >= 0; i-- {
				elems[i] = f.pop()
			}
			f.push(ArrVal(&ArrayValue{Elements: elems}))

		case OP_INDEX:
			index := f.pop()
			arr := f.pop()
			if arr.Type != ValArray {
				vm.errorf(f, "index expects an array, got %s", arr)
			}
			if index.Type != ValNumber {
				vm.errorf(f, "index expects a number, got %s", index)
			}
			i := int(index.Num)
			if i >= len(arr.Arr.Elements) {
				vm.errorf(f, "index out of bounds: %d (len %d)", i, len(arr.Arr.Elements))
			}
			if i < 0 {
				f.push(arr.Arr.Elements[len(arr.Arr.Elements)+i])
			} else {
				f.push(arr.Arr.Elements[i])
			}

		case OP_SET_INDEX:
			val := f.pop()
			index := f.pop()
			arr := f.pop()
			if arr.Type != ValArray {
				vm.errorf(f, "index assignment expects an array, got %s", arr)
			}
			if index.Type != ValNumber {
				vm.errorf(f, "index expects a number, got %s", index)
			}
			i := int(index.Num)
			if i < 0 {
				i = len(arr.Arr.Elements) + i
			}
			if i < 0 || i >= len(arr.Arr.Elements) {
				vm.errorf(f, "index out of bounds: %d (len %d)", i, len(arr.Arr.Elements))
			}
			arr.Arr.Elements[i] = val
			f.push(val)

		case OP_CONCAT:
			b := f.pop()
			a := f.pop()
			if a.Type != ValArray || b.Type != ValArray {
				vm.errorf(f, "++ expects arrays, got %s and %s", a, b)
			}
			elems := make([]Value, len(a.Arr.Elements)+len(b.Arr.Elements))
			copy(elems, a.Arr.Elements)
			copy(elems[len(a.Arr.Elements):], b.Arr.Elements)
			f.push(ArrVal(&ArrayValue{Elements: elems}))

		default:
			vm.errorf(f, "unknown opcode %d", op)
		}
	}

	if len(vm.frames) > 0 {
		f := vm.frames[len(vm.frames)-1]
		if len(f.stack) > 0 {
			return f.peek(), nil
		}
	}
	return NilVal(), nil
}
