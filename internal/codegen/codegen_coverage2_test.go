package codegen

// codegen_coverage2_test.go — additional tests targeting uncovered branches
// identified via go tool cover.  Focus areas:
//   - collectMainVariables (SUB/FUNCTION in program, ForStatement, DimStatement,
//     ReadStatement, IfStatement, WhileStatement, DoLoopStatement, SelectCase,
//     DefFnDeclaration)
//   - emitDoLoop  (WHILE/UNTIL at top and bottom, no condition)
//   - emitExit    (FOR, DO, WHILE, LOOP, SUB, FUNCTION, DEF)
//   - emitIncr/emitDecr with Amount
//   - emitInputFromStdin  (multiple vars, LINE INPUT, no prompt, string var)
//   - emitFileInput  (LineInput, numeric var)
//   - emitView  (viewport variant with optional coords)
//   - widenType / numericWidth / zeroValueForType (called directly)
//   - goTypeForExpr  (NumberLiteral with typed NumType, GroupExpr, BinaryExpr,
//     FnCallExpression, FieldAccessExpression, nil)
//   - goTypeForBuiltin  (LEN, CINT, CLNG, CSNG, LEFT$, etc.)
//   - emitBinaryExpr  (all operators)
//   - emitFunctionCall  (many builtins)
//   - emitSelectCase  (IS comparison, range cases)
//   - emitCircle  (with Start/End/Aspect)
//   - emitFilePrint  (no expressions)
//   - emitFileWrite  (no expressions)
//   - emitRedim  (multi-dim)
//   - emitPrint  (comma separators, multiple exprs, trailing sep)
//   - emitStatement  (uncovered branches)

import (
	"strings"
	"testing"

	"github.com/loabletech/go-basic/internal/ast"
)

// ===========================================================================
// widenType / numericWidth / zeroValueForType — called directly
// ===========================================================================

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

// ===========================================================================
// goTypeForBuiltin
// ===========================================================================

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

// ===========================================================================
// goTypeForExpr — various expression types
// ===========================================================================

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
		// default NumType (0 == NumInt — actually we test the "default" branch
		// by using a value that doesn't match the switch)
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

// ===========================================================================
// emitBinaryExpr — all operators
// ===========================================================================

func TestEmitBinaryExprAllOps(t *testing.T) {
	cases := []struct {
		op      string
		contain string
	}{
		{"+", "+"},
		{"-", "-"},
		{"*", "*"},
		{"/", "/"},
		{"\\", "int("},    // integer division
		{"MOD", "% int("},  // modulo (Go uses %)
		{"^", "math.Pow"},
		{"=", "=="},
		{"<>", "!="},
		{"<", " < "},
		{">", " > "},
		{"<=", "<="},
		{">=", ">="},
		{"AND", "& int("},
		{"OR", "| int("},
		{"XOR", "^ int("},
		{"EQV", "^(int("},
		{"IMP", "(^int("},
	}
	for _, tc := range cases {
		stmts := []ast.Statement{
			&ast.PrintStatement{
				Expressions: []ast.Expression{
					&ast.BinaryExpr{
						Operator: tc.op,
						Left:     &ast.NumberLiteral{Value: 4, OriginalText: "4"},
						Right:    &ast.NumberLiteral{Value: 2, OriginalText: "2"},
					},
				},
				Separators: []string{""},
			},
		}
		out := generate(t, stmts)
		if !strings.Contains(out, tc.contain) {
			t.Errorf("emitBinaryExpr op=%q: expected %q in output, got:\n%s", tc.op, tc.contain, out)
		}
	}
}

func TestEmitBinaryExprStringConcat(t *testing.T) {
	// String + String should not wrap in type casts
	stmts := []ast.Statement{
		&ast.PrintStatement{
			Expressions: []ast.Expression{
				&ast.BinaryExpr{
					Operator: "+",
					Left:     &ast.StringLiteral{Value: "hello"},
					Right:    &ast.StringLiteral{Value: " world"},
				},
			},
			Separators: []string{""},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, `"hello"`) {
		t.Errorf("expected string literal in output, got:\n%s", out)
	}
}

func TestEmitBinaryExprUnknownOp(t *testing.T) {
	// Unknown operator should still produce output with comment
	stmts := []ast.Statement{
		&ast.PrintStatement{
			Expressions: []ast.Expression{
				&ast.BinaryExpr{
					Operator: "&",
					Left:     &ast.NumberLiteral{Value: 1, OriginalText: "1"},
					Right:    &ast.NumberLiteral{Value: 2, OriginalText: "2"},
				},
			},
			Separators: []string{""},
		},
	}
	out := generate(t, stmts)
	// The default branch emits a comment with the operator
	if !strings.Contains(out, "&") {
		t.Errorf("expected & in output for unknown op, got:\n%s", out)
	}
}

// ===========================================================================
// emitFunctionCall — all builtins
// ===========================================================================

func makeCallPrint(name string, args ...ast.Expression) []ast.Statement {
	return []ast.Statement{
		&ast.PrintStatement{
			Expressions: []ast.Expression{
				&ast.FunctionCall{Name: name, Args: args},
			},
			Separators: []string{""},
		},
	}
}

func TestEmitFunctionCallMath(t *testing.T) {
	num := &ast.NumberLiteral{Value: 4, OriginalText: "4"}
	cases := []struct{ name, want string }{
		{"ABS", "rt.Abs("},
		{"SGN", "rt.Sgn("},
		{"INT", "rt.IntFloor("},
		{"FIX", "rt.Fix("},
		{"CEIL", "rt.Ceil("},
		{"SQR", "rt.Sqr("},
		{"EXP", "rt.Exp("},
		{"EXP2", "rt.Exp2("},
		{"EXP10", "rt.Exp10("},
		{"LOG", "rt.Log("},
		{"LOG2", "rt.Log2("},
		{"LOG10", "rt.Log10("},
		{"SIN", "rt.Sin("},
		{"COS", "rt.Cos("},
		{"TAN", "rt.Tan("},
		{"ATN", "rt.Atn("},
	}
	for _, tc := range cases {
		out := generate(t, makeCallPrint(tc.name, num))
		if !strings.Contains(out, tc.want) {
			t.Errorf("FunctionCall %q: expected %q in output, got:\n%s", tc.name, tc.want, out)
		}
	}
}

func TestEmitFunctionCallConversions(t *testing.T) {
	num := &ast.NumberLiteral{Value: 3, OriginalText: "3"}
	cases := []struct{ name, want string }{
		{"CINT", "rt.Cint("},
		{"CLNG", "rt.Clng("},
		{"CSNG", "rt.Csng("},
		{"CDBL", "rt.Cdbl("},
	}
	for _, tc := range cases {
		out := generate(t, makeCallPrint(tc.name, num))
		if !strings.Contains(out, tc.want) {
			t.Errorf("FunctionCall %q: expected %q in output, got:\n%s", tc.name, tc.want, out)
		}
	}
}

func TestEmitFunctionCallStrings(t *testing.T) {
	str := &ast.StringLiteral{Value: "hello"}
	num := &ast.NumberLiteral{Value: 3, OriginalText: "3"}

	cases := []struct {
		name string
		args []ast.Expression
		want string
	}{
		{"LEFT$", []ast.Expression{str, num}, "rt.Left("},
		{"RIGHT$", []ast.Expression{str, num}, "rt.Right("},
		{"MID$", []ast.Expression{str, num, num}, "rt.Mid("},
		{"MID$", []ast.Expression{str, num}, "rt.Mid("},
		{"LEN", []ast.Expression{str}, "rt.Len("},
		{"INSTR", []ast.Expression{str, str}, "rt.Instr("},
		{"INSTR", []ast.Expression{num, str, str}, "rt.Instr("},
		{"ASC", []ast.Expression{str}, "rt.Asc("},
		{"CHR$", []ast.Expression{num}, "rt.Chr("},
		{"STR$", []ast.Expression{num}, "rt.Str("},
		{"VAL", []ast.Expression{str}, "rt.Val("},
		{"HEX$", []ast.Expression{num}, "rt.Hex("},
		{"OCT$", []ast.Expression{num}, "rt.Oct("},
		{"BIN$", []ast.Expression{num}, "rt.Bin("},
		{"UCASE$", []ast.Expression{str}, "rt.UCase("},
		{"LCASE$", []ast.Expression{str}, "rt.LCase("},
		{"LTRIM$", []ast.Expression{str}, "rt.LTrim("},
		{"RTRIM$", []ast.Expression{str}, "rt.RTrim("},
		{"TRIM$", []ast.Expression{str}, "rt.Trim("},
		{"SPACE$", []ast.Expression{num}, "rt.Space("},
		{"STRING$", []ast.Expression{num, num}, "rt.StringRepeat("},
	}
	for _, tc := range cases {
		out := generate(t, makeCallPrint(tc.name, tc.args...))
		if !strings.Contains(out, tc.want) {
			t.Errorf("FunctionCall %q: expected %q in output, got:\n%s", tc.name, tc.want, out)
		}
	}
}

func TestEmitFunctionCallMkCv(t *testing.T) {
	str := &ast.StringLiteral{Value: "ab"}
	num := &ast.NumberLiteral{Value: 1, OriginalText: "1"}
	cases := []struct {
		name string
		args []ast.Expression
		want string
	}{
		{"MKI$", []ast.Expression{num}, "rt.Mki("},
		{"MKL$", []ast.Expression{num}, "rt.Mkl("},
		{"MKS$", []ast.Expression{num}, "rt.Mks("},
		{"MKD$", []ast.Expression{num}, "rt.Mkd("},
		{"CVI", []ast.Expression{str}, "rt.Cvi("},
		{"CVL", []ast.Expression{str}, "rt.Cvl("},
		{"CVS", []ast.Expression{str}, "rt.Cvs("},
		{"CVD", []ast.Expression{str}, "rt.Cvd("},
	}
	for _, tc := range cases {
		out := generate(t, makeCallPrint(tc.name, tc.args...))
		if !strings.Contains(out, tc.want) {
			t.Errorf("FunctionCall %q: expected %q in output, got:\n%s", tc.name, tc.want, out)
		}
	}
}

func TestEmitFunctionCallRndTimer(t *testing.T) {
	num := &ast.NumberLiteral{Value: 1, OriginalText: "1"}

	// RND with no args
	out := generate(t, makeCallPrint("RND"))
	if !strings.Contains(out, "rng.Rnd(") {
		t.Errorf("RND (no args): expected rng.Rnd( in output, got:\n%s", out)
	}

	// RND with arg
	out = generate(t, makeCallPrint("RND", num))
	if !strings.Contains(out, "rng.Rnd(") {
		t.Errorf("RND (with arg): expected rng.Rnd( in output, got:\n%s", out)
	}

	// TIMER
	out = generate(t, makeCallPrint("TIMER"))
	if !strings.Contains(out, "rt.Timer()") {
		t.Errorf("TIMER: expected rt.Timer() in output, got:\n%s", out)
	}

	// DATE$
	out = generate(t, makeCallPrint("DATE$"))
	if !strings.Contains(out, "rt.DateStr()") {
		t.Errorf("DATE$: expected rt.DateStr() in output, got:\n%s", out)
	}

	// TIME$
	out = generate(t, makeCallPrint("TIME$"))
	if !strings.Contains(out, "rt.TimeStr()") {
		t.Errorf("TIME$: expected rt.TimeStr() in output, got:\n%s", out)
	}

	// COMMAND$
	out = generate(t, makeCallPrint("COMMAND$"))
	if !strings.Contains(out, "rt.CommandStr()") {
		t.Errorf("COMMAND$: expected rt.CommandStr() in output, got:\n%s", out)
	}

	// ENVIRON$
	out = generate(t, makeCallPrint("ENVIRON$", &ast.StringLiteral{Value: "PATH"}))
	if !strings.Contains(out, "rt.EnvironGet(") {
		t.Errorf("ENVIRON$: expected rt.EnvironGet( in output, got:\n%s", out)
	}
}

func TestEmitFunctionCallTabSpc(t *testing.T) {
	num := &ast.NumberLiteral{Value: 10, OriginalText: "10"}

	out := generate(t, makeCallPrint("TAB", num))
	if !strings.Contains(out, "rt.Spc(") {
		t.Errorf("TAB: expected rt.Spc( in output, got:\n%s", out)
	}

	out = generate(t, makeCallPrint("SPC", num))
	if !strings.Contains(out, "rt.Spc(") {
		t.Errorf("SPC: expected rt.Spc( in output, got:\n%s", out)
	}
}

func TestEmitFunctionCallFrePeek(t *testing.T) {
	num := &ast.NumberLiteral{Value: 0, OriginalText: "0"}

	// FRE with no args
	out := generate(t, makeCallPrint("FRE"))
	if !strings.Contains(out, "rt.Fre(") {
		t.Errorf("FRE (no args): expected rt.Fre( in output, got:\n%s", out)
	}

	// FRE with arg
	out = generate(t, makeCallPrint("FRE", num))
	if !strings.Contains(out, "rt.Fre(") {
		t.Errorf("FRE (with arg): expected rt.Fre( in output, got:\n%s", out)
	}

	// PEEK with no args
	out = generate(t, makeCallPrint("PEEK"))
	if !strings.Contains(out, "rt.Peek(") {
		t.Errorf("PEEK (no args): expected rt.Peek( in output, got:\n%s", out)
	}

	// PEEK with arg
	out = generate(t, makeCallPrint("PEEK", num))
	if !strings.Contains(out, "rt.Peek(") {
		t.Errorf("PEEK (with arg): expected rt.Peek( in output, got:\n%s", out)
	}
}

func TestEmitFunctionCallEOF(t *testing.T) {
	num := &ast.NumberLiteral{Value: 1, OriginalText: "1"}

	// EOF with arg
	out := generate(t, makeCallPrint("EOF", num))
	if !strings.Contains(out, "fm.Eof(") {
		t.Errorf("EOF (with arg): expected fm.Eof( in output, got:\n%s", out)
	}

	// EOF without arg
	out = generate(t, makeCallPrint("EOF"))
	if !strings.Contains(out, "float64(0)") {
		t.Errorf("EOF (no args): expected float64(0) in output, got:\n%s", out)
	}
}

func TestEmitFunctionCallUserDefined(t *testing.T) {
	num := &ast.NumberLiteral{Value: 5, OriginalText: "5"}

	out := generate(t, makeCallPrint("myFunc", num))
	if !strings.Contains(out, "myFunc(") {
		t.Errorf("user-defined function: expected myFunc( in output, got:\n%s", out)
	}
}

// ===========================================================================
// emitDoLoop — all variants
// ===========================================================================

func TestEmitDoLoopNoCondition(t *testing.T) {
	stmts := []ast.Statement{
		&ast.DoLoopStatement{
			Condition: nil,
			Body: []ast.Statement{
				&ast.ExitStatement{ExitType: "DO"},
			},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "for {") {
		t.Errorf("expected infinite loop 'for {' in output, got:\n%s", out)
	}
}

func TestEmitDoLoopWhileAtTop(t *testing.T) {
	stmts := []ast.Statement{
		&ast.DoLoopStatement{
			Condition: &ast.NumberLiteral{Value: 1, OriginalText: "1"},
			TestAtTop: true,
			IsUntil:   false,
			Body:      []ast.Statement{},
		},
	}
	out := generate(t, stmts)
	// WHILE at top: for <condition> { ... }
	if !strings.Contains(out, "for") {
		t.Errorf("expected 'for' loop in output, got:\n%s", out)
	}
}

func TestEmitDoLoopUntilAtTop(t *testing.T) {
	stmts := []ast.Statement{
		&ast.DoLoopStatement{
			Condition: &ast.NumberLiteral{Value: 0, OriginalText: "0"},
			TestAtTop: true,
			IsUntil:   true,
			Body:      []ast.Statement{},
		},
	}
	out := generate(t, stmts)
	// UNTIL at top: for !(<condition>) { ... }
	if !strings.Contains(out, "!(") {
		t.Errorf("expected negated condition '!(' in output, got:\n%s", out)
	}
}

func TestEmitDoLoopWhileAtBottom(t *testing.T) {
	stmts := []ast.Statement{
		&ast.DoLoopStatement{
			Condition: &ast.NumberLiteral{Value: 1, OriginalText: "1"},
			TestAtTop: false,
			IsUntil:   false,
			Body:      []ast.Statement{},
		},
	}
	out := generate(t, stmts)
	// WHILE at bottom: if !(<condition>) { break }
	if !strings.Contains(out, "break") {
		t.Errorf("expected break in output for WHILE at bottom, got:\n%s", out)
	}
	if !strings.Contains(out, "!(") {
		t.Errorf("expected negated condition '!(' in output, got:\n%s", out)
	}
}

func TestEmitDoLoopUntilAtBottom(t *testing.T) {
	stmts := []ast.Statement{
		&ast.DoLoopStatement{
			Condition: &ast.NumberLiteral{Value: 1, OriginalText: "1"},
			TestAtTop: false,
			IsUntil:   true,
			Body:      []ast.Statement{},
		},
	}
	out := generate(t, stmts)
	// UNTIL at bottom: if <condition> { break }
	if !strings.Contains(out, "break") {
		t.Errorf("expected break in output for UNTIL at bottom, got:\n%s", out)
	}
}

// ===========================================================================
// emitExit — all exit types
// ===========================================================================

func TestEmitExitFor(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ExitStatement{ExitType: "FOR"},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "break") {
		t.Errorf("EXIT FOR: expected 'break' in output, got:\n%s", out)
	}
}

func TestEmitExitDo(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ExitStatement{ExitType: "DO"},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "break") {
		t.Errorf("EXIT DO: expected 'break' in output, got:\n%s", out)
	}
}

func TestEmitExitWhile(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ExitStatement{ExitType: "WHILE"},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "break") {
		t.Errorf("EXIT WHILE: expected 'break' in output, got:\n%s", out)
	}
}

func TestEmitExitLoop(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ExitStatement{ExitType: "LOOP"},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "break") {
		t.Errorf("EXIT LOOP: expected 'break' in output, got:\n%s", out)
	}
}

func TestEmitExitSub(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ExitStatement{ExitType: "SUB"},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "return") {
		t.Errorf("EXIT SUB: expected 'return' in output, got:\n%s", out)
	}
}

func TestEmitExitFunction(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ExitStatement{ExitType: "FUNCTION"},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "return") {
		t.Errorf("EXIT FUNCTION: expected 'return' in output, got:\n%s", out)
	}
}

func TestEmitExitDef(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ExitStatement{ExitType: "DEF"},
	}
	out := generate(t, stmts)
	// Not inDefFn context: should emit "return"
	if !strings.Contains(out, "return") {
		t.Errorf("EXIT DEF (not in defFn): expected 'return' in output, got:\n%s", out)
	}
}

func TestEmitExitUnknown(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ExitStatement{ExitType: "UNKNOWN_TYPE"},
	}
	out := generate(t, stmts)
	// default branch should emit a comment
	if !strings.Contains(out, "EXIT") {
		t.Errorf("EXIT UNKNOWN: expected 'EXIT' in output comment, got:\n%s", out)
	}
}

// ===========================================================================
// emitIncr / emitDecr with Amount
// ===========================================================================

func TestEmitIncrWithAmount(t *testing.T) {
	stmts := []ast.Statement{
		&ast.IncrStatement{
			Variable: &ast.Identifier{Name: "counter"},
			Amount:   &ast.NumberLiteral{Value: 5, OriginalText: "5"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "+= 5") {
		t.Errorf("INCR with amount: expected '+= 5' in output, got:\n%s", out)
	}
}

func TestEmitDecrWithAmount(t *testing.T) {
	stmts := []ast.Statement{
		&ast.DecrStatement{
			Variable: &ast.Identifier{Name: "counter"},
			Amount:   &ast.NumberLiteral{Value: 3, OriginalText: "3"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "-= 3") {
		t.Errorf("DECR with amount: expected '-= 3' in output, got:\n%s", out)
	}
}

// ===========================================================================
// emitInputFromStdin — various patterns
// ===========================================================================

func TestEmitInputFromStdinNoPrompt(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ReadStatement{
			IsInput:   true,
			Prompt:    "", // no prompt
			Variables: []ast.Expression{&ast.Identifier{Name: "x"}},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "scanner_") {
		t.Errorf("INPUT no prompt: expected scanner_ in output, got:\n%s", out)
	}
	// Should NOT have fmt.Print for prompt
	if strings.Contains(out, `fmt.Print("`)  {
		t.Errorf("INPUT no prompt: should NOT have fmt.Print prompt, got:\n%s", out)
	}
}

func TestEmitInputFromStdinStringVar(t *testing.T) {
	// String variable should use scanner_.Text() directly
	stmts := []ast.Statement{
		&ast.ReadStatement{
			IsInput:   true,
			Prompt:    "Enter name: ",
			Variables: []ast.Expression{&ast.Identifier{Name: "name", TypeSuffix: "$"}},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "scanner_.Text()") {
		t.Errorf("INPUT string var: expected scanner_.Text() in output, got:\n%s", out)
	}
}

func TestEmitInputFromStdinMultipleVars(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ReadStatement{
			IsInput: true,
			Prompt:  "Enter a, b: ",
			Variables: []ast.Expression{
				&ast.Identifier{Name: "a"},
				&ast.Identifier{Name: "b"},
			},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "InputSplitLine") {
		t.Errorf("INPUT multiple vars: expected InputSplitLine in output, got:\n%s", out)
	}
	if !strings.Contains(out, "parts_") {
		t.Errorf("INPUT multiple vars: expected parts_ in output, got:\n%s", out)
	}
}

func TestEmitInputFromStdinLineInput(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ReadStatement{
			IsInput:     true,
			IsLineInput: true,
			Prompt:      "",
			Variables:   []ast.Expression{&ast.Identifier{Name: "line", TypeSuffix: "$"}},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "scanner_") {
		t.Errorf("LINE INPUT: expected scanner_ in output, got:\n%s", out)
	}
	if !strings.Contains(out, "scanner_.Text()") {
		t.Errorf("LINE INPUT: expected scanner_.Text() in output, got:\n%s", out)
	}
}

// ===========================================================================
// emitFileInput — variants
// ===========================================================================

func TestEmitFileInputLineInput(t *testing.T) {
	stmts := []ast.Statement{
		&ast.FileInputStatement{
			FileNum:     &ast.NumberLiteral{Value: 1, OriginalText: "1"},
			IsLineInput: true,
			Variables:   []ast.Expression{&ast.Identifier{Name: "line", TypeSuffix: "$"}},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "FileLineInput") {
		t.Errorf("FILE LINE INPUT: expected FileLineInput in output, got:\n%s", out)
	}
}

func TestEmitFileInputRegularString(t *testing.T) {
	stmts := []ast.Statement{
		&ast.FileInputStatement{
			FileNum:     &ast.NumberLiteral{Value: 1, OriginalText: "1"},
			IsLineInput: false,
			Variables:   []ast.Expression{&ast.Identifier{Name: "s", TypeSuffix: "$"}},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "FileInput") {
		t.Errorf("FILE INPUT string: expected FileInput in output, got:\n%s", out)
	}
}

func TestEmitFileInputRegularNumeric(t *testing.T) {
	stmts := []ast.Statement{
		&ast.FileInputStatement{
			FileNum:     &ast.NumberLiteral{Value: 2, OriginalText: "2"},
			IsLineInput: false,
			Variables:   []ast.Expression{&ast.Identifier{Name: "n"}},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "FileInput") {
		t.Errorf("FILE INPUT numeric: expected FileInput in output, got:\n%s", out)
	}
	if !strings.Contains(out, "ParseFloat") {
		t.Errorf("FILE INPUT numeric: expected ParseFloat in output, got:\n%s", out)
	}
}

func TestEmitFileInputLineInputNumeric(t *testing.T) {
	stmts := []ast.Statement{
		&ast.FileInputStatement{
			FileNum:     &ast.NumberLiteral{Value: 1, OriginalText: "1"},
			IsLineInput: true,
			Variables:   []ast.Expression{&ast.Identifier{Name: "n"}},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "FileLineInput") {
		t.Errorf("FILE LINE INPUT numeric: expected FileLineInput in output, got:\n%s", out)
	}
	if !strings.Contains(out, "ParseFloat") {
		t.Errorf("FILE LINE INPUT numeric: expected ParseFloat in output, got:\n%s", out)
	}
}

// ===========================================================================
// emitView — viewport variant
// ===========================================================================

func TestEmitViewport(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ViewStatement{
			IsPrint: false,
			X1:      &ast.NumberLiteral{Value: 0, OriginalText: "0"},
			Y1:      &ast.NumberLiteral{Value: 0, OriginalText: "0"},
			X2:      &ast.NumberLiteral{Value: 319, OriginalText: "319"},
			Y2:      &ast.NumberLiteral{Value: 199, OriginalText: "199"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "ViewPort") {
		t.Errorf("VIEW (viewport): expected 'ViewPort' in output, got:\n%s", out)
	}
}

func TestEmitViewportWithColors(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ViewStatement{
			IsPrint:     false,
			X1:          &ast.NumberLiteral{Value: 10, OriginalText: "10"},
			Y1:          &ast.NumberLiteral{Value: 10, OriginalText: "10"},
			X2:          &ast.NumberLiteral{Value: 200, OriginalText: "200"},
			Y2:          &ast.NumberLiteral{Value: 150, OriginalText: "150"},
			FillColor:   &ast.NumberLiteral{Value: 1, OriginalText: "1"},
			BorderColor: &ast.NumberLiteral{Value: 14, OriginalText: "14"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "ViewPort") {
		t.Errorf("VIEW with colors: expected 'ViewPort' in output, got:\n%s", out)
	}
}

func TestEmitViewportNoCoords(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ViewStatement{
			IsPrint: false,
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "ViewPort") {
		t.Errorf("VIEW no coords: expected 'ViewPort' in output, got:\n%s", out)
	}
}

// ===========================================================================
// emitCircle — with optional Start / End / Aspect
// ===========================================================================

func TestEmitCircleWithStartEnd(t *testing.T) {
	stmts := []ast.Statement{
		&ast.CircleStmt{
			X:      &ast.NumberLiteral{Value: 100, OriginalText: "100"},
			Y:      &ast.NumberLiteral{Value: 100, OriginalText: "100"},
			Radius: &ast.NumberLiteral{Value: 50, OriginalText: "50"},
			Color:  &ast.NumberLiteral{Value: 4, OriginalText: "4"},
			Start:  &ast.NumberLiteral{Value: 0, OriginalText: "0"},
			End:    &ast.NumberLiteral{Value: 6, OriginalText: "6"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "Circle") {
		t.Errorf("CIRCLE with Start/End: expected 'Circle' in output, got:\n%s", out)
	}
}

func TestEmitCircleWithAspect(t *testing.T) {
	stmts := []ast.Statement{
		&ast.CircleStmt{
			X:      &ast.NumberLiteral{Value: 160, OriginalText: "160"},
			Y:      &ast.NumberLiteral{Value: 100, OriginalText: "100"},
			Radius: &ast.NumberLiteral{Value: 50, OriginalText: "50"},
			Aspect: &ast.NumberLiteral{Value: 2, OriginalText: "2"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "Circle") {
		t.Errorf("CIRCLE with Aspect: expected 'Circle' in output, got:\n%s", out)
	}
}

// ===========================================================================
// emitFilePrint — no expressions
// ===========================================================================

func TestEmitFilePrintNoExpressions(t *testing.T) {
	stmts := []ast.Statement{
		&ast.FilePrintStatement{
			FileNum:     &ast.NumberLiteral{Value: 1, OriginalText: "1"},
			Expressions: []ast.Expression{},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "FilePrint(") {
		t.Errorf("FILE PRINT no exprs: expected 'FilePrint(' in output, got:\n%s", out)
	}
}

// ===========================================================================
// emitFileWrite — no expressions
// ===========================================================================

func TestEmitFileWriteNoExpressions(t *testing.T) {
	stmts := []ast.Statement{
		&ast.FileWriteStatement{
			FileNum:     &ast.NumberLiteral{Value: 1, OriginalText: "1"},
			Expressions: []ast.Expression{},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "FileWrite(") {
		t.Errorf("FILE WRITE no exprs: expected 'FileWrite(' in output, got:\n%s", out)
	}
}

// ===========================================================================
// emitRedim — multi-dimensional
// ===========================================================================

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

// ===========================================================================
// emitPrint — comma separators and multiple expressions
// ===========================================================================

func TestEmitPrintComma(t *testing.T) {
	// Comma separator triggers zone printing
	stmts := []ast.Statement{
		&ast.PrintStatement{
			Expressions: []ast.Expression{
				&ast.NumberLiteral{Value: 1, OriginalText: "1"},
				&ast.NumberLiteral{Value: 2, OriginalText: "2"},
			},
			Separators: []string{",", ""},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "PrintZone") {
		t.Errorf("PRINT with comma: expected 'PrintZone' in output, got:\n%s", out)
	}
}

func TestEmitPrintCommaTrailingSep(t *testing.T) {
	// Comma separator with trailing sep (no newline at end)
	stmts := []ast.Statement{
		&ast.PrintStatement{
			Expressions: []ast.Expression{
				&ast.NumberLiteral{Value: 42, OriginalText: "42"},
			},
			Separators:     []string{","},
			HasTrailingSep: true,
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "PrintZone") {
		t.Errorf("PRINT comma trailing sep: expected 'PrintZone' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "fmt.Print(out_)") {
		t.Errorf("PRINT comma trailing sep: expected 'fmt.Print(out_)' in output, got:\n%s", out)
	}
}

func TestEmitPrintMultipleNoSep(t *testing.T) {
	// Multiple expressions without trailing sep → fmt.Println
	stmts := []ast.Statement{
		&ast.PrintStatement{
			Expressions: []ast.Expression{
				&ast.NumberLiteral{Value: 1, OriginalText: "1"},
				&ast.NumberLiteral{Value: 2, OriginalText: "2"},
			},
			Separators:     []string{";", ""},
			HasTrailingSep: false,
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "fmt.Println(") {
		t.Errorf("PRINT multiple no trailing sep: expected 'fmt.Println(' in output, got:\n%s", out)
	}
}

func TestEmitPrintMultipleTrailingSep(t *testing.T) {
	// Multiple expressions with trailing sep → fmt.Print
	stmts := []ast.Statement{
		&ast.PrintStatement{
			Expressions: []ast.Expression{
				&ast.NumberLiteral{Value: 1, OriginalText: "1"},
				&ast.NumberLiteral{Value: 2, OriginalText: "2"},
			},
			Separators:     []string{";", ";"},
			HasTrailingSep: true,
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "fmt.Print(") {
		t.Errorf("PRINT multiple trailing sep: expected 'fmt.Print(' in output, got:\n%s", out)
	}
}

// ===========================================================================
// emitSelectCase — IS comparison and range cases
// ===========================================================================

func TestEmitSelectCaseIsComparison(t *testing.T) {
	stmts := []ast.Statement{
		&ast.SelectCaseStatement{
			TestExpr: &ast.NumberLiteral{Value: 5, OriginalText: "5"},
			Cases: []ast.CaseClause{
				{
					Values: []ast.CaseValue{
						{
							Value:        &ast.NumberLiteral{Value: 3, OriginalText: "3"},
							IsComparison: true,
							Comparison:   ">",
						},
					},
					Body: []ast.Statement{
						&ast.PrintStatement{
							Expressions: []ast.Expression{&ast.StringLiteral{Value: "gt3"}},
							Separators:  []string{""},
						},
					},
				},
			},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "switch {") {
		t.Errorf("SELECT CASE IS: expected 'switch {' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "> 3") {
		t.Errorf("SELECT CASE IS: expected '> 3' in output, got:\n%s", out)
	}
}

func TestEmitSelectCaseRange(t *testing.T) {
	stmts := []ast.Statement{
		&ast.SelectCaseStatement{
			TestExpr: &ast.NumberLiteral{Value: 5, OriginalText: "5"},
			Cases: []ast.CaseClause{
				{
					Values: []ast.CaseValue{
						{
							Value:    &ast.NumberLiteral{Value: 1, OriginalText: "1"},
							EndValue: &ast.NumberLiteral{Value: 10, OriginalText: "10"},
							IsRange:  true,
						},
					},
					Body: []ast.Statement{},
				},
			},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, ">=") {
		t.Errorf("SELECT CASE range: expected '>=' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "<=") {
		t.Errorf("SELECT CASE range: expected '<=' in output, got:\n%s", out)
	}
}

func TestEmitSelectCaseElseBlock(t *testing.T) {
	stmts := []ast.Statement{
		&ast.SelectCaseStatement{
			TestExpr: &ast.NumberLiteral{Value: 99, OriginalText: "99"},
			Cases:    []ast.CaseClause{},
			ElseBlock: []ast.Statement{
				&ast.PrintStatement{
					Expressions: []ast.Expression{&ast.StringLiteral{Value: "else"}},
					Separators:  []string{""},
				},
			},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "default:") {
		t.Errorf("SELECT CASE else: expected 'default:' in output, got:\n%s", out)
	}
}

// ===========================================================================
// collectMainVariables — indirect coverage via GOTO hoisting
// ===========================================================================

func TestCollectMainVariablesWithGoto(t *testing.T) {
	// A GOTO forces collectMainVariables to run.
	// Include ForStatement, DimStatement, ReadStatement (IsInput), WhileStatement,
	// DoLoopStatement, SelectCaseStatement, IfStatement all in one program.
	stmts := []ast.Statement{
		&ast.GotoStatement{Target: "99"},
		&ast.LineNumberStatement{Number: 10},
		&ast.ForStatement{
			Counter: &ast.Identifier{Name: "i"},
			Start:   &ast.NumberLiteral{Value: 1},
			End:     &ast.NumberLiteral{Value: 5},
			Body: []ast.Statement{
				&ast.LetStatement{
					Name:  &ast.Identifier{Name: "inner"},
					Value: &ast.NumberLiteral{Value: 0},
				},
			},
		},
		&ast.DimStatement{
			Declarations: []ast.DimDecl{
				{Name: "scalar"},
			},
		},
		&ast.DimStatement{
			Declarations: []ast.DimDecl{
				{
					Name: "arr1d",
					Dimensions: []ast.DimRange{
						{Upper: &ast.NumberLiteral{Value: 10}},
					},
				},
			},
		},
		&ast.DimStatement{
			Declarations: []ast.DimDecl{
				{
					Name: "arr2d",
					Dimensions: []ast.DimRange{
						{Upper: &ast.NumberLiteral{Value: 5}},
						{Upper: &ast.NumberLiteral{Value: 5}},
					},
				},
			},
		},
		&ast.ReadStatement{
			IsInput: true,
			Variables: []ast.Expression{
				&ast.Identifier{Name: "inputVar"},
			},
		},
		&ast.WhileStatement{
			Condition: &ast.NumberLiteral{Value: 0},
			Body: []ast.Statement{
				&ast.LetStatement{
					Name:  &ast.Identifier{Name: "whileInner"},
					Value: &ast.NumberLiteral{Value: 0},
				},
			},
		},
		&ast.DoLoopStatement{
			Condition: &ast.NumberLiteral{Value: 0},
			Body: []ast.Statement{
				&ast.LetStatement{
					Name:  &ast.Identifier{Name: "doInner"},
					Value: &ast.NumberLiteral{Value: 0},
				},
			},
		},
		&ast.SelectCaseStatement{
			TestExpr: &ast.NumberLiteral{Value: 1},
			Cases: []ast.CaseClause{
				{
					Values: []ast.CaseValue{{Value: &ast.NumberLiteral{Value: 1}}},
					Body: []ast.Statement{
						&ast.LetStatement{
							Name:  &ast.Identifier{Name: "caseVar"},
							Value: &ast.NumberLiteral{Value: 0},
						},
					},
				},
			},
		},
		&ast.IfStatement{
			Condition: &ast.NumberLiteral{Value: 1},
			ThenBlock: []ast.Statement{
				&ast.LetStatement{
					Name:  &ast.Identifier{Name: "thenVar"},
					Value: &ast.NumberLiteral{Value: 0},
				},
			},
			ElseIfClauses: []ast.ElseIfClause{
				{
					Condition: &ast.NumberLiteral{Value: 0},
					Body: []ast.Statement{
						&ast.LetStatement{
							Name:  &ast.Identifier{Name: "elseIfVar"},
							Value: &ast.NumberLiteral{Value: 0},
						},
					},
				},
			},
			ElseBlock: []ast.Statement{
				&ast.LetStatement{
					Name:  &ast.Identifier{Name: "elseVar"},
					Value: &ast.NumberLiteral{Value: 0},
				},
			},
		},
		&ast.LineNumberStatement{Number: 99},
		&ast.PrintStatement{
			Expressions: []ast.Expression{&ast.StringLiteral{Value: "done"}},
			Separators:  []string{""},
		},
	}

	out := generate(t, stmts)
	// The program should compile; hoisting should have occurred
	if !strings.Contains(out, "// Hoisted variable declarations") {
		t.Errorf("Expected hoisted variable declarations comment in output:\n%s", out)
	}
}

func TestCollectMainVariablesWithDefFn(t *testing.T) {
	// DefFnDeclaration inside a GOTO program also exercises
	// the DefFnDeclaration branch of collectMainVariables.
	stmts := []ast.Statement{
		&ast.GotoStatement{Target: "end"},
		&ast.DefFnDeclaration{
			Name: "FNtest",
			Params: []ast.Parameter{
				{Name: "x"},
			},
			Body: []ast.Statement{
				&ast.LetStatement{
					Name:  &ast.Identifier{Name: "FNtest"},
					Value: &ast.Identifier{Name: "x"},
				},
			},
		},
		&ast.LabelStatement{Name: "end"},
	}

	out := generate(t, stmts)
	if !strings.Contains(out, "FNtest") || !strings.Contains(out, "func") {
		t.Errorf("Expected FNtest function in output, got:\n%s", out)
	}
}

// ===========================================================================
// emitStatement — uncovered branches
// ===========================================================================

func TestEmitStatementEnd(t *testing.T) {
	stmts := []ast.Statement{&ast.EndStatement{}}
	out := generate(t, stmts)
	if !strings.Contains(out, "os.Exit(0)") {
		t.Errorf("END: expected os.Exit(0) in output, got:\n%s", out)
	}
}

func TestEmitStatementStop(t *testing.T) {
	stmts := []ast.Statement{&ast.StopStatement{}}
	out := generate(t, stmts)
	if !strings.Contains(out, "os.Exit(0)") {
		t.Errorf("STOP: expected os.Exit(0) in output, got:\n%s", out)
	}
}

func TestEmitStatementSystem(t *testing.T) {
	stmts := []ast.Statement{&ast.SystemStatement{}}
	out := generate(t, stmts)
	if !strings.Contains(out, "os.Exit(0)") {
		t.Errorf("SYSTEM: expected os.Exit(0) in output, got:\n%s", out)
	}
}

func TestEmitStatementBeep(t *testing.T) {
	stmts := []ast.Statement{&ast.BeepStatement{}}
	out := generate(t, stmts)
	if !strings.Contains(out, "BEEP") {
		t.Errorf("BEEP: expected BEEP in output, got:\n%s", out)
	}
}

func TestEmitStatementCls(t *testing.T) {
	stmts := []ast.Statement{&ast.ClsStatement{}}
	out := generate(t, stmts)
	if !strings.Contains(out, "AnsiCls") {
		t.Errorf("CLS: expected AnsiCls in output, got:\n%s", out)
	}
}

func TestEmitStatementData(t *testing.T) {
	stmts := []ast.Statement{
		&ast.DataStatement{
			Values: []ast.Expression{
				&ast.NumberLiteral{Value: 1, OriginalText: "1"},
			},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "DATA") {
		t.Errorf("DATA: expected 'DATA' comment in output, got:\n%s", out)
	}
}

func TestEmitStatementOptionBase(t *testing.T) {
	stmts := []ast.Statement{
		&ast.OptionBaseStatement{Value: 1},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "OPTION BASE") {
		t.Errorf("OPTION BASE: expected 'OPTION BASE' comment in output, got:\n%s", out)
	}
}

func TestEmitStatementDefType(t *testing.T) {
	stmts := []ast.Statement{
		&ast.DefTypeStatement{Type: "DEFINT"},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "DEFINT") {
		t.Errorf("DEFTYPE: expected 'DEFINT' in output, got:\n%s", out)
	}
}

func TestEmitStatementScope(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ScopeStatement{
			Modifier:  "SHARED",
			Variables: []string{"x", "y"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "SHARED") {
		t.Errorf("SCOPE: expected 'SHARED' in output, got:\n%s", out)
	}
}

func TestEmitStatementPoke(t *testing.T) {
	stmts := []ast.Statement{
		&ast.PokeStatement{
			Address: &ast.NumberLiteral{Value: 1000, OriginalText: "1000"},
			Value:   &ast.NumberLiteral{Value: 42, OriginalText: "42"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "POKE") {
		t.Errorf("POKE: expected 'POKE' comment in output, got:\n%s", out)
	}
}

func TestEmitStatementConst(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ConstStatement{
			Name:  "PI",
			Value: &ast.NumberLiteral{Value: 3, OriginalText: "3"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "CONST") {
		t.Errorf("CONST: expected 'CONST' in output, got:\n%s", out)
	}
}

func TestEmitStatementClear(t *testing.T) {
	stmts := []ast.Statement{&ast.ClearStatement{}}
	out := generate(t, stmts)
	if !strings.Contains(out, "CLEAR") {
		t.Errorf("CLEAR: expected 'CLEAR' in output, got:\n%s", out)
	}
}

func TestEmitStatementDraw(t *testing.T) {
	stmts := []ast.Statement{
		&ast.DrawStmt{
			CommandString: &ast.StringLiteral{Value: "BM100,100"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "rt.Draw(") {
		t.Errorf("DRAW: expected 'rt.Draw(' in output, got:\n%s", out)
	}
}

func TestEmitStatementScreen(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ScreenStatement{
			Mode: &ast.NumberLiteral{Value: 12, OriginalText: "12"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "ScreenMode") {
		t.Errorf("SCREEN: expected 'ScreenMode' in output, got:\n%s", out)
	}
}

func TestEmitStatementOnErrorGoto(t *testing.T) {
	stmts := []ast.Statement{
		&ast.OnErrorGotoStatement{Target: "errHandler"},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "errState.SetHandler") {
		t.Errorf("ON ERROR GOTO: expected 'errState.SetHandler' in output, got:\n%s", out)
	}
}

func TestEmitStatementOnErrorGotoDisable(t *testing.T) {
	stmts := []ast.Statement{
		&ast.OnErrorGotoStatement{Target: "0"},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "ON ERROR GOTO 0") {
		t.Errorf("ON ERROR GOTO 0: expected 'ON ERROR GOTO 0' comment in output, got:\n%s", out)
	}
}

func TestEmitStatementPlay(t *testing.T) {
	stmts := []ast.Statement{
		&ast.PlayStatement{
			CommandString: &ast.StringLiteral{Value: "T2L4C"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "rt.Play(") {
		t.Errorf("PLAY: expected 'rt.Play(' in output, got:\n%s", out)
	}
}

func TestEmitStatementSound(t *testing.T) {
	stmts := []ast.Statement{
		&ast.SoundStatement{
			Frequency: &ast.NumberLiteral{Value: 440, OriginalText: "440"},
			Duration:  &ast.NumberLiteral{Value: 18, OriginalText: "18"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "rt.Sound(") {
		t.Errorf("SOUND: expected 'rt.Sound(' in output, got:\n%s", out)
	}
}

func TestEmitStatementError(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ErrorStatement{
			Code: &ast.NumberLiteral{Value: 5, OriginalText: "5"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "TriggerError") {
		t.Errorf("ERROR: expected 'TriggerError' in output, got:\n%s", out)
	}
}

func TestEmitStatementResume(t *testing.T) {
	// RESUME NEXT
	stmts := []ast.Statement{
		&ast.ResumeStatement{Type: "NEXT"},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "RESUME NEXT") {
		t.Errorf("RESUME NEXT: expected 'RESUME NEXT' in output, got:\n%s", out)
	}
}

func TestEmitStatementResumeEmpty(t *testing.T) {
	// RESUME (empty type)
	stmts := []ast.Statement{
		&ast.ResumeStatement{Type: ""},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "RESUME") {
		t.Errorf("RESUME: expected 'RESUME' in output, got:\n%s", out)
	}
}

func TestEmitStatementResumeLabel(t *testing.T) {
	// RESUME <label>
	stmts := []ast.Statement{
		&ast.ResumeStatement{Type: "recoverPoint"},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "goto") {
		t.Errorf("RESUME label: expected 'goto' in output, got:\n%s", out)
	}
}

func TestEmitStatementFieldAssign(t *testing.T) {
	stmts := []ast.Statement{
		&ast.FieldAssignStatement{
			Object: &ast.Identifier{Name: "obj"},
			Field:  "x",
			Value:  &ast.NumberLiteral{Value: 42, OriginalText: "42"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "obj.x") {
		t.Errorf("FIELD ASSIGN: expected 'obj.x' in output, got:\n%s", out)
	}
}

func TestEmitStatementFnAssign(t *testing.T) {
	// FnAssignStatement outside defFn context
	stmts := []ast.Statement{
		&ast.FnAssignStatement{
			Name:  "FNtest",
			Value: &ast.NumberLiteral{Value: 99, OriginalText: "99"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "_fn_") {
		t.Errorf("FN ASSIGN: expected '_fn_' in output, got:\n%s", out)
	}
}

func TestEmitStatementGet(t *testing.T) {
	stmts := []ast.Statement{
		&ast.GetStatement{
			FileNum: &ast.NumberLiteral{Value: 1, OriginalText: "1"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "RandomGet") {
		t.Errorf("GET: expected 'RandomGet' in output, got:\n%s", out)
	}
}

func TestEmitStatementPut(t *testing.T) {
	stmts := []ast.Statement{
		&ast.PutStatement{
			FileNum: &ast.NumberLiteral{Value: 1, OriginalText: "1"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "RandomPut") {
		t.Errorf("PUT: expected 'RandomPut' in output, got:\n%s", out)
	}
}

func TestEmitStatementSeek(t *testing.T) {
	stmts := []ast.Statement{
		&ast.SeekStatement{
			FileNum:  &ast.NumberLiteral{Value: 1, OriginalText: "1"},
			Position: &ast.NumberLiteral{Value: 100, OriginalText: "100"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "FileSeek") {
		t.Errorf("SEEK: expected 'FileSeek' in output, got:\n%s", out)
	}
}

func TestEmitStatementField(t *testing.T) {
	stmts := []ast.Statement{
		&ast.FieldStatement{
			FileNum: &ast.NumberLiteral{Value: 1, OriginalText: "1"},
			Fields: []ast.FieldDef{
				{VarName: "name$", Length: &ast.NumberLiteral{Value: 20, OriginalText: "20"}},
			},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "fm.Field(") {
		t.Errorf("FIELD: expected 'fm.Field(' in output, got:\n%s", out)
	}
}

func TestEmitStatementLset(t *testing.T) {
	stmts := []ast.Statement{
		&ast.LsetStatement{
			Variable: "field1",
			Value:    &ast.StringLiteral{Value: "hello"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "Lset") {
		t.Errorf("LSET: expected 'Lset' in output, got:\n%s", out)
	}
}

// ===========================================================================
// emitExpr — GroupExpr and FieldAccessExpression
// ===========================================================================

func TestEmitExprGroupExpr(t *testing.T) {
	stmts := []ast.Statement{
		&ast.PrintStatement{
			Expressions: []ast.Expression{
				&ast.GroupExpr{
					Inner: &ast.NumberLiteral{Value: 42, OriginalText: "42"},
				},
			},
			Separators: []string{""},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "(42)") {
		t.Errorf("GroupExpr: expected '(42)' in output, got:\n%s", out)
	}
}

func TestEmitExprFieldAccessExpression(t *testing.T) {
	stmts := []ast.Statement{
		&ast.PrintStatement{
			Expressions: []ast.Expression{
				&ast.FieldAccessExpression{
					Object: &ast.Identifier{Name: "myStruct"},
					Field:  "value",
				},
			},
			Separators: []string{""},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "myStruct.value") {
		t.Errorf("FieldAccessExpression: expected 'myStruct.value' in output, got:\n%s", out)
	}
}

func TestEmitExprNil(t *testing.T) {
	// nil expression returns "0"
	gen := New()
	got := gen.emitExpr(nil)
	if got != "0" {
		t.Errorf("emitExpr(nil) = %q, want 0", got)
	}
}

// ===========================================================================
// toBoolExpr — various expression strings
// ===========================================================================

func TestToBoolExprAlreadyBool(t *testing.T) {
	gen := New()

	cases := []struct {
		expr string
		want string
	}{
		{"(x == 0)", "(x == 0)"},     // contains ==
		{"(x != y)", "(x != y)"},     // contains !=
		{"(a <= b)", "(a <= b)"},     // contains <=
		{"(a >= b)", "(a >= b)"},     // contains >=
		{"(p && q)", "(p && q)"},     // contains &&
		{"(p || q)", "(p || q)"},     // contains ||
		{"!(x)", "!(x)"},             // contains !(
		{"true", "true"},             // contains true
		{"false", "false"},           // contains false
		{"(a < b)", "(a < b)"},       // contains <
		{"(a > b)", "(a > b)"},       // contains >
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

// ===========================================================================
// emitPrintUsing with HasTrailingSep
// ===========================================================================

func TestEmitPrintUsingTrailingSep(t *testing.T) {
	stmts := []ast.Statement{
		&ast.PrintStatement{
			Format: &ast.StringLiteral{Value: "##.##"},
			Expressions: []ast.Expression{
				&ast.NumberLiteral{Value: 3, OriginalText: "3"},
			},
			Separators:     []string{""},
			HasTrailingSep: true, // suppress the extra newline
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "PrintUsing") {
		t.Errorf("PRINT USING trailing sep: expected 'PrintUsing' in output, got:\n%s", out)
	}
	// Should NOT have the extra fmt.Println() call
	if strings.Count(out, "fmt.Println()") > 0 && strings.Contains(out, "PrintUsing") {
		// only the extra newline should be absent
		if strings.Contains(out, "fmt.Println()\n\tfmt.Println()") {
			t.Errorf("PRINT USING trailing sep: should not have extra fmt.Println(), got:\n%s", out)
		}
	}
}

