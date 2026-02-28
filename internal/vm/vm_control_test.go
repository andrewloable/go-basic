package vm

import (
	"strings"
	"testing"
)

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

func TestExecToLong(t *testing.T) {
	// OpToLong truncates a float to int.
	chunk := makeChunk(
		[]Value{FloatVal(3.9)},
		[]Instruction{
			{Op: OpPush, Operand: 0},
			{Op: OpToLong},
			{Op: OpHalt},
		},
	)
	vm, _ := runVM(t, chunk)
	if vm.stack[0].asInt() != 3 {
		t.Errorf("execToLong(3.9): expected 3, got %v", vm.stack[0])
	}
}

func TestExecToSingle(t *testing.T) {
	// OpToSingle converts to float32 precision.
	chunk := makeChunk(
		[]Value{FloatVal(3.14159265358979)},
		[]Instruction{
			{Op: OpPush, Operand: 0},
			{Op: OpToSingle},
			{Op: OpHalt},
		},
	)
	vm, _ := runVM(t, chunk)
	// float32(3.14159265358979) ≈ 3.1415927
	if vm.stack[0].Type != ValFloat {
		t.Errorf("execToSingle: expected float result, got %v", vm.stack[0])
	}
}

func TestExecToDouble(t *testing.T) {
	// OpToDouble converts to float64.
	chunk := makeChunk(
		[]Value{IntVal(7)},
		[]Instruction{
			{Op: OpPush, Operand: 0},
			{Op: OpToDouble},
			{Op: OpHalt},
		},
	)
	vm, _ := runVM(t, chunk)
	if vm.stack[0].asFloat() != 7.0 {
		t.Errorf("execToDouble(7): expected 7.0, got %v", vm.stack[0])
	}
}

func TestExecToString(t *testing.T) {
	// OpToString converts a numeric value to its string representation.
	chunk := makeChunk(
		[]Value{IntVal(42)},
		[]Instruction{
			{Op: OpPush, Operand: 0},
			{Op: OpToString},
			{Op: OpHalt},
		},
	)
	vm, _ := runVM(t, chunk)
	if vm.stack[0].Type != ValString {
		t.Errorf("execToString: expected string result, got %v", vm.stack[0])
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

func TestArrayOutOfBounds(t *testing.T) {
	src := `DIM a(3)
PRINT a(10)`
	l := mustParse(t, src)
	v := NewVM(l)
	var buf strings.Builder
	v.SetOutput(&buf)
	err := v.Run()
	if err == nil {
		t.Error("expected runtime error for array out of bounds")
	}
}

func TestExecStoreArrayOutOfBounds(t *testing.T) {
	// Writing to an out-of-bounds array index should produce a runtime error.
	src := `DIM a(3)
a(10) = 99`
	chunk, compiler := compileSource(t, src)
	v := NewVM(chunk)
	var buf strings.Builder
	v.SetOutput(&buf)
	if pool := compiler.DataPool(); len(pool) > 0 {
		v.SetDataPool(pool)
	}
	err := v.Run()
	if err == nil {
		t.Error("expected runtime error for out-of-bounds array write")
	}
}

// ---------------------------------------------------------------------------
// Test: FOR-like loop
// ---------------------------------------------------------------------------

// This test simulates:
//
//	sum = 0
//	i = 1
//	top:
//	  sum = sum + i
//	  i = i + 1
//	  IF i <= 10 GOTO top
//	result: sum = 55
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

// Simulates:
//
//	FOR i = 3 TO 1 STEP -1
//	  PRINT i
//	NEXT i
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

// Simulates:
//
//	n = 5
//	result = 1
//	top:
//	  result = result * n
//	  n = n - 1
//	  IF n > 0 GOTO top
//	=> result = 120
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
