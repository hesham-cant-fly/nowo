package main

import (
	"fmt"
	"os"

	"github.com/hesham-cant-fly/haste-lang/vm"
)

func main() {
	src := readEntireFile("example.nowo")

	comp, err := vm.Compile(src)
	if err != nil {
		fmt.Println(err)
		return
	}

	// fmt.Println(comp.AST())
	// fmt.Println()
	// fmt.Print(comp.Chunk().Disassemble())

	result, err := comp.Run()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("=>", result)
}

func readEntireFile(path string) string {
	result, err := os.ReadFile(path)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return string(result)
}
