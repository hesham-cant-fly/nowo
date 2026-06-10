package vm

import "fmt"

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

func (vm *VM) Run(main *Chunk) (result Value, err error) {
	defer func() {
		if r := recover(); r != nil {
			result = NilVal()
			err = fmt.Errorf("runtime error: %v", r)
		}
	}()

	proto := &FuncValue{Name: "main", Chunk: main, Env: NewEnv(nil)}
	vm.frames = append(vm.frames, &Frame{fn: proto, stack: make([]Value, 0, 256), env: proto.Env})

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
				panic("+ expects numbers")
			}
			f.push(NumVal(a.Num + b.Num))

		case OP_SUB:
			b := f.pop()
			a := f.pop()
			if a.Type != ValNumber || b.Type != ValNumber {
				panic("- expects numbers")
			}
			f.push(NumVal(a.Num - b.Num))

		case OP_MUL:
			b := f.pop()
			a := f.pop()
			if a.Type != ValNumber || b.Type != ValNumber {
				panic("* expects numbers")
			}
			f.push(NumVal(a.Num * b.Num))

		case OP_DIV:
			b := f.pop()
			a := f.pop()
			if a.Type != ValNumber || b.Type != ValNumber {
				panic("/ expects numbers")
			}
			if b.Num == 0 {
				panic("division by zero")
			}
			f.push(NumVal(a.Num / b.Num))

		case OP_NEG:
			a := f.pop()
			if a.Type != ValNumber {
				panic("- expects a number")
			}
			f.push(NumVal(-a.Num))

		case OP_EQ:
			b := f.pop()
			a := f.pop()
			if a.Num == b.Num {
				f.push(NumVal(1))
			} else {
				f.push(NumVal(0))
			}

		case OP_LT:
			b := f.pop()
			a := f.pop()
			f.push(BoolVal(a.Num < b.Num))

		case OP_GT:
			b := f.pop()
			a := f.pop()
			f.push(BoolVal(a.Num > b.Num))

		case OP_LOAD:
			name := f.fn.Chunk.Names[operand]
			v, ok := f.env.Get(name)
			if !ok {
				panic(fmt.Sprintf("undefined variable %q", name))
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
			if callee.Type != ValFunction {
				panic("not callable")
			}
			fn := callee.Fn
			if nargs != len(fn.Params) {
				panic(fmt.Sprintf("expected %d args, got %d", len(fn.Params), nargs))
			}
			env := NewEnv(fn.Env)
			for i, p := range fn.Params {
				env.SetLocal(p, args[i])
			}
			nf := &Frame{fn: fn, stack: make([]Value, 0, 64), env: env}
			vm.frames = append(vm.frames, nf)

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
			fn := &FuncValue{Name: tmpl.Name, Chunk: tmpl.Chunk, Params: tmpl.Params, Env: f.env}
			f.push(FnVal(fn))

		default:
			panic(fmt.Sprintf("unknown opcode %d", op))
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
