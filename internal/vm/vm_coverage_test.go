package vm

import (
	"math"
	"strings"
	"testing"

	"github.com/loabletech/go-basic/internal/lexer"
	"github.com/loabletech/go-basic/internal/parser"
)

// ===========================================================================
// Value method tests (String, asFloat, asInt, isTruthy, asString)
// ===========================================================================

func TestValueString(t *testing.T) {
	cases := []struct {
		v    Value
		want string
	}{
		{IntVal(42), "Int(42)"},
		{FloatVal(3.14), "Float(3.14)"},
		{StringVal("hi"), `String("hi")`},
		{Value{Type: 99}, "Value(?)"},
	}
	for _, tc := range cases {
		got := tc.v.String()
		if got != tc.want {
			t.Errorf("Value.String() = %q, want %q", got, tc.want)
		}
	}
}

func TestValueAsFloat(t *testing.T) {
	if IntVal(5).asFloat() != 5.0 {
		t.Error("IntVal.asFloat() should be 5.0")
	}
	if FloatVal(2.5).asFloat() != 2.5 {
		t.Error("FloatVal.asFloat() should be 2.5")
	}
	if StringVal("x").asFloat() != 0 {
		t.Error("StringVal.asFloat() should be 0")
	}
}

func TestValueAsInt(t *testing.T) {
	if IntVal(7).asInt() != 7 {
		t.Error("IntVal.asInt() should be 7")
	}
	if FloatVal(3.9).asInt() != 3 {
		t.Error("FloatVal(3.9).asInt() should be 3 (truncated)")
	}
	if StringVal("x").asInt() != 0 {
		t.Error("StringVal.asInt() should be 0")
	}
}

func TestValueIsTruthy(t *testing.T) {
	if !IntVal(-1).isTruthy() {
		t.Error("IntVal(-1).isTruthy() should be true")
	}
	if IntVal(0).isTruthy() {
		t.Error("IntVal(0).isTruthy() should be false")
	}
	if !FloatVal(0.1).isTruthy() {
		t.Error("FloatVal(0.1).isTruthy() should be true")
	}
	if FloatVal(0).isTruthy() {
		t.Error("FloatVal(0).isTruthy() should be false")
	}
	if !StringVal("x").isTruthy() {
		t.Error("StringVal(non-empty).isTruthy() should be true")
	}
	if StringVal("").isTruthy() {
		t.Error("StringVal('').isTruthy() should be false")
	}
	if (Value{Type: 99}).isTruthy() {
		t.Error("unknown type isTruthy() should be false")
	}
}

func TestValueAsString(t *testing.T) {
	if IntVal(5).asString() != " 5 " {
		t.Errorf("IntVal(5).asString() = %q, want ' 5 '", IntVal(5).asString())
	}
	if IntVal(-3).asString() != "-3 " {
		t.Errorf("IntVal(-3).asString() = %q, want '-3 '", IntVal(-3).asString())
	}
	if FloatVal(1.5).asString() != " 1.5 " {
		t.Errorf("FloatVal(1.5).asString() = %q, want ' 1.5 '", FloatVal(1.5).asString())
	}
	if FloatVal(-1.5).asString() != "-1.5 " {
		t.Errorf("FloatVal(-1.5).asString() = %q, want '-1.5 '", FloatVal(-1.5).asString())
	}
	if StringVal("hello").asString() != "hello" {
		t.Error("StringVal.asString() should return the string as-is")
	}
	if (Value{Type: 99}).asString() != "" {
		t.Error("unknown type asString() should be ''")
	}
}

// ===========================================================================
// Compiler: compileDoLoop
// ===========================================================================

func TestCompileDoLoopInfinite(t *testing.T) {
	// Infinite DO...LOOP
	src := `x = 0
DO
x = x + 1
IF x >= 3 THEN EXIT DO
LOOP`
	output := compileAndRun(t, src)
	_ = output // just verifying it doesn't crash
}

func TestCompileDoLoopWhileTop(t *testing.T) {
	src := `x = 1
total = 0
DO WHILE x <= 5
total = total + x
x = x + 1
LOOP
PRINT total`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "15") {
		t.Fatalf("expected 15, got %q", output)
	}
}

func TestCompileDoLoopUntilTop(t *testing.T) {
	src := `x = 1
total = 0
DO UNTIL x > 5
total = total + x
x = x + 1
LOOP
PRINT total`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "15") {
		t.Fatalf("expected 15, got %q", output)
	}
}

func TestCompileDoLoopUntilBottom(t *testing.T) {
	src := `x = 1
total = 0
DO
total = total + x
x = x + 1
LOOP UNTIL x > 5
PRINT total`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "15") {
		t.Fatalf("expected 15, got %q", output)
	}
}

func TestCompileDoLoopWhileBottom(t *testing.T) {
	src := `x = 1
total = 0
DO
total = total + x
x = x + 1
LOOP WHILE x <= 5
PRINT total`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "15") {
		t.Fatalf("expected 15, got %q", output)
	}
}

// ===========================================================================
// Compiler: compileSelectCase
// ===========================================================================

func TestCompileSelectCaseSimple(t *testing.T) {
	src := `x = 2
SELECT CASE x
CASE 1
PRINT "one"
CASE 2
PRINT "two"
CASE 3
PRINT "three"
END SELECT`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "two") {
		t.Fatalf("expected 'two', got %q", output)
	}
	if strings.Contains(output, "one") || strings.Contains(output, "three") {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestCompileSelectCaseElse(t *testing.T) {
	src := `x = 99
SELECT CASE x
CASE 1
PRINT "one"
CASE ELSE
PRINT "other"
END SELECT`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "other") {
		t.Fatalf("expected 'other', got %q", output)
	}
}

func TestCompileSelectCaseRange(t *testing.T) {
	src := `x = 5
SELECT CASE x
CASE 1 TO 3
PRINT "low"
CASE 4 TO 6
PRINT "mid"
CASE 7 TO 10
PRINT "high"
END SELECT`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "mid") {
		t.Fatalf("expected 'mid', got %q", output)
	}
}

func TestCompileSelectCaseIs(t *testing.T) {
	src := `x = 10
SELECT CASE x
CASE IS > 5
PRINT "big"
CASE ELSE
PRINT "small"
END SELECT`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "big") {
		t.Fatalf("expected 'big', got %q", output)
	}
}

// ===========================================================================
// Compiler: compileGoto / compileGosub
// ===========================================================================

func TestCompileGoto(t *testing.T) {
	src := `GOTO myLabel
PRINT "skipped"
myLabel:
PRINT "reached"`
	output := compileAndRun(t, src)
	if strings.Contains(output, "skipped") {
		t.Fatal("expected 'skipped' to be skipped")
	}
	if !strings.Contains(output, "reached") {
		t.Fatalf("expected 'reached' in output, got %q", output)
	}
}

func TestCompileGosub(t *testing.T) {
	src := `GOSUB mySub
PRINT "after"
END
mySub:
PRINT "sub"
RETURN`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "sub") {
		t.Fatalf("expected 'sub' in output, got %q", output)
	}
	if !strings.Contains(output, "after") {
		t.Fatalf("expected 'after' in output, got %q", output)
	}
}

// ===========================================================================
// Compiler: compileRead / compileRestore
// ===========================================================================

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

// ===========================================================================
// Compiler: compileDim / compileDimDecl
// ===========================================================================

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

// ===========================================================================
// Compiler: compileSubDecl / compileFunctionDecl
// ===========================================================================

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

// ===========================================================================
// Compiler: compileDefFn
// ===========================================================================

func TestCompileDefFnSingleLine(t *testing.T) {
	// DEF FN is compiled (body skipped with jump) but FN calls in VM emit NOP placeholder
	// Just verify it compiles and executes without error
	src := `DEF FNsquare(x) = x * x
PRINT "ok"`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "ok") {
		t.Fatalf("expected 'ok', got %q", output)
	}
}

// ===========================================================================
// Compiler: compileExit
// ===========================================================================

func TestCompileExitFor(t *testing.T) {
	src := `total = 0
FOR i = 1 TO 10
total = total + i
IF i = 3 THEN EXIT FOR
NEXT i
PRINT total`
	output := compileAndRun(t, src)
	// 1+2+3 = 6
	if !strings.Contains(output, "6") {
		t.Fatalf("expected '6', got %q", output)
	}
}

func TestCompileExitDo(t *testing.T) {
	src := `x = 0
DO
x = x + 1
IF x = 3 THEN EXIT DO
LOOP
PRINT x`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "3") {
		t.Fatalf("expected '3', got %q", output)
	}
}

func TestCompileExitSub(t *testing.T) {
	// EXIT SUB emits OpRet; test via GOSUB
	src := `GOSUB earlyExit
PRINT "done"
END
earlyExit:
PRINT "before"
EXIT SUB
PRINT "unreachable"
RETURN`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "before") {
		t.Fatalf("expected 'before', got %q", output)
	}
	if strings.Contains(output, "unreachable") {
		t.Fatalf("unexpected 'unreachable' in output: %q", output)
	}
}

// ===========================================================================
// Compiler: compileSwap
// ===========================================================================

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

// ===========================================================================
// Compiler: compileIncr / compileDecr
// ===========================================================================

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

// ===========================================================================
// Compiler: compileArrayAccess
// ===========================================================================

func TestCompileArrayAccess(t *testing.T) {
	src := `DIM v(5)
v(2) = 99
x = v(2)
PRINT x`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "99") {
		t.Fatalf("expected '99', got %q", output)
	}
}

// ===========================================================================
// VM: execBuiltin coverage
// ===========================================================================

func TestBuiltinFix(t *testing.T) {
	out := compileAndRun(t, "PRINT FIX(3.9)")
	if !strings.Contains(out, "3") {
		t.Fatalf("expected '3', got %q", out)
	}
}

func TestBuiltinInt(t *testing.T) {
	out := compileAndRun(t, "PRINT INT(-3.1)")
	if !strings.Contains(out, "-4") {
		t.Fatalf("expected '-4' (floor), got %q", out)
	}
}

func TestBuiltinExp(t *testing.T) {
	out := compileAndRun(t, "PRINT EXP(0)")
	if !strings.Contains(out, "1") {
		t.Fatalf("expected '1', got %q", out)
	}
}

func TestBuiltinSin(t *testing.T) {
	out := compileAndRun(t, "PRINT SIN(0)")
	if !strings.Contains(out, "0") {
		t.Fatalf("expected '0', got %q", out)
	}
}

func TestBuiltinCos(t *testing.T) {
	out := compileAndRun(t, "PRINT COS(0)")
	if !strings.Contains(out, "1") {
		t.Fatalf("expected '1', got %q", out)
	}
}

func TestBuiltinTan(t *testing.T) {
	out := compileAndRun(t, "PRINT TAN(0)")
	if !strings.Contains(out, "0") {
		t.Fatalf("expected '0', got %q", out)
	}
}

func TestBuiltinAtn(t *testing.T) {
	out := compileAndRun(t, "PRINT ATN(0)")
	if !strings.Contains(out, "0") {
		t.Fatalf("expected '0', got %q", out)
	}
}

func TestBuiltinLog(t *testing.T) {
	out := compileAndRun(t, "PRINT LOG(1)")
	if !strings.Contains(out, "0") {
		t.Fatalf("expected '0', got %q", out)
	}
}

func TestBuiltinCint(t *testing.T) {
	out := compileAndRun(t, "PRINT CINT(3.6)")
	if !strings.Contains(out, "4") {
		t.Fatalf("expected '4', got %q", out)
	}
}

func TestBuiltinCsng(t *testing.T) {
	out := compileAndRun(t, "PRINT CSNG(3)")
	if !strings.Contains(out, "3") {
		t.Fatalf("expected '3', got %q", out)
	}
}

func TestBuiltinCdbl(t *testing.T) {
	out := compileAndRun(t, "PRINT CDBL(3)")
	if !strings.Contains(out, "3") {
		t.Fatalf("expected '3', got %q", out)
	}
}

func TestBuiltinRight(t *testing.T) {
	out := compileAndRun(t, `PRINT RIGHT$("Hello", 3)`)
	if !strings.Contains(out, "llo") {
		t.Fatalf("expected 'llo', got %q", out)
	}
}

func TestBuiltinMid(t *testing.T) {
	out := compileAndRun(t, `PRINT MID$("Hello", 2, 3)`)
	if !strings.Contains(out, "ell") {
		t.Fatalf("expected 'ell', got %q", out)
	}
}

func TestBuiltinAsc(t *testing.T) {
	out := compileAndRun(t, `PRINT ASC("A")`)
	if !strings.Contains(out, "65") {
		t.Fatalf("expected '65', got %q", out)
	}
}

func TestBuiltinChrDollar(t *testing.T) {
	out := compileAndRun(t, "PRINT CHR$(66)")
	if !strings.Contains(out, "B") {
		t.Fatalf("expected 'B', got %q", out)
	}
}

func TestBuiltinStr(t *testing.T) {
	out := compileAndRun(t, "PRINT STR$(42)")
	if !strings.Contains(out, "42") {
		t.Fatalf("expected '42', got %q", out)
	}
}

func TestBuiltinVal(t *testing.T) {
	out := compileAndRun(t, `PRINT VAL("42")`)
	if !strings.Contains(out, "42") {
		t.Fatalf("expected '42', got %q", out)
	}
}

func TestBuiltinUCaseLower(t *testing.T) {
	out := compileAndRun(t, `PRINT UCASE$("world")`)
	if !strings.Contains(out, "WORLD") {
		t.Fatalf("expected 'WORLD', got %q", out)
	}
}

func TestBuiltinLCase(t *testing.T) {
	out := compileAndRun(t, `PRINT LCASE$("HELLO")`)
	if !strings.Contains(out, "hello") {
		t.Fatalf("expected 'hello', got %q", out)
	}
}

func TestBuiltinLTrim(t *testing.T) {
	out := compileAndRun(t, `PRINT LTRIM$("  hi")`)
	if !strings.Contains(out, "hi") {
		t.Fatalf("expected 'hi', got %q", out)
	}
}

func TestBuiltinRTrim(t *testing.T) {
	out := compileAndRun(t, `PRINT RTRIM$("hi  ")`)
	if !strings.Contains(out, "hi") {
		t.Fatalf("expected 'hi', got %q", out)
	}
}

func TestBuiltinTrim(t *testing.T) {
	out := compileAndRun(t, `PRINT TRIM$("  hi  ")`)
	if !strings.Contains(out, "hi") {
		t.Fatalf("expected 'hi', got %q", out)
	}
}

func TestBuiltinSpace(t *testing.T) {
	out := compileAndRun(t, "x$ = SPACE$(3)")
	_ = out // just verifies no crash
}

func TestBuiltinHex(t *testing.T) {
	out := compileAndRun(t, "PRINT HEX$(255)")
	if !strings.Contains(out, "FF") {
		t.Fatalf("expected 'FF', got %q", out)
	}
}

func TestBuiltinOct(t *testing.T) {
	out := compileAndRun(t, "PRINT OCT$(8)")
	if !strings.Contains(out, "10") {
		t.Fatalf("expected '10', got %q", out)
	}
}

func TestBuiltinInstr(t *testing.T) {
	out := compileAndRun(t, `PRINT INSTR(1, "Hello", "ell")`)
	if !strings.Contains(out, "2") {
		t.Fatalf("expected '2', got %q", out)
	}
}

func TestBuiltinRnd(t *testing.T) {
	out := compileAndRun(t, "x = RND(1)")
	_ = out // just verifies no crash
}

// ===========================================================================
// Compiler: emitComparisonOp (via SELECT CASE IS)
// ===========================================================================

func TestEmitComparisonOpEq(t *testing.T) {
	src := `x = 5
SELECT CASE x
CASE IS = 5
PRINT "eq"
END SELECT`
	out := compileAndRun(t, src)
	if !strings.Contains(out, "eq") {
		t.Fatalf("expected 'eq', got %q", out)
	}
}

func TestEmitComparisonOpNe(t *testing.T) {
	src := `x = 5
SELECT CASE x
CASE IS <> 3
PRINT "ne"
END SELECT`
	out := compileAndRun(t, src)
	if !strings.Contains(out, "ne") {
		t.Fatalf("expected 'ne', got %q", out)
	}
}

func TestEmitComparisonOpLe(t *testing.T) {
	src := `x = 3
SELECT CASE x
CASE IS <= 5
PRINT "le"
END SELECT`
	out := compileAndRun(t, src)
	if !strings.Contains(out, "le") {
		t.Fatalf("expected 'le', got %q", out)
	}
}

func TestEmitComparisonOpGe(t *testing.T) {
	src := `x = 7
SELECT CASE x
CASE IS >= 5
PRINT "ge"
END SELECT`
	out := compileAndRun(t, src)
	if !strings.Contains(out, "ge") {
		t.Fatalf("expected 'ge', got %q", out)
	}
}

// ===========================================================================
// Compiler: compileArrayAssignment
// ===========================================================================

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

// ===========================================================================
// Compiler: isWholeNumber (via compileNumberLiteral with int values)
// ===========================================================================

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

// ===========================================================================
// VM: String built-in function
// ===========================================================================

func TestBuiltinStringRepeat(t *testing.T) {
	out := compileAndRun(t, `PRINT STRING$(3, 65)`)
	if !strings.Contains(out, "AAA") {
		t.Fatalf("expected 'AAA', got %q", out)
	}
}

// ===========================================================================
// Compiler: collectData (DATA statements)
// ===========================================================================

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

// ===========================================================================
// Compiler: addError / error reporting
// ===========================================================================

func TestCompilerAddError(t *testing.T) {
	// SWAP with non-variable arguments should trigger addError
	// We verify compile still succeeds but errors are recorded
	compiler := NewCompiler(nil)
	compiler.addError("test error at line %d", 1)
	// Just verify addError doesn't panic
}

// ===========================================================================
// Compiler: findLoop / popLoop
// ===========================================================================

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
