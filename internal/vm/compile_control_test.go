package vm

import (
	"strings"
	"testing"
)

// ===========================================================================
// Compiler: compileIf — IF/ELSEIF/ELSE blocks
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

func TestCompileExitWhileLoop(t *testing.T) {
	// EXIT WHILE exercises the "WHILE" case in compileExit.
	src := `
x = 0
WHILE x < 10
  x = x + 1
  IF x = 5 THEN EXIT WHILE
WEND
PRINT x
`
	out := compileAndRun(t, src)
	if !strings.Contains(out, "5") {
		t.Errorf("EXIT WHILE: expected '5', got %q", out)
	}
}

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
// Compiler: compileGoto — forward label
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
// ON GOTO / ON GOSUB
// ===========================================================================

func TestOnComputedGoto(t *testing.T) {
	tests := []struct {
		n       int
		want    string
		notWant string
	}{
		{1, "one", ""},
		{2, "two", ""},
		{3, "three", ""},
		{0, "fallthrough", ""},  // out of range — falls through to GOTO done
		{4, "fallthrough", ""},  // out of range — falls through to GOTO done
	}
	for _, tt := range tests {
		src := formatOnGotoSrc(tt.n)
		out := compileAndRun(t, src)
		if !strings.Contains(out, tt.want) {
			t.Errorf("ON %d GOTO: expected %q, got %q", tt.n, tt.want, out)
		}
	}
}

func formatOnGotoSrc(n int) string {
	switch n {
	case 1:
		return `
n = 1
ON n GOTO lbl1, lbl2, lbl3
PRINT "fallthrough"
GOTO done
lbl1:
PRINT "one"
GOTO done
lbl2:
PRINT "two"
GOTO done
lbl3:
PRINT "three"
done:
`
	case 2:
		return `
n = 2
ON n GOTO lbl1, lbl2, lbl3
PRINT "fallthrough"
GOTO done
lbl1:
PRINT "one"
GOTO done
lbl2:
PRINT "two"
GOTO done
lbl3:
PRINT "three"
done:
`
	case 3:
		return `
n = 3
ON n GOTO lbl1, lbl2, lbl3
PRINT "fallthrough"
GOTO done
lbl1:
PRINT "one"
GOTO done
lbl2:
PRINT "two"
GOTO done
lbl3:
PRINT "three"
done:
`
	case 0:
		return `
n = 0
ON n GOTO lbl1, lbl2, lbl3
PRINT "fallthrough"
GOTO done
lbl1:
PRINT "one"
GOTO done
lbl2:
PRINT "two"
GOTO done
lbl3:
PRINT "three"
done:
`
	default:
		return `
n = 4
ON n GOTO lbl1, lbl2, lbl3
PRINT "fallthrough"
GOTO done
lbl1:
PRINT "one"
GOTO done
lbl2:
PRINT "two"
GOTO done
lbl3:
PRINT "three"
done:
`
	}
}

func TestOnComputedGosub(t *testing.T) {
	src := `
n = 2
ON n GOSUB sub1, sub2, sub3
PRINT "back"
END
sub1:
PRINT "s1"
RETURN
sub2:
PRINT "s2"
RETURN
sub3:
PRINT "s3"
RETURN
`
	out := compileAndRun(t, src)
	if !strings.Contains(out, "s2") {
		t.Errorf("ON 2 GOSUB: expected 's2', got %q", out)
	}
	if !strings.Contains(out, "back") {
		t.Errorf("ON GOSUB: expected 'back' after return, got %q", out)
	}
	if strings.Contains(out, "s1") || strings.Contains(out, "s3") {
		t.Errorf("ON GOSUB: unexpected sub called, got %q", out)
	}
}

// ===========================================================================
// Compiler: SWAP statement
// ===========================================================================

func TestCompileSwapScalars(t *testing.T) {
	src := `a = 10
b = 20
SWAP a, b
PRINT a
PRINT b`
	output := compileAndRun(t, src)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected 2 lines of output, got %q", output)
	}
	if !strings.Contains(lines[0], "20") {
		t.Errorf("expected first line to contain '20', got %q", lines[0])
	}
	if !strings.Contains(lines[1], "10") {
		t.Errorf("expected second line to contain '10', got %q", lines[1])
	}
}

func TestCompileSwapArrayElements(t *testing.T) {
	src := `DIM arr(5)
arr(1) = 100
arr(2) = 200
SWAP arr(1), arr(2)
PRINT arr(1)
PRINT arr(2)`
	output := compileAndRun(t, src)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected 2 lines of output, got %q", output)
	}
	if !strings.Contains(lines[0], "200") {
		t.Errorf("expected first line to contain '200', got %q", lines[0])
	}
	if !strings.Contains(lines[1], "100") {
		t.Errorf("expected second line to contain '100', got %q", lines[1])
	}
}

func TestCompileSwapMixed(t *testing.T) {
	src := `DIM arr(5)
a = 10
arr(1) = 20
SWAP a, arr(1)
PRINT a
PRINT arr(1)`
	output := compileAndRun(t, src)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected 2 lines of output, got %q", output)
	}
	if !strings.Contains(lines[0], "20") {
		t.Errorf("expected first line to contain '20', got %q", lines[0])
	}
	if !strings.Contains(lines[1], "10") {
		t.Errorf("expected second line to contain '10', got %q", lines[1])
	}
}
