package vm

import (
	"fmt"
	"strconv"

	"github.com/hesham-cant-fly/haste-lang/lexer"
	"github.com/hesham-cant-fly/haste-lang/parser"
)

type Compiler struct {
	src   string
	ast   parser.AstProgram
	chunk *Chunk
}

func Compile(file, src string) (*Compiler, error) {
	tokens := lexer.Tokenize(src)
	ast, err := parser.Parse(file, src, tokens)
	if err != nil {
		return nil, err
	}

	bc := &bytecodeCompiler{chunk: NewChunk("main"), env: StdEnv()}
	for i, node := range ast.Nodes {
		bc.compileExpr(node)
		if i < len(ast.Nodes)-1 {
			bc.chunk.EmitSimple(OP_POP, 0)
		}
	}
	bc.chunk.EmitSimple(OP_RET, 0)

	return &Compiler{src: src, ast: ast, chunk: bc.chunk}, nil
}

func (c *Compiler) AST() parser.AstProgram { return c.ast }
func (c *Compiler) Chunk() *Chunk          { return c.chunk }
func (c *Compiler) Source() string         { return c.src }

func (c *Compiler) Run() (Value, error) {
	return New().Run(c.chunk)
}

type bytecodeCompiler struct {
	chunk        *Chunk
	env          *Environment
	funcName     string
	matchCounter int
}

func astLine(n parser.Ast) int {
	switch v := n.(type) {
	case parser.AstNumber:     return v.Line
	case parser.AstIdentifier: return v.Line
	case parser.AstUnary:      return v.Line
	case parser.AstBinary:     return v.Line
	case parser.AstTernary:    return v.Line
	case parser.AstGrouping:   return v.Line
	case parser.AstBind:       return v.Line
	case parser.AstDecl:       return v.Line
	case parser.AstAssign:     return v.Line
	case parser.AstCall:       return v.Line
	case parser.AstArray:      return v.Line
	case parser.AstSubscript:  return v.Line
	case parser.AstFunction:   return v.Line
	case parser.AstMatch:      return v.Line
	case parser.AstPatLiteral: return v.Line
	case parser.AstPatBind:    return v.Line
	case parser.AstPatArray:   return v.Line
	}
	return 0
}

func (c *bytecodeCompiler) compileExpr(node parser.Ast) {
	line := astLine(node)

	switch n := node.(type) {
	case parser.AstNumber:
		val, _ := strconv.ParseFloat(n.Value, 64)
		c.chunk.Emit(OP_CONST, c.chunk.AddConst(NumVal(val)), line)

	case parser.AstIdentifier:
		c.chunk.Emit(OP_LOAD, c.chunk.AddName(n.Value()), line)

	case parser.AstUnary:
		c.compileExpr(n.Rhs)
		switch n.Op {
		case lexer.MINUS:
			c.chunk.EmitSimple(OP_NEG, line)
		case lexer.EXCLAMATION_MARK:
			c.chunk.EmitSimple(OP_NOT, line)
		case lexer.HASHTAG:
			c.chunk.EmitSimple(OP_LEN, line)
		}

	case parser.AstBinary:
		c.compileExpr(n.Lhs)
		c.compileExpr(n.Rhs)
		switch n.Op {
		case lexer.EQEQ:
			c.chunk.EmitSimple(OP_EQ, line)
		case lexer.LESS:
			c.chunk.EmitSimple(OP_LT, line)
		case lexer.GREATER:
			c.chunk.EmitSimple(OP_GT, line)
		case lexer.LESS_EQ:
			c.chunk.EmitSimple(OP_LE, line)
		case lexer.GREATER_EQ:
			c.chunk.EmitSimple(OP_GE, line)
		case lexer.BANG_EQ:
			c.chunk.EmitSimple(OP_NE, line)
		case lexer.PLUS:
			c.chunk.EmitSimple(OP_ADD, line)
		case lexer.MINUS:
			c.chunk.EmitSimple(OP_SUB, line)
		case lexer.STAR:
			c.chunk.EmitSimple(OP_MUL, line)
		case lexer.SLASH:
			c.chunk.EmitSimple(OP_DIV, line)
		case lexer.PLUS_PLUS:
			c.chunk.EmitSimple(OP_CONCAT, line)
		}

	case parser.AstTernary:
		c.compileExpr(n.Cond)
		elseJmp := len(c.chunk.Code)
		c.chunk.Emit(OP_JIF, 0, line)
		c.compileExpr(n.Then)
		endJmp := len(c.chunk.Code)
		c.chunk.Emit(OP_JMP, 0, line)

		c.chunk.Code[elseJmp] = MakeInst(OP_JIF, len(c.chunk.Code)-(elseJmp+1))

		c.compileExpr(n.Else)
		c.chunk.Code[endJmp] = MakeInst(OP_JMP, len(c.chunk.Code)-(endJmp+1))

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
				c.chunk.EmitSimple(OP_POP, line)
			}
		}

	case parser.AstDecl:
		name := n.Lhs.(parser.AstIdentifier).Value()
		oldName := c.funcName
		c.funcName = name
		c.compileExpr(n.Value)
		c.funcName = oldName
		c.chunk.Emit(OP_STORE, c.chunk.AddName(name), line)

	case parser.AstAssign:
		switch lhs := n.Lhs.(type) {
		case parser.AstIdentifier:
			c.compileExpr(n.Value)
			c.chunk.Emit(OP_STORE, c.chunk.AddName(lhs.Value()), line)
		case parser.AstSubscript:
			c.compileExpr(lhs.Array)
			c.compileExpr(lhs.Index)
			c.compileExpr(n.Value)
			c.chunk.EmitSimple(OP_SET_INDEX, line)
		}

	case parser.AstCall:
		c.compileExpr(n.Callee)
		for _, arg := range n.Args {
			c.compileExpr(arg)
		}
		c.chunk.Emit(OP_CALL, len(n.Args), line)

	case parser.AstArray:
		for _, elem := range n.Elements {
			c.compileExpr(elem)
		}
		c.chunk.Emit(OP_ARRAY, len(n.Elements), line)

	case parser.AstSubscript:
		c.compileExpr(n.Array)
		c.compileExpr(n.Index)
		c.chunk.EmitSimple(OP_INDEX, line)

	case parser.AstFunction:
		fnName := c.funcName
		if fnName == "" {
			fnName = "_fn"
		}
		fnChunk := NewChunk(fnName)
		fnCompiler := &bytecodeCompiler{chunk: fnChunk, env: c.env, funcName: fnName}
		hardArity := 0

		for _, arg := range n.Args {
			nm := fnChunk.AddName(arg.Name)

			if arg.DefaultValue != nil {
				fnChunk.Emit(OP_LOAD, nm, line)
				fnChunk.EmitSimple(OP_ISNIL, line)
				elseJmp := len(fnChunk.Code)
				fnChunk.Emit(OP_JIF, 0, line)
				fnCompiler.compileExpr(arg.DefaultValue)
				fnChunk.Emit(OP_STORE, nm, line)

				fnChunk.Code[elseJmp] = MakeInst(OP_JIF, len(fnChunk.Code)-(elseJmp+1))
				continue
			}
			hardArity += 1
		}

		fnCompiler.compileExpr(n.Body)
		fnChunk.EmitSimple(OP_RET, line)

		proto := &FuncValue{
			Name:      fnName,
			Chunk:     fnChunk,
			HardArity: hardArity,
			Params:    paramNames(n.Args),
		}
		idx := c.chunk.AddConst(FnVal(proto))
		c.chunk.Emit(OP_MKFN, idx, line)
		c.chunk.AddSub(fnChunk)

	case parser.AstMatch:
		c.compileExpr(n.Scrutinee)
		matchName := c.chunk.AddName(fmt.Sprintf("__match_%d", c.matchCounter))
		c.chunk.Emit(OP_STORE, matchName, line)
		c.matchCounter += 1
		c.chunk.EmitSimple(OP_POP, line)

		var endsIndicies []int
		for _, pattern := range n.Arms {
			elseIndicies := c.compilePatternCheck(pattern.Pattern, matchName)

			c.compileExpr(pattern.Body)

			endIndex := len(c.chunk.Code)
			c.chunk.Emit(OP_JMP, 0, line)
			endsIndicies = append(endsIndicies, endIndex)

			if len(elseIndicies) == 0 {
				continue
			}

			for _, elseIndex := range elseIndicies {
				c.chunk.Code[elseIndex] = MakeInst(OP_JIF, len(c.chunk.Code)-(elseIndex + 1))
			}
		}

		for _, endIndex := range endsIndicies {
			ln := len(c.chunk.Code)
			c.chunk.Code[endIndex] = MakeInst(OP_JMP, ln - endIndex)
		}

		// the default branch
		c.chunk.EmitSimple(OP_NIL, line)
	}
}

func (c *bytecodeCompiler) compilePatternCheck(pattern parser.Ast, matchName int) (elseIndex []int) {
	line := astLine(pattern)
	if matchName != -1 {
		c.chunk.Emit(OP_LOAD, matchName, line)
	}

	switch pattern := pattern.(type) {
	case parser.AstPatLiteral:
		// c.chunk.EmitSimple(OP_DUP, line)
		c.compileExpr(pattern.Value)
		c.chunk.EmitSimple(OP_EQ, line)
		elseIndex = append(elseIndex, len(c.chunk.Code))
		c.chunk.Emit(OP_JIF, 0, line)

	case parser.AstPatBind:
		// c.chunk.EmitSimple(OP_DUP, line)
		name := c.chunk.AddName(pattern.Name)
		c.chunk.Emit(OP_STORE, name, line)
		c.chunk.EmitSimple(OP_POP, line)

	case parser.AstPatArray:
		c.chunk.EmitSimple(OP_DUP, line)
		c.chunk.EmitSimple(OP_LEN, line)
		elementsLen := len(pattern.Elements)

		// NOTE: exact vs fixed

		if pattern.HasRest {
			c.chunk.Emit(OP_CONST, c.chunk.AddConst(NumVal(float64(elementsLen - 1))), line)
			c.chunk.EmitSimple(OP_GE, line)
		} else {
			c.chunk.Emit(OP_CONST, c.chunk.AddConst(NumVal(float64(elementsLen))), line)
			c.chunk.EmitSimple(OP_EQ, line)
		}

		elseIndex = append(elseIndex, len(c.chunk.Code))
		c.chunk.Emit(OP_JIF, 0, line)

		for i, element := range pattern.Elements {
			c.chunk.EmitSimple(OP_DUP, line)
			if element.Rest {
				start := c.chunk.AddConst(NumVal(float64(i)))
				c.chunk.Emit(OP_CONST, start, line)
				end := c.chunk.AddConst(NumVal(float64(-1)))
				c.chunk.Emit(OP_CONST, end, line)

				c.chunk.EmitSimple(OP_SLICE, line)

				for _, x := range c.compilePatternCheck(element.Pattern, -1) {
					elseIndex = append(elseIndex, x)
				}
				break
			}

			c.chunk.Emit(OP_CONST, c.chunk.AddConst(NumVal(float64(i))), line)
			c.chunk.EmitSimple(OP_INDEX, line)

			for _, x := range c.compilePatternCheck(element.Pattern, -1) {
				elseIndex = append(elseIndex, x)
			}
		}

	default: panic("oops.. this is not supposed to happen :P")
	}

	// c.chunk.EmitSimple(OP_POP, line)


	return
}

func paramNames(args []parser.AstFunctionArg) []string {
	names := make([]string, len(args))
	for i, a := range args {
		names[i] = a.Name
	}
	return names
}
