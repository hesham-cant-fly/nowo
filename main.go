package main

import (
	"fmt"

	"github.com/hesham-cant-fly/haste-lang/lexer"
	"github.com/hesham-cant-fly/haste-lang/parser"
)

/*
identity := (x)
  x; x
add := (x, y) x + y
*/

func main() {
	// src := "identity := (x) x"
	// src := `main := ()
	// x := 1;
	// y := 2;
	// print(add(x, y))`
	src := "main := () x := a; x"
	tokens := lexer.Tokenize(src)
	ast, err := parser.Parse(tokens)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(ast)
}
