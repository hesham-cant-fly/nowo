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
	v := runCode(t, `f := () { 1; 2; 3 }
f()`)
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

func TestComparison(t *testing.T) {
	tests := []struct {
		src string
		val float64
	}{
		{"1 < 2", 1},
		{"2 < 1", 0},
		{"1 < 1", 0},
		{"1 <= 2", 1},
		{"2 <= 1", 0},
		{"1 <= 1", 1},
		{"2 > 1", 1},
		{"1 > 2", 0},
		{"1 > 1", 0},
		{"2 >= 1", 1},
		{"1 >= 2", 0},
		{"1 >= 1", 1},
		{"1 == 1", 1},
		{"1 == 2", 0},
		{"1 != 2", 1},
		{"1 != 1", 0},
	}
	for _, tt := range tests {
		v := runCode(t, tt.src)
		if v.Type != ValNumber || v.Num != tt.val {
			t.Fatalf("%q: expected %v, got %v", tt.src, tt.val, v)
		}
	}
}

func TestUnaryNot(t *testing.T) {
	tests := []struct {
		src string
		val float64
	}{
		{"!0", 1},
		{"!1", 0},
		{"!42", 0},
		{"!!0", 0},
		{"!!1", 1},
	}
	for _, tt := range tests {
		v := runCode(t, tt.src)
		if v.Type != ValNumber || v.Num != tt.val {
			t.Fatalf("%q: expected %v, got %v", tt.src, tt.val, v)
		}
	}
}

func TestBuiltinPrint(t *testing.T) {
	v := runCode(t, `print(42)`)
	if v.Type != ValNil {
		t.Fatalf("expected nil, got %v", v)
	}
}

func TestBuiltinPrintMultipleArgs(t *testing.T) {
	v := runCode(t, `print(1, 2, 3)`)
	if v.Type != ValNil {
		t.Fatalf("expected nil, got %v", v)
	}
}

func TestBuiltinPrintOverridable(t *testing.T) {
	v := runCode(t, `
print := (x) x + 1
print(41)
`)
	if v.Type != ValNumber || v.Num != 42 {
		t.Fatalf("expected 42, got %v", v)
	}
}

func TestBuiltinPrintInFunction(t *testing.T) {
	v := runCode(t, `
f := () print(42)
f()
`)
	if v.Type != ValNil {
		t.Fatalf("expected nil, got %v", v)
	}
}

func TestArrayLiteral(t *testing.T) {
	v := runCode(t, `.[1, 2, 3]`)
	if v.Type != ValArray {
		t.Fatalf("expected array, got %v", v)
	}
	if len(v.Arr.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(v.Arr.Elements))
	}
	if v.Arr.Elements[0].Num != 1 || v.Arr.Elements[1].Num != 2 || v.Arr.Elements[2].Num != 3 {
		t.Fatalf("expected [1 2 3], got %v", v)
	}
}

func TestArrayEmpty(t *testing.T) {
	v := runCode(t, `.[]`)
	if v.Type != ValArray || len(v.Arr.Elements) != 0 {
		t.Fatalf("expected empty array, got %v", v)
	}
}

func TestArraySubscript(t *testing.T) {
	v := runCode(t, `.[10, 20, 30][1]`)
	if v.Type != ValNumber || v.Num != 20 {
		t.Fatalf("expected 20, got %v", v)
	}
}

func TestArraySubscriptOutOfBounds(t *testing.T) {
	comp, err := Compile(`.[1, 2, 3][5]`)
	if err != nil {
		t.Fatal("compile should succeed")
	}
	_, err = comp.Run()
	if err == nil {
		t.Fatal("expected runtime error")
	}
}

func TestArraySubscriptNegative(t *testing.T) {
	v := runCode(t, `.[1, 2, 3][-1]`)
	if v.Type != ValNumber || v.Num != 3 {
		t.Fatalf("expected 3, got %v", v)
	}
}

func TestArrayConcat(t *testing.T) {
	v := runCode(t, `.[1, 2] ++ .[3, 4]`)
	if v.Type != ValArray {
		t.Fatalf("expected array, got %v", v)
	}
	if len(v.Arr.Elements) != 4 {
		t.Fatalf("expected 4 elements, got %d", len(v.Arr.Elements))
	}
	if v.Arr.Elements[0].Num != 1 || v.Arr.Elements[1].Num != 2 ||
		v.Arr.Elements[2].Num != 3 || v.Arr.Elements[3].Num != 4 {
		t.Fatalf("expected [1 2 3 4], got %v", v)
	}
}

func TestArrayInFunction(t *testing.T) {
	v := runCode(t, `
first := (arr) arr[0]
first(.[10, 20, 30])
`)
	if v.Type != ValNumber || v.Num != 10 {
		t.Fatalf("expected 10, got %v", v)
	}
}

func TestArrayNested(t *testing.T) {
	v := runCode(t, `.[.[1, 2], .[3, 4]][0][1]`)
	if v.Type != ValNumber || v.Num != 2 {
		t.Fatalf("expected 2, got %v", v)
	}
}

func TestDefaultParam(t *testing.T) {
	v := runCode(t, `
f := (y, x = 1) x
f(42)
`)
	if v.Type != ValNumber || v.Num != 1 {
		t.Fatalf("expected 1, got %v", v)
	}
}

func TestDefaultParamProvided(t *testing.T) {
	v := runCode(t, `
f := (y, x = 1) x
f(42, 99)
`)
	if v.Type != ValNumber || v.Num != 99 {
		t.Fatalf("expected 99, got %v", v)
	}
}

func TestDefaultParamRequired(t *testing.T) {
	comp, err := Compile(`
f := (y, x = 1) y
f()
`)
	if err != nil {
		t.Fatal("compile should succeed")
	}
	_, err = comp.Run()
	if err == nil {
		t.Fatal("expected runtime error for missing required param")
	}
}

func TestDefaultParamAllDefault(t *testing.T) {
	v := runCode(t, `
f := (x = 10, y = 20) x + y
f()
`)
	if v.Type != ValNumber || v.Num != 30 {
		t.Fatalf("expected 30, got %v", v)
	}
}

func TestDefaultParamExpression(t *testing.T) {
	v := runCode(t, `
f := (x, m = 2, n = m + 1) x + m + n
f(1)
`)
	if v.Type != ValNumber || v.Num != 6 {
		t.Fatalf("expected 6, got %v", v)
	}
}

func TestLength(t *testing.T) {
	v := runCode(t, `#.[10, 20, 30]`)
	if v.Type != ValNumber || v.Num != 3 {
		t.Fatalf("expected 3, got %v", v)
	}
}

func TestLengthEmpty(t *testing.T) {
	v := runCode(t, `#.[]`)
	if v.Type != ValNumber || v.Num != 0 {
		t.Fatalf("expected 0, got %v", v)
	}
}

func TestNilLiteral(t *testing.T) {
	v := runCode(t, `nil`)
	if v.Type != ValNil {
		t.Fatalf("expected nil, got %v", v)
	}
}

func TestNilInArray(t *testing.T) {
	v := runCode(t, `.[nil, 1]`)
	if v.Type != ValArray {
		t.Fatalf("expected array, got %v", v)
	}
	if len(v.Arr.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(v.Arr.Elements))
	}
	if v.Arr.Elements[0].Type != ValNil {
		t.Fatalf("expected nil, got %v", v.Arr.Elements[0])
	}
}

func TestNilShadowable(t *testing.T) {
	v := runCode(t, `
nil := 42
nil
`)
	if v.Type != ValNumber || v.Num != 42 {
		t.Fatalf("expected 42, got %v", v)
	}
}

func TestPipeOperator(t *testing.T) {
	v := runCode(t, `42:(x) x + 1`)
	if v.Type != ValNumber || v.Num != 43 {
		t.Fatalf("expected 43, got %v", v)
	}
}

func TestPipeMultiArg(t *testing.T) {
	v := runCode(t, `5:(x, y = 10) x + y`)
	if v.Type != ValNumber || v.Num != 15 {
		t.Fatalf("expected 15, got %v", v)
	}
}

func TestGroupBind(t *testing.T) {
	v := runCode(t, `{1; 2; 3}`)
	if v.Type != ValNumber || v.Num != 3 {
		t.Fatalf("expected 3, got %v", v)
	}
}

func TestGroupBindInFunction(t *testing.T) {
	v := runCode(t, `
f := (x) { x; x + 1 }
f(42)
`)
	if v.Type != ValNumber || v.Num != 43 {
		t.Fatalf("expected 43, got %v", v)
	}
}

func TestNotOnNil(t *testing.T) {
	v := runCode(t, `!nil`)
	if v.Type != ValNumber || v.Num != 1 {
		t.Fatalf("expected 1, got %v", v)
	}
}

func TestArithExpr(t *testing.T) {
	v := runCode(t, `{2 + 3} * 4`)
	if v.Type != ValNumber || v.Num != 20 {
		t.Fatalf("expected 20, got %v", v)
	}
}

func TestDeclInBody(t *testing.T) {
	v := runCode(t, `f := (x) { g := (y) x + y; g(1) }
f(41)
`)
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
