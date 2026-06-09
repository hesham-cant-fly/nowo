package parser

import "github.com/hesham-cant-fly/haste-lang/lexer"

type Ast interface{}

type AstProgram struct {
	Nodes []Ast
}

type AstIdentifier struct {
	value string
}

type AstDecl struct {
	Lhs   Ast
	Value Ast
}

type AstAssign struct {
	Lhs   Ast
	Value Ast
}

type AstFunction struct {
	Args []AstFunctionArg
	Body Ast
}

type AstFunctionArg struct {
	Name         string
	DefaultValue Ast
}

type AstBind struct {
	Lhs Ast
	Rhs Ast
}

type AstGrouping struct {
	Child Ast
}

type AstBinary struct {
	Lhs Ast
	Rhs Ast
	Op  lexer.TokenKind
}

type AstUnary struct {
	Rhs Ast
	Op  lexer.TokenKind
}

type AstAccess struct {
	Lhs   Ast
	Field string
}

type AstCall struct {
	Callee Ast
	Args   []Ast
}

type AstNumber struct {
	Value string
}
