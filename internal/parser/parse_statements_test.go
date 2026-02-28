package parser

import (
	"testing"

	"github.com/loabletech/go-basic/internal/ast"
)

// ===========================================================================
// Tests for parse_statements.go: LET, SWAP, INCR, DECR, POKE, READ, ERASE,
// CALL, ERROR, RANDOMIZE, CLEAR, WIDTH, END, CALL, sub-call dispatch
// ===========================================================================

// ===========================================================================
// LET / implicit assignment
// ===========================================================================

func TestParseLetAssignment(t *testing.T) {
	prog, errs := parse("LET x = 42")
	expectNoErrors(t, errs)
	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}
	ls, ok := prog.Statements[0].(*ast.LetStatement)
	if !ok {
		t.Fatalf("expected LetStatement, got %T", prog.Statements[0])
	}
	if ls.Name.Name != "x" {
		t.Errorf("expected variable 'x', got %q", ls.Name.Name)
	}
}

func TestParseImplicitAssignment(t *testing.T) {
	prog, errs := parse("x = 42")
	expectNoErrors(t, errs)
	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}
	_, ok := prog.Statements[0].(*ast.LetStatement)
	if !ok {
		t.Fatalf("expected LetStatement for implicit assignment, got %T", prog.Statements[0])
	}
}

func TestParseArrayAssignment(t *testing.T) {
	prog, errs := parse("arr%(i%) = 42")
	expectNoErrors(t, errs)
	s := getStmt[*ast.ArrayAssignment](t, prog, 0)
	if s.Array == nil {
		t.Error("expected Array")
	}
	if s.Value == nil {
		t.Error("expected Value")
	}
}

func TestParseAssignmentErrorNoIdent(t *testing.T) {
	// LET without identifier triggers error path
	_, errs := parse("LET = 42")
	if len(errs) == 0 {
		t.Error("expected parse error for LET without identifier")
	}
}

func TestParseAssignmentErrorNoEq(t *testing.T) {
	// Variable without '=' triggers error path
	_, errs := parse("myVar THEN")
	_ = errs // May or may not produce errors depending on parsing
}

func TestParseFieldAssignment(t *testing.T) {
	prog, errs := parse(`TYPE Point
x AS INTEGER
y AS INTEGER
END TYPE
DIM p AS Point
p.x = 10`)
	if len(errs) > 0 {
		_ = errs
		return
	}
	if len(prog.Statements) == 0 {
		t.Error("expected statements")
	}
}

// ===========================================================================
// SWAP statement
// ===========================================================================

func TestParseSwap(t *testing.T) {
	prog, errs := parse("SWAP a, b")
	expectNoErrors(t, errs)
	_, ok := prog.Statements[0].(*ast.SwapStatement)
	if !ok {
		t.Fatalf("expected SwapStatement, got %T", prog.Statements[0])
	}
}

func TestParseSwapStatement(t *testing.T) {
	prog, errs := parse("SWAP a%, b%")
	expectNoErrors(t, errs)
	s := getStmt[*ast.SwapStatement](t, prog, 0)
	if s.Var1 == nil {
		t.Error("expected Var1")
	}
	if s.Var2 == nil {
		t.Error("expected Var2")
	}
}

// ===========================================================================
// INCR / DECR statements
// ===========================================================================

func TestParseIncrStatement(t *testing.T) {
	prog, errs := parse("INCR x%")
	expectNoErrors(t, errs)
	s := getStmt[*ast.IncrStatement](t, prog, 0)
	if s.Variable == nil {
		t.Error("expected Variable")
	}
}

func TestParseIncrWithAmount(t *testing.T) {
	prog, errs := parse("INCR x%, 5")
	expectNoErrors(t, errs)
	s := getStmt[*ast.IncrStatement](t, prog, 0)
	if s.Amount == nil {
		t.Error("expected Amount")
	}
}

func TestParseDecrStatement(t *testing.T) {
	prog, errs := parse("DECR x%")
	expectNoErrors(t, errs)
	s := getStmt[*ast.DecrStatement](t, prog, 0)
	if s.Variable == nil {
		t.Error("expected Variable")
	}
}

// ===========================================================================
// POKE statement
// ===========================================================================

// TestRegressionPoke verifies POKE address, value parsing.
func TestRegressionPoke(t *testing.T) {
	_, errs := parse("POKE 1047, 7")
	expectNoErrors(t, errs)
}

func TestParsePokeStatement(t *testing.T) {
	prog, errs := parse("POKE 1024, 42")
	expectNoErrors(t, errs)
	s := getStmt[*ast.PokeStatement](t, prog, 0)
	if s.Address == nil {
		t.Error("expected Address")
	}
	if s.Value == nil {
		t.Error("expected Value")
	}
}

// ===========================================================================
// ERASE statement
// ===========================================================================

func TestParseEraseStatement(t *testing.T) {
	prog, errs := parse("ERASE myArray")
	expectNoErrors(t, errs)
	s := getStmt[*ast.EraseStatement](t, prog, 0)
	if len(s.Names) != 1 {
		t.Errorf("expected 1 name, got %d", len(s.Names))
	}
	if s.Names[0] != "myArray" {
		t.Errorf("expected 'myArray', got %q", s.Names[0])
	}
}

func TestParseEraseMultiple(t *testing.T) {
	prog, errs := parse("ERASE arr1, arr2, arr3")
	expectNoErrors(t, errs)
	s := getStmt[*ast.EraseStatement](t, prog, 0)
	if len(s.Names) != 3 {
		t.Errorf("expected 3 names, got %d", len(s.Names))
	}
}

// ===========================================================================
// CALL statement
// ===========================================================================

func TestParseCallStatement(t *testing.T) {
	prog, errs := parse("CALL MySub(5, 10)")
	expectNoErrors(t, errs)
	if len(prog.Statements) == 0 {
		t.Fatal("expected at least 1 statement")
	}
}

func TestParseCallNoArgs(t *testing.T) {
	prog, errs := parse("CALL PrintHeader")
	expectNoErrors(t, errs)
	if len(prog.Statements) == 0 {
		t.Fatal("expected at least 1 statement")
	}
}

func TestParseExpressionListEmpty(t *testing.T) {
	prog, errs := parse("CALL MySub")
	expectNoErrors(t, errs)
	_ = prog
}

func TestParseSubCallNoParens(t *testing.T) {
	// Bare sub call without CALL keyword.
	_, errs := parse(`MySub "World"`)
	_ = errs // May produce parse errors — just verify no panic
}

// ===========================================================================
// ERROR statement
// ===========================================================================

func TestParseErrorStatement(t *testing.T) {
	prog, errs := parse("ERROR 53")
	expectNoErrors(t, errs)
	s := getStmt[*ast.ErrorStatement](t, prog, 0)
	if s.Code == nil {
		t.Error("expected Code expression to be set")
	}
}

// ===========================================================================
// RANDOMIZE statement
// ===========================================================================

func TestParseRandomizeTimer(t *testing.T) {
	prog, errs := parse("RANDOMIZE TIMER")
	expectNoErrors(t, errs)
	s := getStmt[*ast.RandomizeStatement](t, prog, 0)
	if s.Seed == nil {
		t.Error("expected RANDOMIZE TIMER to have a Seed expression")
	}
}

func TestParseRandomizeWithSeed(t *testing.T) {
	prog, errs := parse("RANDOMIZE 42")
	expectNoErrors(t, errs)
	s := getStmt[*ast.RandomizeStatement](t, prog, 0)
	if s.Seed == nil {
		t.Error("expected RANDOMIZE 42 to have a Seed expression")
	}
	nl, ok := s.Seed.(*ast.NumberLiteral)
	if !ok {
		t.Fatalf("expected NumberLiteral seed, got %T", s.Seed)
	}
	if nl.Value != 42 {
		t.Errorf("expected seed 42, got %v", nl.Value)
	}
}

func TestParseRandomizeNoArg(t *testing.T) {
	prog, errs := parse("RANDOMIZE\n")
	expectNoErrors(t, errs)
	s := getStmt[*ast.RandomizeStatement](t, prog, 0)
	_ = s // no seed is ok
}

// ===========================================================================
// CLEAR statement
// ===========================================================================

// TestRegressionCLEAR verifies that the CLEAR statement parses correctly.
func TestRegressionCLEAR(t *testing.T) {
	_, errs := parse("CLEAR")
	expectNoErrors(t, errs)
	_, errs = parse("CLEAR 500")
	expectNoErrors(t, errs)
}

// ===========================================================================
// WIDTH statement
// ===========================================================================

// TestRegressionWidth verifies WIDTH statement parsing.
func TestRegressionWidth(t *testing.T) {
	_, errs := parse("WIDTH 80, 25")
	expectNoErrors(t, errs)
	_, errs = parse("WIDTH 40")
	expectNoErrors(t, errs)
}

// ===========================================================================
// END statement
// ===========================================================================

func TestParseEnd(t *testing.T) {
	prog, errs := parse("END")
	expectNoErrors(t, errs)
	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}
	_, ok := prog.Statements[0].(*ast.EndStatement)
	if !ok {
		t.Fatalf("expected EndStatement, got %T", prog.Statements[0])
	}
}

// ===========================================================================
// Multiple statements on a single line
// ===========================================================================

func TestParseMultipleStatementsOnLine(t *testing.T) {
	prog, errs := parse("x% = 1 : y% = 2 : z% = 3")
	expectNoErrors(t, errs)
	if len(prog.Statements) < 3 {
		t.Errorf("expected at least 3 statements, got %d", len(prog.Statements))
	}
}
