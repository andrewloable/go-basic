package vm

import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// Test: DEF FN single-line inline call
// ---------------------------------------------------------------------------

func TestDefFnSingleLine(t *testing.T) {
	src := `
DEF FNSquare(x) = x * x
PRINT FNSquare(5)
`
	out := compileAndRun(t, src)
	if !strings.Contains(out, "25") {
		t.Errorf("FNSquare(5): expected '25', got %q", out)
	}
}

func TestDefFnRestoresParam(t *testing.T) {
	// Verify that the caller's variable is restored after the DEF FN call.
	src := `
x = 10
DEF FNDouble(x) = x * 2
PRINT FNDouble(7)
PRINT x
`
	out := compileAndRun(t, src)
	if !strings.Contains(out, "14") {
		t.Errorf("FNDouble(7): expected '14', got %q", out)
	}
	if !strings.Contains(out, "10") {
		t.Errorf("x after call: expected '10', got %q", out)
	}
}

func TestDefFnMultiParam(t *testing.T) {
	src := `
DEF FNAdd(a, b) = a + b
PRINT FNAdd(3, 4)
`
	out := compileAndRun(t, src)
	if !strings.Contains(out, "7") {
		t.Errorf("FNAdd(3,4): expected '7', got %q", out)
	}
}

// ---------------------------------------------------------------------------
// Test: compileFnCallExpr — FN name(args) with space syntax
// ---------------------------------------------------------------------------

func TestFnCallExprSpacedSyntax(t *testing.T) {
	// FN Square(x) with explicit space between FN and name exercises
	// compileFnCallExpr (the FnCallExpression path), not the FunctionCall path.
	src := `
DEF FN Square(x) = x * x
PRINT FN Square(5)
`
	out := compileAndRun(t, src)
	if !strings.Contains(out, "25") {
		t.Errorf("FN Square(5): expected '25', got %q", out)
	}
}

func TestFnCallExprMultiArg(t *testing.T) {
	src := `
DEF FN Add(a, b) = a + b
PRINT FN Add(3, 4)
`
	out := compileAndRun(t, src)
	if !strings.Contains(out, "7") {
		t.Errorf("FN Add(3,4): expected '7', got %q", out)
	}
}

func TestFnCallExprUnknown(t *testing.T) {
	// Unknown DEF FN name — compiler pushes zero as placeholder, no crash.
	src := `
PRINT FN Unknown(5)
`
	out := compileAndRun(t, src)
	if !strings.Contains(out, "0") {
		t.Errorf("FN Unknown(5): expected '0' fallback, got %q", out)
	}
}

// ---------------------------------------------------------------------------
// Test: compileFieldAccessExpr — TYPE field read via spaced-dot syntax
// ---------------------------------------------------------------------------

func TestFieldAccessExprSpaced(t *testing.T) {
	// Using "p . x" (spaces around dot) forces separate tokens:
	//   IDENTIFIER("p")  TOKEN_ILLEGAL(".")  IDENTIFIER("x")
	// The parser then creates a FieldAccessExpression, exercising
	// compileFieldAccessExpr.
	src := `
TYPE Point
  x AS INTEGER
  y AS INTEGER
END TYPE
p . x = 10
p . y = 20
PRINT p . x
PRINT p . y
`
	out := compileAndRun(t, src)
	if !strings.Contains(out, "10") {
		t.Errorf("p.x read: expected '10', got %q", out)
	}
	if !strings.Contains(out, "20") {
		t.Errorf("p.y read: expected '20', got %q", out)
	}
}

func TestFieldAccessExprInExpression(t *testing.T) {
	// Field access in an arithmetic expression context.
	src := `
p . x = 3
p . y = 4
result = p . x + p . y
PRINT result
`
	out := compileAndRun(t, src)
	if !strings.Contains(out, "7") {
		t.Errorf("p.x + p.y: expected '7', got %q", out)
	}
}

func TestFieldAccessExprComplexObject(t *testing.T) {
	// arr(0) . x exercises the "!ok" path in compileFieldAccessExpr, which
	// emits a zero placeholder since complex objects are not supported.
	// Verify no crash.
	src := `
DIM arr(5)
PRINT arr(0) . x
`
	out := compileAndRun(t, src)
	// Complex object path pushes FloatVal(0) as placeholder.
	if !strings.Contains(out, "0") {
		t.Errorf("complex field access: expected '0' placeholder, got %q", out)
	}
}

func TestFieldAssignComplexObject(t *testing.T) {
	// arr(0) . x = 5 exercises the "!ok" path in compileFieldAssign.
	// The complex object case emits a NOP — verify no crash.
	src := `
DIM arr(3)
arr(0) . x = 99
PRINT "ok"
`
	out := compileAndRun(t, src)
	if !strings.Contains(out, "ok") {
		t.Errorf("complex field assign: expected 'ok', got %q", out)
	}
}

// ---------------------------------------------------------------------------
// Test: compileDefFn (compiler path)
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// Test: compileArrayAccess
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// Test: compileBinaryExpr — uncovered operators
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// Test: compileUnaryExpr — NOT and non-constant negation
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// Test: compileExpression — GroupExpr path
// ---------------------------------------------------------------------------

func TestCompileGroupExpr(t *testing.T) {
	output := compileAndRun(t, `PRINT (2 + 3) * 4`)
	if !strings.Contains(output, "20") {
		t.Errorf("(2+3)*4 = %q, want 20", output)
	}
}

// ---------------------------------------------------------------------------
// Test: FunctionCall via ArrayAccess node
// ---------------------------------------------------------------------------

func TestFunctionCallViaArrayAccessNode(t *testing.T) {
	// When a user-defined FUNCTION is called and the parser emits an ArrayAccess
	// node (because it doesn't recognise the name), the compiler should still
	// route to OpCall and return the correct value.
	src := `
DECLARE FUNCTION Twice (n)
PRINT Twice(7)
END
FUNCTION Twice (n)
  Twice = n * 2
END FUNCTION
`
	out := compileAndRun(t, src)
	if !strings.Contains(out, "14") {
		t.Errorf("Twice(7): expected '14', got %q", out)
	}
}
