package parser

import (
	"testing"

	"github.com/loabletech/go-basic/internal/ast"
	"github.com/loabletech/go-basic/internal/lexer"
)

// parse creates a parser from the given input string and runs ParseProgram.
func parse(input string) (*ast.Program, []string) {
	l := lexer.New(input)
	p := New(l)
	prog := p.ParseProgram()
	return prog, p.Errors()
}

// expectNoErrors fails the test if any parse errors were produced.
func expectNoErrors(t *testing.T, errors []string) {
	t.Helper()
	if len(errors) > 0 {
		t.Fatalf("unexpected parse errors: %v", errors)
	}
}

// getStmt retrieves the statement at index idx from prog and asserts it is of type T.
func getStmt[T ast.Statement](t *testing.T, prog *ast.Program, idx int) T {
	t.Helper()
	if idx >= len(prog.Statements) {
		t.Fatalf("expected at least %d statements, got %d", idx+1, len(prog.Statements))
	}
	s, ok := prog.Statements[idx].(T)
	if !ok {
		t.Fatalf("statement[%d]: expected %T, got %T", idx, *new(T), prog.Statements[idx])
	}
	return s
}

// newParserFromInput creates a Parser for direct method testing.
func newParserFromInput(input string) *Parser {
	l := lexer.New(input)
	p := New(l)
	return p
}
