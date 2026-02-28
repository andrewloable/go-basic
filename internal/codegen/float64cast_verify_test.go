package codegen

import (
	"strings"
	"testing"

	"github.com/loabletech/go-basic/internal/lexer"
	"github.com/loabletech/go-basic/internal/parser"
	"github.com/loabletech/go-basic/internal/semantic"
)

// TestMathFunctionFloat32ArgCast checks that when a float32 (SINGLE) variable
// is passed to a math builtin (SQR, SIN, COS, etc.), the emitted Go wraps the
// argument in float64(...) so the generated code compiles without error.
func TestMathFunctionFloat32ArgCast(t *testing.T) {
	// Use BASIC type-suffix notation (!) for SINGLE so the variable is named X_sng.
	src := `
DIM X! AS SINGLE
DIM N% AS INTEGER
X! = 25.0
N% = 16
PRINT SQR(X!)
PRINT SIN(X!)
PRINT COS(X!)
PRINT ABS(N%)
PRINT INT(X!)
END
`
	l := lexer.New(src)
	p := parser.New(l)
	program := p.ParseProgram()
	resolver := semantic.NewResolver(program)
	table, _ := resolver.Resolve()
	gen := New()
	goSrc, err := gen.Generate(program, table)
	if err != nil {
		t.Fatalf("generate error: %v", err)
	}

	t.Logf("Generated source:\n%s", goSrc)

	// Verify that float64 casts are present for the typed variables.
	for _, want := range []string{
		"float64(X_sng)",
		"float64(N_pct)",
	} {
		if !strings.Contains(goSrc, want) {
			t.Errorf("expected cast %q in generated code, got:\n%s", want, goSrc)
		}
	}
}

// TestMathFunctionFloat32ArgCastNoCastForFloat64 verifies that when an
// argument is already float64, no redundant float64(...) cast is emitted.
func TestMathFunctionFloat32ArgCastNoCastForFloat64(t *testing.T) {
	src := `
DIM Y# AS DOUBLE
Y# = 9.0
PRINT SQR(Y#)
END
`
	l := lexer.New(src)
	p := parser.New(l)
	program := p.ParseProgram()
	resolver := semantic.NewResolver(program)
	table, _ := resolver.Resolve()
	gen := New()
	goSrc, err := gen.Generate(program, table)
	if err != nil {
		t.Fatalf("generate error: %v", err)
	}

	t.Logf("Generated source:\n%s", goSrc)

	// For a float64 argument, float64() cast should NOT be emitted (it is redundant).
	if strings.Contains(goSrc, "float64(Y_dbl)") {
		t.Errorf("unexpected redundant float64 cast for already-float64 variable Y# in:\n%s", goSrc)
	}
	// The variable should be passed directly.
	if !strings.Contains(goSrc, "rt.Sqr(Y_dbl)") {
		t.Errorf("expected rt.Sqr(Y_dbl) (no cast) in generated code, got:\n%s", goSrc)
	}
}
