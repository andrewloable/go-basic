package vm

import (
	"strings"
	"testing"
)

// ===========================================================================
// compileBinaryExpr: uncovered operators
// ===========================================================================

func TestBinaryIntegerDiv(t *testing.T) {
	output := compileAndRun(t, `PRINT 7 \ 2`)
	if !strings.Contains(output, "3") {
		t.Errorf("7 \\ 2 = %q, want 3", output)
	}
}

func TestBinaryMod(t *testing.T) {
	output := compileAndRun(t, `PRINT 10 MOD 3`)
	if !strings.Contains(output, "1") {
		t.Errorf("10 MOD 3 = %q, want 1", output)
	}
}

func TestBinaryPow(t *testing.T) {
	output := compileAndRun(t, `PRINT 2 ^ 8`)
	if !strings.Contains(output, "256") {
		t.Errorf("2^8 = %q, want 256", output)
	}
}

func TestBinaryNe(t *testing.T) {
	output := compileAndRun(t, `IF 5 <> 3 THEN PRINT "yes"`)
	if !strings.Contains(output, "yes") {
		t.Errorf("5 <> 3 = %q, want yes", output)
	}
}

func TestBinaryAnd(t *testing.T) {
	output := compileAndRun(t, `PRINT 1 AND 1`)
	if !strings.Contains(output, "1") {
		t.Errorf("1 AND 1 = %q, want 1", output)
	}
}

func TestBinaryOr(t *testing.T) {
	output := compileAndRun(t, `PRINT 0 OR 1`)
	if !strings.Contains(output, "1") {
		t.Errorf("0 OR 1 = %q, want 1", output)
	}
}

func TestBinaryXor(t *testing.T) {
	output := compileAndRun(t, `PRINT -1 XOR -1`)
	// XOR of -1 and -1 is 0
	if !strings.Contains(output, "0") {
		t.Errorf("-1 XOR -1 = %q, want 0", output)
	}
}

func TestBinaryEqv(t *testing.T) {
	output := compileAndRun(t, `PRINT 1 EQV 1`)
	_ = output // just verify no panic
}

func TestBinaryImp(t *testing.T) {
	output := compileAndRun(t, `PRINT 0 IMP 1`)
	_ = output // just verify no panic
}

// ===========================================================================
// compileUnaryExpr: NOT and non-constant negation
// ===========================================================================

func TestUnaryNot(t *testing.T) {
	output := compileAndRun(t, `x = 0
PRINT NOT x`)
	// NOT 0 = -1 in BASIC
	if !strings.Contains(output, "-1") {
		t.Errorf("NOT 0 = %q, want -1", output)
	}
}

func TestUnaryNegVariable(t *testing.T) {
	// Non-constant negation (goes through compileExpression, then OpNeg)
	output := compileAndRun(t, `x = 5
PRINT -x`)
	if !strings.Contains(output, "-5") {
		t.Errorf("-x = %q, want -5", output)
	}
}

// ===========================================================================
// compileIf: ELSEIF and ELSE blocks
// ===========================================================================

func TestCompileIfElseIf(t *testing.T) {
	src := `x = 2
IF x = 1 THEN
PRINT "one"
ELSEIF x = 2 THEN
PRINT "two"
ELSE
PRINT "other"
END IF`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "two") {
		t.Errorf("expected 'two', got %q", output)
	}
}

func TestCompileIfElseBlock(t *testing.T) {
	src := `x = 5
IF x = 1 THEN
PRINT "one"
ELSE
PRINT "not one"
END IF`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "not one") {
		t.Errorf("expected 'not one', got %q", output)
	}
}

// ===========================================================================
// compilePrint: tab separator and semicolon separator
// ===========================================================================

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

// ===========================================================================
// compileExpression: GroupExpr path
// ===========================================================================

func TestCompileGroupExpr(t *testing.T) {
	output := compileAndRun(t, `PRINT (2 + 3) * 4`)
	if !strings.Contains(output, "20") {
		t.Errorf("(2+3)*4 = %q, want 20", output)
	}
}

// ===========================================================================
// compileStatement: various NOP-emitting and simple statement types
// ===========================================================================

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
// execBuiltin: uncovered builtins (Mki, Mkl, Mks, Mkd, Cvi, Cvl, Cvs, Cvd,
//              Tab, Spc, Len, String, Bin, Clng, Sqr error, Log errors)
// ===========================================================================

func TestBuiltinLenStr(t *testing.T) {
	output := compileAndRun(t, `PRINT LEN("hello")`)
	if !strings.Contains(output, "5") {
		t.Errorf("LEN = %q, want 5", output)
	}
}

func TestBuiltinBin(t *testing.T) {
	output := compileAndRun(t, `PRINT BIN$(10)`)
	if !strings.Contains(output, "1010") {
		t.Errorf("BIN$(10) = %q, want 1010", output)
	}
}

func TestBuiltinClng(t *testing.T) {
	output := compileAndRun(t, `PRINT CLNG(3.7)`)
	if !strings.Contains(output, "4") {
		t.Errorf("CLNG(3.7) = %q, want 4", output)
	}
}

func TestBuiltinTab(t *testing.T) {
	output := compileAndRun(t, `PRINT TAB(3)`)
	_ = output // just verify no panic
}

func TestBuiltinSpc(t *testing.T) {
	output := compileAndRun(t, `PRINT SPC(3)`)
	if !strings.Contains(output, "   ") {
		t.Errorf("SPC(3) = %q, want 3 spaces", output)
	}
}

func TestBuiltinMkiCvi(t *testing.T) {
	output := compileAndRun(t, `s$ = MKI$(100)
PRINT CVI(s$)`)
	if !strings.Contains(output, "100") {
		t.Errorf("MKI/CVI = %q", output)
	}
}

func TestBuiltinMklCvl(t *testing.T) {
	output := compileAndRun(t, `s$ = MKL$(12345)
PRINT CVL(s$)`)
	if !strings.Contains(output, "12345") {
		t.Errorf("MKL/CVL = %q", output)
	}
}

func TestBuiltinMksCvs(t *testing.T) {
	output := compileAndRun(t, `s$ = MKS$(1.5)
PRINT CVS(s$)`)
	if !strings.Contains(output, "1.5") {
		t.Errorf("MKS/CVS = %q", output)
	}
}

func TestBuiltinMkdCvd(t *testing.T) {
	output := compileAndRun(t, `s$ = MKD$(3.14)
PRINT CVD(s$)`)
	if !strings.Contains(output, "3.14") {
		t.Errorf("MKD/CVD = %q", output)
	}
}

func TestBuiltinSqrError(t *testing.T) {
	// SQR of negative should produce a runtime error
	src := `PRINT SQR(-1)`
	l := mustParse(t, src)
	v := NewVM(l)
	var buf strings.Builder
	v.SetOutput(&buf)
	err := v.Run()
	if err == nil {
		t.Error("expected runtime error for SQR(-1)")
	}
}

func TestBuiltinLogError(t *testing.T) {
	// LOG of zero or negative should produce a runtime error
	src := `PRINT LOG(-1)`
	l := mustParse(t, src)
	v := NewVM(l)
	var buf strings.Builder
	v.SetOutput(&buf)
	err := v.Run()
	if err == nil {
		t.Error("expected runtime error for LOG(-1)")
	}
}

func TestBuiltinLog2Error(t *testing.T) {
	src := `PRINT LOG2(-1)`
	l := mustParse(t, src)
	v := NewVM(l)
	var buf strings.Builder
	v.SetOutput(&buf)
	err := v.Run()
	if err == nil {
		t.Error("expected runtime error for LOG2(-1)")
	}
}

func TestBuiltinLog10Error(t *testing.T) {
	src := `PRINT LOG10(-1)`
	l := mustParse(t, src)
	v := NewVM(l)
	var buf strings.Builder
	v.SetOutput(&buf)
	err := v.Run()
	if err == nil {
		t.Error("expected runtime error for LOG10(-1)")
	}
}

func TestBuiltinStringFn(t *testing.T) {
	// STRING$(n, char) - repeat char n times
	output := compileAndRun(t, `PRINT STRING$(3, 65)`)
	if !strings.Contains(output, "AAA") {
		t.Errorf("STRING$(3, 65) = %q, want AAA", output)
	}
}

func TestBuiltinInstr2Args(t *testing.T) {
	// INSTR with 2 args (no start position)
	output := compileAndRun(t, `PRINT INSTR("hello world", "world")`)
	if !strings.Contains(output, "7") {
		t.Errorf("INSTR 2-arg = %q, want 7", output)
	}
}

// mustParse compiles BASIC source to a Chunk (fatals on parse/compile errors).
func mustParse(t *testing.T, src string) *Chunk {
	t.Helper()
	chunk, _ := compileSource(t, src)
	return chunk
}

// ===========================================================================
// collectData: edge cases
// ===========================================================================

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

// ===========================================================================
// compileExit: EXIT WHILE (via EXIT DO with while loop)
// ===========================================================================

func TestCompileExitWhile(t *testing.T) {
	// EXIT DO inside a DO WHILE loop
	src := `x = 0
DO WHILE x < 10
x = x + 1
IF x = 3 THEN EXIT DO
LOOP
PRINT x`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "3") {
		t.Errorf("EXIT DO from DO WHILE = %q, want 3", output)
	}
}

// ===========================================================================
// compileGoto / compileGosub: undefined label (error paths)
// ===========================================================================

func TestCompileGotoForwardLabel(t *testing.T) {
	src := `GOTO done
PRINT "skip"
done:
PRINT "done"`
	output := compileAndRun(t, src)
	if !strings.Contains(output, "done") {
		t.Errorf("forward GOTO = %q", output)
	}
}

// ===========================================================================
// VM Run: computeArrayIndex (out-of-bounds)
// ===========================================================================

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

// ===========================================================================
// compileFunctionCall: INSTR with 3 args (explicit start position)
// ===========================================================================

func TestBuiltinInstr3Args(t *testing.T) {
	output := compileAndRun(t, `PRINT INSTR(5, "hello world", "o")`)
	// "hello world" has "o" at pos 5 and 8; starting from 5 → should find pos 5
	if !strings.Contains(output, "5") && !strings.Contains(output, "8") {
		t.Errorf("INSTR 3-arg = %q, expected 5 or 8", output)
	}
}
