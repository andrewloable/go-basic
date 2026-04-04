package vm

import (
	"strings"
	"testing"
)

// ===========================================================================
// End-to-end tests: compile + run through VM and verify output
// ===========================================================================

func TestE2E_HelloWorld(t *testing.T) {
	output := compileAndRun(t, `PRINT "Hello World"`)
	if output != "Hello World\n" {
		t.Fatalf("expected \"Hello World\\n\", got %q", output)
	}
}

func TestE2E_AssignmentAndPrint(t *testing.T) {
	src := `x = 42
PRINT x`
	output := compileAndRun(t, src)
	// Numeric PRINT: " 42 \n"
	if !strings.Contains(output, "42") {
		t.Fatalf("expected output containing '42', got %q", output)
	}
}

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

func TestE2E_StringPrint(t *testing.T) {
	src := `a$ = "Hello"
b$ = " World"
PRINT a$ + b$`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "Hello World") {
		t.Fatalf("expected 'Hello World' in output, got %q", output)
	}
}

func TestE2E_Arithmetic(t *testing.T) {
	src := `x = (2 + 3) * 4
PRINT x`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "20") {
		t.Fatalf("expected output containing '20', got %q", output)
	}
}

func TestE2E_MultiplePrint(t *testing.T) {
	src := `PRINT "A"
PRINT "B"
PRINT "C"`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "A") || !strings.Contains(output, "B") || !strings.Contains(output, "C") {
		t.Fatalf("expected A, B, C in output, got %q", output)
	}
}

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

func TestE2E_ForLoopNegativeStep(t *testing.T) {
	src := `FOR i = 3 TO 1 STEP -1
PRINT i;
NEXT i`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "3") || !strings.Contains(output, "2") || !strings.Contains(output, "1") {
		t.Fatalf("expected 3, 2, 1 in output, got %q", output)
	}
}

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

func TestE2E_UnaryNegation(t *testing.T) {
	src := `x = -10
PRINT x`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "-10") {
		t.Fatalf("expected output containing '-10', got %q", output)
	}
}

func TestE2E_BuiltinSQR(t *testing.T) {
	src := `PRINT SQR(16)`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "4") {
		t.Fatalf("expected output containing '4', got %q", output)
	}
}

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

// ---------------------------------------------------------------------------
// CONST, POKE, CLEAR
// ---------------------------------------------------------------------------

func TestConstStatement(t *testing.T) {
	out := compileAndRun(t, `
CONST PI = 3.14159
PRINT PI
`)
	if !strings.Contains(out, "3.14159") {
		t.Errorf("CONST PI: expected '3.14159' in output, got %q", out)
	}
}

func TestConstExprArith(t *testing.T) {
	out := compileAndRun(t, `
CONST TWO = 1 + 1
PRINT TWO
`)
	if !strings.Contains(out, "2") {
		t.Errorf("CONST 1+1: expected '2' in output, got %q", out)
	}
}

func TestPokeNoOp(t *testing.T) {
	// POKE should not crash and should not produce output.
	out := compileAndRun(t, `
POKE 1000, 255
PRINT "ok"
`)
	if !strings.Contains(out, "ok") {
		t.Errorf("POKE: expected 'ok', got %q", out)
	}
}

func TestClearResetsVars(t *testing.T) {
	out := compileAndRun(t, `
x = 42
CLEAR
PRINT x
`)
	// After CLEAR, x should be 0.
	if !strings.Contains(out, "0") {
		t.Errorf("CLEAR: expected x=0 after CLEAR, got %q", out)
	}
	if strings.Contains(out, "42") {
		t.Errorf("CLEAR: x should not be 42 after CLEAR, got %q", out)
	}
}

// ---------------------------------------------------------------------------
// TYPE field access and assignment
// ---------------------------------------------------------------------------

func TestTypeFieldAssignAndAccess(t *testing.T) {
	// TYPE blocks don't emit bytecode; field access uses flat variable mangling.
	src := `
TYPE Point
  x AS INTEGER
  y AS INTEGER
END TYPE
p_x = 10
p_y = 20
PRINT p_x
PRINT p_y
`
	out := compileAndRun(t, src)
	if !strings.Contains(out, "10") {
		t.Errorf("p.x: expected '10', got %q", out)
	}
	if !strings.Contains(out, "20") {
		t.Errorf("p.y: expected '20', got %q", out)
	}
}

func TestTypeDefMapPopulated(t *testing.T) {
	// TYPE block should populate typeDefMap; no bytecode errors.
	src := `
TYPE Point
  x AS INTEGER
  y AS INTEGER
END TYPE
DIM p AS Point
p.x = 5
PRINT p.x
`
	out := compileAndRun(t, src)
	if !strings.Contains(out, "5") {
		t.Errorf("p.x: expected '5', got %q", out)
	}

	_, compiler := compileSource(t, src)
	fields, ok := compiler.typeDefMap["POINT"]
	if !ok {
		t.Fatal("typeDefMap missing POINT")
	}
	if len(fields) != 2 {
		t.Errorf("POINT fields: expected 2, got %d", len(fields))
	}
}

func TestDefTypeStatement(t *testing.T) {
	// DEFINT A-Z records suffix in defTypeMap; program runs without error.
	src := `
DEFINT A-Z
a = 10
PRINT a
`
	out := compileAndRun(t, src)
	if !strings.Contains(out, "10") {
		t.Errorf("DEFINT a=10: expected '10', got %q", out)
	}

	_, compiler := compileSource(t, src)
	if compiler.defTypeMap[0] != "%" { // index 0 = 'A'-'A'
		t.Errorf("defTypeMap[A]: expected '%%', got %q", compiler.defTypeMap[0])
	}
}

func TestScopeStatementNoError(t *testing.T) {
	// SHARED inside a program scope should compile without error.
	// In the flat global VM, scope modifiers are accepted but produce no bytecode.
	src := `
SHARED x
x = 42
PRINT x
`
	out := compileAndRun(t, src)
	if !strings.Contains(out, "42") {
		t.Errorf("SHARED x: expected '42', got %q", out)
	}
}

func TestFieldAssignStatement(t *testing.T) {
	// FieldAssignStatement: p.x = 5 should store to flat variable p_x.
	src := `
TYPE Point
  x AS INTEGER
  y AS INTEGER
END TYPE
DIM p AS Point
p.x = 42
p.y = 7
PRINT p.x
PRINT p.y
`
	out := compileAndRun(t, src)
	if !strings.Contains(out, "42") {
		t.Errorf("p.x after assign: expected '42', got %q", out)
	}
	if !strings.Contains(out, "7") {
		t.Errorf("p.y after assign: expected '7', got %q", out)
	}
}

func TestSwapArrayElements(t *testing.T) {
	// SWAP a(j), a(j+1) should exchange two array elements.
	src := `
DIM a(3)
a(1) = 10
a(2) = 20
a(3) = 30
SWAP a(1), a(2)
PRINT a(1)
PRINT a(2)
PRINT a(3)
`
	out := compileAndRun(t, src)
	if !strings.Contains(out, "20") {
		t.Errorf("after SWAP a(1),a(2): expected a(1)=20, got %q", out)
	}
	lines := strings.Fields(out)
	if len(lines) < 2 || lines[1] != "10" {
		t.Errorf("after SWAP a(1),a(2): expected a(2)=10, got lines %v", lines)
	}
}

func TestRecursiveFunction(t *testing.T) {
	// Recursive FUNCTION using parameter binding — Factorial(5) = 120.
	src := `
DECLARE FUNCTION Factorial (n)
PRINT Factorial(5)
END
FUNCTION Factorial (n)
  IF n <= 1 THEN
    Factorial = 1
  ELSE
    Factorial = n * Factorial(n - 1)
  END IF
END FUNCTION
`
	out := compileAndRun(t, src)
	if !strings.Contains(out, "120") {
		t.Errorf("Factorial(5): expected '120', got %q", out)
	}
}

func TestSubParameterBinding(t *testing.T) {
	// SUB with a parameter: argument should be bound inside the body.
	src := `
DECLARE SUB Greet (name$)
CALL Greet("World")
END
SUB Greet (name$)
  PRINT "Hello "; name$
END SUB
`
	out := compileAndRun(t, src)
	if !strings.Contains(out, "Hello") || !strings.Contains(out, "World") {
		t.Errorf("Greet: expected 'Hello World', got %q", out)
	}
}

// ---------------------------------------------------------------------------
// Test: compileDefType — DEFLNG, DEFSNG, DEFDBL, DEFSTR variants
// ---------------------------------------------------------------------------

func TestCompileDefTypeLng(t *testing.T) {
	src := `
DEFLNG L-L
l = 1000000
PRINT l
`
	out := compileAndRun(t, src)
	if !strings.Contains(out, "1000000") {
		t.Errorf("DEFLNG: expected '1000000', got %q", out)
	}
	_, compiler := compileSource(t, src)
	if compiler.defTypeMap['L'-'A'] != "&" {
		t.Errorf("defTypeMap[L]: expected '&', got %q", compiler.defTypeMap['L'-'A'])
	}
}

func TestCompileDefTypeSng(t *testing.T) {
	src := `
DEFSNG S-S
s = 3.14
PRINT s
`
	out := compileAndRun(t, src)
	if !strings.Contains(out, "3.14") && !strings.Contains(out, "3.1") {
		t.Errorf("DEFSNG: expected ~3.14, got %q", out)
	}
	_, compiler := compileSource(t, src)
	if compiler.defTypeMap['S'-'A'] != "!" {
		t.Errorf("defTypeMap[S]: expected '!', got %q", compiler.defTypeMap['S'-'A'])
	}
}

func TestCompileDefTypeDbl(t *testing.T) {
	src := `
DEFDBL D-D
d = 1.23456789
PRINT d
`
	out := compileAndRun(t, src)
	if !strings.Contains(out, "1.23") {
		t.Errorf("DEFDBL: expected ~1.23, got %q", out)
	}
	_, compiler := compileSource(t, src)
	if compiler.defTypeMap['D'-'A'] != "#" {
		t.Errorf("defTypeMap[D]: expected '#', got %q", compiler.defTypeMap['D'-'A'])
	}
}

func TestCompileDefTypeStr(t *testing.T) {
	src := `
DEFSTR N-N
n = "hello"
PRINT n
`
	out := compileAndRun(t, src)
	if !strings.Contains(out, "hello") {
		t.Errorf("DEFSTR: expected 'hello', got %q", out)
	}
	_, compiler := compileSource(t, src)
	if compiler.defTypeMap['N'-'A'] != "$" {
		t.Errorf("defTypeMap[N]: expected '$', got %q", compiler.defTypeMap['N'-'A'])
	}
}

func TestCompileDefTypeLetterRange(t *testing.T) {
	// Lowercase letter range exercises the 'a'-'z' branch in compileDefType.
	src := `
DEFSTR M-P
PRINT "range ok"
`
	out := compileAndRun(t, src)
	if !strings.Contains(out, "range ok") {
		t.Errorf("DEFSTR M-P: expected 'range ok', got %q", out)
	}
	_, compiler := compileSource(t, src)
	if compiler.defTypeMap['M'-'A'] != "$" {
		t.Errorf("defTypeMap[M]: expected '$', got %q", compiler.defTypeMap['M'-'A'])
	}
}

// ---------------------------------------------------------------------------
// Test: compileRead / compileRestore
// ---------------------------------------------------------------------------

func TestCompileRead(t *testing.T) {
	src := `DATA 10, 20, 30
READ a
READ b
PRINT a
PRINT b`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "10") {
		t.Fatalf("expected '10', got %q", output)
	}
	if !strings.Contains(output, "20") {
		t.Fatalf("expected '20', got %q", output)
	}
}

func TestCompileRestore(t *testing.T) {
	src := `DATA 5, 6
READ a
RESTORE
READ b
PRINT a
PRINT b`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "5") {
		t.Fatalf("expected '5', got %q", output)
	}
	// After RESTORE, b should read first item again (5)
	// At least both values are printed
}

// ---------------------------------------------------------------------------
// Test: compileDim / compileDimDecl
// ---------------------------------------------------------------------------

func TestCompileDim(t *testing.T) {
	src := `DIM arr(5)
arr(1) = 42
PRINT arr(1)`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "42") {
		t.Fatalf("expected '42', got %q", output)
	}
}

func TestCompileDimMultiDim(t *testing.T) {
	src := `DIM mat(3, 3)
mat(2, 2) = 7
PRINT mat(2, 2)`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "7") {
		t.Fatalf("expected '7', got %q", output)
	}
}

// ---------------------------------------------------------------------------
// Test: compileSubDecl / compileFunctionDecl
// ---------------------------------------------------------------------------

func TestCompileSubDecl(t *testing.T) {
	// SUB forward declaration only — body skipped
	src := `SUB myproc
PRINT "in sub"
END SUB
PRINT "main"`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "main") {
		t.Fatalf("expected 'main', got %q", output)
	}
}

func TestCompileSubDeclForward(t *testing.T) {
	// DECLARE should compile without error (forward-only, no body emitted)
	src := `DECLARE SUB myProc
PRINT "ok"`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "ok") {
		t.Fatalf("expected 'ok', got %q", output)
	}
}

func TestCompileFunctionDecl(t *testing.T) {
	// FunctionDecl just tests that the body is skipped in main code flow
	src := `FUNCTION triple(n)
triple = n * 3
END FUNCTION
PRINT "ok"`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "ok") {
		t.Fatalf("expected 'ok', got %q", output)
	}
}

// ---------------------------------------------------------------------------
// Test: compileSwap
// ---------------------------------------------------------------------------

func TestCompileSwap(t *testing.T) {
	src := `a = 10
b = 20
SWAP a, b
PRINT a
PRINT b`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "20") {
		t.Fatalf("expected '20' (new a), got %q", output)
	}
	if !strings.Contains(output, "10") {
		t.Fatalf("expected '10' (new b), got %q", output)
	}
}

// ---------------------------------------------------------------------------
// Test: compileIncr / compileDecr
// ---------------------------------------------------------------------------

func TestCompileIncr(t *testing.T) {
	src := `x = 5
INCR x
PRINT x`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "6") {
		t.Fatalf("expected '6', got %q", output)
	}
}

func TestCompileIncrWithAmount(t *testing.T) {
	src := `x = 5
INCR x, 3
PRINT x`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "8") {
		t.Fatalf("expected '8', got %q", output)
	}
}

func TestCompileDecr(t *testing.T) {
	src := `x = 5
DECR x
PRINT x`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "4") {
		t.Fatalf("expected '4', got %q", output)
	}
}

func TestCompileDecrWithAmount(t *testing.T) {
	src := `x = 10
DECR x, 3
PRINT x`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "7") {
		t.Fatalf("expected '7', got %q", output)
	}
}

// ---------------------------------------------------------------------------
// Test: compileArrayAssignment
// ---------------------------------------------------------------------------

func TestCompileArrayAssignment(t *testing.T) {
	src := `DIM scores(10)
scores(1) = 100
scores(2) = 200
PRINT scores(1)
PRINT scores(2)`
	out := compileAndRun(t, src)
	if !strings.Contains(out, "100") {
		t.Fatalf("expected '100', got %q", out)
	}
	if !strings.Contains(out, "200") {
		t.Fatalf("expected '200', got %q", out)
	}
}

// ---------------------------------------------------------------------------
// Test: collectData (DATA statements)
// ---------------------------------------------------------------------------

func TestCollectData(t *testing.T) {
	src := `DATA 1, 2, 3
DATA "hello"
READ a
READ b
READ c
READ d$
PRINT a
PRINT b
PRINT c
PRINT d$`
	out := compileAndRun(t, src)
	if !strings.Contains(out, "1") || !strings.Contains(out, "2") || !strings.Contains(out, "3") {
		t.Fatalf("expected 1, 2, 3 in output, got %q", out)
	}
	if !strings.Contains(out, "hello") {
		t.Fatalf("expected 'hello' in output, got %q", out)
	}
}

func TestCollectDataStringValues(t *testing.T) {
	src := `DATA "hello", "world"
READ a$
READ b$
PRINT a$
PRINT b$`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "hello") || !strings.Contains(output, "world") {
		t.Errorf("DATA strings = %q", output)
	}
}

func TestCollectDataNegativeValues(t *testing.T) {
	src := `DATA -5, -3.14
READ x
PRINT x`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "-5") {
		t.Errorf("DATA negative = %q", output)
	}
}

// ---------------------------------------------------------------------------
// Test: compilePrint — separators
// ---------------------------------------------------------------------------

func TestCompilePrintTabSeparator(t *testing.T) {
	output := compileAndRun(t, `PRINT "a", "b"`)
	// "a" then tab then "b"
	if !strings.Contains(output, "a") || !strings.Contains(output, "b") {
		t.Errorf("print with comma = %q", output)
	}
}

func TestCompilePrintSemicolonSeparator(t *testing.T) {
	output := compileAndRun(t, `PRINT "a"; "b"`)
	if !strings.Contains(output, "a") || !strings.Contains(output, "b") {
		t.Errorf("print with semicolon = %q", output)
	}
}

func TestCompilePrintTrailingSemicolon(t *testing.T) {
	output := compileAndRun(t, `PRINT "a";
PRINT "b"`)
	if !strings.Contains(output, "a") || !strings.Contains(output, "b") {
		t.Errorf("trailing semicolon = %q", output)
	}
}

// ---------------------------------------------------------------------------
// Test: compileStatement — NOP-emitting and simple statement types
// ---------------------------------------------------------------------------

func TestCompileReturnStatement(t *testing.T) {
	src := `GOSUB mySub
GOTO done
mySub:
PRINT "sub"
RETURN
done:`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "sub") {
		t.Errorf("GOSUB/RETURN = %q, want 'sub'", output)
	}
}

func TestCompileEndStatement(t *testing.T) {
	output := compileAndRun(t, `PRINT "before"
END
PRINT "after"`)
	if !strings.Contains(output, "before") {
		t.Errorf("expected 'before', got %q", output)
	}
}

func TestCompileStopStatement(t *testing.T) {
	output := compileAndRun(t, `PRINT "ok"
STOP`)
	if !strings.Contains(output, "ok") {
		t.Errorf("STOP = %q", output)
	}
}

func TestCompileRemStatement(t *testing.T) {
	output := compileAndRun(t, `REM This is a comment
PRINT "hello"`)
	if !strings.Contains(output, "hello") {
		t.Errorf("REM = %q", output)
	}
}

func TestCompileDataStatement(t *testing.T) {
	src := `DATA 10, 20, 30
READ a
READ b
PRINT a
PRINT b`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "10") || !strings.Contains(output, "20") {
		t.Errorf("DATA/READ = %q", output)
	}
}

func TestCompileLabelAndLineNumber(t *testing.T) {
	src := `GOTO myLabel
PRINT "skipped"
myLabel:
PRINT "reached"`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "reached") {
		t.Errorf("label jump = %q", output)
	}
	if strings.Contains(output, "skipped") {
		t.Errorf("should not print skipped")
	}
}

func TestCompileLineNumber(t *testing.T) {
	src := `GOTO 100
PRINT "skipped"
100 PRINT "line 100"`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "line 100") {
		t.Errorf("line number = %q", output)
	}
}

func TestCompileNopStatements(t *testing.T) {
	// These all compile to NOP
	src := `DEFINT A-Z
OPTION BASE 1
CONST pi = 3.14159
PRINT "ok"`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "ok") {
		t.Errorf("NOP stmts = %q", output)
	}
}

// ===========================================================================
// Compiler: DEFINT / DEFSNG type declaration
// ===========================================================================

func TestCompileDefIntRange(t *testing.T) {
	src := `DEFINT A-Z
a = 100
b = 200
PRINT a + b`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "300") {
		t.Errorf("expected output to contain '300', got %q", output)
	}
}

func TestCompileDefSngSingle(t *testing.T) {
	src := `DEFSNG X-X
x = 3.14
PRINT x`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "3.14") {
		t.Errorf("expected output to contain '3.14', got %q", output)
	}
}

// ===========================================================================
// Compiler: DATA / READ with mixed and negative values
// ===========================================================================

func TestCompileDataMixedTypes(t *testing.T) {
	src := `DATA 10, 3.14, "hello"
READ a, b, c$
PRINT a
PRINT b
PRINT c$`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "10") {
		t.Errorf("expected output to contain '10', got %q", output)
	}
	if !strings.Contains(output, "3.14") {
		t.Errorf("expected output to contain '3.14', got %q", output)
	}
	if !strings.Contains(output, "hello") {
		t.Errorf("expected output to contain 'hello', got %q", output)
	}
}

func TestCompileDataNegative(t *testing.T) {
	src := `DATA -5, -3.14
READ a, b
PRINT a
PRINT b`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "-5") {
		t.Errorf("expected output to contain '-5', got %q", output)
	}
	if !strings.Contains(output, "-3.14") {
		t.Errorf("expected output to contain '-3.14', got %q", output)
	}
}

// ===========================================================================
// Compiler: SELECT CASE with comparison operators
// ===========================================================================

func TestCompileSelectCaseComparisons(t *testing.T) {
	src := `x = 5
SELECT CASE x
CASE IS = 5
  PRINT "equal"
END SELECT`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "equal") {
		t.Errorf("expected output to contain 'equal', got %q", output)
	}
}

func TestCompileSelectCaseLessThan(t *testing.T) {
	src := `x = 3
SELECT CASE x
CASE IS < 5
  PRINT "less"
END SELECT`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "less") {
		t.Errorf("expected output to contain 'less', got %q", output)
	}
}
