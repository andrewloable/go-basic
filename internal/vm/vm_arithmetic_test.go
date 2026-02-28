package vm

import (
	"math"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Test: arithmetic operations
// ---------------------------------------------------------------------------

func TestArithmeticInt(t *testing.T) {
	tests := []struct {
		name   string
		op     Opcode
		a, b   int64
		expect int64
	}{
		{"add", OpAdd, 10, 20, 30},
		{"sub", OpSub, 50, 20, 30},
		{"mul", OpMul, 6, 7, 42},
		{"idiv", OpIDiv, 17, 5, 3},
		{"mod", OpMod, 17, 5, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunk := makeChunk(
				[]Value{IntVal(tt.a), IntVal(tt.b)},
				[]Instruction{
					{Op: OpPush, Operand: 0},
					{Op: OpPush, Operand: 1},
					{Op: tt.op},
					{Op: OpHalt},
				},
			)
			vm, _ := runVM(t, chunk)
			if vm.sp != 1 {
				t.Fatalf("expected sp=1, got %d", vm.sp)
			}
			result := vm.stack[0]
			if result.Type != ValInt || result.Int != tt.expect {
				t.Fatalf("expected Int(%d), got %s", tt.expect, result)
			}
		})
	}
}

func TestArithmeticFloat(t *testing.T) {
	// Division always produces float.
	chunk := makeChunk(
		[]Value{FloatVal(10.0), FloatVal(3.0)},
		[]Instruction{
			{Op: OpPush, Operand: 0},
			{Op: OpPush, Operand: 1},
			{Op: OpDiv},
			{Op: OpHalt},
		},
	)
	vm, _ := runVM(t, chunk)
	result := vm.stack[0]
	if result.Type != ValFloat {
		t.Fatalf("expected float result, got %s", result)
	}
	expected := 10.0 / 3.0
	if math.Abs(result.Float-expected) > 1e-10 {
		t.Fatalf("expected %g, got %g", expected, result.Float)
	}
}

func TestArithmeticMixed(t *testing.T) {
	// Int + Float should produce Float.
	chunk := makeChunk(
		[]Value{IntVal(5), FloatVal(2.5)},
		[]Instruction{
			{Op: OpPush, Operand: 0},
			{Op: OpPush, Operand: 1},
			{Op: OpAdd},
			{Op: OpHalt},
		},
	)
	vm, _ := runVM(t, chunk)
	result := vm.stack[0]
	if result.Type != ValFloat {
		t.Fatalf("expected float result, got %s", result)
	}
	if result.Float != 7.5 {
		t.Fatalf("expected 7.5, got %g", result.Float)
	}
}

func TestPow(t *testing.T) {
	chunk := makeChunk(
		[]Value{IntVal(2), IntVal(10)},
		[]Instruction{
			{Op: OpPush, Operand: 0},
			{Op: OpPush, Operand: 1},
			{Op: OpPow},
			{Op: OpHalt},
		},
	)
	vm, _ := runVM(t, chunk)
	result := vm.stack[0]
	if result.Float != 1024.0 {
		t.Fatalf("expected 1024, got %g", result.Float)
	}
}

func TestNeg(t *testing.T) {
	chunk := makeChunk(
		[]Value{IntVal(42)},
		[]Instruction{
			{Op: OpPush, Operand: 0},
			{Op: OpNeg},
			{Op: OpHalt},
		},
	)
	vm, _ := runVM(t, chunk)
	result := vm.stack[0]
	if result.Type != ValInt || result.Int != -42 {
		t.Fatalf("expected Int(-42), got %s", result)
	}
}

func TestDivisionByZero(t *testing.T) {
	chunk := makeChunk(
		[]Value{IntVal(10), IntVal(0)},
		[]Instruction{
			{Op: OpPush, Operand: 0},
			{Op: OpPush, Operand: 1},
			{Op: OpDiv},
			{Op: OpHalt},
		},
	)
	vm := NewVM(chunk)
	err := vm.Run()
	if err == nil {
		t.Fatal("expected division by zero error")
	}
	if !strings.Contains(err.Error(), "division by zero") {
		t.Fatalf("expected 'division by zero' in error, got: %s", err)
	}
}

// ---------------------------------------------------------------------------
// Test: comparison operations
// ---------------------------------------------------------------------------

func TestComparisons(t *testing.T) {
	tests := []struct {
		name   string
		op     Opcode
		a, b   int64
		expect int64 // -1 for true, 0 for false
	}{
		{"eq_true", OpEq, 5, 5, -1},
		{"eq_false", OpEq, 5, 6, 0},
		{"ne_true", OpNe, 5, 6, -1},
		{"ne_false", OpNe, 5, 5, 0},
		{"lt_true", OpLt, 3, 5, -1},
		{"lt_false", OpLt, 5, 3, 0},
		{"gt_true", OpGt, 5, 3, -1},
		{"gt_false", OpGt, 3, 5, 0},
		{"le_true_less", OpLe, 3, 5, -1},
		{"le_true_eq", OpLe, 5, 5, -1},
		{"le_false", OpLe, 6, 5, 0},
		{"ge_true_greater", OpGe, 5, 3, -1},
		{"ge_true_eq", OpGe, 5, 5, -1},
		{"ge_false", OpGe, 3, 5, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunk := makeChunk(
				[]Value{IntVal(tt.a), IntVal(tt.b)},
				[]Instruction{
					{Op: OpPush, Operand: 0},
					{Op: OpPush, Operand: 1},
					{Op: tt.op},
					{Op: OpHalt},
				},
			)
			vm, _ := runVM(t, chunk)
			result := vm.stack[0]
			if result.Int != tt.expect {
				t.Fatalf("expected %d, got %d", tt.expect, result.Int)
			}
		})
	}
}

func TestStringComparison(t *testing.T) {
	chunk := makeChunk(
		[]Value{StringVal("abc"), StringVal("def")},
		[]Instruction{
			{Op: OpPush, Operand: 0},
			{Op: OpPush, Operand: 1},
			{Op: OpLt},
			{Op: OpHalt},
		},
	)
	vm, _ := runVM(t, chunk)
	result := vm.stack[0]
	// "abc" < "def" should be true (-1)
	if result.Int != -1 {
		t.Fatalf("expected -1 (true), got %d", result.Int)
	}
}

// ---------------------------------------------------------------------------
// Test: logical operations
// ---------------------------------------------------------------------------

func TestLogical(t *testing.T) {
	tests := []struct {
		name   string
		op     Opcode
		a, b   int64
		expect int64
	}{
		{"and", OpAnd, -1, 0, 0},
		{"or", OpOr, -1, 0, -1},
		{"xor", OpXor, -1, -1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunk := makeChunk(
				[]Value{IntVal(tt.a), IntVal(tt.b)},
				[]Instruction{
					{Op: OpPush, Operand: 0},
					{Op: OpPush, Operand: 1},
					{Op: tt.op},
					{Op: OpHalt},
				},
			)
			vm, _ := runVM(t, chunk)
			result := vm.stack[0]
			if result.Int != tt.expect {
				t.Fatalf("expected %d, got %d", tt.expect, result.Int)
			}
		})
	}
}

func TestNot(t *testing.T) {
	// NOT 0 = -1 (all bits flipped)
	chunk := makeChunk(
		[]Value{IntVal(0)},
		[]Instruction{
			{Op: OpPush, Operand: 0},
			{Op: OpNot},
			{Op: OpHalt},
		},
	)
	vm, _ := runVM(t, chunk)
	result := vm.stack[0]
	if result.Int != -1 {
		t.Fatalf("expected -1, got %d", result.Int)
	}
}

// ---------------------------------------------------------------------------
// Test: string operations
// ---------------------------------------------------------------------------

func TestStringConcat(t *testing.T) {
	chunk := makeChunk(
		[]Value{StringVal("Hello, "), StringVal("World!")},
		[]Instruction{
			{Op: OpPush, Operand: 0},
			{Op: OpPush, Operand: 1},
			{Op: OpConcat},
			{Op: OpHalt},
		},
	)
	vm, _ := runVM(t, chunk)
	result := vm.stack[0]
	if result.Type != ValString || result.Str != "Hello, World!" {
		t.Fatalf("expected String(\"Hello, World!\"), got %s", result)
	}
}

func TestStringAddition(t *testing.T) {
	// OpAdd should also concatenate strings.
	chunk := makeChunk(
		[]Value{StringVal("foo"), StringVal("bar")},
		[]Instruction{
			{Op: OpPush, Operand: 0},
			{Op: OpPush, Operand: 1},
			{Op: OpAdd},
			{Op: OpHalt},
		},
	)
	vm, _ := runVM(t, chunk)
	result := vm.stack[0]
	if result.Str != "foobar" {
		t.Fatalf("expected \"foobar\", got %q", result.Str)
	}
}

// ---------------------------------------------------------------------------
// Test: EQV and IMP
// ---------------------------------------------------------------------------

func TestEqvImp(t *testing.T) {
	// EQV: NOT (a XOR b)
	// -1 EQV -1 = NOT (0) = -1
	chunk := makeChunk(
		[]Value{IntVal(-1)},
		[]Instruction{
			{Op: OpPush, Operand: 0},
			{Op: OpPush, Operand: 0},
			{Op: OpEqv},
			{Op: OpHalt},
		},
	)
	vm, _ := runVM(t, chunk)
	if vm.stack[0].Int != -1 {
		t.Fatalf("EQV: expected -1, got %d", vm.stack[0].Int)
	}

	// IMP: (NOT a) OR b
	// 0 IMP -1 = (NOT 0) OR -1 = -1 OR -1 = -1
	chunk2 := makeChunk(
		[]Value{IntVal(0), IntVal(-1)},
		[]Instruction{
			{Op: OpPush, Operand: 0},
			{Op: OpPush, Operand: 1},
			{Op: OpImp},
			{Op: OpHalt},
		},
	)
	vm2, _ := runVM(t, chunk2)
	if vm2.stack[0].Int != -1 {
		t.Fatalf("IMP: expected -1, got %d", vm2.stack[0].Int)
	}
}

// ---------------------------------------------------------------------------
// Test: execSub and execMul float paths
// ---------------------------------------------------------------------------

func TestExecSubFloat(t *testing.T) {
	// Subtraction with at least one float operand uses the float branch.
	src := `PRINT 5.5 - 2.5`
	out := compileAndRun(t, src)
	if !strings.Contains(out, "3") {
		t.Errorf("5.5 - 2.5: expected '3', got %q", out)
	}
}

func TestExecMulFloat(t *testing.T) {
	// Multiplication with float operands uses the float branch.
	src := `PRINT 2.5 * 4.0`
	out := compileAndRun(t, src)
	if !strings.Contains(out, "10") {
		t.Errorf("2.5 * 4.0: expected '10', got %q", out)
	}
}
