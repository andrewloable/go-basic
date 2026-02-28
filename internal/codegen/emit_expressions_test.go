package codegen

import (
	"strings"
	"testing"

	"github.com/loabletech/go-basic/internal/ast"
	"github.com/loabletech/go-basic/internal/semantic"
)

// ---------------------------------------------------------------------------
// Tests for emit_expressions.go — binary ops, unary ops, function calls,
// array access, field access, etc.
// ---------------------------------------------------------------------------

// makeCallPrint is a local helper to build PRINT <funcCall(args...)> statements.
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

func TestFunctionCallMappings(t *testing.T) {
	stmts := []ast.Statement{
		&ast.PrintStatement{
			Expressions: []ast.Expression{
				&ast.FunctionCall{
					Name: "ABS",
					Args: []ast.Expression{&ast.NumberLiteral{Value: -5, NumType: ast.NumInt}},
				},
			},
			Separators: []string{""},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "rt.Abs(") {
		t.Errorf("expected rt.Abs(, got:\n%s", out)
	}
}

// TestDollarSuffixedBuiltins verifies that CHR$, LEFT$, RIGHT$, MID$, HEX$,
// OCT$, BIN$, STR$, STRING$, ENVIRON$, UCASE$, LCASE$, LTRIM$, RTRIM$, and
// SPACE$ are emitted as runtime function calls (rt.*) and NOT as array accesses.
func TestDollarSuffixedBuiltins(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
		args     []ast.Expression
		want     string
	}{
		{
			name:     "CHR$",
			funcName: "CHR$",
			args:     []ast.Expression{&ast.NumberLiteral{Value: 65, NumType: ast.NumInt}},
			want:     "rt.Chr(",
		},
		{
			name:     "LEFT$",
			funcName: "LEFT$",
			args: []ast.Expression{
				&ast.StringLiteral{Value: "hello"},
				&ast.NumberLiteral{Value: 3, NumType: ast.NumInt},
			},
			want: "rt.Left(",
		},
		{
			name:     "RIGHT$",
			funcName: "RIGHT$",
			args: []ast.Expression{
				&ast.StringLiteral{Value: "hello"},
				&ast.NumberLiteral{Value: 3, NumType: ast.NumInt},
			},
			want: "rt.Right(",
		},
		{
			name:     "MID$",
			funcName: "MID$",
			args: []ast.Expression{
				&ast.StringLiteral{Value: "hello"},
				&ast.NumberLiteral{Value: 2, NumType: ast.NumInt},
				&ast.NumberLiteral{Value: 3, NumType: ast.NumInt},
			},
			want: "rt.Mid(",
		},
		{
			name:     "HEX$",
			funcName: "HEX$",
			args:     []ast.Expression{&ast.NumberLiteral{Value: 255, NumType: ast.NumInt}},
			want:     "rt.Hex(",
		},
		{
			name:     "OCT$",
			funcName: "OCT$",
			args:     []ast.Expression{&ast.NumberLiteral{Value: 8, NumType: ast.NumInt}},
			want:     "rt.Oct(",
		},
		{
			name:     "BIN$",
			funcName: "BIN$",
			args:     []ast.Expression{&ast.NumberLiteral{Value: 5, NumType: ast.NumInt}},
			want:     "rt.Bin(",
		},
		{
			name:     "STR$",
			funcName: "STR$",
			args:     []ast.Expression{&ast.NumberLiteral{Value: 42, NumType: ast.NumInt}},
			want:     "rt.Str(",
		},
		{
			name:     "STRING$",
			funcName: "STRING$",
			args: []ast.Expression{
				&ast.NumberLiteral{Value: 3, NumType: ast.NumInt},
				&ast.NumberLiteral{Value: 42, NumType: ast.NumInt},
			},
			want: "rt.StringRepeat(",
		},
		{
			name:     "ENVIRON$",
			funcName: "ENVIRON$",
			args:     []ast.Expression{&ast.StringLiteral{Value: "PATH"}},
			want:     "rt.EnvironGet(",
		},
		{
			name:     "UCASE$",
			funcName: "UCASE$",
			args:     []ast.Expression{&ast.StringLiteral{Value: "hello"}},
			want:     "rt.UCase(",
		},
		{
			name:     "LCASE$",
			funcName: "LCASE$",
			args:     []ast.Expression{&ast.StringLiteral{Value: "HELLO"}},
			want:     "rt.LCase(",
		},
		{
			name:     "LTRIM$",
			funcName: "LTRIM$",
			args:     []ast.Expression{&ast.StringLiteral{Value: "  hi"}},
			want:     "rt.LTrim(",
		},
		{
			name:     "RTRIM$",
			funcName: "RTRIM$",
			args:     []ast.Expression{&ast.StringLiteral{Value: "hi  "}},
			want:     "rt.RTrim(",
		},
		{
			name:     "SPACE$",
			funcName: "SPACE$",
			args:     []ast.Expression{&ast.NumberLiteral{Value: 5, NumType: ast.NumInt}},
			want:     "rt.Space(",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stmts := []ast.Statement{
				&ast.PrintStatement{
					Expressions: []ast.Expression{
						&ast.FunctionCall{
							Name: tt.funcName,
							Args: tt.args,
						},
					},
					Separators: []string{""},
				},
			}
			out := generate(t, stmts)
			if !strings.Contains(out, tt.want) {
				t.Errorf("expected %q in output\nGot:\n%s", tt.want, out)
			}
			// Ensure it is NOT an array access (e.g. CHR_str[...]).
			badPattern := strings.ToUpper(strings.TrimSuffix(tt.funcName, "$")) + "_str["
			if strings.Contains(out, badPattern) {
				t.Errorf("found array access pattern %q — builtin emitted as array access instead of function call\nGot:\n%s", badPattern, out)
			}
		})
	}
}

func TestBinaryExprPower(t *testing.T) {
	stmts := []ast.Statement{
		&ast.PrintStatement{
			Expressions: []ast.Expression{
				&ast.BinaryExpr{
					Left:     &ast.NumberLiteral{Value: 2, NumType: ast.NumInt},
					Operator: "^",
					Right:    &ast.NumberLiteral{Value: 3, NumType: ast.NumInt},
				},
			},
			Separators: []string{""},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "math.Pow(") {
		t.Errorf("expected math.Pow(, got:\n%s", out)
	}
}

func TestBuiltinInkeyDollar(t *testing.T) {
	stmts := []ast.Statement{
		&ast.PrintStatement{
			Expressions: []ast.Expression{
				&ast.Identifier{Name: "INKEY", TypeSuffix: "$"},
			},
			Separators: []string{""},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "rt.Inkey()") {
		t.Errorf("expected rt.Inkey(), got:\n%s", out)
	}
}

func TestBuiltinDateDollar(t *testing.T) {
	stmts := []ast.Statement{
		&ast.PrintStatement{
			Expressions: []ast.Expression{
				&ast.Identifier{Name: "DATE", TypeSuffix: "$"},
			},
			Separators: []string{""},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "rt.DateStr()") {
		t.Errorf("expected rt.DateStr(), got:\n%s", out)
	}
}

func TestBuiltinTimeDollar(t *testing.T) {
	stmts := []ast.Statement{
		&ast.PrintStatement{
			Expressions: []ast.Expression{
				&ast.Identifier{Name: "TIME", TypeSuffix: "$"},
			},
			Separators: []string{""},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "rt.TimeStr()") {
		t.Errorf("expected rt.TimeStr(), got:\n%s", out)
	}
}

func TestBuiltinRndIdentifier(t *testing.T) {
	stmts := []ast.Statement{
		&ast.LetStatement{
			Name:  &ast.Identifier{Name: "x"},
			Value: &ast.Identifier{Name: "RND"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "rng.Rnd(1)") {
		t.Errorf("expected rng.Rnd(1), got:\n%s", out)
	}
	if !strings.Contains(out, "rng := rt.NewRNG()") {
		t.Errorf("expected rng preamble, got:\n%s", out)
	}
}

func TestBuiltinTimerIdentifier(t *testing.T) {
	stmts := []ast.Statement{
		&ast.LetStatement{
			Name:  &ast.Identifier{Name: "x"},
			Value: &ast.Identifier{Name: "TIMER"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "rt.Timer()") {
		t.Errorf("expected rt.Timer(), got:\n%s", out)
	}
}

func TestBuiltinErrIdentifier(t *testing.T) {
	stmts := []ast.Statement{
		&ast.PrintStatement{
			Expressions: []ast.Expression{
				&ast.Identifier{Name: "ERR"},
			},
			Separators: []string{""},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "float64(errState.Err())") {
		t.Errorf("expected float64(errState.Err()), got:\n%s", out)
	}
	if !strings.Contains(out, "errState := rt.NewErrorState()") {
		t.Errorf("expected errState preamble, got:\n%s", out)
	}
}

func TestBuiltinErlIdentifier(t *testing.T) {
	stmts := []ast.Statement{
		&ast.PrintStatement{
			Expressions: []ast.Expression{
				&ast.Identifier{Name: "ERL"},
			},
			Separators: []string{""},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "float64(errState.Erl())") {
		t.Errorf("expected float64(errState.Erl()), got:\n%s", out)
	}
	if !strings.Contains(out, "errState := rt.NewErrorState()") {
		t.Errorf("expected errState preamble, got:\n%s", out)
	}
}

func TestBuiltinEradrIdentifier(t *testing.T) {
	stmts := []ast.Statement{
		&ast.LetStatement{
			Name:  &ast.Identifier{Name: "x"},
			Value: &ast.Identifier{Name: "ERADR"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "float64(0)") {
		t.Errorf("expected float64(0) for ERADR, got:\n%s", out)
	}
}

func TestFrePeekEofEmission(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
		args     []ast.Expression
		want     string
	}{
		{
			name:     "FRE with arg",
			funcName: "FRE",
			args:     []ast.Expression{&ast.NumberLiteral{Value: 0}},
			want:     "rt.Fre(",
		},
		{
			name:     "FRE no args",
			funcName: "FRE",
			args:     nil,
			want:     "rt.Fre(0)",
		},
		{
			name:     "PEEK with arg",
			funcName: "PEEK",
			args:     []ast.Expression{&ast.NumberLiteral{Value: 1000}},
			want:     "rt.Peek(",
		},
		{
			name:     "PEEK no args",
			funcName: "PEEK",
			args:     nil,
			want:     "rt.Peek(0)",
		},
		{
			name:     "EOF with file number",
			funcName: "EOF",
			args:     []ast.Expression{&ast.NumberLiteral{Value: 1}},
			want:     "fm.Eof(",
		},
		{
			name:     "EOF no args",
			funcName: "EOF",
			args:     nil,
			want:     "float64(0)",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stmts := []ast.Statement{
				&ast.PrintStatement{
					Expressions: []ast.Expression{
						&ast.FunctionCall{
							Name: tc.funcName,
							Args: tc.args,
						},
					},
					Separators: []string{""},
				},
			}
			out := generate(t, stmts)
			if !strings.Contains(out, tc.want) {
				t.Errorf("expected %q in output, got:\n%s", tc.want, out)
			}
		})
	}
}

func TestEmitMultiDimArray(t *testing.T) {
	stmts := []ast.Statement{
		&ast.DimStatement{
			Declarations: []ast.DimDecl{
				{
					Name: "matrix",
					Dimensions: []ast.DimRange{
						{Upper: &ast.NumberLiteral{Value: 5, OriginalText: "5"}},
						{Upper: &ast.NumberLiteral{Value: 10, OriginalText: "10"}},
					},
				},
			},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "matrix") {
		t.Errorf("expected 'matrix' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "make(") {
		t.Errorf("expected 'make(' in output, got:\n%s", out)
	}
}

func TestEmitArrayAccess1D(t *testing.T) {
	stmts := []ast.Statement{
		&ast.PrintStatement{
			Expressions: []ast.Expression{
				&ast.ArrayAccess{
					Name:    "arr",
					Indices: []ast.Expression{&ast.NumberLiteral{Value: 1, OriginalText: "1"}},
				},
			},
			Separators: []string{""},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "arr[") {
		t.Errorf("expected 'arr[' in output, got:\n%s", out)
	}
}

func TestEmitArrayAccessMultiDim(t *testing.T) {
	stmts := []ast.Statement{
		&ast.PrintStatement{
			Expressions: []ast.Expression{
				&ast.ArrayAccess{
					Name: "mat",
					Indices: []ast.Expression{
						&ast.NumberLiteral{Value: 1, OriginalText: "1"},
						&ast.NumberLiteral{Value: 2, OriginalText: "2"},
					},
				},
			},
			Separators: []string{""},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "mat[") {
		t.Errorf("expected 'mat[' in output, got:\n%s", out)
	}
}

func TestEmitArrayAccessNoIndices(t *testing.T) {
	// Array passed by reference (no indices) - should emit just the name
	stmts := []ast.Statement{
		&ast.PrintStatement{
			Expressions: []ast.Expression{
				&ast.ArrayAccess{
					Name:    "arr",
					Indices: []ast.Expression{},
				},
			},
			Separators: []string{""},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "arr") {
		t.Errorf("expected 'arr' in output, got:\n%s", out)
	}
}

func TestEmitArrayAccessFromFunction(t *testing.T) {
	// Array access where symbol table says it's a function
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.PrintStatement{
				Expressions: []ast.Expression{
					&ast.ArrayAccess{
						Name:    "myFunc",
						Indices: []ast.Expression{&ast.NumberLiteral{Value: 5, OriginalText: "5"}},
					},
				},
				Separators: []string{""},
			},
		},
	}
	table := semantic.NewSymbolTable()
	_ = table.Define("MYFUNC", &semantic.Symbol{Type: semantic.SymFunction})
	gen := New()
	out, err := gen.Generate(prog, table)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if !strings.Contains(out, "myFunc(") {
		t.Errorf("expected 'myFunc(' in output, got:\n%s", out)
	}
}

func TestEmitUnaryNeg(t *testing.T) {
	stmts := []ast.Statement{
		&ast.PrintStatement{
			Expressions: []ast.Expression{
				&ast.UnaryExpr{
					Operator: "-",
					Operand:  &ast.NumberLiteral{Value: 5, OriginalText: "5"},
				},
			},
			Separators: []string{""},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "-") {
		t.Errorf("expected unary minus in output, got:\n%s", out)
	}
}

func TestEmitUnaryPos(t *testing.T) {
	stmts := []ast.Statement{
		&ast.PrintStatement{
			Expressions: []ast.Expression{
				&ast.UnaryExpr{
					Operator: "+",
					Operand:  &ast.NumberLiteral{Value: 3, OriginalText: "3"},
				},
			},
			Separators: []string{""},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "+") {
		t.Errorf("expected unary plus in output, got:\n%s", out)
	}
}

func TestEmitUnaryNot(t *testing.T) {
	stmts := []ast.Statement{
		&ast.PrintStatement{
			Expressions: []ast.Expression{
				&ast.UnaryExpr{
					Operator: "NOT",
					Operand:  &ast.NumberLiteral{Value: 1, OriginalText: "1"},
				},
			},
			Separators: []string{""},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "^int(") {
		t.Errorf("expected '^int(' (NOT) in output, got:\n%s", out)
	}
}

func TestEmitUnaryDefault(t *testing.T) {
	stmts := []ast.Statement{
		&ast.PrintStatement{
			Expressions: []ast.Expression{
				&ast.UnaryExpr{
					Operator: "~",
					Operand:  &ast.NumberLiteral{Value: 1, OriginalText: "1"},
				},
			},
			Separators: []string{""},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "~") {
		t.Errorf("expected '~' in output, got:\n%s", out)
	}
}

func TestEmitFnCallExpression(t *testing.T) {
	stmts := []ast.Statement{
		&ast.PrintStatement{
			Expressions: []ast.Expression{
				&ast.FnCallExpression{
					Name: "FNsquare",
					Args: []ast.Expression{&ast.NumberLiteral{Value: 4, OriginalText: "4"}},
				},
			},
			Separators: []string{""},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "fn_FNsquare(") {
		t.Errorf("expected 'fn_FNsquare(' in output, got:\n%s", out)
	}
}

func TestEmitBinaryExprAllOps(t *testing.T) {
	cases := []struct {
		op      string
		contain string
	}{
		{"+", "+"},
		{"-", "-"},
		{"*", "*"},
		{"/", "/"},
		{"\\", "int("},   // integer division
		{"MOD", "% int("}, // modulo (Go uses %)
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
	if !strings.Contains(out, "&") {
		t.Errorf("expected & in output for unknown op, got:\n%s", out)
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

func TestEmitArrayAssignment1D(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ArrayAssignment{
			Array: &ast.ArrayAccess{
				Name:    "arr",
				Indices: []ast.Expression{&ast.NumberLiteral{Value: 2, OriginalText: "2"}},
			},
			Value: &ast.NumberLiteral{Value: 42, OriginalText: "42"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "arr") {
		t.Errorf("expected 'arr' in output, got:\n%s", out)
	}
}

func TestEmitArrayAssignment2D(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ArrayAssignment{
			Array: &ast.ArrayAccess{
				Name: "mat",
				Indices: []ast.Expression{
					&ast.NumberLiteral{Value: 1, OriginalText: "1"},
					&ast.NumberLiteral{Value: 2, OriginalText: "2"},
				},
			},
			Value: &ast.NumberLiteral{Value: 99, OriginalText: "99"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "mat") {
		t.Errorf("expected 'mat' in output, got:\n%s", out)
	}
}
