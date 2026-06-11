package parser

import (
	"errors"
	"fmt"

	"github.com/hesham-cant-fly/haste-lang/lexer"
)

type (
	prefixFn func(*parser) (Ast, error)
	infixFn  func(*parser, Ast) (Ast, error)
)

type rule struct {
	prefix     prefixFn
	infix      infixFn
	prec       int
	rightAssoc bool
}

const (
	PREC_NONE      = iota
	PREC_BIND      // ;
	PREC_ASSIGN    // = :=
	PREC_TERNARY   // ? !
	PREC_CMP       // == != < > <= >=
	PREC_TERM      // + -
	PREC_CONCAT    // ++
	PREC_FACTOR    // * /
	PREC_SUBSCRIPT // [
	PREC_PIPE      // :
	PREC_ACCESS    // .
	PREC_UNARY     // prefix - +
	PREC_PRIMARY
)
const PREC_LOWEST = PREC_BIND

type parser struct {
	tokens   []lexer.Token
	current  int
	previous lexer.Token
	hasError bool
}

func Parse(tokens []lexer.Token) (AstProgram, error) {
	p := &parser{tokens: tokens}
	var program []Ast

	for !p.ended() {
		node, err := p.parseNode()
		if err != nil {
			fmt.Println(err)
			p.hasError = true
			p.advance()
			continue
		}
		program = append(program, node)
	}

	if p.hasError {
		return AstProgram{}, errors.New("Cannot parse this source code")
	}

	return AstProgram{Nodes: program}, nil
}

func (p *parser) parseNode() (Ast, error) {
	// if p.check(lexer.IDENT) && p.checkNext(lexer.OPEN_PAREN) {
	// 	return p.parseFunctionDecl()
	// }

	// if p.check(lexer.IDENT) && p.checkNext(lexer.EQ) {
	// 	name := p.advance()
	// 	p.advance()
	// 	value, err := p.parseExpr(PREC_TERM)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	if p.match(lexer.SEMICOLON) {
	// 	}
	// 	return AstDecl{Name: name.Lexem, Value: value}, nil
	// }

	expr, err := p.parseExpr(PREC_LOWEST)
	if err != nil {
		return nil, err
	}
	// if p.match(lexer.SEMICOLON) {
	// }
	return expr, nil
}

func parseFunction(p *parser) (Ast, error) {
	args, err := p.parseArguments()
	if err != nil {
		return nil, err
	}

	if !p.match(lexer.CLOSE_PAREN) {
		return nil, errors.New("expected ) after parameters")
	}

	body, err := p.parseExpr(PREC_ASSIGN)
	if err != nil {
		return nil, err
	}

	return AstFunction{Args: args, Body: body}, nil
}

func (p *parser) parseArguments() ([]AstFunctionArg, error) {
	var args []AstFunctionArg

	for !p.check(lexer.CLOSE_PAREN) && !p.ended() {
		tok := p.advance()
		if tok.Kind != lexer.IDENT {
			return nil, errors.New("expected parameter name")
		}

		var defaultVal Ast
		if p.match(lexer.EQ) {
			defaultVal, _ = p.parseExpr(PREC_LOWEST)
		}

		args = append(args, AstFunctionArg{Name: tok.Lexem, DefaultValue: defaultVal})

		if p.check(lexer.COMMA) {
			p.advance()
		}
	}

	return args, nil
}

func (p *parser) parseExpr(prec int) (Ast, error) {
	if p.ended() {
		return nil, errors.New("expected expression")
	}

	tok := p.advance()

	prefixRule := getRule(tok.Kind).prefix
	if prefixRule == nil {
		return nil, fmt.Errorf("expected expression, got '%s'", tok.Lexem)
	}

	left, err := prefixRule(p)
	if err != nil {
		return nil, err
	}

	for !p.ended() {
		r := getRule(p.peek().Kind)
		if prec > r.prec || r.infix == nil {
			break
		}
		p.advance()
		left, err = r.infix(p, left)
		if err != nil {
			return nil, err
		}
	}

	return left, nil
}

func getRule(kind lexer.TokenKind) rule {
	switch kind {
	case lexer.SEMICOLON:
		return rule{infix: parseBinding, prec: PREC_BIND}
	case lexer.HASHTAG:
		return rule{prefix: parseUnary, prec: PREC_UNARY}
	case lexer.COLON_EQ:
		return rule{infix: parseDecl, prec: PREC_ASSIGN}
	case lexer.COLON:
		return rule{infix: parsePipe, prec: PREC_PIPE}
	case lexer.OPEN_PAREN:
		return rule{prefix: parseFunction, infix: parseCall, prec: PREC_PRIMARY}
	case lexer.OPEN_BRACE:
		return rule{prefix: parseGrouped, prec: PREC_PRIMARY}
	case lexer.QUESTION_MARK:
		return rule{infix: parseTernary, prec: PREC_TERNARY, rightAssoc: true}
	case lexer.EQEQ:
		return rule{infix: parseBinary, prec: PREC_CMP}
	case lexer.LESS:
		return rule{infix: parseBinary, prec: PREC_CMP}
	case lexer.GREATER:
		return rule{infix: parseBinary, prec: PREC_CMP}
	case lexer.LESS_EQ:
		return rule{infix: parseBinary, prec: PREC_CMP}
	case lexer.GREATER_EQ:
		return rule{infix: parseBinary, prec: PREC_CMP}
	case lexer.BANG_EQ:
		return rule{infix: parseBinary, prec: PREC_CMP}
	case lexer.EXCLAMATION_MARK:
		return rule{prefix: parseUnary, prec: PREC_UNARY}
	case lexer.PLUS:
		return rule{prefix: parseUnary, infix: parseBinary, prec: PREC_TERM}
	case lexer.MINUS:
		return rule{prefix: parseUnary, infix: parseBinary, prec: PREC_TERM}
	case lexer.STAR:
		return rule{infix: parseBinary, prec: PREC_FACTOR}
	case lexer.SLASH:
		return rule{infix: parseBinary, prec: PREC_FACTOR}
	case lexer.PLUS_PLUS:
		return rule{infix: parseBinary, prec: PREC_CONCAT}
	case lexer.DOT:
		return rule{prefix: parseArrayLiteral, prec: PREC_PRIMARY}
	case lexer.OPEN_BRACKET:
		return rule{infix: parseSubscript, prec: PREC_SUBSCRIPT}
	case lexer.IDENT:
		return rule{prefix: parseIdent, prec: PREC_PRIMARY}
	case lexer.INT:
		return rule{prefix: parseNumber, prec: PREC_PRIMARY}
	case lexer.FLOAT:
		return rule{prefix: parseNumber, prec: PREC_PRIMARY}
	default:
		return rule{}
	}
}

func parseBinding(p *parser, lhs Ast) (Ast, error) {
	rhs, err := p.parseExpr(PREC_BIND)
	if err != nil {
		return nil, err
	}

	return AstBind{
		Lhs: lhs,
		Rhs: rhs,
	}, nil
}

func parseDecl(p *parser, lhs Ast) (Ast, error) {
	if _, ok := lhs.(AstIdentifier); !ok {
		return nil, fmt.Errorf("Expected an identifier here")
	}

	value, err := p.parseExpr(PREC_ASSIGN)
	if err != nil {
		return nil, err
	}

	return AstDecl{
		Lhs:   lhs,
		Value: value,
	}, nil
}

func parsePipe(p *parser, lhs Ast) (Ast, error) {
	rhs, err := p.parseExpr(PREC_PIPE + 1)
	if err != nil {
		return nil, err
	}

	switch r := rhs.(type) {
	case AstCall:
		return AstCall{Callee: r.Callee, Args: append([]Ast{lhs}, r.Args...)}, nil
	default:
		return AstCall{Callee: rhs, Args: []Ast{lhs}}, nil
	}
}

func parseArrayLiteral(p *parser) (Ast, error) {
	if !p.match(lexer.OPEN_BRACKET) {
		return nil, fmt.Errorf("expected '[' after '.'")
	}
	var elements []Ast
	for !p.check(lexer.CLOSE_BRACKET) && !p.ended() {
		elem, err := p.parseExpr(PREC_LOWEST)
		if err != nil {
			return nil, err
		}
		elements = append(elements, elem)
		if !p.check(lexer.CLOSE_BRACKET) && !p.ended() {
			if !p.match(lexer.COMMA) {
				return nil, fmt.Errorf("expected ',' or ']' in array literal")
			}
		}
	}
	if !p.match(lexer.CLOSE_BRACKET) {
		return nil, fmt.Errorf("expected ']'")
	}
	return AstArray{Elements: elements}, nil
}

func parseSubscript(p *parser, lhs Ast) (Ast, error) {
	index, err := p.parseExpr(PREC_LOWEST)
	if err != nil {
		return nil, err
	}
	if !p.match(lexer.CLOSE_BRACKET) {
		return nil, fmt.Errorf("expected ']'")
	}
	return AstSubscript{Array: lhs, Index: index}, nil
}

func parseIdent(p *parser) (Ast, error) {
	return AstIdentifier{value: p.previous.Lexem}, nil
}

func parseNumber(p *parser) (Ast, error) {
	return AstNumber{Value: p.previous.Lexem}, nil
}

func parseGrouped(p *parser) (Ast, error) {
	expr, err := p.parseExpr(PREC_LOWEST)
	if err != nil {
		return nil, err
	}
	if !p.match(lexer.CLOSE_BRACE) {
		return nil, errors.New("expected }")
	}
	return AstGrouping{Child: expr}, nil
}

func parseTernary(p *parser, cond Ast) (Ast, error) {
	// op := p.previous.Kind
	// prec := getRule(op).prec
	then, err := p.parseExpr(PREC_LOWEST)
	if err != nil {
		return nil, err
	}

	if !p.match(lexer.EXCLAMATION_MARK) {
		return nil, fmt.Errorf("Expected `!`")
	}

	else_, err := p.parseExpr(PREC_LOWEST)
	if err != nil {
		return nil, err
	}

	return AstTernary{
		Cond: cond,
		Then: then,
		Else: else_,
	}, nil
}

func parseUnary(p *parser) (Ast, error) {
	op := p.previous.Kind
	rhs, err := p.parseExpr(PREC_UNARY)
	if err != nil {
		return nil, err
	}
	return AstUnary{Op: op, Rhs: rhs}, nil
}

func parseBinary(p *parser, left Ast) (Ast, error) {
	op := p.previous.Kind
	r := getRule(op)
	prec := r.prec + 1
	if r.rightAssoc {
		prec = r.prec
	}
	right, err := p.parseExpr(prec)
	if err != nil {
		return nil, err
	}
	return AstBinary{Lhs: left, Rhs: right, Op: op}, nil
}

// func parseAccess(p *parser, left Ast) (Ast, error) {
// 	tok := p.peek()
// 	if !p.match(lexer.IDENT) {
// 		return nil, errors.New("expected field name after .")
// 	}
// 	return AstAccess{Lhs: left, Field: tok.Lexem}, nil
// }

func parseCall(p *parser, callee Ast) (Ast, error) {
	var args []Ast
	for !p.check(lexer.CLOSE_PAREN) && !p.ended() {
		arg, err := p.parseExpr(PREC_LOWEST)
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
		if p.check(lexer.COMMA) {
			p.advance()
		} else if !p.check(lexer.CLOSE_PAREN) {
			p.advance()
		}
	}
	if !p.match(lexer.CLOSE_PAREN) {
		return nil, errors.New("expected ) after arguments")
	}
	return AstCall{Callee: callee, Args: args}, nil
}

func (p *parser) advance() lexer.Token {
	if p.ended() {
		return lexer.Token{}
	}
	tok := p.tokens[p.current]
	p.current++
	p.previous = tok
	return tok
}

func (p *parser) peek() lexer.Token {
	if p.ended() {
		return lexer.Token{}
	}
	return p.tokens[p.current]
}

func (p *parser) check(kind lexer.TokenKind) bool {
	return !p.ended() && p.tokens[p.current].Kind == kind
}

func (p *parser) checkNext(kind lexer.TokenKind) bool {
	if p.ended() || p.current+1 >= len(p.tokens) {
		return false
	}
	return p.tokens[p.current+1].Kind == kind
}

func (p *parser) match(kind lexer.TokenKind) bool {
	if p.check(kind) {
		p.advance()
		return true
	}
	return false
}

func (p *parser) ended() bool {
	return p.current >= len(p.tokens)
}
