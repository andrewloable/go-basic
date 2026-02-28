package semantic

import (
	"testing"

	"github.com/loabletech/go-basic/internal/ast"
)

// ===========================================================================
// Expression resolver tests (resolve_expressions.go)
// ===========================================================================

func TestResolveExpressionUnaryExpr(t *testing.T) {
	// A unary expression should be resolved without error.
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LetStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Name:    &ast.Identifier{Name: "x"},
				Value: &ast.UnaryExpr{
					BasePos:  ast.Position{Line: 1, Column: 5},
					Operator: "-",
					Operand:  &ast.NumberLiteral{Value: 5, NumType: ast.NumInt},
				},
			},
		},
	}
	rr := NewResolver(prog)
	_, errs := rr.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for UnaryExpr resolution, got: %v", errs)
	}
}

func TestResolveExpressionGroupExpr(t *testing.T) {
	// A grouped expression should be resolved without error.
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LetStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Name:    &ast.Identifier{Name: "y"},
				Value: &ast.GroupExpr{
					BasePos: ast.Position{Line: 1, Column: 5},
					Inner:   &ast.NumberLiteral{Value: 10, NumType: ast.NumInt},
				},
			},
		},
	}
	rr := NewResolver(prog)
	_, errs := rr.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for GroupExpr resolution, got: %v", errs)
	}
}

// ===========================================================================
// resolveElementType, paramDataType, resolveReturnType helper tests
// (these helpers live in resolve_expressions.go)
// ===========================================================================

func TestResolverArrayAccessExpression(t *testing.T) {
	// DIM scores%(10) / x% = scores%(3)
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.DimStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Declarations: []ast.DimDecl{
					{
						Name:       "scores",
						TypeSuffix: "%",
						Dimensions: []ast.DimRange{
							{Upper: &ast.NumberLiteral{Value: 10, OriginalText: "10"}},
						},
					},
				},
			},
			&ast.LetStatement{
				BasePos: ast.Position{Line: 2, Column: 1},
				Name:    &ast.Identifier{Name: "x", TypeSuffix: "%"},
				Value: &ast.ArrayAccess{
					Name:       "scores",
					TypeSuffix: "%",
					Indices:    []ast.Expression{&ast.NumberLiteral{Value: 3, OriginalText: "3"}},
				},
			},
		},
	}
	resolver := NewResolver(prog)
	table, errs := resolver.Resolve()
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	sym := table.Lookup("scores%")
	if sym == nil {
		t.Fatal("scores% should be in symbol table")
	}
	if !sym.Used {
		t.Error("scores% should be marked as Used after array access")
	}
}

func TestResolveElementTypeWithASKeyword(t *testing.T) {
	// DIM x AS INTEGER (elementType = "INTEGER")
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.DimStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Declarations: []ast.DimDecl{
					{Name: "x", TypeSuffix: "", ElementType: "INTEGER"},
					{Name: "y", TypeSuffix: "", ElementType: "LONG"},
					{Name: "z", TypeSuffix: "", ElementType: "SINGLE"},
					{Name: "w", TypeSuffix: "", ElementType: "DOUBLE"},
					{Name: "s", TypeSuffix: "", ElementType: "STRING"},
				},
			},
		},
	}
	resolver := NewResolver(prog)
	table, errs := resolver.Resolve()
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	tests := []struct {
		name string
		want DataType
	}{
		{"x", TypeInteger},
		{"y", TypeLong},
		{"z", TypeSingle},
		{"w", TypeDouble},
		{"s", TypeString},
	}
	for _, tc := range tests {
		sym := table.Lookup(tc.name)
		if sym == nil {
			t.Fatalf("%s should be in symbol table", tc.name)
		}
		if sym.DataType != tc.want {
			t.Errorf("%s: expected %v, got %v", tc.name, tc.want, sym.DataType)
		}
	}
}

func TestParamDataTypeAllVariants(t *testing.T) {
	st := NewSymbolTable()
	tests := []struct {
		typeStr string
		name    string
		want    DataType
	}{
		{"INTEGER", "x", TypeInteger},
		{"%", "x", TypeInteger},
		{"LONG", "x", TypeLong},
		{"&", "x", TypeLong},
		{"SINGLE", "x", TypeSingle},
		{"!", "x", TypeSingle},
		{"DOUBLE", "x", TypeDouble},
		{"#", "x", TypeDouble},
		{"STRING", "x", TypeString},
		{"$", "x", TypeString},
		{"", "count%", TypeInteger}, // fallback via suffix
		{"", "ratio", TypeSingle},  // fallback default
	}
	for _, tc := range tests {
		got := paramDataType(tc.typeStr, tc.name, st)
		if got != tc.want {
			t.Errorf("paramDataType(%q, %q) = %v, want %v", tc.typeStr, tc.name, got, tc.want)
		}
	}
}

func TestResolveReturnTypeAllVariants(t *testing.T) {
	st := NewSymbolTable()
	tests := []struct {
		retType string
		name    string
		want    DataType
	}{
		{"INTEGER", "fn", TypeInteger},
		{"%", "fn", TypeInteger},
		{"LONG", "fn", TypeLong},
		{"&", "fn", TypeLong},
		{"SINGLE", "fn", TypeSingle},
		{"!", "fn", TypeSingle},
		{"DOUBLE", "fn", TypeDouble},
		{"#", "fn", TypeDouble},
		{"STRING", "fn", TypeString},
		{"$", "fn", TypeString},
		{"", "fnResult%", TypeInteger}, // fallback via suffix
		{"", "fnResult", TypeSingle},  // fallback default
	}
	for _, tc := range tests {
		got := resolveReturnType(tc.retType, tc.name, st)
		if got != tc.want {
			t.Errorf("resolveReturnType(%q, %q) = %v, want %v", tc.retType, tc.name, got, tc.want)
		}
	}
}

func TestResolveReturnTypeFromNameSuffix(t *testing.T) {
	st := NewSymbolTable()
	// Name ends in # → double
	got := resolveReturnType("", "Result#", st)
	if got != TypeDouble {
		t.Errorf("expected TypeDouble from # suffix in name, got %v", got)
	}
	// Name ends in $ → string
	got = resolveReturnType("", "Name$", st)
	if got != TypeString {
		t.Errorf("expected TypeString from $ suffix in name, got %v", got)
	}
}
