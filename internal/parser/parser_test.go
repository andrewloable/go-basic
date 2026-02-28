package parser

import (
	"testing"

	"github.com/loabletech/go-basic/internal/lexer"
)

// ===========================================================================
// Core parser: New(), ParseProgram(), peekTokenIs, expectPeek, peekError
// ===========================================================================

func TestParseEmpty(t *testing.T) {
	prog, errs := parse("")
	expectNoErrors(t, errs)
	if len(prog.Statements) != 0 {
		t.Errorf("expected 0 statements, got %d", len(prog.Statements))
	}
}

func TestParseErrors(t *testing.T) {
	// This should produce parse errors but not panic
	_, errs := parse("IF THEN")
	if len(errs) == 0 {
		t.Error("expected parse errors for 'IF THEN'")
	}
}

func TestParseComplexProgram(t *testing.T) {
	input := `10 DIM a(10)
20 FOR i = 1 TO 10
30   a(i) = i * i
40 NEXT i
50 PRINT "Squares:"
60 FOR i = 1 TO 10
70   PRINT a(i)
80 NEXT i
90 END`
	prog, errs := parse(input)
	expectNoErrors(t, errs)
	if len(prog.Statements) < 5 {
		t.Errorf("expected at least 5 statements, got %d", len(prog.Statements))
	}
}

func TestParseMultiStatementLine(t *testing.T) {
	prog, errs := parse("x = 1 : y = 2 : PRINT x + y")
	expectNoErrors(t, errs)
	if len(prog.Statements) != 3 {
		t.Fatalf("expected 3 statements, got %d", len(prog.Statements))
	}
}

// ===========================================================================
// peekTokenIs / expectPeek / peekError — direct unit tests
// ===========================================================================

func TestPeekTokenIs_Match(t *testing.T) {
	// "x% = 42" — curToken is x%, peekToken is =
	p := newParserFromInput("x% = 42")
	// After New(), curToken = x%, peekToken = =
	if !p.peekTokenIs(lexer.TOKEN_EQ) {
		t.Error("expected peekTokenIs(TOKEN_EQ) to be true")
	}
}

func TestPeekTokenIs_NoMatch(t *testing.T) {
	p := newParserFromInput("x% = 42")
	if p.peekTokenIs(lexer.TOKEN_PLUS) {
		t.Error("expected peekTokenIs(TOKEN_PLUS) to be false")
	}
}

func TestExpectPeek_Success(t *testing.T) {
	// curToken = x%, peekToken = =
	p := newParserFromInput("x% = 42")
	if !p.expectPeek(lexer.TOKEN_EQ) {
		t.Error("expected expectPeek to return true when peek matches")
	}
	// After success, curToken should have advanced to =
	if !p.curTokenIs(lexer.TOKEN_EQ) {
		t.Errorf("expected curToken to be EQ after expectPeek, got %s", lexer.TokenName(p.curToken.Type))
	}
}

func TestExpectPeek_Failure(t *testing.T) {
	// curToken = x%, peekToken = = (not PLUS), so expectPeek(PLUS) fails
	p := newParserFromInput("x% = 42")
	result := p.expectPeek(lexer.TOKEN_PLUS)
	if result {
		t.Error("expected expectPeek to return false when peek does not match")
	}
	// Should have recorded an error via peekError
	if len(p.Errors()) == 0 {
		t.Error("expected peekError to record an error")
	}
}

func TestPeekError_RecordsError(t *testing.T) {
	// Directly call peekError to ensure it records a message
	p := newParserFromInput("x% = 42")
	initialErrors := len(p.Errors())
	p.peekError(lexer.TOKEN_PLUS)
	if len(p.Errors()) <= initialErrors {
		t.Error("expected peekError to add an error")
	}
}
