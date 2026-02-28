package codegen

import (
	"strings"
	"testing"

	"github.com/loabletech/go-basic/internal/ast"
	"github.com/loabletech/go-basic/internal/semantic"
)

// helper builds a Program and SymbolTable, then calls Generate.
func generate(t *testing.T, stmts []ast.Statement) string {
	t.Helper()
	prog := &ast.Program{Statements: stmts}
	table := semantic.NewSymbolTable()
	gen := New()
	out, err := gen.Generate(prog, table)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	return out
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestEmptyProgram(t *testing.T) {
	out := generate(t, nil)
	if !strings.Contains(out, "package main") {
		t.Error("output missing 'package main'")
	}
	if !strings.Contains(out, "func main()") {
		t.Error("output missing 'func main()'")
	}
	// Should import the runtime.
	if !strings.Contains(out, `rt "github.com/loabletech/go-basic/internal/runtime"`) {
		t.Error("output missing runtime import")
	}
}

func TestPrintNoArgs(t *testing.T) {
	stmts := []ast.Statement{
		&ast.PrintStatement{},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "fmt.Println()") {
		t.Errorf("expected fmt.Println(), got:\n%s", out)
	}
}

func TestPrintWithString(t *testing.T) {
	stmts := []ast.Statement{
		&ast.PrintStatement{
			Expressions: []ast.Expression{
				&ast.StringLiteral{Value: "Hello, World!"},
			},
			Separators: []string{""},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, `fmt.Println("Hello, World!")`) {
		t.Errorf("expected fmt.Println with string, got:\n%s", out)
	}
}

func TestPrintTrailingSemicolon(t *testing.T) {
	stmts := []ast.Statement{
		&ast.PrintStatement{
			Expressions: []ast.Expression{
				&ast.StringLiteral{Value: "no newline"},
			},
			Separators:     []string{";"},
			HasTrailingSep: true,
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "fmt.Print(") {
		t.Errorf("expected fmt.Print (no newline), got:\n%s", out)
	}
}

func TestLetAssignment(t *testing.T) {
	stmts := []ast.Statement{
		&ast.LetStatement{
			Name:  &ast.Identifier{Name: "x"},
			Value: &ast.NumberLiteral{Value: 42, NumType: ast.NumInt},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "var x") {
		t.Errorf("expected variable declaration for x, got:\n%s", out)
	}
	if !strings.Contains(out, "42") {
		t.Errorf("expected value 42, got:\n%s", out)
	}
}

func TestLetStringVariable(t *testing.T) {
	stmts := []ast.Statement{
		&ast.LetStatement{
			Name:       &ast.Identifier{Name: "name", TypeSuffix: "$"},
			Value:      &ast.StringLiteral{Value: "Alice"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "name_str") {
		t.Errorf("expected mangled name name_str, got:\n%s", out)
	}
	if !strings.Contains(out, "string") {
		t.Errorf("expected string type, got:\n%s", out)
	}
}

func TestIfThenElse(t *testing.T) {
	stmts := []ast.Statement{
		&ast.IfStatement{
			Condition: &ast.BinaryExpr{
				Left:     &ast.Identifier{Name: "x"},
				Operator: ">",
				Right:    &ast.NumberLiteral{Value: 10, NumType: ast.NumInt},
			},
			ThenBlock: []ast.Statement{
				&ast.PrintStatement{
					Expressions: []ast.Expression{
						&ast.StringLiteral{Value: "big"},
					},
					Separators: []string{""},
				},
			},
			ElseBlock: []ast.Statement{
				&ast.PrintStatement{
					Expressions: []ast.Expression{
						&ast.StringLiteral{Value: "small"},
					},
					Separators: []string{""},
				},
			},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "if") {
		t.Errorf("expected 'if' statement, got:\n%s", out)
	}
	if !strings.Contains(out, "} else {") {
		t.Errorf("expected '} else {', got:\n%s", out)
	}
	if !strings.Contains(out, `"big"`) {
		t.Errorf("expected then-branch string, got:\n%s", out)
	}
	if !strings.Contains(out, `"small"`) {
		t.Errorf("expected else-branch string, got:\n%s", out)
	}
}

func TestIfElseIf(t *testing.T) {
	stmts := []ast.Statement{
		&ast.IfStatement{
			Condition: &ast.BinaryExpr{
				Left:     &ast.Identifier{Name: "x"},
				Operator: "=",
				Right:    &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
			},
			ThenBlock: []ast.Statement{
				&ast.PrintStatement{
					Expressions: []ast.Expression{&ast.StringLiteral{Value: "one"}},
					Separators:  []string{""},
				},
			},
			ElseIfClauses: []ast.ElseIfClause{
				{
					Condition: &ast.BinaryExpr{
						Left:     &ast.Identifier{Name: "x"},
						Operator: "=",
						Right:    &ast.NumberLiteral{Value: 2, NumType: ast.NumInt},
					},
					Body: []ast.Statement{
						&ast.PrintStatement{
							Expressions: []ast.Expression{&ast.StringLiteral{Value: "two"}},
							Separators:  []string{""},
						},
					},
				},
			},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "} else if") {
		t.Errorf("expected '} else if', got:\n%s", out)
	}
}

func TestForLoop(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ForStatement{
			Counter: &ast.Identifier{Name: "i"},
			Start:   &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
			End:     &ast.NumberLiteral{Value: 10, NumType: ast.NumInt},
			Body: []ast.Statement{
				&ast.PrintStatement{
					Expressions: []ast.Expression{
						&ast.Identifier{Name: "i"},
					},
					Separators: []string{""},
				},
			},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "for i =") {
		t.Errorf("expected for loop with counter i, got:\n%s", out)
	}
}

func TestForLoopWithStep(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ForStatement{
			Counter: &ast.Identifier{Name: "i"},
			Start:   &ast.NumberLiteral{Value: 0, NumType: ast.NumInt},
			End:     &ast.NumberLiteral{Value: 20, NumType: ast.NumInt},
			Step:    &ast.NumberLiteral{Value: 2, NumType: ast.NumInt},
			Body:    []ast.Statement{},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "step_") {
		t.Errorf("expected step variable, got:\n%s", out)
	}
}

func TestWhileLoop(t *testing.T) {
	stmts := []ast.Statement{
		&ast.WhileStatement{
			Condition: &ast.BinaryExpr{
				Left:     &ast.Identifier{Name: "x"},
				Operator: "<",
				Right:    &ast.NumberLiteral{Value: 100, NumType: ast.NumInt},
			},
			Body: []ast.Statement{
				&ast.IncrStatement{
					Variable: &ast.Identifier{Name: "x"},
				},
			},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "for (x < ") {
		t.Errorf("expected while loop as for loop, got:\n%s", out)
	}
}

func TestDoLoopTopWhile(t *testing.T) {
	stmts := []ast.Statement{
		&ast.DoLoopStatement{
			Condition: &ast.BinaryExpr{
				Left:     &ast.Identifier{Name: "x"},
				Operator: ">",
				Right:    &ast.NumberLiteral{Value: 0, NumType: ast.NumInt},
			},
			TestAtTop: true,
			IsUntil:   false,
			Body:      []ast.Statement{},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "for (x >") {
		t.Errorf("expected DO WHILE as for loop, got:\n%s", out)
	}
}

func TestDoLoopBottomUntil(t *testing.T) {
	stmts := []ast.Statement{
		&ast.DoLoopStatement{
			Condition: &ast.BinaryExpr{
				Left:     &ast.Identifier{Name: "done"},
				Operator: "=",
				Right:    &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
			},
			TestAtTop: false,
			IsUntil:   true,
			Body: []ast.Statement{
				&ast.PrintStatement{
					Expressions: []ast.Expression{&ast.StringLiteral{Value: "loop"}},
					Separators:  []string{""},
				},
			},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "for {") {
		t.Errorf("expected DO ... LOOP UNTIL as for { ... break }, got:\n%s", out)
	}
	if !strings.Contains(out, "break") {
		t.Errorf("expected break for UNTIL, got:\n%s", out)
	}
}

func TestEndStatement(t *testing.T) {
	stmts := []ast.Statement{
		&ast.EndStatement{},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "os.Exit(0)") {
		t.Errorf("expected os.Exit(0), got:\n%s", out)
	}
}

func TestRemComment(t *testing.T) {
	stmts := []ast.Statement{
		&ast.RemStatement{Text: "this is a comment"},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "// this is a comment") {
		t.Errorf("expected Go comment, got:\n%s", out)
	}
}

func TestGoto(t *testing.T) {
	stmts := []ast.Statement{
		&ast.LineNumberStatement{Number: 100},
		&ast.GotoStatement{Target: "100"},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "line_100:") {
		t.Errorf("expected label line_100:, got:\n%s", out)
	}
	if !strings.Contains(out, "goto line_100") {
		t.Errorf("expected goto line_100, got:\n%s", out)
	}
}

func TestLabelStatement(t *testing.T) {
	stmts := []ast.Statement{
		&ast.LabelStatement{Name: "myLabel"},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "label_myLabel:") {
		t.Errorf("expected label_myLabel:, got:\n%s", out)
	}
}

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

func TestSelectCase(t *testing.T) {
	stmts := []ast.Statement{
		&ast.SelectCaseStatement{
			TestExpr: &ast.Identifier{Name: "x"},
			Cases: []ast.CaseClause{
				{
					Values: []ast.CaseValue{
						{Value: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt}},
					},
					Body: []ast.Statement{
						&ast.PrintStatement{
							Expressions: []ast.Expression{&ast.StringLiteral{Value: "one"}},
							Separators:  []string{""},
						},
					},
				},
				{
					Values: []ast.CaseValue{
						{Value: &ast.NumberLiteral{Value: 2, NumType: ast.NumInt}},
					},
					Body: []ast.Statement{
						&ast.PrintStatement{
							Expressions: []ast.Expression{&ast.StringLiteral{Value: "two"}},
							Separators:  []string{""},
						},
					},
				},
			},
			ElseBlock: []ast.Statement{
				&ast.PrintStatement{
					Expressions: []ast.Expression{&ast.StringLiteral{Value: "other"}},
					Separators:  []string{""},
				},
			},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "sel_") {
		t.Errorf("expected select temp variable, got:\n%s", out)
	}
	if !strings.Contains(out, "} else {") {
		t.Errorf("expected else block for CASE ELSE, got:\n%s", out)
	}
}

func TestSwapStatement(t *testing.T) {
	stmts := []ast.Statement{
		&ast.SwapStatement{
			Var1: &ast.Identifier{Name: "a"},
			Var2: &ast.Identifier{Name: "b"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "a, b = b, a") {
		t.Errorf("expected Go swap idiom, got:\n%s", out)
	}
}

func TestIncrDecr(t *testing.T) {
	stmts := []ast.Statement{
		&ast.IncrStatement{
			Variable: &ast.Identifier{Name: "x"},
		},
		&ast.DecrStatement{
			Variable: &ast.Identifier{Name: "y"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "x++") {
		t.Errorf("expected x++, got:\n%s", out)
	}
	if !strings.Contains(out, "y--") {
		t.Errorf("expected y--, got:\n%s", out)
	}
}

func TestExitFor(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ExitStatement{ExitType: "FOR"},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "break") {
		t.Errorf("expected break for EXIT FOR, got:\n%s", out)
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
					Name:  &ast.Identifier{Name: "Double"},
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
		{"return", "b_return"},  // reserved word
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

func TestFunctionCallMappings(t *testing.T) {
	// Test that built-in function calls map to the runtime package.
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

func TestOutputContainsPackageMain(t *testing.T) {
	out := generate(t, nil)
	if !strings.HasPrefix(out, "package main\n") {
		t.Errorf("output should start with 'package main', got:\n%.100s", out)
	}
}

func TestDataRead(t *testing.T) {
	stmts := []ast.Statement{
		&ast.DataStatement{
			Values: []ast.Expression{
				&ast.NumberLiteral{Value: 10, NumType: ast.NumInt},
				&ast.NumberLiteral{Value: 20, NumType: ast.NumInt},
			},
		},
		&ast.ReadStatement{
			Variables: []ast.Expression{
				&ast.Identifier{Name: "a"},
			},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "dataPool") {
		t.Errorf("expected dataPool, got:\n%s", out)
	}
	if !strings.Contains(out, "dataIdx") {
		t.Errorf("expected dataIdx, got:\n%s", out)
	}
}

func TestBeepStatement(t *testing.T) {
	stmts := []ast.Statement{
		&ast.BeepStatement{},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, `"\a"`) {
		t.Errorf("expected bell character for BEEP, got:\n%s", out)
	}
}

// ---------------------------------------------------------------------------
// No-argument builtin identifiers (INKEY$, DATE$, TIME$, RND, TIMER, ERR, ERL, ERADR)
// ---------------------------------------------------------------------------

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

// TestFrePeekEofEmission verifies that FRE, PEEK, and EOF are emitted as
// runtime/fm calls rather than falling through to the user-function default.
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

// TestSharedVarsPackageLevel verifies that SHARED variables inside SUBs
// are emitted as package-level var declarations, not inside main() or
// the SUB body.
func TestSharedVarsPackageLevel(t *testing.T) {
	stmts := []ast.Statement{
		// DIM Name1$ in main
		&ast.DimStatement{
			Declarations: []ast.DimDecl{
				{Name: "Name1", TypeSuffix: "$", ElementType: "STRING"},
			},
		},
		// LET Name1$ = "Alice" in main
		&ast.LetStatement{
			Name:  &ast.Identifier{Name: "Name1", TypeSuffix: "$"},
			Value: &ast.StringLiteral{Value: "Alice"},
		},
		// SUB ShowInfo with SHARED Name1$
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

	// Name1_str should be declared at package level.
	if !strings.Contains(out, "var Name1_str string") {
		t.Errorf("expected package-level 'var Name1_str string', got:\n%s", out)
	}

	// The package-level declaration should appear BEFORE func main().
	pkgIdx := strings.Index(out, "var Name1_str string")
	mainIdx := strings.Index(out, "func main()")
	if pkgIdx > mainIdx {
		t.Errorf("package-level var should appear before func main(), pkgIdx=%d mainIdx=%d", pkgIdx, mainIdx)
	}

	// Name1_str should NOT be re-declared inside main() (assignment only).
	mainBody := out[mainIdx:]
	if strings.Contains(mainBody, "var Name1_str string") {
		t.Errorf("Name1_str should not be re-declared inside main(), got:\n%s", mainBody)
	}

	// The SUB body should use Name1_str without declaring it.
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
		// SUB with a local variable
		&ast.SubDeclaration{
			Name: "MySub",
			Body: []ast.Statement{
				&ast.LetStatement{
					Name:  &ast.Identifier{Name: "LocalVar", TypeSuffix: "$"},
					Value: &ast.StringLiteral{Value: "hello"},
				},
			},
		},
		// Same variable name in main should get its own declaration
		&ast.LetStatement{
			Name:  &ast.Identifier{Name: "LocalVar", TypeSuffix: "$"},
			Value: &ast.StringLiteral{Value: "world"},
		},
	}
	out := generate(t, stmts)

	// LocalVar_str should be declared in both main and the SUB.
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

// ---------------------------------------------------------------------------
// File I/O statement emission tests
// ---------------------------------------------------------------------------

func TestFilePrintStatement(t *testing.T) {
	stmts := []ast.Statement{
		&ast.FilePrintStatement{
			FileNum:     &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
			Expressions: []ast.Expression{&ast.StringLiteral{Value: "hello"}},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "fm.FilePrint(") {
		t.Errorf("expected fm.FilePrint call, got:\n%s", out)
	}
	if !strings.Contains(out, "fm := rt.NewFileManager()") {
		t.Errorf("expected FileManager declaration, got:\n%s", out)
	}
}

func TestFileInputStatement(t *testing.T) {
	stmts := []ast.Statement{
		&ast.FileInputStatement{
			FileNum:   &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
			Variables: []ast.Expression{&ast.Identifier{Name: "a", TypeSuffix: "$"}},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "fm.FileInput(") {
		t.Errorf("expected fm.FileInput call, got:\n%s", out)
	}
}

func TestFileWriteStatement(t *testing.T) {
	stmts := []ast.Statement{
		&ast.FileWriteStatement{
			FileNum:     &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
			Expressions: []ast.Expression{&ast.StringLiteral{Value: "test"}, &ast.NumberLiteral{Value: 100, NumType: ast.NumInt}},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "fm.FileWrite(") {
		t.Errorf("expected fm.FileWrite call, got:\n%s", out)
	}
}

func TestFieldStatement(t *testing.T) {
	stmts := []ast.Statement{
		&ast.FieldStatement{
			FileNum: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
			Fields: []ast.FieldDef{
				{Length: &ast.NumberLiteral{Value: 20, NumType: ast.NumInt}, VarName: "A$"},
				{Length: &ast.NumberLiteral{Value: 10, NumType: ast.NumInt}, VarName: "B$"},
			},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "fm.Field(") {
		t.Errorf("expected fm.Field call, got:\n%s", out)
	}
	if !strings.Contains(out, "rt.FieldDef") {
		t.Errorf("expected rt.FieldDef in output, got:\n%s", out)
	}
}

func TestLsetStatementEmission(t *testing.T) {
	stmts := []ast.Statement{
		&ast.LsetStatement{
			Variable: "A$",
			Value:    &ast.StringLiteral{Value: "hello"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "fm.Lset(") {
		t.Errorf("expected fm.Lset call, got:\n%s", out)
	}
}

func TestPutStatement(t *testing.T) {
	stmts := []ast.Statement{
		&ast.PutStatement{
			FileNum:     &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
			RecordOrPos: &ast.NumberLiteral{Value: 5, NumType: ast.NumInt},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "fm.RandomPut(") {
		t.Errorf("expected fm.RandomPut call, got:\n%s", out)
	}
}

func TestGetStatementEmission(t *testing.T) {
	stmts := []ast.Statement{
		&ast.GetStatement{
			FileNum:     &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
			RecordOrPos: &ast.NumberLiteral{Value: 3, NumType: ast.NumInt},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "fm.RandomGet(") {
		t.Errorf("expected fm.RandomGet call, got:\n%s", out)
	}
}

func TestSeekStatementEmission(t *testing.T) {
	stmts := []ast.Statement{
		&ast.SeekStatement{
			FileNum:  &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
			Position: &ast.NumberLiteral{Value: 10, NumType: ast.NumInt},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "fm.FileSeek(") {
		t.Errorf("expected fm.FileSeek call, got:\n%s", out)
	}
}

func TestScreenStatement(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ScreenStatement{
			Mode: &ast.NumberLiteral{Value: 13, NumType: ast.NumInt},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "rt.ScreenMode(int(13))") {
		t.Errorf("expected rt.ScreenMode(int(13)), got:\n%s", out)
	}
}
