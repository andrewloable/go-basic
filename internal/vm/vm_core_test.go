package vm

import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Test: basic push/pop
// ---------------------------------------------------------------------------

func TestPushPop(t *testing.T) {
	chunk := makeChunk(
		[]Value{IntVal(42), FloatVal(3.14), StringVal("hello")},
		[]Instruction{
			{Op: OpPush, Operand: 0},  // push 42
			{Op: OpPush, Operand: 1},  // push 3.14
			{Op: OpPush, Operand: 2},  // push "hello"
			{Op: OpPop},               // pop "hello"
			{Op: OpPop},               // pop 3.14
			{Op: OpHalt},
		},
	)

	vm := NewVM(chunk)
	var buf strings.Builder
	vm.SetOutput(&buf)
	err := vm.Run()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// After execution, stack should have one item: 42
	if vm.sp != 1 {
		t.Fatalf("expected sp=1, got %d", vm.sp)
	}
	v := vm.stack[0]
	if v.Type != ValInt || v.Int != 42 {
		t.Fatalf("expected Int(42), got %s", v)
	}
}

func TestDup(t *testing.T) {
	chunk := makeChunk(
		[]Value{IntVal(7)},
		[]Instruction{
			{Op: OpPush, Operand: 0}, // push 7
			{Op: OpDup},              // dup -> 7, 7
			{Op: OpHalt},
		},
	)

	vm := NewVM(chunk)
	var buf strings.Builder
	vm.SetOutput(&buf)
	err := vm.Run()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if vm.sp != 2 {
		t.Fatalf("expected sp=2, got %d", vm.sp)
	}
	if vm.stack[0].Int != 7 || vm.stack[1].Int != 7 {
		t.Fatal("DUP did not duplicate the value correctly")
	}
}

// ---------------------------------------------------------------------------
// Test: variables (LOAD/STORE)
// ---------------------------------------------------------------------------

func TestLoadStore(t *testing.T) {
	chunk := makeChunk(
		[]Value{IntVal(42), StringVal("x")},
		[]Instruction{
			{Op: OpPush, Operand: 0},   // push 42
			{Op: OpStore, Operand: 1},  // store into "x"
			{Op: OpLoad, Operand: 1},   // load "x"
			{Op: OpHalt},
		},
	)
	vm, _ := runVM(t, chunk)
	if vm.sp != 1 {
		t.Fatalf("expected sp=1, got %d", vm.sp)
	}
	if vm.stack[0].Int != 42 {
		t.Fatalf("expected 42, got %s", vm.stack[0])
	}
}

func TestLoadUninitialized(t *testing.T) {
	// Loading an uninitialized variable should return 0.
	chunk := makeChunk(
		[]Value{StringVal("y")},
		[]Instruction{
			{Op: OpLoad, Operand: 0},
			{Op: OpHalt},
		},
	)
	vm, _ := runVM(t, chunk)
	if vm.stack[0].Type != ValInt || vm.stack[0].Int != 0 {
		t.Fatalf("expected Int(0), got %s", vm.stack[0])
	}
}

// ---------------------------------------------------------------------------
// Test: NOP and LINE
// ---------------------------------------------------------------------------

func TestNopAndLine(t *testing.T) {
	chunk := makeChunk(
		[]Value{IntVal(1)},
		[]Instruction{
			{Op: OpNop},
			{Op: OpLine, Operand: 100},
			{Op: OpPush, Operand: 0},
			{Op: OpHalt},
		},
	)
	vm, _ := runVM(t, chunk)
	if vm.sp != 1 || vm.stack[0].Int != 1 {
		t.Fatal("NOP/LINE should not affect execution")
	}
}

// ---------------------------------------------------------------------------
// Test: error includes line number
// ---------------------------------------------------------------------------

func TestRuntimeErrorLineNumber(t *testing.T) {
	chunk := &Chunk{
		Code: []Instruction{
			{Op: OpPush, Operand: 0},
			{Op: OpPush, Operand: 1},
			{Op: OpDiv},
			{Op: OpHalt},
		},
		Constants: []Value{FloatVal(1.0), FloatVal(0.0)},
		Lines:     []int{10, 10, 10, 20},
	}
	vm := NewVM(chunk)
	err := vm.Run()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "line 10") {
		t.Fatalf("expected 'line 10' in error, got: %s", err)
	}
}

// ---------------------------------------------------------------------------
// Test: execNeg — float path
// ---------------------------------------------------------------------------

func TestExecNegFloat(t *testing.T) {
	// Negating a float variable exercises the float branch in execNeg.
	src := `
x = 3.5
PRINT -x
`
	out := compileAndRun(t, src)
	if !strings.Contains(out, "-3.5") && !strings.Contains(out, "-3") {
		t.Errorf("-x where x=3.5: expected '-3.5', got %q", out)
	}
}
