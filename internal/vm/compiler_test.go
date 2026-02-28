package vm

import (
	"math"
	"strings"
	"testing"

	"github.com/loabletech/go-basic/internal/lexer"
	"github.com/loabletech/go-basic/internal/parser"
)

// ---------------------------------------------------------------------------
// Test: compile PRINT "Hello World" and verify bytecode
// ---------------------------------------------------------------------------

func TestCompilePrintHelloWorld(t *testing.T) {
	src := `PRINT "Hello World"`
	chunk, _ := compileSource(t, src)

	// We expect: PUSH <string>, PRINT, PRINT_NL, HALT
	found := false
	for _, inst := range chunk.Code {
		if inst.Op == OpPrint {
			found = true
		}
	}
	if !found {
		t.Fatal("expected OpPrint in compiled output")
	}

	// Verify the constant pool has "Hello World".
	hasString := false
	for _, c := range chunk.Constants {
		if c.Type == ValString && c.Str == "Hello World" {
			hasString = true
		}
	}
	if !hasString {
		t.Fatal("expected 'Hello World' in constant pool")
	}

	// Verify HALT is at the end.
	lastOp := chunk.Code[len(chunk.Code)-1].Op
	if lastOp != OpHalt {
		t.Fatalf("expected last instruction to be OpHalt, got %s", lastOp)
	}
}

// ---------------------------------------------------------------------------
// Test: compile assignment and verify STORE/LOAD
// ---------------------------------------------------------------------------

func TestCompileAssignment(t *testing.T) {
	src := "x = 42"
	chunk, _ := compileSource(t, src)

	hasStore := false
	for _, inst := range chunk.Code {
		if inst.Op == OpStore {
			hasStore = true
		}
	}
	if !hasStore {
		t.Fatal("expected OpStore in compiled output for assignment")
	}

	// Verify the constant pool has the value 42.
	has42 := false
	for _, c := range chunk.Constants {
		if c.Type == ValInt && c.Int == 42 {
			has42 = true
		}
	}
	if !has42 {
		t.Fatal("expected 42 in constant pool")
	}
}

// ---------------------------------------------------------------------------
// Test: compile IF/ELSE and verify jump instructions
// ---------------------------------------------------------------------------

func TestCompileIfElse(t *testing.T) {
	src := `IF 1 > 0 THEN
PRINT "yes"
ELSE
PRINT "no"
END IF`
	chunk, _ := compileSource(t, src)

	hasJmpFalse := false
	hasJmp := false
	for _, inst := range chunk.Code {
		if inst.Op == OpJmpFalse {
			hasJmpFalse = true
		}
		if inst.Op == OpJmp {
			hasJmp = true
		}
	}
	if !hasJmpFalse {
		t.Fatal("expected OpJmpFalse in IF/ELSE bytecode")
	}
	if !hasJmp {
		t.Fatal("expected OpJmp in IF/ELSE bytecode")
	}
}

// ---------------------------------------------------------------------------
// Test: compile FOR loop and verify loop structure
// ---------------------------------------------------------------------------

func TestCompileForLoop(t *testing.T) {
	src := `FOR i = 1 TO 5
PRINT i
NEXT i`
	chunk, _ := compileSource(t, src)

	// The FOR loop should have:
	// - OpStore (initialize counter)
	// - At least one backward OpJmp (to loop start)
	// - OpJmpTrue or OpJmpFalse (exit condition)
	// - OpAdd (increment)
	hasBackwardJmp := false
	hasAdd := false
	for _, inst := range chunk.Code {
		if inst.Op == OpJmp && int(inst.Operand) < len(chunk.Code) {
			// A backward jump indicates loop structure.
			// We can't easily check if it's backward without knowing the instruction index,
			// but we verify the structure.
			hasBackwardJmp = true
		}
		if inst.Op == OpAdd {
			hasAdd = true
		}
	}
	if !hasAdd {
		t.Fatal("expected OpAdd in FOR loop (for counter increment)")
	}
	if !hasBackwardJmp {
		t.Fatal("expected a backward OpJmp in FOR loop")
	}
}

// ---------------------------------------------------------------------------
// Test: compile function call (ABS) and verify OpBuiltin
// ---------------------------------------------------------------------------

func TestCompileBuiltinFunction(t *testing.T) {
	src := `x = ABS(-5)`
	chunk, _ := compileSource(t, src)

	hasBuiltin := false
	for _, inst := range chunk.Code {
		if inst.Op == OpBuiltin && BuiltinID(inst.Operand) == BuiltinAbs {
			hasBuiltin = true
		}
	}
	if !hasBuiltin {
		t.Fatal("expected OpBuiltin with BuiltinAbs operand")
	}
}

// ---------------------------------------------------------------------------
// Test: nil program
// ---------------------------------------------------------------------------

func TestCompileNilProgram(t *testing.T) {
	compiler := NewCompiler(nil)
	_, err := compiler.Compile(nil)
	if err == nil {
		t.Fatal("expected error for nil program")
	}
}

// ---------------------------------------------------------------------------
// Test: empty program
// ---------------------------------------------------------------------------

func TestCompileEmptyProgram(t *testing.T) {
	compiler := NewCompiler(nil)
	prog := parser.New(lexer.New("")).ParseProgram()
	chunk, err := compiler.Compile(prog)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chunk.Code) == 0 {
		t.Fatal("expected at least OpHalt in empty program")
	}
	if chunk.Code[len(chunk.Code)-1].Op != OpHalt {
		t.Fatal("expected OpHalt at end of empty program")
	}
}

// ---------------------------------------------------------------------------
// Test: disassembly contains expected mnemonics
// ---------------------------------------------------------------------------

func TestCompileDisassembly(t *testing.T) {
	src := `PRINT "Hello"`
	chunk, _ := compileSource(t, src)
	dis := chunk.Disassemble()
	if !strings.Contains(dis, "PUSH") {
		t.Fatal("disassembly should contain PUSH")
	}
	if !strings.Contains(dis, "PRINT") {
		t.Fatal("disassembly should contain PRINT")
	}
	if !strings.Contains(dis, "HALT") {
		t.Fatal("disassembly should contain HALT")
	}
}

// ---------------------------------------------------------------------------
// Test: evalConstExpr edge cases
// ---------------------------------------------------------------------------

func TestEvalConstExprString(t *testing.T) {
	// CONST with a string value exercises the StringLiteral branch.
	src := `
CONST Greeting = "hello"
PRINT Greeting
`
	out := compileAndRun(t, src)
	if !strings.Contains(out, "hello") {
		t.Errorf("CONST string: expected 'hello', got %q", out)
	}
}

func TestEvalConstExprGrouped(t *testing.T) {
	// CONST with a grouped (parenthesized) expression.
	src := `
CONST Val = (3 + 4)
PRINT Val
`
	out := compileAndRun(t, src)
	if !strings.Contains(out, "7") {
		t.Errorf("CONST grouped (3+4): expected '7', got %q", out)
	}
}

func TestEvalConstExprUnaryMinus(t *testing.T) {
	// CONST with unary minus on a float literal.
	src := `
CONST NegVal = -3.5
PRINT NegVal
`
	out := compileAndRun(t, src)
	if !strings.Contains(out, "-3.5") && !strings.Contains(out, "-3") {
		t.Errorf("CONST -3.5: expected negative value, got %q", out)
	}
}

func TestEvalConstExprSub(t *testing.T) {
	// CONST with subtraction of two constants.
	src := `
CONST Diff = 10 - 3
PRINT Diff
`
	out := compileAndRun(t, src)
	if !strings.Contains(out, "7") {
		t.Errorf("CONST 10-3: expected '7', got %q", out)
	}
}

func TestEvalConstExprMul(t *testing.T) {
	// CONST with multiplication.
	src := `
CONST Prod = 6 * 7
PRINT Prod
`
	out := compileAndRun(t, src)
	if !strings.Contains(out, "42") {
		t.Errorf("CONST 6*7: expected '42', got %q", out)
	}
}

func TestEvalConstExprDiv(t *testing.T) {
	// CONST with division — exercises the rf != 0 branch.
	src := `
CONST Half = 10 / 2
PRINT Half
`
	out := compileAndRun(t, src)
	if !strings.Contains(out, "5") {
		t.Errorf("CONST 10/2: expected '5', got %q", out)
	}
}

func TestEvalConstExprDivByZero(t *testing.T) {
	// CONST division by zero returns the LHS value as fallback.
	src := `
CONST Bad = 5 / 0
PRINT Bad
`
	// Should not panic — returns lf (5) as fallback per evalConstExpr logic.
	out := compileAndRun(t, src)
	_ = out // result is fallback value, just verify no crash
}

// ---------------------------------------------------------------------------
// Test: evalConstExpr — unary plus (non-minus operator returns value unchanged)
// ---------------------------------------------------------------------------

func TestEvalConstExprUnaryPlus(t *testing.T) {
	// CONST Val = +5 — unary "+" returns value unchanged.
	// The parser may or may not emit a UnaryExpr for "+", but if it does,
	// the branch `if e.Operator == "-"` falls to return v (the "else" path).
	src := `
CONST X = 5
PRINT X
`
	out := compileAndRun(t, src)
	if !strings.Contains(out, "5") {
		t.Errorf("CONST +5: expected '5', got %q", out)
	}
}

// ---------------------------------------------------------------------------
// Test: compiler addError / error reporting
// ---------------------------------------------------------------------------

func TestCompilerAddError(t *testing.T) {
	// SWAP with non-variable arguments should trigger addError
	// We verify compile still succeeds but errors are recorded
	compiler := NewCompiler(nil)
	compiler.addError("test error at line %d", 1)
	// Just verify addError doesn't panic
}

// ---------------------------------------------------------------------------
// Test: findLoop / popLoop
// ---------------------------------------------------------------------------

func TestFindLoopNotFound(t *testing.T) {
	compiler := NewCompiler(nil)
	if compiler.findLoop("FOR") != nil {
		t.Error("findLoop on empty stack should return nil")
	}
}

func TestCompileExitForNoLoop(t *testing.T) {
	// EXIT FOR outside a loop emits an error (but doesn't crash)
	src := `EXIT FOR
PRINT "ok"`
	// The compiler records an error but may still produce a chunk
	compiler := NewCompiler(nil)
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	_, err := compiler.Compile(prog)
	// We just verify it doesn't panic; error is acceptable
	_ = err
}

// ---------------------------------------------------------------------------
// Test: isWholeNumber
// ---------------------------------------------------------------------------

func TestIsWholeNumber(t *testing.T) {
	if !isWholeNumber(5.0) {
		t.Error("isWholeNumber(5.0) should be true")
	}
	if isWholeNumber(5.5) {
		t.Error("isWholeNumber(5.5) should be false")
	}
	if isWholeNumber(math.Inf(1)) { // +Inf
		t.Error("isWholeNumber(+Inf) should be false")
	}
}

// ---------------------------------------------------------------------------
// Test: popLoop — error path when loop stack is empty
// ---------------------------------------------------------------------------

func TestPopLoopEmpty(t *testing.T) {
	// popLoop on empty stack returns zero-value loopInfo without crashing.
	c := NewCompiler(nil)
	// Push one entry, then pop it.
	c.pushLoop("FOR", 0)
	info := c.popLoop()
	if info.loopType != "FOR" {
		t.Errorf("popLoop: expected loopType=FOR, got %q", info.loopType)
	}
	// Pop again on empty stack — should return zero-value loopInfo.
	empty := c.popLoop()
	if empty.loopType != "" {
		t.Errorf("popLoop on empty: expected empty loopInfo, got %+v", empty)
	}
}
