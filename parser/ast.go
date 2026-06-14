package parser

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/hesham-cant-fly/haste-lang/lexer"
)

type Ast interface{}

type AstProgram struct {
	Nodes []Ast
}

type AstIdentifier struct {
	value string
	Line  int
}

func (n AstIdentifier) Value() string { return n.value }

type AstDecl struct {
	Lhs   Ast `ast:"child"`
	Value Ast `ast:"child"`
	Line  int
}

type AstAssign struct {
	Lhs   Ast `ast:"child"`
	Value Ast `ast:"child"`
	Line  int
}

type AstFunction struct {
	Args []AstFunctionArg
	Body Ast
	Line int
}

type AstFunctionArg struct {
	Name         string
	DefaultValue Ast
}

type AstBind struct {
	Lhs  Ast
	Rhs  Ast
	Line int
}

type AstTernary struct {
	Cond Ast `ast:"child"`
	Then Ast `ast:"child"`
	Else Ast `ast:"child"`
	Line int
}

type AstGrouping struct {
	Child Ast `ast:"child"`
	Line  int
}

type AstBinary struct {
	Lhs  Ast             `ast:"child"`
	Rhs  Ast             `ast:"child"`
	Op   lexer.TokenKind `ast:"label"`
	Line int
}

type AstUnary struct {
	Rhs  Ast             `ast:"child"`
	Op   lexer.TokenKind `ast:"label"`
	Line int
}

type AstAccess struct {
	Lhs   Ast    `ast:"child"`
	Field string `ast:"label"`
	Line  int
}

type AstCall struct {
	Callee Ast
	Args   []Ast
	Line   int
}

type AstArray struct {
	Elements []Ast
	Line     int
}

type AstMatch struct {
	Scrutinee Ast `ast:"child"`
	Arms      []AstMatchArm
	Line      int
}

type AstMatchArm struct {
	Pattern Ast `ast:"child"`
	Body    Ast `ast:"child"`
}

type AstPatLiteral struct {
	Value Ast
	Line  int
}

type AstPatBind struct {
	Name string
	Line int
}

type AstPatArray struct {
	Elements []AstPatArrayElement
	HasRest  bool
	Line     int
}

type AstPatArrayElement struct {
	Pattern Ast
	Rest    bool
}

func (n AstArray) FormatAst(indent string) string {
	var b strings.Builder
	b.WriteString("Array\n")
	for i, elem := range n.Elements {
		prefix := "├─ "
		childPrefix := indent + "│  "
		if i == len(n.Elements)-1 {
			prefix = "└─ "
			childPrefix = indent + "   "
		}
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(indent + prefix)
		b.WriteString(formatAst(elem, childPrefix))
	}
	return b.String()
}

func (n AstMatch) FormatAst(indent string) string {
	var b strings.Builder
	b.WriteString("Match\n")
	b.WriteString(indent + "├─ ")
	b.WriteString(formatAst(n.Scrutinee, indent+"│  "))
	for i, arm := range n.Arms {
		armPrefix := "├─ "
		armChildIndent := indent + "│  "
		if i == len(n.Arms)-1 {
			armPrefix = "└─ "
			armChildIndent = indent + "   "
		}
		b.WriteString("\n" + indent + armPrefix + "Arm\n")
		b.WriteString(armChildIndent + "├─ Pattern: ")
		b.WriteString(formatAst(arm.Pattern, armChildIndent+"│  "))
		b.WriteString("\n" + armChildIndent + "└─ Body: ")
		b.WriteString(formatAst(arm.Body, armChildIndent+"   "))
	}
	return b.String()
}

func (n AstPatLiteral) FormatAst(indent string) string {
	return "Num " + formatAstOneLine(n.Value)
}

func (n AstPatBind) FormatAst(indent string) string {
	return fmt.Sprintf("Bind %q", n.Name)
}

func (n AstPatArray) FormatAst(indent string) string {
	var b strings.Builder
	b.WriteString("Array\n")
	for i, elem := range n.Elements {
		prefix := "├─ "
		childIndent := indent + "│  "
		if i == len(n.Elements)-1 {
			prefix = "└─ "
			childIndent = indent + "   "
		}
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(indent + prefix)
		if elem.Rest {
			b.WriteString("*")
		}
		b.WriteString(formatAst(elem.Pattern, childIndent))
	}
	return b.String()
}

type AstSubscript struct {
	Array Ast `ast:"child"`
	Index Ast `ast:"child"`
	Line  int
}

type AstNumber struct {
	Value string
	Line  int
}

type CustomFormatter interface {
	FormatAst(indent string) string
}

func (p AstProgram) String() string {
	var b strings.Builder
	b.WriteString("Program\n")
	for i, node := range p.Nodes {
		prefix := "├─ "
		childPrefix := "│  "
		if i == len(p.Nodes)-1 {
			prefix = "└─ "
			childPrefix = "   "
		}
		b.WriteString(prefix)
		b.WriteString(formatAst(node, childPrefix))
		if i < len(p.Nodes)-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func (n AstIdentifier) FormatAst(indent string) string {
	return fmt.Sprintf("Ident %q", n.value)
}

func (n AstNumber) FormatAst(indent string) string {
	return fmt.Sprintf("Num %s", n.Value)
}

func (n AstBind) FormatAst(indent string) string {
	exprs := flattenBind(&n)
	var b strings.Builder
	b.WriteString("Seq ;\n")
	for i, expr := range exprs {
		prefix := "├─ "
		childPrefix := indent + "│  "
		if i == len(exprs)-1 {
			prefix = "└─ "
			childPrefix = indent + "   "
		}
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(indent + prefix)
		b.WriteString(formatAst(expr, childPrefix))
	}
	return b.String()
}

func (n AstCall) FormatAst(indent string) string {
	var b strings.Builder
	b.WriteString("Call\n")
	b.WriteString(indent + "├─ Callee: ")
	b.WriteString(formatAst(n.Callee, indent+"│        "))
	if len(n.Args) > 0 {
		b.WriteString("\n" + indent + "└─ Args\n")
		for i, arg := range n.Args {
			prefix := "├─ "
			childPrefix := "│  "
			if i == len(n.Args)-1 {
				prefix = "└─ "
				childPrefix = "   "
			}
			b.WriteString(indent + "   " + prefix)
			b.WriteString(formatAst(arg, indent+"   "+childPrefix))
			if i < len(n.Args)-1 {
				b.WriteByte('\n')
			}
		}
	} else {
		b.WriteString("\n" + indent + "└─ Args: none")
	}
	return b.String()
}

func (n AstFunction) FormatAst(indent string) string {
	var b strings.Builder
	b.WriteString("Function\n")
	b.WriteString(indent + "├─ Params: ")
	params := make([]string, len(n.Args))
	for i, arg := range n.Args {
		s := arg.Name
		if arg.DefaultValue != nil {
			s += " = " + formatAstOneLine(arg.DefaultValue)
		}
		params[i] = s
	}
	b.WriteString(strings.Join(params, ", "))
	b.WriteString("\n" + indent + "└─ Body\n")
	b.WriteString(indent + "   └─ ")
	b.WriteString(formatAst(n.Body, indent+"      "))
	return b.String()
}

func formatAst(node Ast, indent string) string {
	if node == nil {
		return "nil"
	}

	if cf, ok := node.(CustomFormatter); ok {
		return cf.FormatAst(indent)
	}

	v := reflect.ValueOf(node)
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return "nil"
		}
		v = v.Elem()
	}
	t := v.Type()

	label := deriveLabel(t, v)
	children := collectChildren(t, v)
	return buildTree(label, children, indent)
}

func deriveLabel(t reflect.Type, v reflect.Value) string {
	var label strings.Builder
	label.WriteString(strings.TrimPrefix(t.Name(), "Ast"))

	for i := range t.NumField() {
		f := t.Field(i)
		if f.Tag.Get("ast") == "label" {
			fv := v.Field(i)
			if f.Type.Kind() == reflect.String {
				label.WriteString(" ." + fv.String())
			} else if f.Type.Name() == "TokenKind" {
				kind := lexer.TokenKind(fv.Int())
				label.WriteString(" " + tokenSymbol(kind))
			}
		}
	}

	return label.String()
}

func collectChildren(t reflect.Type, v reflect.Value) []Ast {
	var children []Ast
	for i := range t.NumField() {
		f := t.Field(i)
		tag := f.Tag.Get("ast")
		if !strings.HasPrefix(tag, "child") {
			continue
		}
		fv := v.Field(i)
		children = append(children, fv.Interface().(Ast))
	}
	return children
}

func buildTree(label string, children []Ast, indent string) string {
	if len(children) == 0 {
		return label
	}

	var b strings.Builder
	b.WriteString(label + "\n")
	for i, child := range children {
		prefix := "├─ "
		childIndent := indent + "│  "
		if i == len(children)-1 {
			prefix = "└─ "
			childIndent = indent + "   "
		}
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(indent + prefix)
		b.WriteString(formatAst(child, childIndent))
	}
	return b.String()
}

func flattenBind(n *AstBind) []Ast {
	var exprs []Ast
	var walk func(node Ast)
	walk = func(node Ast) {
		switch n := node.(type) {
		case *AstBind:
			walk(n.Lhs)
			walk(n.Rhs)
		case AstBind:
			walk(n.Lhs)
			walk(n.Rhs)
		default:
			exprs = append(exprs, node)
		}
	}
	walk(n)
	return exprs
}

func formatAstOneLine(node Ast) string {
	switch n := node.(type) {
	case *AstIdentifier:
		return fmt.Sprintf("%q", n.value)
	case AstIdentifier:
		return fmt.Sprintf("%q", n.value)
	case *AstNumber:
		return n.Value
	case AstNumber:
		return n.Value
	case *AstFunction:
		return fmt.Sprintf("(%s) ...", formatParams(n.Args))
	case AstFunction:
		return fmt.Sprintf("(%s) ...", formatParams(n.Args))
	case *AstBinary:
		return fmt.Sprintf("(%s %s %s)", formatAstOneLine(n.Lhs), tokenSymbol(n.Op), formatAstOneLine(n.Rhs))
	case AstBinary:
		return fmt.Sprintf("(%s %s %s)", formatAstOneLine(n.Lhs), tokenSymbol(n.Op), formatAstOneLine(n.Rhs))
	default:
		return "..."
	}
}

func formatParams(args []AstFunctionArg) string {
	names := make([]string, len(args))
	for i, a := range args {
		names[i] = a.Name
	}
	return strings.Join(names, ", ")
}

func tokenSymbol(kind lexer.TokenKind) string {
	switch kind {
	case lexer.PLUS:
		return "+"
	case lexer.MINUS:
		return "-"
	case lexer.STAR:
		return "*"
	case lexer.SLASH:
		return "/"
	case lexer.LESS_EQ:
		return "<="
	case lexer.GREATER_EQ:
		return ">="
	case lexer.BANG_EQ:
		return "!="
	case lexer.EXCLAMATION_MARK:
		return "!"
	case lexer.PLUS_PLUS:
		return "++"
	case lexer.OPEN_BRACKET:
		return "["
	case lexer.CLOSE_BRACKET:
		return "]"
	default:
		return fmt.Sprintf("TK(%d)", kind)
	}
}
