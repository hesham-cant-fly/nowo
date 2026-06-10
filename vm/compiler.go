package vm

import (
	"strconv"

	"github.com/hesham-cant-fly/haste-lang/lexer"
	"github.com/hesham-cant-fly/haste-lang/parser"
)

type Compiler struct {
	src   string
	ast   parser.AstProgram
	chunk *Chunk
}

func Compile(src string) (*Compiler, error) {
	tokens := lexer.Tokenize(src)
	ast, err := parser.Parse(tokens)
	if err != nil {
		return nil, err
	}

	bc := &bytecodeCompiler{chunk: NewChunk("main"), env: NewEnv(nil)}
	for i, node := range ast.Nodes {
		bc.compileExpr(node)
		if i < len(ast.Nodes)-1 {
			bc.chunk.EmitSimple(OP_POP)
		}
	}
	bc.chunk.EmitSimple(OP_RET)

	return &Compiler{src: src, ast: ast, chunk: bc.chunk}, nil
}

func (c *Compiler) AST() parser.AstProgram { return c.ast }
func (c *Compiler) Chunk() *Chunk          { return c.chunk }
func (c *Compiler) Source() string         { return c.src }

func (c *Compiler) Run() (Value, error) {
	return New().Run(c.chunk)
}

type bytecodeCompiler struct {
	chunk    *Chunk
	env      *Environment
	funcName string
}

func (c *bytecodeCompiler) compileExpr(node parser.Ast) {
	switch n := node.(type) {
	case parser.AstNumber:
		val, _ := strconv.ParseFloat(n.Value, 64)
		c.chunk.Emit(OP_CONST, c.chunk.AddConst(NumVal(val)))

	case parser.AstIdentifier:
		c.chunk.Emit(OP_LOAD, c.chunk.AddName(n.Value()))

	case parser.AstUnary:
		c.compileExpr(n.Rhs)
		switch n.Op {
		case lexer.MINUS:
			c.chunk.EmitSimple(OP_NEG)
		}

	case parser.AstBinary:
		c.compileExpr(n.Lhs)
		c.compileExpr(n.Rhs)
		switch n.Op {
		case lexer.EQEQ:    c.chunk.EmitSimple(OP_EQ)
		case lexer.LESS:    c.chunk.EmitSimple(OP_LT)
		case lexer.GREATER: c.chunk.EmitSimple(OP_GT)
		case lexer.PLUS:    c.chunk.EmitSimple(OP_ADD)
		case lexer.MINUS:   c.chunk.EmitSimple(OP_SUB)
		case lexer.STAR:    c.chunk.EmitSimple(OP_MUL)
		case lexer.SLASH:   c.chunk.EmitSimple(OP_DIV)
		}

	case parser.AstTernary:
		c.compileExpr(n.Cond)
		elseJmp := len(c.chunk.Code)
		c.chunk.Emit(OP_JIF, 0) // placeholder
		c.compileExpr(n.Then)
		endJmp := len(c.chunk.Code)
		c.chunk.Emit(OP_JMP, 0)

		c.chunk.Code[elseJmp] = MakeInst(OP_JIF, len(c.chunk.Code) - (elseJmp + 1))

		c.compileExpr(n.Else)
		c.chunk.Code[endJmp] = MakeInst(OP_JMP, len(c.chunk.Code) - (endJmp + 1))

	case parser.AstGrouping:
		c.compileExpr(n.Child)

	case parser.AstBind:
		var exprs []parser.Ast
		var walk func(parser.Ast)
		walk = func(n parser.Ast) {
			switch b := n.(type) {
			case parser.AstBind:
				walk(b.Lhs)
				walk(b.Rhs)
			default:
				exprs = append(exprs, n)
			}
		}
		walk(n)
		for i, expr := range exprs {
			c.compileExpr(expr)
			if i < len(exprs)-1 {
				c.chunk.EmitSimple(OP_POP)
			}
		}

	case parser.AstDecl:
		name := n.Lhs.(parser.AstIdentifier).Value()
		oldName := c.funcName
		c.funcName = name
		c.compileExpr(n.Value)
		c.funcName = oldName
		c.chunk.Emit(OP_STORE, c.chunk.AddName(name))

	case parser.AstAssign:
		c.compileExpr(n.Value)
		name := n.Lhs.(parser.AstIdentifier).Value()
		c.chunk.Emit(OP_STORE, c.chunk.AddName(name))

	case parser.AstCall:
		c.compileExpr(n.Callee)
		for _, arg := range n.Args {
			c.compileExpr(arg)
		}
		c.chunk.Emit(OP_CALL, len(n.Args))

	case parser.AstFunction:
		fnName := c.funcName
		if fnName == "" {
			fnName = "_fn"
		}
		fnChunk := NewChunk(fnName)
		fnCompiler := &bytecodeCompiler{chunk: fnChunk, env: c.env, funcName: fnName}

		for _, arg := range n.Args {
			fnChunk.AddName(arg.Name)
		}
		fnCompiler.compileExpr(n.Body)
		fnChunk.EmitSimple(OP_RET)

		proto := &FuncValue{
			Name:   fnName,
			Chunk:  fnChunk,
			Params: paramNames(n.Args),
		}
		idx := c.chunk.AddConst(FnVal(proto))
		c.chunk.Emit(OP_MKFN, idx)
		c.chunk.AddSub(fnChunk)
	}
}

func paramNames(args []parser.AstFunctionArg) []string {
	names := make([]string, len(args))
	for i, a := range args {
		names[i] = a.Name
	}
	return names
}
