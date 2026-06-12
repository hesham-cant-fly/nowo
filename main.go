package main

import (
	"fmt"
	"os"

	"github.com/hesham-cant-fly/haste-lang/vm"
)

func main() {
	userArgs := os.Args[1:]
	if len(userArgs) == 0 {
		fmt.Println("specify a file to run")
		return
	}

	file := userArgs[0]
	src := readEntireFile(file)

	comp, err := vm.Compile(file, src)
	if err != nil {
		fmt.Println(err)
		return
	}

	// fmt.Println(comp.AST())
	// fmt.Println()
	// fmt.Print(comp.Chunk().Disassemble())

	_, err = comp.Run()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	// fmt.Println("=>", result)
}

func readEntireFile(path string) string {
	result, err := os.ReadFile(path)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return string(result)
}
