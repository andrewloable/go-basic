package codegen

import (
	"strings"
	"testing"

	"github.com/loabletech/go-basic/internal/lexer"
	"github.com/loabletech/go-basic/internal/parser"
	"github.com/loabletech/go-basic/internal/semantic"
)

// TestEmitLetTypeCast verifies that emitLet wraps the RHS expression in a type cast.
func TestEmitLetTypeCast(t *testing.T) {
	src := `
DIM A%
DIM B!
DIM C#
A% = 5
B! = 3.14
C# = 2.718281828
A% = B!
B! = C#
`
	l := lexer.New(src)
	p := parser.New(l)
	program := p.ParseProgram()

	resolver := semantic.NewResolver(program)
	table, _ := resolver.Resolve()

	gen := New()
	out, err := gen.Generate(program, table)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	// A% should be int16, so assignments must use int16(...)
	if !strings.Contains(out, "int16(") {
		t.Errorf("expected int16() cast for A%% variable, got:\n%s", out)
	}
	// B! should be float32, so assignments must use float32(...)
	if !strings.Contains(out, "float32(") {
		t.Errorf("expected float32() cast for B! variable, got:\n%s", out)
	}
	// C# should be float64, so assignments must use float64(...)
	if !strings.Contains(out, "float64(") {
		t.Errorf("expected float64() cast for C# variable, got:\n%s", out)
	}

	t.Logf("Generated code:\n%s", out)
}

// TestEmitLetTypeCastString verifies that string variables are NOT wrapped in a type cast.
func TestEmitLetTypeCastString(t *testing.T) {
	src := `
DIM S$
S$ = "hello"
`
	l := lexer.New(src)
	p := parser.New(l)
	program := p.ParseProgram()

	resolver := semantic.NewResolver(program)
	table, _ := resolver.Resolve()

	gen := New()
	out, err := gen.Generate(program, table)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	// String variables should not have string(...) cast
	if strings.Contains(out, "string(\"hello\")") {
		t.Errorf("string variable should not have string() cast, got:\n%s", out)
	}

	// But it should have a string declaration (name mangled: S$ -> S_str)
	if !strings.Contains(out, "var S_str string") {
		t.Errorf("expected 'var S_str string' declaration, got:\n%s", out)
	}

	t.Logf("Generated code:\n%s", out)
}

// TestDataReadTypeCast verifies that READ statements cast tv_ to the target type.
func TestDataReadTypeCast(t *testing.T) {
	src := `
DATA 10, 20, 30
DIM A%
DIM B!
READ A%
READ B!
`
	l := lexer.New(src)
	p := parser.New(l)
	program := p.ParseProgram()

	resolver := semantic.NewResolver(program)
	table, _ := resolver.Resolve()

	gen := New()
	out, err := gen.Generate(program, table)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	// int16 target should have int16(tv_) cast
	if !strings.Contains(out, "= int16(tv_)") {
		t.Errorf("expected int16(tv_) cast for READ into A%%, got:\n%s", out)
	}
	// float32 target should have float32(tv_) cast
	if !strings.Contains(out, "= float32(tv_)") {
		t.Errorf("expected float32(tv_) cast for READ into B!, got:\n%s", out)
	}

	t.Logf("Generated code:\n%s", out)
}
// TestBinaryExprTypePromotion verifies that emitBinaryExpr inserts widening
// casts when operands have mismatched numeric types.
func TestBinaryExprTypePromotion(t *testing.T) {
	// float32 variable + float64 variable -> float64(B_sng) or float32(C_dbl) cast
	src := `
DIM B!
DIM C#
DIM D#
B! = 1.5
C# = 2.5
D# = B! + C#
`
	l := lexer.New(src)
	p := parser.New(l)
	program := p.ParseProgram()

	resolver := semantic.NewResolver(program)
	table, _ := resolver.Resolve()

	gen := New()
	out, err := gen.Generate(program, table)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	// B! is float32, C# is float64. The addition B! + C# must cast
	// the narrower (B! / float32) to float64 so Go accepts it.
	if !strings.Contains(out, "float64(B_sng)") {
		t.Errorf("expected float64(B_sng) widening cast in B! + C#, got:\n%s", out)
	}

	t.Logf("Generated code:\n%s", out)
}

// TestBinaryExprInt16Float32Promotion verifies int16 + float32 -> float32 cast.
func TestBinaryExprInt16Float32Promotion(t *testing.T) {
	src := `
DIM A%
DIM B!
DIM C!
A% = 3
B! = 1.5
C! = A% + B!
`
	l := lexer.New(src)
	p := parser.New(l)
	program := p.ParseProgram()

	resolver := semantic.NewResolver(program)
	table, _ := resolver.Resolve()

	gen := New()
	out, err := gen.Generate(program, table)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	// A% is int16, B! is float32. The narrower (A% / int16) must be cast to float32.
	if !strings.Contains(out, "float32(A_pct)") {
		t.Errorf("expected float32(A_pct) widening cast in A%% + B!, got:\n%s", out)
	}

	t.Logf("Generated code:\n%s", out)
}

// TestBinaryExprSameTypesNoExtraCast verifies that same-type expressions
// do NOT get extra casts inserted in the binary expression itself.
func TestBinaryExprSameTypesNoExtraCast(t *testing.T) {
	src := `
DIM A!
DIM B!
DIM C!
A! = 1.0
B! = 2.0
C! = A! + B!
`
	l := lexer.New(src)
	p := parser.New(l)
	program := p.ParseProgram()

	resolver := semantic.NewResolver(program)
	table, _ := resolver.Resolve()

	gen := New()
	out, err := gen.Generate(program, table)
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	// Same types (float32 + float32): the binary expression should not contain
	// extra float32(...) casts around A_sng or B_sng in the addition.
	addLine := ""
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "C_sng =") {
			addLine = line
		}
	}
	if strings.Contains(addLine, "float32(A_sng)") {
		t.Errorf("unexpected extra float32() cast for same-type operands, line: %s", addLine)
	}

	t.Logf("Generated code:\n%s", out)
}
