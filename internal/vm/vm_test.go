package vm

import (
	"math"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Helper: build a chunk from instructions and constants
// ---------------------------------------------------------------------------

func makeChunk(constants []Value, instructions []Instruction) *Chunk {
	lines := make([]int, len(instructions))
	for i := range lines {
		lines[i] = i + 1
	}
	return &Chunk{
		Code:      instructions,
		Constants: constants,
		Lines:     lines,
	}
}

func runVM(t *testing.T, chunk *Chunk) (*VM, string) {
	t.Helper()
	vm := NewVM(chunk)
	var buf strings.Builder
	vm.SetOutput(&buf)
	if err := vm.Run(); err != nil {
		t.Fatalf("VM.Run() error: %v", err)
	}
	return vm, buf.String()
}

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
// Test: jump instructions
// ---------------------------------------------------------------------------

func TestJmp(t *testing.T) {
	// Push 1, jump over push 2, push 3 => stack should have [1, 3]
	chunk := makeChunk(
		[]Value{IntVal(1), IntVal(2), IntVal(3)},
		[]Instruction{
			{Op: OpPush, Operand: 0},  // 0: push 1
			{Op: OpJmp, Operand: 3},   // 1: jump to 3
			{Op: OpPush, Operand: 1},  // 2: push 2 (skipped)
			{Op: OpPush, Operand: 2},  // 3: push 3
			{Op: OpHalt},              // 4: halt
		},
	)
	vm, _ := runVM(t, chunk)
	if vm.sp != 2 {
		t.Fatalf("expected sp=2, got %d", vm.sp)
	}
	if vm.stack[0].Int != 1 {
		t.Fatalf("expected stack[0]=1, got %d", vm.stack[0].Int)
	}
	if vm.stack[1].Int != 3 {
		t.Fatalf("expected stack[1]=3, got %d", vm.stack[1].Int)
	}
}

func TestJmpFalse(t *testing.T) {
	// Push 0 (false), JmpFalse to instruction 3, push 99 (skipped), push 42
	chunk := makeChunk(
		[]Value{IntVal(0), IntVal(99), IntVal(42)},
		[]Instruction{
			{Op: OpPush, Operand: 0},      // 0: push 0
			{Op: OpJmpFalse, Operand: 3},  // 1: jump to 3 if false
			{Op: OpPush, Operand: 1},      // 2: push 99 (skipped)
			{Op: OpPush, Operand: 2},      // 3: push 42
			{Op: OpHalt},                  // 4: halt
		},
	)
	vm, _ := runVM(t, chunk)
	if vm.sp != 1 {
		t.Fatalf("expected sp=1, got %d", vm.sp)
	}
	if vm.stack[0].Int != 42 {
		t.Fatalf("expected 42, got %d", vm.stack[0].Int)
	}
}

func TestJmpTrue(t *testing.T) {
	// Push -1 (true), JmpTrue to instruction 3
	chunk := makeChunk(
		[]Value{IntVal(-1), IntVal(99), IntVal(42)},
		[]Instruction{
			{Op: OpPush, Operand: 0},     // 0: push -1 (true)
			{Op: OpJmpTrue, Operand: 3},  // 1: jump to 3 if true
			{Op: OpPush, Operand: 1},     // 2: push 99 (skipped)
			{Op: OpPush, Operand: 2},     // 3: push 42
			{Op: OpHalt},                 // 4: halt
		},
	)
	vm, _ := runVM(t, chunk)
	if vm.sp != 1 {
		t.Fatalf("expected sp=1, got %d", vm.sp)
	}
	if vm.stack[0].Int != 42 {
		t.Fatalf("expected 42, got %d", vm.stack[0].Int)
	}
}

// ---------------------------------------------------------------------------
// Test: GOSUB/RETURN
// ---------------------------------------------------------------------------

func TestGosubReturn(t *testing.T) {
	// Main: push 1, gosub to 4, push 3, halt
	// Sub at 4: push 2, return
	// Expected stack: [1, 2, 3]
	chunk := makeChunk(
		[]Value{IntVal(1), IntVal(3), IntVal(2)},
		[]Instruction{
			{Op: OpPush, Operand: 0},    // 0: push 1
			{Op: OpGosub, Operand: 4},   // 1: gosub to 4
			{Op: OpPush, Operand: 1},    // 2: push 3
			{Op: OpHalt},                // 3: halt
			{Op: OpPush, Operand: 2},    // 4: push 2
			{Op: OpReturn},              // 5: return
		},
	)
	vm, _ := runVM(t, chunk)
	if vm.sp != 3 {
		t.Fatalf("expected sp=3, got %d", vm.sp)
	}
	if vm.stack[0].Int != 1 || vm.stack[1].Int != 2 || vm.stack[2].Int != 3 {
		t.Fatalf("expected [1,2,3], got [%d,%d,%d]",
			vm.stack[0].Int, vm.stack[1].Int, vm.stack[2].Int)
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
// Test: Hello World bytecode program
// ---------------------------------------------------------------------------

func TestHelloWorld(t *testing.T) {
	chunk := makeChunk(
		[]Value{StringVal("Hello, World!")},
		[]Instruction{
			{Op: OpPush, Operand: 0},
			{Op: OpPrint},
			{Op: OpPrintNewline},
			{Op: OpHalt},
		},
	)
	_, output := runVM(t, chunk)
	if output != "Hello, World!\n" {
		t.Fatalf("expected \"Hello, World!\\n\", got %q", output)
	}
}

func TestPrintNumbers(t *testing.T) {
	chunk := makeChunk(
		[]Value{IntVal(42)},
		[]Instruction{
			{Op: OpPush, Operand: 0},
			{Op: OpPrint},
			{Op: OpPrintNewline},
			{Op: OpHalt},
		},
	)
	_, output := runVM(t, chunk)
	// BASIC prints positive numbers with a leading and trailing space.
	if output != " 42 \n" {
		t.Fatalf("expected \" 42 \\n\", got %q", output)
	}
}

// ---------------------------------------------------------------------------
// Test: a loop (FOR-like pattern)
// ---------------------------------------------------------------------------

// This test simulates:
//   sum = 0
//   i = 1
//   top:
//     sum = sum + i
//     i = i + 1
//     IF i <= 10 GOTO top
//   result: sum = 55

func TestForLikeLoop(t *testing.T) {
	chunk := makeChunk(
		[]Value{
			IntVal(0),        // 0: initial sum
			StringVal("sum"), // 1: variable name "sum"
			IntVal(1),        // 2: initial i / increment
			StringVal("i"),   // 3: variable name "i"
			IntVal(10),       // 4: loop limit
		},
		[]Instruction{
			// sum = 0
			{Op: OpPush, Operand: 0},   // 0: push 0
			{Op: OpStore, Operand: 1},  // 1: store to "sum"

			// i = 1
			{Op: OpPush, Operand: 2},   // 2: push 1
			{Op: OpStore, Operand: 3},  // 3: store to "i"

			// top (instruction 4):
			// sum = sum + i
			{Op: OpLoad, Operand: 1},   // 4: load sum
			{Op: OpLoad, Operand: 3},   // 5: load i
			{Op: OpAdd},                // 6: sum + i
			{Op: OpStore, Operand: 1},  // 7: store to sum

			// i = i + 1
			{Op: OpLoad, Operand: 3},   // 8: load i
			{Op: OpPush, Operand: 2},   // 9: push 1
			{Op: OpAdd},                // 10: i + 1
			{Op: OpStore, Operand: 3},  // 11: store to i

			// IF i <= 10 GOTO top
			{Op: OpLoad, Operand: 3},      // 12: load i
			{Op: OpPush, Operand: 4},      // 13: push 10
			{Op: OpLe},                    // 14: i <= 10?
			{Op: OpJmpTrue, Operand: 4},   // 15: if true, jump to top (4)

			// After loop: push sum onto stack for checking.
			{Op: OpLoad, Operand: 1},      // 16: load sum
			{Op: OpHalt},                  // 17: halt
		},
	)
	vm, _ := runVM(t, chunk)
	if vm.sp != 1 {
		t.Fatalf("expected sp=1, got %d", vm.sp)
	}
	result := vm.stack[0]
	if result.Type != ValInt || result.Int != 55 {
		t.Fatalf("expected sum=55, got %s", result)
	}
}

// ---------------------------------------------------------------------------
// Test: Countdown loop with PRINT
// ---------------------------------------------------------------------------

// Simulates:
//   FOR i = 3 TO 1 STEP -1
//     PRINT i
//   NEXT i

func TestCountdownLoop(t *testing.T) {
	chunk := makeChunk(
		[]Value{
			IntVal(3),        // 0: initial i
			StringVal("i"),   // 1: variable name "i"
			IntVal(1),        // 2: limit / step
		},
		[]Instruction{
			// i = 3
			{Op: OpPush, Operand: 0},   // 0: push 3
			{Op: OpStore, Operand: 1},  // 1: store to "i"

			// top (instruction 2):
			// PRINT i
			{Op: OpLoad, Operand: 1},   // 2: load i
			{Op: OpPrint},              // 3: print
			{Op: OpPrintNewline},       // 4: newline

			// i = i - 1
			{Op: OpLoad, Operand: 1},   // 5: load i
			{Op: OpPush, Operand: 2},   // 6: push 1
			{Op: OpSub},                // 7: i - 1
			{Op: OpStore, Operand: 1},  // 8: store to i

			// IF i >= 1 GOTO top
			{Op: OpLoad, Operand: 1},      // 9: load i
			{Op: OpPush, Operand: 2},      // 10: push 1
			{Op: OpGe},                    // 11: i >= 1?
			{Op: OpJmpTrue, Operand: 2},   // 12: if true, jump to 2

			{Op: OpHalt},                  // 13: halt
		},
	)
	_, output := runVM(t, chunk)
	expected := " 3 \n 2 \n 1 \n"
	if output != expected {
		t.Fatalf("expected %q, got %q", expected, output)
	}
}

// ---------------------------------------------------------------------------
// Test: type conversions
// ---------------------------------------------------------------------------

func TestTypeConversions(t *testing.T) {
	// ToInt on a float should truncate.
	chunk := makeChunk(
		[]Value{FloatVal(3.99)},
		[]Instruction{
			{Op: OpPush, Operand: 0},
			{Op: OpToInt},
			{Op: OpHalt},
		},
	)
	vm, _ := runVM(t, chunk)
	result := vm.stack[0]
	if result.Type != ValInt || result.Int != 3 {
		t.Fatalf("expected Int(3), got %s", result)
	}
}

// ---------------------------------------------------------------------------
// Test: built-in functions
// ---------------------------------------------------------------------------

func TestBuiltinAbs(t *testing.T) {
	chunk := makeChunk(
		[]Value{FloatVal(-42.5)},
		[]Instruction{
			{Op: OpPush, Operand: 0},
			{Op: OpBuiltin, Operand: int32(BuiltinAbs)},
			{Op: OpHalt},
		},
	)
	vm, _ := runVM(t, chunk)
	result := vm.stack[0]
	if result.Float != 42.5 {
		t.Fatalf("expected 42.5, got %g", result.Float)
	}
}

func TestBuiltinLen(t *testing.T) {
	chunk := makeChunk(
		[]Value{StringVal("Hello")},
		[]Instruction{
			{Op: OpPush, Operand: 0},
			{Op: OpBuiltin, Operand: int32(BuiltinLen)},
			{Op: OpHalt},
		},
	)
	vm, _ := runVM(t, chunk)
	result := vm.stack[0]
	if result.Type != ValInt || result.Int != 5 {
		t.Fatalf("expected Int(5), got %s", result)
	}
}

func TestBuiltinLeft(t *testing.T) {
	chunk := makeChunk(
		[]Value{StringVal("Hello, World!"), IntVal(5)},
		[]Instruction{
			{Op: OpPush, Operand: 0}, // push string
			{Op: OpPush, Operand: 1}, // push n
			{Op: OpBuiltin, Operand: int32(BuiltinLeft)},
			{Op: OpHalt},
		},
	)
	vm, _ := runVM(t, chunk)
	result := vm.stack[0]
	if result.Str != "Hello" {
		t.Fatalf("expected \"Hello\", got %q", result.Str)
	}
}

func TestBuiltinUCase(t *testing.T) {
	chunk := makeChunk(
		[]Value{StringVal("hello")},
		[]Instruction{
			{Op: OpPush, Operand: 0},
			{Op: OpBuiltin, Operand: int32(BuiltinUCase)},
			{Op: OpHalt},
		},
	)
	vm, _ := runVM(t, chunk)
	result := vm.stack[0]
	if result.Str != "HELLO" {
		t.Fatalf("expected \"HELLO\", got %q", result.Str)
	}
}

func TestBuiltinSqr(t *testing.T) {
	chunk := makeChunk(
		[]Value{FloatVal(16.0)},
		[]Instruction{
			{Op: OpPush, Operand: 0},
			{Op: OpBuiltin, Operand: int32(BuiltinSqr)},
			{Op: OpHalt},
		},
	)
	vm, _ := runVM(t, chunk)
	result := vm.stack[0]
	if result.Float != 4.0 {
		t.Fatalf("expected 4.0, got %g", result.Float)
	}
}

func TestBuiltinSqrNegative(t *testing.T) {
	chunk := makeChunk(
		[]Value{FloatVal(-1.0)},
		[]Instruction{
			{Op: OpPush, Operand: 0},
			{Op: OpBuiltin, Operand: int32(BuiltinSqr)},
			{Op: OpHalt},
		},
	)
	vm := NewVM(chunk)
	err := vm.Run()
	if err == nil {
		t.Fatal("expected error for SQR of negative number")
	}
}

func TestBuiltinChr(t *testing.T) {
	chunk := makeChunk(
		[]Value{IntVal(65)},
		[]Instruction{
			{Op: OpPush, Operand: 0},
			{Op: OpBuiltin, Operand: int32(BuiltinChr)},
			{Op: OpHalt},
		},
	)
	vm, _ := runVM(t, chunk)
	result := vm.stack[0]
	if result.Str != "A" {
		t.Fatalf("expected \"A\", got %q", result.Str)
	}
}

func TestBuiltinSgn(t *testing.T) {
	tests := []struct {
		input  float64
		expect int64
	}{
		{-5.0, -1},
		{0.0, 0},
		{3.14, 1},
	}
	for _, tt := range tests {
		chunk := makeChunk(
			[]Value{FloatVal(tt.input)},
			[]Instruction{
				{Op: OpPush, Operand: 0},
				{Op: OpBuiltin, Operand: int32(BuiltinSgn)},
				{Op: OpHalt},
			},
		)
		vm, _ := runVM(t, chunk)
		result := vm.stack[0]
		if result.Int != tt.expect {
			t.Fatalf("SGN(%g): expected %d, got %d", tt.input, tt.expect, result.Int)
		}
	}
}

// ---------------------------------------------------------------------------
// Test: DATA / READ / RESTORE
// ---------------------------------------------------------------------------

func TestDataReadRestore(t *testing.T) {
	chunk := makeChunk(
		[]Value{StringVal("a"), StringVal("b")},
		[]Instruction{
			{Op: OpRead, Operand: 0},    // read into "a"
			{Op: OpRead, Operand: 1},    // read into "b"
			{Op: OpRestore, Operand: -1}, // restore to beginning
			{Op: OpRead, Operand: 0},    // read into "a" again
			{Op: OpHalt},
		},
	)
	vm := NewVM(chunk)
	vm.SetDataPool([]Value{IntVal(10), IntVal(20)})
	var buf strings.Builder
	vm.SetOutput(&buf)
	err := vm.Run()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// After: a=10 (re-read after restore), b=20
	if vm.globals["a"].Int != 10 {
		t.Fatalf("expected a=10, got %s", vm.globals["a"])
	}
	if vm.globals["b"].Int != 20 {
		t.Fatalf("expected b=20, got %s", vm.globals["b"])
	}
}

func TestReadOutOfData(t *testing.T) {
	chunk := makeChunk(
		[]Value{StringVal("x")},
		[]Instruction{
			{Op: OpRead, Operand: 0},
			{Op: OpHalt},
		},
	)
	vm := NewVM(chunk)
	// No data pool set — should error.
	err := vm.Run()
	if err == nil {
		t.Fatal("expected 'out of DATA' error")
	}
	if !strings.Contains(err.Error(), "out of DATA") {
		t.Fatalf("expected 'out of DATA' in error, got: %s", err)
	}
}

// ---------------------------------------------------------------------------
// Test: DIM array operations
// ---------------------------------------------------------------------------

func TestArrayDimLoadStore(t *testing.T) {
	// DIM arr(4)  — 5 elements (0..4)
	// arr(2) = 99
	// result = arr(2)
	chunk := makeChunk(
		[]Value{
			IntVal(4),          // 0: dimension size (upper bound)
			StringVal("arr"),   // 1: array name
			IntVal(2),          // 2: index
			IntVal(99),         // 3: value to store
		},
		[]Instruction{
			// DIM arr(4)
			{Op: OpPush, Operand: 0},      // push dimension size 4
			{Op: OpPush, Operand: 1},      // push array name
			{Op: OpDimArray, Operand: 1},  // 1 dimension

			// arr(2) = 99
			// StoreArray pops value first, then indices
			{Op: OpPush, Operand: 2},      // push index 2
			{Op: OpPush, Operand: 3},      // push 99 (value)
			{Op: OpStoreArray, Operand: 1}, // store into arr

			// load arr(2)
			{Op: OpPush, Operand: 2},      // push index 2
			{Op: OpLoadArray, Operand: 1}, // load from arr

			{Op: OpHalt},
		},
	)
	vm, _ := runVM(t, chunk)
	result := vm.stack[0]
	if result.Type != ValInt || result.Int != 99 {
		t.Fatalf("expected Int(99), got %s", result)
	}
}

// ---------------------------------------------------------------------------
// Test: CALL/RET
// ---------------------------------------------------------------------------

func TestCallRet(t *testing.T) {
	// Main: push 10, call sub at 4, push 30, halt
	// Sub at 4: push 20, ret
	// Stack should be [10, 20, 30]
	chunk := makeChunk(
		[]Value{IntVal(10), IntVal(30), IntVal(20)},
		[]Instruction{
			{Op: OpPush, Operand: 0},    // 0: push 10
			{Op: OpCall, Operand: 4},    // 1: call sub at 4
			{Op: OpPush, Operand: 1},    // 2: push 30
			{Op: OpHalt},                // 3: halt
			{Op: OpPush, Operand: 2},    // 4: push 20
			{Op: OpRet},                 // 5: return
		},
	)
	vm, _ := runVM(t, chunk)
	if vm.sp != 3 {
		t.Fatalf("expected sp=3, got %d", vm.sp)
	}
	if vm.stack[0].Int != 10 || vm.stack[1].Int != 20 || vm.stack[2].Int != 30 {
		t.Fatalf("expected [10,20,30], got [%d,%d,%d]",
			vm.stack[0].Int, vm.stack[1].Int, vm.stack[2].Int)
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
// Test: Opcode String()
// ---------------------------------------------------------------------------

func TestOpcodeString(t *testing.T) {
	if OpPush.String() != "PUSH" {
		t.Fatalf("expected PUSH, got %s", OpPush.String())
	}
	if OpHalt.String() != "HALT" {
		t.Fatalf("expected HALT, got %s", OpHalt.String())
	}
	// Unknown opcode.
	unknown := Opcode(255)
	s := unknown.String()
	if !strings.Contains(s, "UNKNOWN") {
		t.Fatalf("expected UNKNOWN in string, got %s", s)
	}
}

// ---------------------------------------------------------------------------
// Test: BuiltinID String()
// ---------------------------------------------------------------------------

func TestBuiltinIDString(t *testing.T) {
	if BuiltinAbs.String() != "ABS" {
		t.Fatalf("expected ABS, got %s", BuiltinAbs.String())
	}
	unknown := BuiltinID(255)
	s := unknown.String()
	if !strings.Contains(s, "UNKNOWN") {
		t.Fatalf("expected UNKNOWN in string, got %s", s)
	}
}

// ---------------------------------------------------------------------------
// Test: Chunk helpers
// ---------------------------------------------------------------------------

func TestChunkAddConstant(t *testing.T) {
	c := &Chunk{}
	idx0 := c.AddConstant(IntVal(1))
	idx1 := c.AddConstant(StringVal("foo"))
	if idx0 != 0 || idx1 != 1 {
		t.Fatalf("expected indices 0, 1 but got %d, %d", idx0, idx1)
	}
	if len(c.Constants) != 2 {
		t.Fatalf("expected 2 constants, got %d", len(c.Constants))
	}
}

func TestChunkEmit(t *testing.T) {
	c := &Chunk{}
	pos := c.Emit(OpPush, 0, 10)
	if pos != 0 {
		t.Fatalf("expected position 0, got %d", pos)
	}
	if len(c.Code) != 1 {
		t.Fatalf("expected 1 instruction, got %d", len(c.Code))
	}
	if c.Lines[0] != 10 {
		t.Fatalf("expected line 10, got %d", c.Lines[0])
	}
}

func TestChunkDisassemble(t *testing.T) {
	c := &Chunk{}
	c.Emit(OpPush, 0, 1)
	c.Emit(OpHalt, 0, 1)
	dis := c.Disassemble()
	if !strings.Contains(dis, "PUSH") || !strings.Contains(dis, "HALT") {
		t.Fatalf("disassembly missing expected instructions: %s", dis)
	}
}

// ---------------------------------------------------------------------------
// Test: Value helpers
// ---------------------------------------------------------------------------

func TestValueHelpers(t *testing.T) {
	iv := IntVal(10)
	if iv.asFloat() != 10.0 {
		t.Fatal("IntVal.asFloat failed")
	}
	if iv.asInt() != 10 {
		t.Fatal("IntVal.asInt failed")
	}
	if !iv.isTruthy() {
		t.Fatal("IntVal(10) should be truthy")
	}

	fv := FloatVal(0.0)
	if fv.isTruthy() {
		t.Fatal("FloatVal(0.0) should not be truthy")
	}

	sv := StringVal("")
	if sv.isTruthy() {
		t.Fatal("StringVal(\"\") should not be truthy")
	}
	sv2 := StringVal("x")
	if !sv2.isTruthy() {
		t.Fatal("StringVal(\"x\") should be truthy")
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
// Test: complex program — compute factorial of 5
// ---------------------------------------------------------------------------

// Simulates:
//   n = 5
//   result = 1
//   top:
//     result = result * n
//     n = n - 1
//     IF n > 0 GOTO top
//   => result = 120

func TestFactorial(t *testing.T) {
	chunk := makeChunk(
		[]Value{
			IntVal(5),             // 0: initial n
			StringVal("n"),        // 1: variable name "n"
			IntVal(1),             // 2: initial result / decrement
			StringVal("result"),   // 3: variable name "result"
			IntVal(0),             // 4: zero for comparison
		},
		[]Instruction{
			// n = 5
			{Op: OpPush, Operand: 0},   // 0: push 5
			{Op: OpStore, Operand: 1},  // 1: store to "n"

			// result = 1
			{Op: OpPush, Operand: 2},   // 2: push 1
			{Op: OpStore, Operand: 3},  // 3: store to "result"

			// top (instruction 4):
			// result = result * n
			{Op: OpLoad, Operand: 3},   // 4: load result
			{Op: OpLoad, Operand: 1},   // 5: load n
			{Op: OpMul},                // 6: result * n
			{Op: OpStore, Operand: 3},  // 7: store to result

			// n = n - 1
			{Op: OpLoad, Operand: 1},   // 8: load n
			{Op: OpPush, Operand: 2},   // 9: push 1
			{Op: OpSub},                // 10: n - 1
			{Op: OpStore, Operand: 1},  // 11: store to n

			// IF n > 0 GOTO top
			{Op: OpLoad, Operand: 1},      // 12: load n
			{Op: OpPush, Operand: 4},      // 13: push 0
			{Op: OpGt},                    // 14: n > 0?
			{Op: OpJmpTrue, Operand: 4},   // 15: if true, jump to top (4)

			// Load result for verification
			{Op: OpLoad, Operand: 3},      // 16: load result
			{Op: OpHalt},                  // 17: halt
		},
	)
	vm, _ := runVM(t, chunk)
	result := vm.stack[0]
	if result.Type != ValInt || result.Int != 120 {
		t.Fatalf("expected 120, got %s", result)
	}
}

// ---------------------------------------------------------------------------
// Test: PrintTab and PrintSemicolon
// ---------------------------------------------------------------------------

func TestPrintTabAndSemicolon(t *testing.T) {
	chunk := makeChunk(
		[]Value{StringVal("A"), StringVal("B")},
		[]Instruction{
			{Op: OpPush, Operand: 0},
			{Op: OpPrint},
			{Op: OpPrintTab},
			{Op: OpPush, Operand: 1},
			{Op: OpPrint},
			{Op: OpPrintSemicolon}, // should suppress newline
			{Op: OpHalt},
		},
	)
	_, output := runVM(t, chunk)
	expected := "A\tB"
	if output != expected {
		t.Fatalf("expected %q, got %q", expected, output)
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
