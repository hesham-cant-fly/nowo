package vm

import (
	"fmt"
	"strings"
)

type Chunk struct {
	Code      []Instruction
	Constants []Value
	Names     []string
	Name      string
	Subs      []*Chunk
}

func NewChunk(name string) *Chunk {
	return &Chunk{Name: name}
}

func (c *Chunk) AddSub(sub *Chunk) {
	c.Subs = append(c.Subs, sub)
}

func (c *Chunk) Emit(op Opcode, operand int) {
	c.Code = append(c.Code, MakeInst(op, operand))
}

func (c *Chunk) EmitSimple(op Opcode) {
	c.Code = append(c.Code, MakeInst(op, 0))
}

func (c *Chunk) AddConst(v Value) int {
	c.Constants = append(c.Constants, v)
	return len(c.Constants) - 1
}

func (c *Chunk) AddName(name string) int {
	for i, n := range c.Names {
		if n == name {
			return i
		}
	}
	c.Names = append(c.Names, name)
	return len(c.Names) - 1
}

func (c *Chunk) Disassemble() string { return c.disassemble("") }

func (c *Chunk) disassemble(indent string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s== %s (%d inst, %d const, %d names) ==\n",
		indent, c.Name, len(c.Code), len(c.Constants), len(c.Names))

	for i, k := range c.Constants {
		if k.Type == ValFunction {
			fmt.Fprintf(&b, "%s  const %d: <fn %s>\n", indent, i, k.Fn.Name)
		} else {
			fmt.Fprintf(&b, "%s  const %d: %s\n", indent, i, k)
		}
	}
	for i, n := range c.Names {
		fmt.Fprintf(&b, "%s  name %d: %q\n", indent, i, n)
	}

	for i := 0; i < len(c.Code); i++ {
		inst := c.Code[i]
		op := inst.Opcode()
		operand := inst.Operand()

		fmt.Fprintf(&b, "%s%4d  ", indent, i)

		switch op {
		case OP_CONST:
			fmt.Fprintf(&b, "%-6s %d (%s)\n", OpName(op), operand, c.Constants[operand])
		case OP_LOAD, OP_STORE:
			fmt.Fprintf(&b, "%-6s %d (%q)\n", OpName(op), operand, c.Names[operand])
		case OP_JMP, OP_JIF:
			target := i + 1 + operand
			fmt.Fprintf(&b, "%-6s %d (-> %d)\n", OpName(op), operand, target)
		case OP_CALL, OP_MKFN, OP_ARRAY:
			fmt.Fprintf(&b, "%-6s %d\n", OpName(op), operand)
		default:
			fmt.Fprintf(&b, "%s\n", OpName(op))
		}
	}

	for _, sub := range c.Subs {
		b.WriteByte('\n')
		b.WriteString(sub.disassemble(indent))
	}
	return b.String()
}
