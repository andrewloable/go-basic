package codegen

import (
	"strings"
	"testing"

	"github.com/loabletech/go-basic/internal/ast"
)

// ---------------------------------------------------------------------------
// Tests for emit_declarations.go — SUB, FUNCTION, TYPE, DIM, CONST, etc.
// ---------------------------------------------------------------------------

func TestDimArray(t *testing.T) {
	stmts := []ast.Statement{
		&ast.DimStatement{
			Declarations: []ast.DimDecl{
				{
					Name: "arr",
					Dimensions: []ast.DimRange{
						{Upper: &ast.NumberLiteral{Value: 10, NumType: ast.NumInt}},
					},
				},
			},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "make([]") {
		t.Errorf("expected make([] for array, got:\n%s", out)
	}
}

func TestDimScalar(t *testing.T) {
	stmts := []ast.Statement{
		&ast.DimStatement{
			Declarations: []ast.DimDecl{
				{Name: "count", TypeSuffix: "%"},
			},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "var count_pct int16") {
		t.Errorf("expected 'var count_pct int16', got:\n%s", out)
	}
}

func TestSubDeclaration(t *testing.T) {
	stmts := []ast.Statement{
		&ast.SubDeclaration{
			Name: "MySub",
			Params: []ast.Parameter{
				{Name: "x", Type: "INTEGER"},
			},
			Body: []ast.Statement{
				&ast.PrintStatement{
					Expressions: []ast.Expression{&ast.Identifier{Name: "x"}},
					Separators:  []string{""},
				},
			},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "func MySub(") {
		t.Errorf("expected func MySub(, got:\n%s", out)
	}
}

func TestFunctionDeclaration(t *testing.T) {
	stmts := []ast.Statement{
		&ast.FunctionDeclaration{
			Name:       "Double",
			ReturnType: "SINGLE",
			Params: []ast.Parameter{
				{Name: "n", Type: "SINGLE"},
			},
			Body: []ast.Statement{
				&ast.LetStatement{
					Name: &ast.Identifier{Name: "Double"},
					Value: &ast.BinaryExpr{
						Left:     &ast.Identifier{Name: "n"},
						Operator: "*",
						Right:    &ast.NumberLiteral{Value: 2, NumType: ast.NumInt},
					},
				},
			},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "func Double(") {
		t.Errorf("expected func Double(, got:\n%s", out)
	}
	if !strings.Contains(out, "float32") {
		t.Errorf("expected float32 return type, got:\n%s", out)
	}
}

func TestEmitTypeBlock(t *testing.T) {
	stmts := []ast.Statement{
		&ast.TypeBlockStatement{
			Name: "MyType",
			Fields: []ast.TypeField{
				{Name: "x", TypeName: "INTEGER"},
			},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "MyType") {
		t.Errorf("expected MyType in output, got:\n%s", out)
	}
}

func TestEmitDefFnSingleLine(t *testing.T) {
	stmts := []ast.Statement{
		&ast.DefFnDeclaration{
			Name: "FNsquare",
			Params: []ast.Parameter{
				{Name: "x"},
			},
			SingleLineExpr: &ast.BinaryExpr{
				Operator: "*",
				Left:     &ast.Identifier{Name: "x"},
				Right:    &ast.Identifier{Name: "x"},
			},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "FNsquare") {
		t.Errorf("expected 'FNsquare' in output, got:\n%s", out)
	}
}

func TestEmitDefFnMultiLine(t *testing.T) {
	stmts := []ast.Statement{
		&ast.DefFnDeclaration{
			Name: "FNdouble",
			Params: []ast.Parameter{
				{Name: "n"},
			},
			Body: []ast.Statement{
				&ast.LetStatement{
					Name:  &ast.Identifier{Name: "FNdouble"},
					Value: &ast.BinaryExpr{Operator: "*", Left: &ast.Identifier{Name: "n"}, Right: &ast.NumberLiteral{Value: 2, OriginalText: "2"}},
				},
			},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "FNdouble") {
		t.Errorf("expected 'FNdouble' in output, got:\n%s", out)
	}
}

func TestEmitRedim(t *testing.T) {
	stmts := []ast.Statement{
		&ast.RedimStatement{
			Declarations: []ast.DimDecl{
				{
					Name: "arr",
					Dimensions: []ast.DimRange{
						{Upper: &ast.NumberLiteral{Value: 20, OriginalText: "20"}},
					},
				},
			},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "arr") {
		t.Errorf("expected 'arr' in output, got:\n%s", out)
	}
}

func TestEmitRedimMultiDim(t *testing.T) {
	stmts := []ast.Statement{
		&ast.RedimStatement{
			Declarations: []ast.DimDecl{
				{
					Name: "grid",
					Dimensions: []ast.DimRange{
						{Upper: &ast.NumberLiteral{Value: 10, OriginalText: "10"}},
						{Upper: &ast.NumberLiteral{Value: 10, OriginalText: "10"}},
					},
				},
			},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "grid") {
		t.Errorf("REDIM multi-dim: expected 'grid' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "make(") {
		t.Errorf("REDIM multi-dim: expected 'make(' in output, got:\n%s", out)
	}
}

// TestSharedVarsPackageLevel verifies that SHARED variables inside SUBs
// are emitted as package-level var declarations.
func TestSharedVarsPackageLevel(t *testing.T) {
	stmts := []ast.Statement{
		&ast.DimStatement{
			Declarations: []ast.DimDecl{
				{Name: "Name1", TypeSuffix: "$", ElementType: "STRING"},
			},
		},
		&ast.LetStatement{
			Name:  &ast.Identifier{Name: "Name1", TypeSuffix: "$"},
			Value: &ast.StringLiteral{Value: "Alice"},
		},
		&ast.SubDeclaration{
			Name: "ShowInfo",
			Body: []ast.Statement{
				&ast.ScopeStatement{
					Modifier:  "SHARED",
					Variables: []string{"Name1$"},
				},
				&ast.PrintStatement{
					Expressions: []ast.Expression{
						&ast.Identifier{Name: "Name1", TypeSuffix: "$"},
					},
					Separators: []string{""},
				},
			},
		},
	}
	out := generate(t, stmts)

	if !strings.Contains(out, "var Name1_str string") {
		t.Errorf("expected package-level 'var Name1_str string', got:\n%s", out)
	}

	pkgIdx := strings.Index(out, "var Name1_str string")
	mainIdx := strings.Index(out, "func main()")
	if pkgIdx > mainIdx {
		t.Errorf("package-level var should appear before func main(), pkgIdx=%d mainIdx=%d", pkgIdx, mainIdx)
	}

	mainBody := out[mainIdx:]
	if strings.Contains(mainBody, "var Name1_str string") {
		t.Errorf("Name1_str should not be re-declared inside main(), got:\n%s", mainBody)
	}

	subIdx := strings.Index(out, "func ShowInfo()")
	if subIdx < 0 {
		t.Fatalf("expected func ShowInfo() in output, got:\n%s", out)
	}
	subBody := out[subIdx:]
	if strings.Contains(subBody, "var Name1_str") {
		t.Errorf("Name1_str should not be declared inside SUB, got:\n%s", subBody)
	}
}

// TestSubDeclSeparateScope verifies that variables declared inside a SUB
// do not leak into main's declared set.
func TestSubDeclSeparateScope(t *testing.T) {
	stmts := []ast.Statement{
		&ast.SubDeclaration{
			Name: "MySub",
			Body: []ast.Statement{
				&ast.LetStatement{
					Name:  &ast.Identifier{Name: "LocalVar", TypeSuffix: "$"},
					Value: &ast.StringLiteral{Value: "hello"},
				},
			},
		},
		&ast.LetStatement{
			Name:  &ast.Identifier{Name: "LocalVar", TypeSuffix: "$"},
			Value: &ast.StringLiteral{Value: "world"},
		},
	}
	out := generate(t, stmts)

	mainIdx := strings.Index(out, "func main()")
	subIdx := strings.Index(out, "func MySub()")
	if mainIdx < 0 || subIdx < 0 {
		t.Fatalf("expected func main() and func MySub() in output, got:\n%s", out)
	}

	mainBody := out[mainIdx:subIdx]
	subBody := out[subIdx:]

	if !strings.Contains(mainBody, "var LocalVar_str string") {
		t.Errorf("expected LocalVar_str declaration in main, got:\n%s", mainBody)
	}
	if !strings.Contains(subBody, "var LocalVar_str string") {
		t.Errorf("expected LocalVar_str declaration in SUB, got:\n%s", subBody)
	}
}
