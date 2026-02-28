package codegen

import (
	"strings"
	"testing"

	"github.com/loabletech/go-basic/internal/ast"
	"github.com/loabletech/go-basic/internal/semantic"
)

func TestGotoOverDeclarationHoisting(t *testing.T) {
	// Simulate: 10 GOTO 30 / 20 LET X = 5 / 30 LET Y = 10 / 40 PRINT Y
	stmts := []ast.Statement{
		&ast.LineNumberStatement{Number: 10},
		&ast.GotoStatement{Target: "30"},
		&ast.LineNumberStatement{Number: 20},
		&ast.LetStatement{
			Name:  &ast.Identifier{Name: "X"},
			Value: &ast.NumberLiteral{Value: 5},
		},
		&ast.LineNumberStatement{Number: 30},
		&ast.LetStatement{
			Name:  &ast.Identifier{Name: "Y"},
			Value: &ast.NumberLiteral{Value: 10},
		},
		&ast.PrintStatement{
			Expressions: []ast.Expression{&ast.Identifier{Name: "Y"}},
		},
	}

	prog := &ast.Program{Statements: stmts}
	table := semantic.NewSymbolTable()
	gen := New()
	goSrc, err := gen.Generate(prog, table)
	if err != nil {
		t.Fatalf("Codegen error: %v", err)
	}

	// The output should contain hoisted variable declarations.
	if !strings.Contains(goSrc, "// Hoisted variable declarations") {
		t.Errorf("Expected hoisted variable declarations comment in output:\n%s", goSrc)
	}

	// Variables should NOT be declared inline with var X type = value.
	if strings.Contains(goSrc, "var x float32 =") {
		t.Error("Found inline var declaration for x; should be hoisted")
	}
	if strings.Contains(goSrc, "var y float32 =") {
		t.Error("Found inline var declaration for y; should be hoisted")
	}

	// Hoisted declarations should appear before any goto labels.
	hoistIdx := strings.Index(goSrc, "// Hoisted variable declarations")
	gotoIdx := strings.Index(goSrc, "goto ")
	if hoistIdx < 0 {
		t.Fatal("Missing hoisted declarations")
	}
	if gotoIdx >= 0 && hoistIdx > gotoIdx {
		t.Error("Hoisted declarations should appear before goto statements")
	}

	t.Logf("Generated Go source:\n%s", goSrc)
}

func TestNoHoistingWithoutGoto(t *testing.T) {
	// Program without GOTO should NOT have hoisted declarations.
	stmts := []ast.Statement{
		&ast.LetStatement{
			Name:  &ast.Identifier{Name: "X"},
			Value: &ast.NumberLiteral{Value: 5},
		},
		&ast.PrintStatement{
			Expressions: []ast.Expression{&ast.Identifier{Name: "X"}},
		},
	}

	prog := &ast.Program{Statements: stmts}
	table := semantic.NewSymbolTable()
	gen := New()
	goSrc, err := gen.Generate(prog, table)
	if err != nil {
		t.Fatalf("Codegen error: %v", err)
	}

	// Should NOT have hoisted variables.
	if strings.Contains(goSrc, "// Hoisted variable declarations") {
		t.Errorf("Should not have hoisted declarations without GOTO:\n%s", goSrc)
	}

	// Should have inline var declaration (mangled name is uppercase X).
	if !strings.Contains(goSrc, "var X float32 =") {
		t.Errorf("Expected inline var declaration for X:\n%s", goSrc)
	}
}
