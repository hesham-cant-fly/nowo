package vm

import (
	"testing"
)

func runCode(t *testing.T, src string) Value {
	t.Helper()
	comp, err := Compile(src)
	if err != nil {
		t.Fatal(err)
	}
	v, err := comp.Run()
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestNumber(t *testing.T) {
	v := runCode(t, "42")
	if v.Type != ValNumber || v.Num != 42 {
		t.Fatalf("expected 42, got %v", v)
	}
}

func TestBind(t *testing.T) {
	v := runCode(t, "1;2;3")
	if v.Type != ValNumber || v.Num != 3 {
		t.Fatalf("expected 3, got %v", v)
	}
}

func TestArithmetic(t *testing.T) {
	v := runCode(t, "{2+2}*3")
	if v.Type != ValNumber || v.Num != 12 {
		t.Fatalf("expected 12, got %v", v)
	}
}

func TestDeclAndCall(t *testing.T) {
	src := `
add := (x, y) x + y
main := () add(3, 4)
main()
`
	v := runCode(t, src)
	if v.Type != ValNumber || v.Num != 7 {
		t.Fatalf("expected 7, got %v", v)
	}
}

func TestBindInFunction(t *testing.T) {
	src := `f := () 1; 2; 3
f()`
	v := runCode(t, src)
	if v.Type != ValNumber || v.Num != 3 {
		t.Fatalf("expected 3, got %v", v)
	}
}

func TestNestedFunction(t *testing.T) {
	src := `f := (x) (y) x + y
g := f(10)
g(5)`
	v := runCode(t, src)
	if v.Type != ValNumber || v.Num != 15 {
		t.Fatalf("expected 15, got %v", v)
	}
}

func TestDivisionByZero(t *testing.T) {
	_, err := Compile("1/0")
	if err != nil {
		t.Fatal("compile should succeed")
	}
	comp, _ := Compile("1/0")
	_, err = comp.Run()
	if err == nil {
		t.Fatal("expected runtime error")
	}
}

func TestIdentity(t *testing.T) {
	v := runCode(t, `id := (x) x
id(42)`)
	if v.Type != ValNumber || v.Num != 42 {
		t.Fatalf("expected 42, got %v", v)
	}
}

func TestCompilerInspect(t *testing.T) {
	comp, err := Compile("42")
	if err != nil {
		t.Fatal(err)
	}
	if comp.Source() != "42" {
		t.Fatalf("bad source: %q", comp.Source())
	}
	if comp.Chunk() == nil {
		t.Fatal("chunk is nil")
	}
	if comp.AST().Nodes == nil {
		t.Fatal("AST is nil")
	}
}
