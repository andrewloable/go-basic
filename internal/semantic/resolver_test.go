package semantic

import (
	"strings"
	"testing"

	"github.com/loabletech/go-basic/internal/ast"
)

// ===========================================================================
// Resolver tests (resolver.go — Resolve(), registerForwardDecl, defineGlobal)
// ===========================================================================

func TestResolverSimpleAssignment(t *testing.T) {
	// Simulates: x% = 42
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LetStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Name: &ast.Identifier{
					BasePos:    ast.Position{Line: 1, Column: 1},
					Name:       "x",
					TypeSuffix: "%",
				},
				Value: &ast.NumberLiteral{
					BasePos:      ast.Position{Line: 1, Column: 6},
					Value:        42,
					OriginalText: "42",
				},
			},
		},
	}

	resolver := NewResolver(prog)
	table, errs := resolver.Resolve()
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	sym := table.Lookup("x%")
	if sym == nil {
		t.Fatal("x% should be in the symbol table")
	}
	if sym.DataType != TypeInteger {
		t.Errorf("expected TypeInteger, got %v", sym.DataType)
	}
	if !sym.Defined {
		t.Error("x% should be marked as Defined")
	}
}

func TestResolverForLoop(t *testing.T) {
	// Simulates: FOR i% = 1 TO 10 / NEXT
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ForStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Counter: &ast.Identifier{
					BasePos:    ast.Position{Line: 1, Column: 5},
					Name:       "i",
					TypeSuffix: "%",
				},
				Start: &ast.NumberLiteral{
					BasePos:      ast.Position{Line: 1, Column: 10},
					Value:        1,
					OriginalText: "1",
				},
				End: &ast.NumberLiteral{
					BasePos:      ast.Position{Line: 1, Column: 15},
					Value:        10,
					OriginalText: "10",
				},
			},
		},
	}

	resolver := NewResolver(prog)
	table, errs := resolver.Resolve()
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	sym := table.Lookup("i%")
	if sym == nil {
		t.Fatal("i% should be in the symbol table")
	}
	if sym.DataType != TypeInteger {
		t.Errorf("expected TypeInteger for i%%, got %v", sym.DataType)
	}
}

func TestResolverLabelAndGoto(t *testing.T) {
	// Simulates: myLabel: / GOTO myLabel
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LabelStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Name:    "myLabel",
			},
			&ast.GotoStatement{
				BasePos: ast.Position{Line: 2, Column: 1},
				Target:  "myLabel",
			},
		},
	}

	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestResolverUndefinedLabel(t *testing.T) {
	// GOTO to a non-existent label should produce an error.
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.GotoStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Target:  "missing",
			},
		},
	}

	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) == 0 {
		t.Fatal("expected error for undefined label")
	}
	if !strings.Contains(errs[0], "undefined label") {
		t.Fatalf("expected 'undefined label' error, got: %s", errs[0])
	}
}

func TestResolverSubDeclaration(t *testing.T) {
	// SUB MySub(a%, b$) / END SUB
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.SubDeclaration{
				BasePos: ast.Position{Line: 1, Column: 1},
				Name:    "MySub",
				Params: []ast.Parameter{
					{BasePos: ast.Position{Line: 1, Column: 10}, Name: "a", Type: "%"},
					{BasePos: ast.Position{Line: 1, Column: 14}, Name: "b", Type: "$"},
				},
				Body: []ast.Statement{},
			},
		},
	}

	resolver := NewResolver(prog)
	table, errs := resolver.Resolve()
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	sym := table.Lookup("MySub")
	if sym == nil {
		t.Fatal("MySub should be in the symbol table")
	}
	if sym.Type != SymSub {
		t.Errorf("expected SymSub, got %v", sym.Type)
	}
	if len(sym.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(sym.Params))
	}
	if sym.Params[0] != TypeInteger {
		t.Errorf("expected first param TypeInteger, got %v", sym.Params[0])
	}
	if sym.Params[1] != TypeString {
		t.Errorf("expected second param TypeString, got %v", sym.Params[1])
	}

	// Parameters should be in the MYSUB scope.
	subScope, ok := table.Scopes["MYSUB"]
	if !ok {
		t.Fatal("expected MYSUB scope to exist")
	}
	if subScope.Symbols["A"] == nil {
		t.Error("parameter 'a' should be in the MYSUB scope")
	}
	if subScope.Symbols["B"] == nil {
		t.Error("parameter 'b' should be in the MYSUB scope")
	}
}

func TestResolverFunctionDeclaration(t *testing.T) {
	// FUNCTION Add%(x%, y%) / Add% = x% + y% / END FUNCTION
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.FunctionDeclaration{
				BasePos:    ast.Position{Line: 1, Column: 1},
				Name:       "Add",
				ReturnType: "%",
				Params: []ast.Parameter{
					{BasePos: ast.Position{Line: 1, Column: 14}, Name: "x", Type: "%"},
					{BasePos: ast.Position{Line: 1, Column: 18}, Name: "y", Type: "%"},
				},
				Body: []ast.Statement{
					&ast.LetStatement{
						BasePos: ast.Position{Line: 2, Column: 1},
						Name:    &ast.Identifier{BasePos: ast.Position{Line: 2, Column: 1}, Name: "Add", TypeSuffix: ""},
						Value: &ast.BinaryExpr{
							BasePos:  ast.Position{Line: 2, Column: 8},
							Left:     &ast.Identifier{BasePos: ast.Position{Line: 2, Column: 8}, Name: "x", TypeSuffix: "%"},
							Operator: "+",
							Right:    &ast.Identifier{BasePos: ast.Position{Line: 2, Column: 13}, Name: "y", TypeSuffix: "%"},
						},
					},
				},
			},
		},
	}

	resolver := NewResolver(prog)
	table, errs := resolver.Resolve()
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	sym := table.Lookup("Add")
	if sym == nil {
		t.Fatal("Add should be in the symbol table")
	}
	if sym.Type != SymFunction {
		t.Errorf("expected SymFunction, got %v", sym.Type)
	}
	if sym.ReturnType != TypeInteger {
		t.Errorf("expected TypeInteger return type, got %v", sym.ReturnType)
	}
}

func TestResolverDimArray(t *testing.T) {
	// DIM arr%(10)
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.DimStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Declarations: []ast.DimDecl{
					{
						BasePos:    ast.Position{Line: 1, Column: 5},
						Name:       "arr",
						TypeSuffix: "%",
						Dimensions: []ast.DimRange{
							{
								Upper: &ast.NumberLiteral{
									BasePos:      ast.Position{Line: 1, Column: 9},
									Value:        10,
									OriginalText: "10",
								},
							},
						},
					},
				},
			},
		},
	}

	resolver := NewResolver(prog)
	table, errs := resolver.Resolve()
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	sym := table.Lookup("arr%")
	if sym == nil {
		t.Fatal("arr% should be in the symbol table")
	}
	if sym.Type != SymArray {
		t.Errorf("expected SymArray, got %v", sym.Type)
	}
	if sym.ArrayDims != 1 {
		t.Errorf("expected 1 dimension, got %d", sym.ArrayDims)
	}
}

func TestResolverDuplicateDim(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.DimStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Declarations: []ast.DimDecl{
					{BasePos: ast.Position{Line: 1, Column: 5}, Name: "x", TypeSuffix: "%"},
				},
			},
			&ast.DimStatement{
				BasePos: ast.Position{Line: 2, Column: 1},
				Declarations: []ast.DimDecl{
					{BasePos: ast.Position{Line: 2, Column: 5}, Name: "x", TypeSuffix: "%"},
				},
			},
		},
	}

	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) == 0 {
		t.Fatal("expected error for duplicate DIM")
	}
	if !strings.Contains(errs[0], "duplicate DIM") {
		t.Fatalf("expected duplicate DIM error, got: %s", errs[0])
	}
}

func TestResolverDefType(t *testing.T) {
	// DEFINT I-N then use i without suffix
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.DefTypeStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Type:    "DEFINT",
				LetterRanges: []ast.LetterRange{
					{Start: 'I', End: 'N'},
				},
			},
			&ast.LetStatement{
				BasePos: ast.Position{Line: 2, Column: 1},
				Name:    &ast.Identifier{BasePos: ast.Position{Line: 2, Column: 1}, Name: "index"},
				Value:   &ast.NumberLiteral{BasePos: ast.Position{Line: 2, Column: 9}, Value: 1, OriginalText: "1"},
			},
		},
	}

	resolver := NewResolver(prog)
	table, errs := resolver.Resolve()
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	sym := table.Lookup("index")
	if sym == nil {
		t.Fatal("index should be in symbol table")
	}
	if sym.DataType != TypeInteger {
		t.Errorf("expected TypeInteger (from DEFINT), got %v", sym.DataType)
	}
}

func TestResolverOptionBaseConflict(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.OptionBaseStatement{BasePos: ast.Position{Line: 1, Column: 1}, Value: 1},
			&ast.OptionBaseStatement{BasePos: ast.Position{Line: 2, Column: 1}, Value: 0},
		},
	}

	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) == 0 {
		t.Fatal("expected OPTION BASE conflict error")
	}
	if !strings.Contains(errs[0], "OPTION BASE conflict") {
		t.Fatalf("expected OPTION BASE conflict error, got: %s", errs[0])
	}
}

func TestResolverWrongArgCount(t *testing.T) {
	// FUNCTION Foo%(x%) / END FUNCTION
	// ... = Foo%(1, 2)   <-- too many args
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.FunctionDeclaration{
				BasePos:    ast.Position{Line: 1, Column: 1},
				Name:       "Foo",
				ReturnType: "%",
				Params: []ast.Parameter{
					{BasePos: ast.Position{Line: 1, Column: 14}, Name: "x", Type: "%"},
				},
				Body: []ast.Statement{},
			},
			&ast.LetStatement{
				BasePos: ast.Position{Line: 3, Column: 1},
				Name:    &ast.Identifier{BasePos: ast.Position{Line: 3, Column: 1}, Name: "result", TypeSuffix: "%"},
				Value: &ast.FunctionCall{
					BasePos: ast.Position{Line: 3, Column: 12},
					Name:    "Foo",
					Args: []ast.Expression{
						&ast.NumberLiteral{BasePos: ast.Position{Line: 3, Column: 16}, Value: 1, OriginalText: "1"},
						&ast.NumberLiteral{BasePos: ast.Position{Line: 3, Column: 19}, Value: 2, OriginalText: "2"},
					},
				},
			},
		},
	}

	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) == 0 {
		t.Fatal("expected wrong argument count error")
	}
	if !strings.Contains(errs[0], "wrong number of arguments") {
		t.Fatalf("expected wrong number of arguments error, got: %s", errs[0])
	}
}

func TestResolverImplicitVariable(t *testing.T) {
	// PRINT y (y is not declared, should be implicitly created)
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.PrintStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Expressions: []ast.Expression{
					&ast.Identifier{BasePos: ast.Position{Line: 1, Column: 7}, Name: "y"},
				},
			},
		},
	}

	resolver := NewResolver(prog)
	table, errs := resolver.Resolve()
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	sym := table.Lookup("y")
	if sym == nil {
		t.Fatal("y should be implicitly declared")
	}
	if sym.DataType != TypeSingle {
		t.Errorf("expected default TypeSingle for implicit var, got %v", sym.DataType)
	}
	if !sym.Used {
		t.Error("y should be marked as Used")
	}
	if sym.Defined {
		t.Error("y should NOT be marked as Defined (only referenced, never assigned)")
	}
}

func TestResolverGosubUndefined(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.GosubStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Target:  "handler",
			},
		},
	}

	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) == 0 {
		t.Fatal("expected error for undefined GOSUB target")
	}
	if !strings.Contains(errs[0], "undefined label") {
		t.Fatalf("expected 'undefined label' error, got: %s", errs[0])
	}
}

func TestResolverSharedVariable(t *testing.T) {
	// SUB MySub / SHARED g% / END SUB
	prog := &ast.Program{
		Statements: []ast.Statement{
			// Global assignment.
			&ast.LetStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Name:    &ast.Identifier{BasePos: ast.Position{Line: 1, Column: 1}, Name: "g", TypeSuffix: "%"},
				Value:   &ast.NumberLiteral{BasePos: ast.Position{Line: 1, Column: 6}, Value: 100, OriginalText: "100"},
			},
			&ast.SubDeclaration{
				BasePos: ast.Position{Line: 2, Column: 1},
				Name:    "TestSub",
				Body: []ast.Statement{
					&ast.ScopeStatement{
						BasePos:   ast.Position{Line: 3, Column: 3},
						Modifier:  "SHARED",
						Variables: []string{"g%"},
					},
				},
			},
		},
	}

	resolver := NewResolver(prog)
	table, errs := resolver.Resolve()
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	sym := table.GlobalScope.Symbols["G%"]
	if sym == nil {
		t.Fatal("g% should be in global scope")
	}
	if !sym.IsShared {
		t.Error("g% should be marked as IsShared")
	}
}

func TestResolverStaticVariable(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.SubDeclaration{
				BasePos: ast.Position{Line: 1, Column: 1},
				Name:    "CountSub",
				Body: []ast.Statement{
					&ast.ScopeStatement{
						BasePos:   ast.Position{Line: 2, Column: 3},
						Modifier:  "STATIC",
						Variables: []string{"counter%"},
					},
				},
			},
		},
	}

	resolver := NewResolver(prog)
	table, errs := resolver.Resolve()
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	subScope := table.Scopes["COUNTSUB"]
	if subScope == nil {
		t.Fatal("COUNTSUB scope should exist")
	}
	sym := subScope.Symbols["COUNTER%"]
	if sym == nil {
		t.Fatal("counter% should be in COUNTSUB scope")
	}
	if !sym.IsStatic {
		t.Error("counter% should be marked as IsStatic")
	}
}

func TestResolverLineNumberAsLabel(t *testing.T) {
	// 100 PRINT "hello"
	// GOTO 100
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LineNumberStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Number:  100,
			},
			&ast.PrintStatement{
				BasePos: ast.Position{Line: 1, Column: 5},
				Expressions: []ast.Expression{
					&ast.StringLiteral{BasePos: ast.Position{Line: 1, Column: 11}, Value: "hello"},
				},
			},
			&ast.GotoStatement{
				BasePos: ast.Position{Line: 2, Column: 1},
				Target:  "100",
			},
		},
	}

	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestResolverIfStatement(t *testing.T) {
	// IF x% > 0 THEN y% = 1 ELSE y% = 0
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LetStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Name:    &ast.Identifier{BasePos: ast.Position{Line: 1, Column: 1}, Name: "x", TypeSuffix: "%"},
				Value:   &ast.NumberLiteral{BasePos: ast.Position{Line: 1, Column: 6}, Value: 5, OriginalText: "5"},
			},
			&ast.IfStatement{
				BasePos: ast.Position{Line: 2, Column: 1},
				Condition: &ast.BinaryExpr{
					BasePos:  ast.Position{Line: 2, Column: 4},
					Left:     &ast.Identifier{BasePos: ast.Position{Line: 2, Column: 4}, Name: "x", TypeSuffix: "%"},
					Operator: ">",
					Right:    &ast.NumberLiteral{BasePos: ast.Position{Line: 2, Column: 9}, Value: 0, OriginalText: "0"},
				},
				ThenBlock: []ast.Statement{
					&ast.LetStatement{
						BasePos: ast.Position{Line: 2, Column: 16},
						Name:    &ast.Identifier{BasePos: ast.Position{Line: 2, Column: 16}, Name: "y", TypeSuffix: "%"},
						Value:   &ast.NumberLiteral{BasePos: ast.Position{Line: 2, Column: 21}, Value: 1, OriginalText: "1"},
					},
				},
				ElseBlock: []ast.Statement{
					&ast.LetStatement{
						BasePos: ast.Position{Line: 2, Column: 28},
						Name:    &ast.Identifier{BasePos: ast.Position{Line: 2, Column: 28}, Name: "y", TypeSuffix: "%"},
						Value:   &ast.NumberLiteral{BasePos: ast.Position{Line: 2, Column: 33}, Value: 0, OriginalText: "0"},
					},
				},
			},
		},
	}

	resolver := NewResolver(prog)
	table, errs := resolver.Resolve()
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if table.Lookup("x%") == nil {
		t.Error("x% should be in symbol table")
	}
	if table.Lookup("y%") == nil {
		t.Error("y% should be in symbol table")
	}
}

func TestResolverWhileStatement(t *testing.T) {
	// WHILE x > 0 / x = x - 1 / WEND
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LetStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Name:    &ast.Identifier{Name: "x", TypeSuffix: "%"},
				Value:   &ast.NumberLiteral{Value: 5, OriginalText: "5"},
			},
			&ast.WhileStatement{
				BasePos: ast.Position{Line: 2, Column: 1},
				Condition: &ast.BinaryExpr{
					Left:     &ast.Identifier{Name: "x", TypeSuffix: "%"},
					Operator: ">",
					Right:    &ast.NumberLiteral{Value: 0, OriginalText: "0"},
				},
				Body: []ast.Statement{
					&ast.LetStatement{
						BasePos: ast.Position{Line: 3, Column: 1},
						Name:    &ast.Identifier{Name: "x", TypeSuffix: "%"},
						Value: &ast.BinaryExpr{
							Left:     &ast.Identifier{Name: "x", TypeSuffix: "%"},
							Operator: "-",
							Right:    &ast.NumberLiteral{Value: 1, OriginalText: "1"},
						},
					},
				},
			},
		},
	}
	resolver := NewResolver(prog)
	table, errs := resolver.Resolve()
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if table.Lookup("x%") == nil {
		t.Error("x% should be in symbol table")
	}
}

func TestResolverDoLoopStatement(t *testing.T) {
	// DO WHILE x < 10 / x = x + 1 / LOOP
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LetStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Name:    &ast.Identifier{Name: "x", TypeSuffix: "%"},
				Value:   &ast.NumberLiteral{Value: 0, OriginalText: "0"},
			},
			&ast.DoLoopStatement{
				BasePos:   ast.Position{Line: 2, Column: 1},
				TestAtTop: true,
				Condition: &ast.BinaryExpr{
					Left:     &ast.Identifier{Name: "x", TypeSuffix: "%"},
					Operator: "<",
					Right:    &ast.NumberLiteral{Value: 10, OriginalText: "10"},
				},
				Body: []ast.Statement{
					&ast.IncrStatement{
						BasePos:  ast.Position{Line: 3, Column: 1},
						Variable: &ast.Identifier{Name: "x", TypeSuffix: "%"},
					},
				},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestResolverDoLoopNoCondition(t *testing.T) {
	// Infinite DO / LOOP
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.DoLoopStatement{
				BasePos:   ast.Position{Line: 1, Column: 1},
				Condition: nil,
				Body: []ast.Statement{
					&ast.LetStatement{
						BasePos: ast.Position{Line: 2, Column: 1},
						Name:    &ast.Identifier{Name: "n", TypeSuffix: "%"},
						Value:   &ast.NumberLiteral{Value: 1, OriginalText: "1"},
					},
				},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestResolverSelectCaseStatement(t *testing.T) {
	// SELECT CASE x% / CASE 1 / y% = 1 / CASE 2, 3 / y% = 2 / CASE ELSE / y% = 0 / END SELECT
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LetStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Name:    &ast.Identifier{Name: "x", TypeSuffix: "%"},
				Value:   &ast.NumberLiteral{Value: 2, OriginalText: "2"},
			},
			&ast.SelectCaseStatement{
				BasePos:  ast.Position{Line: 2, Column: 1},
				TestExpr: &ast.Identifier{Name: "x", TypeSuffix: "%"},
				Cases: []ast.CaseClause{
					{
						Values: []ast.CaseValue{{Value: &ast.NumberLiteral{Value: 1, OriginalText: "1"}}},
						Body: []ast.Statement{
							&ast.LetStatement{
								BasePos: ast.Position{},
								Name:    &ast.Identifier{Name: "y", TypeSuffix: "%"},
								Value:   &ast.NumberLiteral{Value: 1, OriginalText: "1"},
							},
						},
					},
				},
				ElseBlock: []ast.Statement{
					&ast.LetStatement{
						BasePos: ast.Position{},
						Name:    &ast.Identifier{Name: "y", TypeSuffix: "%"},
						Value:   &ast.NumberLiteral{Value: 0, OriginalText: "0"},
					},
				},
			},
		},
	}
	resolver := NewResolver(prog)
	table, errs := resolver.Resolve()
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if table.Lookup("y%") == nil {
		t.Error("y% should be in symbol table from CASE body")
	}
}

func TestResolverDefFnDeclaration(t *testing.T) {
	// DEF FNDouble(x!) = x! * 2
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.DefFnDeclaration{
				BasePos: ast.Position{Line: 1, Column: 1},
				Name:    "FNDouble",
				Params: []ast.Parameter{
					{Name: "x", Type: "!"},
				},
				SingleLineExpr: &ast.BinaryExpr{
					Left:     &ast.Identifier{Name: "x", TypeSuffix: "!"},
					Operator: "*",
					Right:    &ast.NumberLiteral{Value: 2, OriginalText: "2"},
				},
			},
		},
	}
	resolver := NewResolver(prog)
	table, errs := resolver.Resolve()
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	sym := table.Lookup("FNDouble")
	if sym == nil {
		t.Fatal("FNDouble should be in symbol table")
	}
	if sym.Type != SymDefFn {
		t.Errorf("expected SymDefFn, got %v", sym.Type)
	}
}

func TestResolverOnErrorGotoStatement(t *testing.T) {
	// ON ERROR GOTO handler / handler:
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.OnErrorGotoStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Target:  "handler",
			},
			&ast.LabelStatement{
				BasePos: ast.Position{Line: 2, Column: 1},
				Name:    "handler",
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestResolverOnEventGosubStatement(t *testing.T) {
	// ON KEY(1) GOSUB handler / handler:
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.OnEventGosubStatement{
				BasePos:    ast.Position{Line: 1, Column: 1},
				EventType:  "KEY",
				EventParam: &ast.NumberLiteral{Value: 1, OriginalText: "1"},
				Target:     "handler",
			},
			&ast.LabelStatement{
				BasePos: ast.Position{Line: 2, Column: 1},
				Name:    "handler",
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestResolverArrayAssignment(t *testing.T) {
	// DIM a%(10) / a%(5) = 42
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.DimStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Declarations: []ast.DimDecl{
					{
						Name:       "a",
						TypeSuffix: "%",
						Dimensions: []ast.DimRange{
							{Upper: &ast.NumberLiteral{Value: 10, OriginalText: "10"}},
						},
					},
				},
			},
			&ast.ArrayAssignment{
				BasePos: ast.Position{Line: 2, Column: 1},
				Array: &ast.ArrayAccess{
					Name:       "a",
					TypeSuffix: "%",
					Indices:    []ast.Expression{&ast.NumberLiteral{Value: 5, OriginalText: "5"}},
				},
				Value: &ast.NumberLiteral{Value: 42, OriginalText: "42"},
			},
		},
	}
	resolver := NewResolver(prog)
	table, errs := resolver.Resolve()
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	sym := table.Lookup("a%")
	if sym == nil {
		t.Fatal("a% should be in symbol table")
	}
	if !sym.Used {
		t.Error("a% should be marked as Used after array assignment")
	}
}

func TestResolverReadStatement(t *testing.T) {
	// DATA 1, 2 / READ x%, y%
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.DataStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Values: []ast.Expression{
					&ast.NumberLiteral{Value: 1, OriginalText: "1"},
					&ast.NumberLiteral{Value: 2, OriginalText: "2"},
				},
			},
			&ast.ReadStatement{
				BasePos: ast.Position{Line: 2, Column: 1},
				Variables: []ast.Expression{
					&ast.Identifier{Name: "x", TypeSuffix: "%"},
					&ast.Identifier{Name: "y", TypeSuffix: "%"},
				},
			},
		},
	}
	resolver := NewResolver(prog)
	table, errs := resolver.Resolve()
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if table.Lookup("x%") == nil {
		t.Error("x% should be in symbol table after READ")
	}
}

func TestResolverSwapStatement(t *testing.T) {
	// a% = 10 / b% = 20 / SWAP a%, b%
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LetStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Name:    &ast.Identifier{Name: "a", TypeSuffix: "%"},
				Value:   &ast.NumberLiteral{Value: 10, OriginalText: "10"},
			},
			&ast.LetStatement{
				BasePos: ast.Position{Line: 2, Column: 1},
				Name:    &ast.Identifier{Name: "b", TypeSuffix: "%"},
				Value:   &ast.NumberLiteral{Value: 20, OriginalText: "20"},
			},
			&ast.SwapStatement{
				BasePos: ast.Position{Line: 3, Column: 1},
				Var1:    &ast.Identifier{Name: "a", TypeSuffix: "%"},
				Var2:    &ast.Identifier{Name: "b", TypeSuffix: "%"},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestResolverIncrStatement(t *testing.T) {
	// x% = 5 / INCR x%, 2
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LetStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Name:    &ast.Identifier{Name: "x", TypeSuffix: "%"},
				Value:   &ast.NumberLiteral{Value: 5, OriginalText: "5"},
			},
			&ast.IncrStatement{
				BasePos:  ast.Position{Line: 2, Column: 1},
				Variable: &ast.Identifier{Name: "x", TypeSuffix: "%"},
				Amount:   &ast.NumberLiteral{Value: 2, OriginalText: "2"},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestResolverDecrStatement(t *testing.T) {
	// x% = 10 / DECR x%
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LetStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Name:    &ast.Identifier{Name: "x", TypeSuffix: "%"},
				Value:   &ast.NumberLiteral{Value: 10, OriginalText: "10"},
			},
			&ast.DecrStatement{
				BasePos:  ast.Position{Line: 2, Column: 1},
				Variable: &ast.Identifier{Name: "x", TypeSuffix: "%"},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestResolverRedimStatement(t *testing.T) {
	// DIM arr%(10) / REDIM arr%(20)
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.DimStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Declarations: []ast.DimDecl{
					{
						Name:       "arr",
						TypeSuffix: "%",
						Dimensions: []ast.DimRange{
							{Upper: &ast.NumberLiteral{Value: 10, OriginalText: "10"}},
						},
					},
				},
			},
			&ast.RedimStatement{
				BasePos: ast.Position{Line: 2, Column: 1},
				Declarations: []ast.DimDecl{
					{
						Name:       "arr",
						TypeSuffix: "%",
						Dimensions: []ast.DimRange{
							{Upper: &ast.NumberLiteral{Value: 20, OriginalText: "20"}},
						},
					},
				},
			},
		},
	}
	resolver := NewResolver(prog)
	table, errs := resolver.Resolve()
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	sym := table.Lookup("arr%")
	if sym == nil {
		t.Fatal("arr% should be in symbol table")
	}
	if sym.ArrayDims != 1 {
		t.Errorf("expected 1 dim after REDIM, got %d", sym.ArrayDims)
	}
}

func TestResolverRedimNewArray(t *testing.T) {
	// REDIM of undeclared array creates it
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.RedimStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Declarations: []ast.DimDecl{
					{
						Name:       "newArr",
						TypeSuffix: "#",
						Dimensions: []ast.DimRange{
							{Upper: &ast.NumberLiteral{Value: 50, OriginalText: "50"}},
						},
					},
				},
			},
		},
	}
	resolver := NewResolver(prog)
	table, errs := resolver.Resolve()
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	sym := table.Lookup("newArr#")
	if sym == nil {
		t.Fatal("newArr# should be created by REDIM")
	}
	if sym.DataType != TypeDouble {
		t.Errorf("expected TypeDouble for # suffix, got %v", sym.DataType)
	}
}

// ===========================================================================
// File I/O resolver tests
// ===========================================================================

func TestResolverOpenStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.OpenStatement{
				BasePos:  ast.Position{Line: 1, Column: 1},
				Filename: &ast.StringLiteral{Value: "test.dat"},
				Mode:     "INPUT",
				FileNum:  &ast.NumberLiteral{Value: 1, OriginalText: "1"},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestResolverOpenWithRecLen(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.OpenStatement{
				BasePos:  ast.Position{Line: 1, Column: 1},
				Filename: &ast.StringLiteral{Value: "data.dat"},
				Mode:     "RANDOM",
				FileNum:  &ast.NumberLiteral{Value: 1, OriginalText: "1"},
				RecLen:   &ast.NumberLiteral{Value: 128, OriginalText: "128"},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestResolverCloseStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.CloseStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				FileNums: []ast.Expression{
					&ast.NumberLiteral{Value: 1, OriginalText: "1"},
				},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestResolverFileInputStatement(t *testing.T) {
	// INPUT #1, x$, y%
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.FileInputStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				FileNum: &ast.NumberLiteral{Value: 1, OriginalText: "1"},
				Variables: []ast.Expression{
					&ast.Identifier{Name: "x", TypeSuffix: "$"},
					&ast.Identifier{Name: "y", TypeSuffix: "%"},
				},
			},
		},
	}
	resolver := NewResolver(prog)
	table, errs := resolver.Resolve()
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if table.Lookup("x$") == nil {
		t.Error("x$ should be in symbol table after INPUT#")
	}
}

func TestResolverFilePrintStatement(t *testing.T) {
	// PRINT #1, x$
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LetStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Name:    &ast.Identifier{Name: "x", TypeSuffix: "$"},
				Value:   &ast.StringLiteral{Value: "hello"},
			},
			&ast.FilePrintStatement{
				BasePos: ast.Position{Line: 2, Column: 1},
				FileNum: &ast.NumberLiteral{Value: 1, OriginalText: "1"},
				Expressions: []ast.Expression{
					&ast.Identifier{Name: "x", TypeSuffix: "$"},
				},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestResolverFilePrintWithFormat(t *testing.T) {
	// PRINT #1, USING "##"; x%
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.FilePrintStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				FileNum: &ast.NumberLiteral{Value: 1, OriginalText: "1"},
				Expressions: []ast.Expression{
					&ast.NumberLiteral{Value: 42, OriginalText: "42"},
				},
				Format: &ast.StringLiteral{Value: "##"},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestResolverFileWriteStatement(t *testing.T) {
	// WRITE #1, x, y$
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.FileWriteStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				FileNum: &ast.NumberLiteral{Value: 1, OriginalText: "1"},
				Expressions: []ast.Expression{
					&ast.Identifier{Name: "x"},
					&ast.Identifier{Name: "y", TypeSuffix: "$"},
				},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestResolverGetStatement(t *testing.T) {
	// GET #1, recNum
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.GetStatement{
				BasePos:     ast.Position{Line: 1, Column: 1},
				FileNum:     &ast.NumberLiteral{Value: 1, OriginalText: "1"},
				RecordOrPos: &ast.Identifier{Name: "recNum"},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestResolverPutStatement(t *testing.T) {
	// PUT #1, recNum
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.PutStatement{
				BasePos:     ast.Position{Line: 1, Column: 1},
				FileNum:     &ast.NumberLiteral{Value: 1, OriginalText: "1"},
				RecordOrPos: &ast.Identifier{Name: "recNum"},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

func TestResolverSeekStatement(t *testing.T) {
	// SEEK #1, pos
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.SeekStatement{
				BasePos:  ast.Position{Line: 1, Column: 1},
				FileNum:  &ast.NumberLiteral{Value: 1, OriginalText: "1"},
				Position: &ast.NumberLiteral{Value: 100, OriginalText: "100"},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
}

// ===========================================================================
// defineGlobal tests (from semantic_coverage_test.go)
// ===========================================================================

func TestDefineGlobalSameTypeNoOp(t *testing.T) {
	// registerForwardDecl followed by a second registration of same type → no error.
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.SubDeclaration{
				BasePos:   ast.Position{Line: 1, Column: 1},
				Name:      "DupSub",
				IsForward: true,
				Body:      []ast.Statement{},
			},
			&ast.SubDeclaration{
				BasePos: ast.Position{Line: 2, Column: 1},
				Name:    "DupSub",
				Body:    []ast.Statement{},
			},
		},
	}
	r := NewResolver(prog)
	_, errs := r.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for forward decl + definition of same type, got: %v", errs)
	}
}

func TestDefineGlobalDifferentTypeError(t *testing.T) {
	// First declare as SUB, then try to declare as FUNCTION → duplicate error.
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.SubDeclaration{
				BasePos: ast.Position{Line: 1, Column: 1},
				Name:    "Conflict",
				Body:    []ast.Statement{},
			},
			&ast.FunctionDeclaration{
				BasePos:    ast.Position{Line: 2, Column: 1},
				Name:       "Conflict",
				ReturnType: "%",
				Body:       []ast.Statement{},
			},
		},
	}
	r := NewResolver(prog)
	_, errs := r.Resolve()
	if len(errs) == 0 {
		t.Fatal("expected duplicate definition error for SUB/FUNCTION with same name")
	}
	found := false
	for _, e := range errs {
		if strings.Contains(e, "duplicate definition") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected 'duplicate definition' error, got: %v", errs)
	}
}
