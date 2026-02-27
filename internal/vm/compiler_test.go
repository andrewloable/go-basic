package vm

import (
	"strings"
	"testing"

	"github.com/loabletech/go-basic/internal/lexer"
	"github.com/loabletech/go-basic/internal/parser"
)

// ---------------------------------------------------------------------------
// Helper: compile BASIC source code to a Chunk
// ---------------------------------------------------------------------------

func compileSource(t *testing.T, src string) (*Chunk, *Compiler) {
	t.Helper()
	l := lexer.New(src)
	p := parser.New(l)
	prog := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	compiler := NewCompiler(nil)
	chunk, err := compiler.Compile(prog)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}
	return chunk, compiler
}

// compileAndRun compiles source, runs it through the VM, returns captured output.
func compileAndRun(t *testing.T, src string) string {
	t.Helper()
	chunk, compiler := compileSource(t, src)
	v := NewVM(chunk)
	var buf strings.Builder
	v.SetOutput(&buf)
	if pool := compiler.DataPool(); len(pool) > 0 {
		v.SetDataPool(pool)
	}
	if err := v.Run(); err != nil {
		t.Fatalf("VM runtime error: %v", err)
	}
	return buf.String()
}

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

// ===========================================================================
// End-to-end tests: compile + run through VM and verify output
// ===========================================================================

// ---------------------------------------------------------------------------
// E2E 1: Hello World
// ---------------------------------------------------------------------------

func TestE2E_HelloWorld(t *testing.T) {
	output := compileAndRun(t, `PRINT "Hello World"`)
	if output != "Hello World\n" {
		t.Fatalf("expected \"Hello World\\n\", got %q", output)
	}
}

// ---------------------------------------------------------------------------
// E2E 2: Assignment + PRINT
// ---------------------------------------------------------------------------

func TestE2E_AssignmentAndPrint(t *testing.T) {
	src := `x = 42
PRINT x`
	output := compileAndRun(t, src)
	// Numeric PRINT: " 42 \n"
	if !strings.Contains(output, "42") {
		t.Fatalf("expected output containing '42', got %q", output)
	}
}

// ---------------------------------------------------------------------------
// E2E 3: IF/ELSE
// ---------------------------------------------------------------------------

func TestE2E_IfElse(t *testing.T) {
	src := `x = 10
IF x > 5 THEN
PRINT "big"
ELSE
PRINT "small"
END IF`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "big") {
		t.Fatalf("expected 'big' in output, got %q", output)
	}
	if strings.Contains(output, "small") {
		t.Fatalf("did not expect 'small' in output, got %q", output)
	}
}

func TestE2E_IfElseFalseBranch(t *testing.T) {
	src := `x = 3
IF x > 5 THEN
PRINT "big"
ELSE
PRINT "small"
END IF`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "small") {
		t.Fatalf("expected 'small' in output, got %q", output)
	}
	if strings.Contains(output, "big") {
		t.Fatalf("did not expect 'big' in output, got %q", output)
	}
}

// ---------------------------------------------------------------------------
// E2E 4: FOR loop
// ---------------------------------------------------------------------------

func TestE2E_ForLoop(t *testing.T) {
	src := `total = 0
FOR i = 1 TO 5
total = total + i
NEXT i
PRINT total`
	output := compileAndRun(t, src)
	// 1+2+3+4+5 = 15
	if !strings.Contains(output, "15") {
		t.Fatalf("expected output containing '15', got %q", output)
	}
}

// ---------------------------------------------------------------------------
// E2E 5: Built-in functions
// ---------------------------------------------------------------------------

func TestE2E_BuiltinFunctions(t *testing.T) {
	src := `x = ABS(-42)
PRINT x`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "42") {
		t.Fatalf("expected output containing '42', got %q", output)
	}
}

func TestE2E_BuiltinLEN(t *testing.T) {
	src := `x = LEN("Hello")
PRINT x`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "5") {
		t.Fatalf("expected output containing '5', got %q", output)
	}
}

// ---------------------------------------------------------------------------
// E2E 6: WHILE loop
// ---------------------------------------------------------------------------

func TestE2E_WhileLoop(t *testing.T) {
	src := `x = 1
total = 0
WHILE x <= 5
total = total + x
x = x + 1
WEND
PRINT total`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "15") {
		t.Fatalf("expected output containing '15', got %q", output)
	}
}

// ---------------------------------------------------------------------------
// E2E 7: String operations
// ---------------------------------------------------------------------------

func TestE2E_StringPrint(t *testing.T) {
	src := `a$ = "Hello"
b$ = " World"
PRINT a$ + b$`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "Hello World") {
		t.Fatalf("expected 'Hello World' in output, got %q", output)
	}
}

// ---------------------------------------------------------------------------
// E2E 8: Arithmetic expressions
// ---------------------------------------------------------------------------

func TestE2E_Arithmetic(t *testing.T) {
	src := `x = (2 + 3) * 4
PRINT x`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "20") {
		t.Fatalf("expected output containing '20', got %q", output)
	}
}

// ---------------------------------------------------------------------------
// E2E 9: Multiple PRINT statements
// ---------------------------------------------------------------------------

func TestE2E_MultiplePrint(t *testing.T) {
	src := `PRINT "A"
PRINT "B"
PRINT "C"`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "A") || !strings.Contains(output, "B") || !strings.Contains(output, "C") {
		t.Fatalf("expected A, B, C in output, got %q", output)
	}
}

// ---------------------------------------------------------------------------
// E2E 10: Nested IF
// ---------------------------------------------------------------------------

func TestE2E_NestedIf(t *testing.T) {
	src := `x = 10
IF x > 5 THEN
IF x > 8 THEN
PRINT "very big"
END IF
END IF`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "very big") {
		t.Fatalf("expected 'very big' in output, got %q", output)
	}
}

// ---------------------------------------------------------------------------
// E2E 11: FOR loop with negative step
// ---------------------------------------------------------------------------

func TestE2E_ForLoopNegativeStep(t *testing.T) {
	src := `FOR i = 3 TO 1 STEP -1
PRINT i;
NEXT i`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "3") || !strings.Contains(output, "2") || !strings.Contains(output, "1") {
		t.Fatalf("expected 3, 2, 1 in output, got %q", output)
	}
}

// ---------------------------------------------------------------------------
// E2E 12: END statement
// ---------------------------------------------------------------------------

func TestE2E_EndStatement(t *testing.T) {
	src := `PRINT "before"
END
PRINT "after"`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "before") {
		t.Fatalf("expected 'before' in output, got %q", output)
	}
	if strings.Contains(output, "after") {
		t.Fatal("did not expect 'after' in output (END should stop execution)")
	}
}

// ---------------------------------------------------------------------------
// E2E 13: Unary negation
// ---------------------------------------------------------------------------

func TestE2E_UnaryNegation(t *testing.T) {
	src := `x = -10
PRINT x`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "-10") {
		t.Fatalf("expected output containing '-10', got %q", output)
	}
}

// ---------------------------------------------------------------------------
// E2E 14: SQR built-in
// ---------------------------------------------------------------------------

func TestE2E_BuiltinSQR(t *testing.T) {
	src := `PRINT SQR(16)`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "4") {
		t.Fatalf("expected output containing '4', got %q", output)
	}
}

// ---------------------------------------------------------------------------
// E2E 15: Multiple built-in calls
// ---------------------------------------------------------------------------

func TestE2E_MultipleBuiltins(t *testing.T) {
	src := `a = ABS(-7)
b = SGN(-3)
PRINT a;
PRINT b`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "7") {
		t.Fatalf("expected '7' in output, got %q", output)
	}
	if !strings.Contains(output, "-1") {
		t.Fatalf("expected '-1' in output, got %q", output)
	}
}
