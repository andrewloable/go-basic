package codegen

import (
	"strings"
	"testing"

	"github.com/loabletech/go-basic/internal/ast"
	"github.com/loabletech/go-basic/internal/semantic"
)

// ===========================================================================
// emitLocate, emitColor
// ===========================================================================

func TestEmitLocate(t *testing.T) {
	stmts := []ast.Statement{
		&ast.LocateStatement{
			Row: &ast.NumberLiteral{Value: 5, OriginalText: "5"},
			Col: &ast.NumberLiteral{Value: 10, OriginalText: "10"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "AnsiLocate") {
		t.Errorf("expected AnsiLocate in output, got:\n%s", out)
	}
}

func TestEmitLocateNoArgs(t *testing.T) {
	stmts := []ast.Statement{
		&ast.LocateStatement{},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "AnsiLocate") {
		t.Errorf("expected AnsiLocate in output, got:\n%s", out)
	}
}

func TestEmitColor(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ColorStatement{
			Foreground: &ast.NumberLiteral{Value: 7, OriginalText: "7"},
			Background: &ast.NumberLiteral{Value: 0, OriginalText: "0"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "AnsiColor") {
		t.Errorf("expected AnsiColor in output, got:\n%s", out)
	}
}

func TestEmitColorNoBackground(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ColorStatement{
			Foreground: &ast.NumberLiteral{Value: 15, OriginalText: "15"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "AnsiColor") {
		t.Errorf("expected AnsiColor in output, got:\n%s", out)
	}
}

// ===========================================================================
// emitOnComputedGoto / emitOnComputedGosub
// ===========================================================================

func TestEmitOnComputedGoto(t *testing.T) {
	stmts := []ast.Statement{
		&ast.OnComputedGotoStatement{
			Expr:    &ast.NumberLiteral{Value: 1, OriginalText: "1"},
			Targets: []string{"label1", "label2"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "_on_idx") {
		t.Errorf("expected _on_idx in output, got:\n%s", out)
	}
	if !strings.Contains(out, "goto") {
		t.Errorf("expected goto in output, got:\n%s", out)
	}
}

func TestEmitOnComputedGosub(t *testing.T) {
	stmts := []ast.Statement{
		&ast.OnComputedGosubStatement{
			Expr:    &ast.NumberLiteral{Value: 2, OriginalText: "2"},
			Targets: []string{"sub1", "sub2", "sub3"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "_on_idx") {
		t.Errorf("expected _on_idx in output, got:\n%s", out)
	}
}

// ===========================================================================
// emitTypeBlock
// ===========================================================================

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

// ===========================================================================
// emitArrayAssignment
// ===========================================================================

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

// ===========================================================================
// emitMultiDimArray
// ===========================================================================

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

// ===========================================================================
// emitArrayAccess
// ===========================================================================

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

// ===========================================================================
// emitUnaryExpr
// ===========================================================================

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

// ===========================================================================
// emitDefFn
// ===========================================================================

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

// ===========================================================================
// emitFnCallExpression
// ===========================================================================

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

// ===========================================================================
// Type helpers
// ===========================================================================

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

func TestAddPackageVar(t *testing.T) {
	gen := New()
	gen.addPackageVar("myVar", "float32")
	if !gen.sharedVars["myVar"] {
		t.Error("expected myVar in sharedVars after addPackageVar")
	}
	// Adding again shouldn't duplicate
	gen.addPackageVar("myVar", "float32")
	count := 0
	for _, v := range gen.packageVars {
		if v.name == "myVar" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected myVar to appear once in packageVars, got %d", count)
	}
}

func TestAddHoistedVar(t *testing.T) {
	gen := New()
	gen.addHoistedVar("v1", "int16")
	gen.addHoistedVar("v1", "int16") // duplicate — should not appear twice
	count := 0
	for _, v := range gen.hoistedVars {
		if v.name == "v1" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected v1 to appear once in hoistedVars, got %d", count)
	}
}

func TestAddHoistedVarSkipsShared(t *testing.T) {
	gen := New()
	gen.sharedVars["sharedV"] = true
	gen.addHoistedVar("sharedV", "float32")
	for _, v := range gen.hoistedVars {
		if v.name == "sharedV" {
			t.Error("expected sharedV to be skipped in hoistedVars (already package-level)")
		}
	}
}

// ===========================================================================
// emitRandomize
// ===========================================================================

func TestEmitRandomizeNoSeed(t *testing.T) {
	stmts := []ast.Statement{
		&ast.RandomizeStatement{},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "Randomize") {
		t.Errorf("expected 'Randomize' in output, got:\n%s", out)
	}
}

func TestEmitRandomizeWithSeed(t *testing.T) {
	stmts := []ast.Statement{
		&ast.RandomizeStatement{
			Seed: &ast.NumberLiteral{Value: 42, OriginalText: "42"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "Randomize") {
		t.Errorf("expected 'Randomize' in output, got:\n%s", out)
	}
}

// ===========================================================================
// emitGosub / emitReturn
// ===========================================================================

func TestEmitGosub(t *testing.T) {
	stmts := []ast.Statement{
		&ast.GosubStatement{Target: "mySub"},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "goto") {
		t.Errorf("expected 'goto' in emitted GOSUB, got:\n%s", out)
	}
}

func TestEmitReturn(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ReturnStatement{},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "return") {
		t.Errorf("expected 'return' in emitted RETURN, got:\n%s", out)
	}
}

// ===========================================================================
// emitRestore
// ===========================================================================

func TestEmitRestore(t *testing.T) {
	stmts := []ast.Statement{
		&ast.RestoreStatement{},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "dataIdx") {
		t.Errorf("expected 'dataIdx' in emitted RESTORE, got:\n%s", out)
	}
}

// ===========================================================================
// emitErase
// ===========================================================================

func TestEmitErase(t *testing.T) {
	stmts := []ast.Statement{
		&ast.EraseStatement{
			Names: []string{"arr1", "arr2"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "arr1") {
		t.Errorf("expected 'arr1' in erase output, got:\n%s", out)
	}
}

// ===========================================================================
// emitCircle, emitLine, emitPset, emitPaint, emitView
// ===========================================================================

func TestEmitCircle(t *testing.T) {
	stmts := []ast.Statement{
		&ast.CircleStmt{
			X:      &ast.NumberLiteral{Value: 100, OriginalText: "100"},
			Y:      &ast.NumberLiteral{Value: 100, OriginalText: "100"},
			Radius: &ast.NumberLiteral{Value: 50, OriginalText: "50"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "Circle") {
		t.Errorf("expected 'Circle' in output, got:\n%s", out)
	}
}

func TestEmitCircleWithColor(t *testing.T) {
	stmts := []ast.Statement{
		&ast.CircleStmt{
			X:      &ast.NumberLiteral{Value: 160, OriginalText: "160"},
			Y:      &ast.NumberLiteral{Value: 100, OriginalText: "100"},
			Radius: &ast.NumberLiteral{Value: 40, OriginalText: "40"},
			Color:  &ast.NumberLiteral{Value: 4, OriginalText: "4"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "Circle") {
		t.Errorf("expected 'Circle' in output, got:\n%s", out)
	}
}

func TestEmitLine(t *testing.T) {
	stmts := []ast.Statement{
		&ast.LineStmt{
			X2: &ast.NumberLiteral{Value: 200, OriginalText: "200"},
			Y2: &ast.NumberLiteral{Value: 100, OriginalText: "100"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "DrawLine") {
		t.Errorf("expected 'DrawLine' in output, got:\n%s", out)
	}
}

func TestEmitLineWithCoords(t *testing.T) {
	stmts := []ast.Statement{
		&ast.LineStmt{
			X1:    &ast.NumberLiteral{Value: 10, OriginalText: "10"},
			Y1:    &ast.NumberLiteral{Value: 20, OriginalText: "20"},
			X2:    &ast.NumberLiteral{Value: 100, OriginalText: "100"},
			Y2:    &ast.NumberLiteral{Value: 200, OriginalText: "200"},
			Color: &ast.NumberLiteral{Value: 14, OriginalText: "14"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "DrawLine") {
		t.Errorf("expected 'DrawLine' in output, got:\n%s", out)
	}
}

func TestEmitPset(t *testing.T) {
	stmts := []ast.Statement{
		&ast.PsetStatement{
			X: &ast.NumberLiteral{Value: 50, OriginalText: "50"},
			Y: &ast.NumberLiteral{Value: 75, OriginalText: "75"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "Pset") {
		t.Errorf("expected 'Pset' in output, got:\n%s", out)
	}
}

func TestEmitPaint(t *testing.T) {
	stmts := []ast.Statement{
		&ast.PaintStmt{
			X: &ast.NumberLiteral{Value: 100, OriginalText: "100"},
			Y: &ast.NumberLiteral{Value: 50, OriginalText: "50"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "Paint") {
		t.Errorf("expected 'Paint' in output, got:\n%s", out)
	}
}

func TestEmitPaintWithColors(t *testing.T) {
	stmts := []ast.Statement{
		&ast.PaintStmt{
			X:           &ast.NumberLiteral{Value: 100, OriginalText: "100"},
			Y:           &ast.NumberLiteral{Value: 50, OriginalText: "50"},
			FillColor:   &ast.NumberLiteral{Value: 2, OriginalText: "2"},
			BorderColor: &ast.NumberLiteral{Value: 4, OriginalText: "4"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "Paint") {
		t.Errorf("expected 'Paint' in output, got:\n%s", out)
	}
}

func TestEmitViewPrint(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ViewStatement{
			IsPrint: true,
			Top:     &ast.NumberLiteral{Value: 1, OriginalText: "1"},
			Bottom:  &ast.NumberLiteral{Value: 24, OriginalText: "24"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "ViewPrint") {
		t.Errorf("expected 'ViewPrint' in output, got:\n%s", out)
	}
}

func TestEmitViewPrintNoArgs(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ViewStatement{
			IsPrint: true,
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "ViewPrint") {
		t.Errorf("expected 'ViewPrint' in output, got:\n%s", out)
	}
}

// ===========================================================================
// emitRedim
// ===========================================================================

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

// ===========================================================================
// emitInputFromStdin (via InputStatement)
// ===========================================================================

func TestEmitInputFromStdin(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ReadStatement{
			IsInput: true,
			Prompt:  "Enter value: ",
			Variables: []ast.Expression{
				&ast.Identifier{Name: "x"},
			},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "Scan") || !strings.Contains(out, "Enter value") {
		t.Errorf("expected Scan and prompt in output, got:\n%s", out)
	}
}

// ===========================================================================
// emitOpen / emitClose
// ===========================================================================

func TestEmitOpen(t *testing.T) {
	stmts := []ast.Statement{
		&ast.OpenStatement{
			Filename: &ast.StringLiteral{Value: "test.txt"},
			FileNum:  &ast.NumberLiteral{Value: 1, OriginalText: "1"},
			Mode:     "INPUT",
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "FileOpen") {
		t.Errorf("expected 'FileOpen' in output, got:\n%s", out)
	}
}

func TestEmitClose(t *testing.T) {
	stmts := []ast.Statement{
		&ast.CloseStatement{
			FileNums: []ast.Expression{
				&ast.NumberLiteral{Value: 1, OriginalText: "1"},
			},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "FileClose") {
		t.Errorf("expected 'FileClose' in output, got:\n%s", out)
	}
}

// ===========================================================================
// emitPrintUsing
// ===========================================================================

func TestEmitPrintUsing(t *testing.T) {
	stmts := []ast.Statement{
		&ast.PrintStatement{
			Format: &ast.StringLiteral{Value: "##.##"},
			Expressions: []ast.Expression{
				&ast.NumberLiteral{Value: 3, OriginalText: "3"},
			},
			Separators: []string{""},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "PrintUsing") {
		t.Errorf("expected 'PrintUsing' in output, got:\n%s", out)
	}
}

// ===========================================================================
// Lset / Rset emit
// ===========================================================================

func TestEmitRset(t *testing.T) {
	stmts := []ast.Statement{
		&ast.RsetStatement{
			Variable: "field1",
			Value:    &ast.StringLiteral{Value: "hello"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "Rset") {
		t.Errorf("expected 'Rset' in output, got:\n%s", out)
	}
}
