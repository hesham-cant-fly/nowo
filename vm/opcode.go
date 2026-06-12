package vm

import "fmt"

type Opcode byte

const (
	OP_NIL Opcode = iota
	OP_TRUE
	OP_FALSE
	OP_CONST
	OP_POP
	OP_ADD
	OP_SUB
	OP_MUL
	OP_DIV
	OP_NEG
	OP_EQ
	OP_LT
	OP_GT
	OP_LE
	OP_GE
	OP_NE
	OP_ISNIL
	OP_NOT
	OP_LEN
	OP_LOAD
	OP_STORE
	OP_JMP
	OP_JIF
	OP_CALL
	OP_RET
	OP_MKFN
	OP_ARRAY
	OP_INDEX
	OP_SET_INDEX
	OP_CONCAT
)

type Instruction uint32

func MakeInst(op Opcode, operand int) Instruction {
	return Instruction(op) | Instruction(operand&0xffffff)<<8
}

func (i Instruction) Opcode() Opcode { return Opcode(i & 0xff) }
func (i Instruction) Operand() int   { return int(i>>8) & 0xffffff }

func OpName(op Opcode) string {
	names := []string{
		"NIL", "TRUE", "FALSE", "CONST", "POP",
		"ADD", "SUB", "MUL", "DIV", "NEG",
		"EQ", "LT", "GT", "LE", "GE", "NE", "ISNIL", "NOT",
		"LEN", "LOAD", "STORE",
		"JMP", "JIF",
		"CALL", "RET", "MKFN",
		"ARRAY", "INDEX", "SET_INDEX", "CONCAT",
	}
	if int(op) < len(names) {
		return names[op]
	}
	return fmt.Sprintf("OP_%d", op)
}
