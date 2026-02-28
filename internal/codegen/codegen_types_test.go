package codegen

import (
	"testing"

	"github.com/loabletech/go-basic/internal/ast"
	"github.com/loabletech/go-basic/internal/semantic"
)

// ---------------------------------------------------------------------------
// Tests for codegen_types.go — type name generation, suffix handling,
// casting, name mangling
// ---------------------------------------------------------------------------

func TestMangleName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"x%", "x_pct"},
		{"name$", "name_str"},
		{"count&", "count_lng"},
		{"rate!", "rate_sng"},
		{"pi#", "pi_dbl"},
		{"x", "x"},
		{"return", "b_return"}, // reserved word
		{"MyVar", "MyVar"},
	}
	for _, tc := range tests {
		got := mangleName(tc.input)
		if got != tc.want {
			t.Errorf("mangleName(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestGoType(t *testing.T) {
	tests := []struct {
		dt   semantic.DataType
		want string
	}{
		{semantic.TypeInteger, "int16"},
		{semantic.TypeLong, "int32"},
		{semantic.TypeSingle, "float32"},
		{semantic.TypeDouble, "float64"},
		{semantic.TypeString, "string"},
		{semantic.TypeUnknown, "float64"},
	}
	for _, tc := range tests {
		got := goType(tc.dt)
		if got != tc.want {
			t.Errorf("goType(%v) = %q, want %q", tc.dt, got, tc.want)
		}
	}
}

func TestGoTypeFromSuffix(t *testing.T) {
	cases := []struct {
		name string
		want string
	}{
		{"x%", "int16"},
		{"x&", "int32"},
		{"x!", "float32"},
		{"x#", "float64"},
		{"x$", "string"},
		{"x", "float64"},
		{"", "float64"},
	}
	for _, tc := range cases {
		got := goTypeFromSuffix(tc.name)
		if got != tc.want {
			t.Errorf("goTypeFromSuffix(%q) = %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestGoTypeForReturnType(t *testing.T) {
	gen := New()
	cases := []struct {
		retType, name, want string
	}{
		{"INTEGER", "f", "int16"},
		{"%", "f", "int16"},
		{"LONG", "f", "int32"},
		{"&", "f", "int32"},
		{"SINGLE", "f", "float32"},
		{"!", "f", "float32"},
		{"DOUBLE", "f", "float64"},
		{"#", "f", "float64"},
		{"STRING", "f", "string"},
		{"$", "f", "string"},
		{"", "f#", "float64"}, // fallback via name suffix
	}
	for _, tc := range cases {
		got := gen.goTypeForReturnType(tc.retType, tc.name)
		if got != tc.want {
			t.Errorf("goTypeForReturnType(%q, %q) = %q, want %q", tc.retType, tc.name, got, tc.want)
		}
	}
}

func TestGoTypeForParamType(t *testing.T) {
	gen := New()
	cases := []struct {
		typeStr, name, want string
	}{
		{"INTEGER", "p", "int16"},
		{"%", "p", "int16"},
		{"LONG", "p", "int32"},
		{"&", "p", "int32"},
		{"SINGLE", "p", "float32"},
		{"!", "p", "float32"},
		{"DOUBLE", "p", "float64"},
		{"#", "p", "float64"},
		{"STRING", "p", "string"},
		{"$", "p", "string"},
		{"", "p%", "int16"}, // fallback via name suffix
	}
	for _, tc := range cases {
		got := gen.goTypeForParamType(tc.typeStr, tc.name)
		if got != tc.want {
			t.Errorf("goTypeForParamType(%q, %q) = %q, want %q", tc.typeStr, tc.name, got, tc.want)
		}
	}
}

func TestSanitizeGoIdent(t *testing.T) {
	cases := []struct {
		input, want string
	}{
		{"hello", "hello"},
		{"hello_world", "hello_world"},
		{"123abc", "_123abc"},
		{"a.b.c", "a_b_c"},
		{"!@#", "_"},
		{"", "_"},
	}
	for _, tc := range cases {
		got := sanitizeGoIdent(tc.input)
		if got != tc.want {
			t.Errorf("sanitizeGoIdent(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestNumericWidth(t *testing.T) {
	cases := []struct {
		typ  string
		want int
	}{
		{"int16", 0},
		{"int32", 1},
		{"float32", 2},
		{"float64", 3},
		{"string", -1},
		{"unknown", -1},
	}
	for _, tc := range cases {
		got := numericWidth(tc.typ)
		if got != tc.want {
			t.Errorf("numericWidth(%q) = %d, want %d", tc.typ, got, tc.want)
		}
	}
}

func TestWidenType(t *testing.T) {
	cases := []struct {
		a, b, want string
	}{
		{"int16", "int32", "int32"},
		{"int32", "int16", "int32"},
		{"float32", "float64", "float64"},
		{"float64", "float32", "float64"},
		{"int16", "int16", "int16"},
		{"string", "float64", "float64"}, // non-numeric → float64
		{"float32", "string", "float64"}, // non-numeric → float64
	}
	for _, tc := range cases {
		got := widenType(tc.a, tc.b)
		if got != tc.want {
			t.Errorf("widenType(%q, %q) = %q, want %q", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestZeroValueForType(t *testing.T) {
	if got := zeroValueForType("string"); got != `""` {
		t.Errorf("zeroValueForType(string) = %q, want \"\"", got)
	}
	if got := zeroValueForType("float64"); got != "0" {
		t.Errorf("zeroValueForType(float64) = %q, want 0", got)
	}
	if got := zeroValueForType("int16"); got != "0" {
		t.Errorf("zeroValueForType(int16) = %q, want 0", got)
	}
}

func TestGoTypeForBuiltin(t *testing.T) {
	cases := []struct {
		name, want string
	}{
		// Integer-returning
		{"LEN", "int"},
		{"INSTR", "int"},
		{"ASC", "int"},
		{"PEEK", "int"},
		// int16-returning
		{"CINT", "int16"},
		{"CVI", "int16"},
		// int32-returning
		{"CLNG", "int32"},
		{"CVL", "int32"},
		// float32-returning
		{"CSNG", "float32"},
		{"CVS", "float32"},
		// String-returning
		{"LEFT$", "string"},
		{"RIGHT$", "string"},
		{"MID$", "string"},
		{"CHR$", "string"},
		{"STR$", "string"},
		{"HEX$", "string"},
		{"OCT$", "string"},
		{"BIN$", "string"},
		{"UCASE$", "string"},
		{"LCASE$", "string"},
		{"LTRIM$", "string"},
		{"RTRIM$", "string"},
		{"TRIM$", "string"},
		{"SPACE$", "string"},
		{"STRING$", "string"},
		{"MKI$", "string"},
		{"MKL$", "string"},
		{"MKS$", "string"},
		{"MKD$", "string"},
		{"DATE$", "string"},
		{"TIME$", "string"},
		{"INKEY$", "string"},
		{"COMMAND$", "string"},
		{"ENVIRON$", "string"},
		{"TAB", "string"},
		{"SPC", "string"},
		// Default float64
		{"ABS", "float64"},
		{"SIN", "float64"},
		{"COS", "float64"},
		{"UNKNOWN_FUNC", "float64"},
	}
	for _, tc := range cases {
		got := goTypeForBuiltin(tc.name)
		if got != tc.want {
			t.Errorf("goTypeForBuiltin(%q) = %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestGoTypeForExprNumberLiterals(t *testing.T) {
	gen := New()

	cases := []struct {
		numType int
		want    string
	}{
		{ast.NumInt, "int16"},
		{ast.NumLong, "int32"},
		{ast.NumSingle, "float32"},
		{ast.NumDouble, "float64"},
	}
	for _, tc := range cases {
		expr := &ast.NumberLiteral{Value: 1, NumType: tc.numType}
		got := gen.goTypeForExpr(expr)
		if got != tc.want {
			t.Errorf("goTypeForExpr(NumberLiteral NumType=%d) = %q, want %q", tc.numType, got, tc.want)
		}
	}

	// nil expression
	if got := gen.goTypeForExpr(nil); got != "float64" {
		t.Errorf("goTypeForExpr(nil) = %q, want float64", got)
	}
}

func TestGoTypeForExprGroupExpr(t *testing.T) {
	gen := New()
	// GroupExpr wrapping a string literal → "string"
	expr := &ast.GroupExpr{
		Inner: &ast.StringLiteral{Value: "hello"},
	}
	got := gen.goTypeForExpr(expr)
	if got != "string" {
		t.Errorf("goTypeForExpr(GroupExpr<string>) = %q, want string", got)
	}
}

func TestGoTypeForExprBinaryExpr(t *testing.T) {
	gen := New()
	// int16 + int32 → int32 (widened)
	expr := &ast.BinaryExpr{
		Operator: "+",
		Left:     &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
		Right:    &ast.NumberLiteral{Value: 2, NumType: ast.NumLong},
	}
	got := gen.goTypeForExpr(expr)
	if got != "int32" {
		t.Errorf("goTypeForExpr(int16+int32) = %q, want int32", got)
	}
}

func TestGoTypeForExprFunctionCall(t *testing.T) {
	gen := New()
	// FunctionCall for LEN → "int"
	expr := &ast.FunctionCall{
		Name: "LEN",
		Args: []ast.Expression{&ast.StringLiteral{Value: "hi"}},
	}
	got := gen.goTypeForExpr(expr)
	if got != "int" {
		t.Errorf("goTypeForExpr(FunctionCall LEN) = %q, want int", got)
	}
}

func TestGoTypeForExprFnCallExpression(t *testing.T) {
	gen := New()
	// FnCallExpression — type inferred from name suffix
	expr := &ast.FnCallExpression{
		Name: "FNresult#",
		Args: []ast.Expression{},
	}
	got := gen.goTypeForExpr(expr)
	if got != "float64" {
		t.Errorf("goTypeForExpr(FnCallExpression with # suffix) = %q, want float64", got)
	}
}

func TestGoTypeForExprFieldAccess(t *testing.T) {
	gen := New()
	// FieldAccessExpression → conservative float64
	expr := &ast.FieldAccessExpression{
		Object: &ast.Identifier{Name: "obj"},
		Field:  "x",
	}
	got := gen.goTypeForExpr(expr)
	if got != "float64" {
		t.Errorf("goTypeForExpr(FieldAccessExpression) = %q, want float64", got)
	}
}

func TestToBoolExprAlreadyBool(t *testing.T) {
	gen := New()

	cases := []struct {
		expr string
		want string
	}{
		{"(x == 0)", "(x == 0)"},
		{"(x != y)", "(x != y)"},
		{"(a <= b)", "(a <= b)"},
		{"(a >= b)", "(a >= b)"},
		{"(p && q)", "(p && q)"},
		{"(p || q)", "(p || q)"},
		{"!(x)", "!(x)"},
		{"true", "true"},
		{"false", "false"},
		{"(a < b)", "(a < b)"},
		{"(a > b)", "(a > b)"},
	}
	for _, tc := range cases {
		got := gen.toBoolExpr(tc.expr)
		if got != tc.want {
			t.Errorf("toBoolExpr(%q) = %q, want %q", tc.expr, got, tc.want)
		}
	}
}

func TestToBoolExprNumeric(t *testing.T) {
	gen := New()
	// A plain numeric expression should be wrapped in != 0
	got := gen.toBoolExpr("x")
	if got != "(x) != 0" {
		t.Errorf("toBoolExpr(x) = %q, want '(x) != 0'", got)
	}
}
