package semantic

import (
	"strings"
	"testing"

	"github.com/loabletech/go-basic/internal/ast"
)

// ===========================================================================
// typechecker.go gap coverage
// ===========================================================================

// checkPrintStatement (0% → covered)

func TestCheckPrintStatementNoExpressions(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.PrintStatement{
				BasePos:     ast.Position{Line: 1, Column: 1},
				Expressions: []ast.Expression{},
			},
		},
	}
	errs := tc.Check(prog)
	if len(errs) != 0 {
		t.Fatalf("expected no errors for empty PRINT, got: %v", errs)
	}
}

func TestCheckPrintStatementWithExpressions(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.PrintStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Expressions: []ast.Expression{
					&ast.NumberLiteral{Value: 42, NumType: ast.NumInt},
					&ast.StringLiteral{Value: "hello"},
				},
			},
		},
	}
	errs := tc.Check(prog)
	if len(errs) != 0 {
		t.Fatalf("expected no errors for PRINT with mixed expressions, got: %v", errs)
	}
}

func TestCheckPrintStatementWithStringFormat(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.PrintStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Expressions: []ast.Expression{
					&ast.NumberLiteral{Value: 3.14, NumType: ast.NumSingle},
				},
				Format: &ast.StringLiteral{BasePos: ast.Position{Line: 1, Column: 10}, Value: "##.##"},
			},
		},
	}
	errs := tc.Check(prog)
	if len(errs) != 0 {
		t.Fatalf("expected no errors for PRINT USING with string format, got: %v", errs)
	}
}

func TestCheckPrintStatementWithNumericFormatError(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.PrintStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Expressions: []ast.Expression{
					&ast.NumberLiteral{Value: 42, NumType: ast.NumInt},
				},
				Format: &ast.NumberLiteral{BasePos: ast.Position{Line: 1, Column: 10}, Value: 123, NumType: ast.NumInt},
			},
		},
	}
	errs := tc.Check(prog)
	if len(errs) == 0 {
		t.Fatal("expected error for PRINT USING with non-string format")
	}
	if !strings.Contains(errs[0], "format must be a string") {
		t.Fatalf("expected format error, got: %s", errs[0])
	}
}

// checkIfStatement ElseIf clauses (60% → covered)

func TestCheckIfStatementElseIfNumeric(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.IfStatement{
				BasePos:   ast.Position{Line: 1, Column: 1},
				Condition: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				ThenBlock: []ast.Statement{},
				ElseIfClauses: []ast.ElseIfClause{
					{
						BasePos:   ast.Position{Line: 2, Column: 1},
						Condition: &ast.NumberLiteral{Value: 0, NumType: ast.NumInt},
						Body:      []ast.Statement{},
					},
				},
				ElseBlock: []ast.Statement{},
			},
		},
	}
	errs := tc.Check(prog)
	if len(errs) != 0 {
		t.Fatalf("expected no errors for IF with numeric ELSEIF, got: %v", errs)
	}
}

func TestCheckIfStatementElseIfStringConditionError(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.IfStatement{
				BasePos:   ast.Position{Line: 1, Column: 1},
				Condition: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				ThenBlock: []ast.Statement{},
				ElseIfClauses: []ast.ElseIfClause{
					{
						BasePos:   ast.Position{Line: 2, Column: 1},
						Condition: &ast.StringLiteral{Value: "yes"},
						Body:      []ast.Statement{},
					},
				},
				ElseBlock: []ast.Statement{},
			},
		},
	}
	errs := tc.Check(prog)
	if len(errs) == 0 {
		t.Fatal("expected error for ELSEIF with string condition")
	}
	if !strings.Contains(errs[0], "ELSEIF condition must be numeric") {
		t.Fatalf("expected ELSEIF condition error, got: %s", errs[0])
	}
}

func TestCheckIfStatementStringConditionError(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.IfStatement{
				BasePos:   ast.Position{Line: 1, Column: 1},
				Condition: &ast.StringLiteral{Value: "bad"},
				ThenBlock: []ast.Statement{},
			},
		},
	}
	errs := tc.Check(prog)
	if len(errs) == 0 {
		t.Fatal("expected error for IF with string condition")
	}
}

// checkForStatement step != nil and string errors (66.7% → covered)

func TestCheckForStatementWithNumericStep(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ForStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Counter: &ast.Identifier{Name: "i"},
				Start:   &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				End:     &ast.NumberLiteral{Value: 10, NumType: ast.NumInt},
				Step:    &ast.NumberLiteral{Value: 2, NumType: ast.NumInt},
				Body:    []ast.Statement{},
			},
		},
	}
	errs := tc.Check(prog)
	if len(errs) != 0 {
		t.Fatalf("expected no errors for FOR with numeric STEP, got: %v", errs)
	}
}

func TestCheckForStatementWithStringStep(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ForStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Counter: &ast.Identifier{Name: "i"},
				Start:   &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				End:     &ast.NumberLiteral{Value: 10, NumType: ast.NumInt},
				Step:    &ast.StringLiteral{Value: "bad"},
				Body:    []ast.Statement{},
			},
		},
	}
	errs := tc.Check(prog)
	if len(errs) == 0 {
		t.Fatal("expected error for FOR with string STEP")
	}
	if !strings.Contains(errs[0], "FOR STEP value must be numeric") {
		t.Fatalf("expected STEP error, got: %s", errs[0])
	}
}

func TestCheckForStatementStringCounterError(t *testing.T) {
	// counter name ending with $ → TypeString → error
	st := buildSymTable(map[string]*Symbol{
		"I$": {Name: "I$", Type: SymVariable, DataType: TypeString},
	})
	tc := NewTypeChecker(st)
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ForStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Counter: &ast.Identifier{Name: "I$"},
				Start:   &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				End:     &ast.NumberLiteral{Value: 10, NumType: ast.NumInt},
				Body:    []ast.Statement{},
			},
		},
	}
	errs := tc.Check(prog)
	if len(errs) == 0 {
		t.Fatal("expected error for FOR with string counter variable")
	}
	if !strings.Contains(errs[0], "FOR counter variable must be numeric") {
		t.Fatalf("expected counter error, got: %s", errs[0])
	}
}

func TestCheckForStatementStringEndError(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ForStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Counter: &ast.Identifier{Name: "i"},
				Start:   &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				End:     &ast.StringLiteral{Value: "bad"},
				Body:    []ast.Statement{},
			},
		},
	}
	errs := tc.Check(prog)
	if len(errs) == 0 {
		t.Fatal("expected error for FOR with string end value")
	}
}

// checkSelectCase IsRange+EndValue covered (75% → covered)

func TestCheckSelectCaseWithRange(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.SelectCaseStatement{
				BasePos:  ast.Position{Line: 1, Column: 1},
				TestExpr: &ast.NumberLiteral{Value: 5, NumType: ast.NumInt},
				Cases: []ast.CaseClause{
					{
						Values: []ast.CaseValue{
							{
								BasePos:  ast.Position{Line: 2, Column: 1},
								Value:    &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
								IsRange:  true,
								EndValue: &ast.NumberLiteral{Value: 10, NumType: ast.NumInt},
							},
						},
						Body: []ast.Statement{},
					},
				},
			},
		},
	}
	errs := tc.Check(prog)
	if len(errs) != 0 {
		t.Fatalf("expected no errors for CASE 1 TO 10 with numeric test, got: %v", errs)
	}
}

func TestCheckSelectCaseRangeTypeMismatch(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	// numeric test, range end is string → mismatch
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.SelectCaseStatement{
				BasePos:  ast.Position{Line: 1, Column: 1},
				TestExpr: &ast.NumberLiteral{Value: 5, NumType: ast.NumInt},
				Cases: []ast.CaseClause{
					{
						Values: []ast.CaseValue{
							{
								BasePos:  ast.Position{Line: 2, Column: 1},
								Value:    &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
								IsRange:  true,
								EndValue: &ast.StringLiteral{Value: "bad"},
							},
						},
						Body: []ast.Statement{},
					},
				},
			},
		},
	}
	errs := tc.Check(prog)
	if len(errs) == 0 {
		t.Fatal("expected type mismatch error for CASE range with string end value")
	}
	if !strings.Contains(errs[0], "CASE TO") {
		t.Fatalf("expected CASE TO error, got: %s", errs[0])
	}
}

func TestCheckSelectCaseElseBlock(t *testing.T) {
	// SELECT CASE with CASE ELSE block containing invalid assignment
	st := buildSymTable(map[string]*Symbol{
		"N%": {Name: "N%", Type: SymVariable, DataType: TypeInteger},
	})
	tc2 := NewTypeChecker(st)
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.SelectCaseStatement{
				BasePos:  ast.Position{Line: 1, Column: 1},
				TestExpr: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				Cases:    []ast.CaseClause{},
				ElseBlock: []ast.Statement{
					&ast.LetStatement{
						BasePos: ast.Position{Line: 2, Column: 1},
						Name:    &ast.Identifier{Name: "N", TypeSuffix: "%"},
						Value:   &ast.StringLiteral{Value: "bad"},
					},
				},
			},
		},
	}
	errs := tc2.Check(prog)
	if len(errs) == 0 {
		t.Fatal("expected type error in CASE ELSE block")
	}
}

// checkIncrDecr amount != nil and string amount error (57.1% → covered)

func TestCheckIncrDecrWithNumericAmount(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.IncrStatement{
				BasePos:  ast.Position{Line: 1, Column: 1},
				Variable: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				Amount:   &ast.NumberLiteral{Value: 5, NumType: ast.NumInt},
			},
		},
	}
	errs := tc.Check(prog)
	if len(errs) != 0 {
		t.Fatalf("expected no errors for INCR with numeric amount, got: %v", errs)
	}
}

func TestCheckIncrDecrWithStringAmountError(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.IncrStatement{
				BasePos:  ast.Position{Line: 1, Column: 1},
				Variable: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				Amount:   &ast.StringLiteral{Value: "bad"},
			},
		},
	}
	errs := tc.Check(prog)
	if len(errs) == 0 {
		t.Fatal("expected error for INCR with string amount")
	}
	if !strings.Contains(errs[0], "amount must be numeric") {
		t.Fatalf("expected amount error, got: %s", errs[0])
	}
}

func TestCheckDecrWithStringAmountError(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.DecrStatement{
				BasePos:  ast.Position{Line: 1, Column: 1},
				Variable: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				Amount:   &ast.StringLiteral{Value: "bad"},
			},
		},
	}
	errs := tc.Check(prog)
	if len(errs) == 0 {
		t.Fatal("expected error for DECR with string amount")
	}
}

func TestCheckDecrWithNumericAmount(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.DecrStatement{
				BasePos:  ast.Position{Line: 1, Column: 1},
				Variable: &ast.NumberLiteral{Value: 5, NumType: ast.NumInt},
				Amount:   &ast.NumberLiteral{Value: 2, NumType: ast.NumInt},
			},
		},
	}
	errs := tc.Check(prog)
	if len(errs) != 0 {
		t.Fatalf("expected no errors for DECR with numeric amount, got: %v", errs)
	}
}

// checkExpression nil case (66.7% → covered)

func TestCheckExpressionNilCase(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	// checkExpression is called via WhileStatement - pass nil condition through DoLoop
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.DoLoopStatement{
				BasePos:   ast.Position{Line: 1, Column: 1},
				Condition: nil,
				Body:      []ast.Statement{},
			},
		},
	}
	errs := tc.Check(prog)
	if len(errs) != 0 {
		t.Fatalf("expected no errors for nil condition, got: %v", errs)
	}
}

// widenType tests (75% → covered)

func TestWidenTypeBothUnknown(t *testing.T) {
	got := widenType(TypeUnknown, TypeUnknown)
	if got != TypeUnknown {
		t.Errorf("widenType(Unknown, Unknown): expected Unknown, got %v", got)
	}
}

func TestWidenTypeAUnknown(t *testing.T) {
	got := widenType(TypeUnknown, TypeInteger)
	if got != TypeInteger {
		t.Errorf("widenType(Unknown, Integer): expected Integer, got %v", got)
	}
}

func TestWidenTypeBUnknown(t *testing.T) {
	got := widenType(TypeDouble, TypeUnknown)
	if got != TypeDouble {
		t.Errorf("widenType(Double, Unknown): expected Double, got %v", got)
	}
}

func TestWidenTypeIntegerAndLong(t *testing.T) {
	// order[Integer]=0, order[Long]=1 → Long is wider
	got := widenType(TypeInteger, TypeLong)
	if got != TypeLong {
		t.Errorf("widenType(Integer, Long): expected Long, got %v", got)
	}
}

func TestWidenTypeLongAndInteger(t *testing.T) {
	// order[Long]=1, order[Integer]=0 → Long is wider (a >= b case)
	got := widenType(TypeLong, TypeInteger)
	if got != TypeLong {
		t.Errorf("widenType(Long, Integer): expected Long, got %v", got)
	}
}

func TestWidenTypeSingleAndDouble(t *testing.T) {
	got := widenType(TypeSingle, TypeDouble)
	if got != TypeDouble {
		t.Errorf("widenType(Single, Double): expected Double, got %v", got)
	}
}

func TestWidenTypeDoubleAndSingle(t *testing.T) {
	got := widenType(TypeDouble, TypeSingle)
	if got != TypeDouble {
		t.Errorf("widenType(Double, Single): expected Double, got %v", got)
	}
}

func TestWidenTypeSameSingle(t *testing.T) {
	got := widenType(TypeSingle, TypeSingle)
	if got != TypeSingle {
		t.Errorf("widenType(Single, Single): expected Single, got %v", got)
	}
}

// isAssignmentCompatible tests (71.4% → covered)

func TestIsAssignmentCompatibleStringToNumeric(t *testing.T) {
	got := isAssignmentCompatible(TypeInteger, TypeString)
	if got {
		t.Error("isAssignmentCompatible(Integer, String): expected false")
	}
}

func TestIsAssignmentCompatibleNumericToString(t *testing.T) {
	got := isAssignmentCompatible(TypeString, TypeInteger)
	if got {
		t.Error("isAssignmentCompatible(String, Integer): expected false")
	}
}

func TestIsAssignmentCompatibleStringToString(t *testing.T) {
	got := isAssignmentCompatible(TypeString, TypeString)
	if !got {
		t.Error("isAssignmentCompatible(String, String): expected true")
	}
}

func TestIsAssignmentCompatibleNumericToNumeric(t *testing.T) {
	got := isAssignmentCompatible(TypeInteger, TypeDouble)
	if !got {
		t.Error("isAssignmentCompatible(Integer, Double): expected true")
	}
}

func TestIsAssignmentCompatibleUnknownVar(t *testing.T) {
	got := isAssignmentCompatible(TypeUnknown, TypeString)
	if !got {
		t.Error("isAssignmentCompatible(Unknown, String): expected true (unknown → permissive)")
	}
}

func TestIsAssignmentCompatibleUnknownExpr(t *testing.T) {
	got := isAssignmentCompatible(TypeInteger, TypeUnknown)
	if !got {
		t.Error("isAssignmentCompatible(Integer, Unknown): expected true (unknown → permissive)")
	}
}

// DataTypeName tests (57.1% → covered)

func TestDataTypeNameInteger(t *testing.T) {
	if got := DataTypeName(TypeInteger); got != "INTEGER" {
		t.Errorf("expected INTEGER, got %s", got)
	}
}

func TestDataTypeNameLong(t *testing.T) {
	if got := DataTypeName(TypeLong); got != "LONG" {
		t.Errorf("expected LONG, got %s", got)
	}
}

func TestDataTypeNameSingle(t *testing.T) {
	if got := DataTypeName(TypeSingle); got != "SINGLE" {
		t.Errorf("expected SINGLE, got %s", got)
	}
}

func TestDataTypeNameDouble(t *testing.T) {
	if got := DataTypeName(TypeDouble); got != "DOUBLE" {
		t.Errorf("expected DOUBLE, got %s", got)
	}
}

func TestDataTypeNameString(t *testing.T) {
	if got := DataTypeName(TypeString); got != "STRING" {
		t.Errorf("expected STRING, got %s", got)
	}
}

func TestDataTypeNameUnknown(t *testing.T) {
	if got := DataTypeName(TypeUnknown); got != "UNKNOWN" {
		t.Errorf("expected UNKNOWN, got %s", got)
	}
}

// ===========================================================================
// symbols.go gap coverage
// ===========================================================================

// resolveStatement – graphics/sound/file statements (39.5% → covered)

func TestResolveStatementGetStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.GetStatement{
				BasePos:     ast.Position{Line: 1, Column: 1},
				FileNum:     &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				RecordOrPos: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				Variable:    nil,
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for GET statement, got: %v", errs)
	}
}

func TestResolveStatementPutStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.PutStatement{
				BasePos:     ast.Position{Line: 1, Column: 1},
				FileNum:     &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				RecordOrPos: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				Variable:    nil,
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for PUT statement, got: %v", errs)
	}
}

func TestResolveStatementSeekStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.SeekStatement{
				BasePos:  ast.Position{Line: 1, Column: 1},
				FileNum:  &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				Position: &ast.NumberLiteral{Value: 100, NumType: ast.NumInt},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for SEEK statement, got: %v", errs)
	}
}

func TestResolveStatementErrorStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ErrorStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Code:    &ast.NumberLiteral{Value: 5, NumType: ast.NumInt},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for ERROR statement, got: %v", errs)
	}
}

func TestResolveStatementLabelStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LabelStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Name:    "myLabel",
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for label statement, got: %v", errs)
	}
}

func TestResolveStatementEndStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.EndStatement{BasePos: ast.Position{Line: 1, Column: 1}},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for END statement, got: %v", errs)
	}
}

func TestResolveStatementBeepStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.BeepStatement{BasePos: ast.Position{Line: 1, Column: 1}},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for BEEP statement, got: %v", errs)
	}
}

func TestResolveStatementEraseStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.EraseStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Names:   []string{"myArr"},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for ERASE statement, got: %v", errs)
	}
}

func TestResolveStatementScreenStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ScreenStatement{
				BasePos:     ast.Position{Line: 1, Column: 1},
				Mode:        &ast.NumberLiteral{Value: 13, NumType: ast.NumInt},
				ColorSwitch: nil,
				ActivePage:  nil,
				VisualPage:  nil,
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for SCREEN statement, got: %v", errs)
	}
}

func TestResolveStatementColorStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ColorStatement{
				BasePos:    ast.Position{Line: 1, Column: 1},
				Foreground: &ast.NumberLiteral{Value: 7, NumType: ast.NumInt},
				Background: &ast.NumberLiteral{Value: 0, NumType: ast.NumInt},
				Border:     nil,
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for COLOR statement, got: %v", errs)
	}
}

func TestResolveStatementPsetStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.PsetStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				X:       &ast.NumberLiteral{Value: 100, NumType: ast.NumInt},
				Y:       &ast.NumberLiteral{Value: 100, NumType: ast.NumInt},
				Color:   &ast.NumberLiteral{Value: 15, NumType: ast.NumInt},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for PSET statement, got: %v", errs)
	}
}

func TestResolveStatementLineStmt(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LineStmt{
				BasePos: ast.Position{Line: 1, Column: 1},
				X1:      &ast.NumberLiteral{Value: 0, NumType: ast.NumInt},
				Y1:      &ast.NumberLiteral{Value: 0, NumType: ast.NumInt},
				X2:      &ast.NumberLiteral{Value: 100, NumType: ast.NumInt},
				Y2:      &ast.NumberLiteral{Value: 100, NumType: ast.NumInt},
				Color:   &ast.NumberLiteral{Value: 15, NumType: ast.NumInt},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for LINE statement, got: %v", errs)
	}
}

func TestResolveStatementCircleStmt(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.CircleStmt{
				BasePos: ast.Position{Line: 1, Column: 1},
				X:       &ast.NumberLiteral{Value: 160, NumType: ast.NumInt},
				Y:       &ast.NumberLiteral{Value: 100, NumType: ast.NumInt},
				Radius:  &ast.NumberLiteral{Value: 50, NumType: ast.NumInt},
				Color:   &ast.NumberLiteral{Value: 15, NumType: ast.NumInt},
				Start:   nil,
				End:     nil,
				Aspect:  nil,
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for CIRCLE statement, got: %v", errs)
	}
}

func TestResolveStatementPaintStmt(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.PaintStmt{
				BasePos:     ast.Position{Line: 1, Column: 1},
				X:           &ast.NumberLiteral{Value: 100, NumType: ast.NumInt},
				Y:           &ast.NumberLiteral{Value: 100, NumType: ast.NumInt},
				FillColor:   &ast.NumberLiteral{Value: 4, NumType: ast.NumInt},
				BorderColor: &ast.NumberLiteral{Value: 15, NumType: ast.NumInt},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for PAINT statement, got: %v", errs)
	}
}

func TestResolveStatementDrawStmt(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.DrawStmt{
				BasePos:       ast.Position{Line: 1, Column: 1},
				CommandString: &ast.StringLiteral{Value: "U10 R10 D10 L10"},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for DRAW statement, got: %v", errs)
	}
}

func TestResolveStatementSoundStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.SoundStatement{
				BasePos:   ast.Position{Line: 1, Column: 1},
				Frequency: &ast.NumberLiteral{Value: 440, NumType: ast.NumSingle},
				Duration:  &ast.NumberLiteral{Value: 18, NumType: ast.NumSingle},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for SOUND statement, got: %v", errs)
	}
}

func TestResolveStatementPlayStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.PlayStatement{
				BasePos:       ast.Position{Line: 1, Column: 1},
				CommandString: &ast.StringLiteral{Value: "CDE"},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for PLAY statement, got: %v", errs)
	}
}

func TestResolveStatementLocateStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LocateStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Row:     &ast.NumberLiteral{Value: 10, NumType: ast.NumInt},
				Col:     &ast.NumberLiteral{Value: 20, NumType: ast.NumInt},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for LOCATE statement, got: %v", errs)
	}
}

func TestResolveStatementClsStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ClsStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Mode:    &ast.NumberLiteral{Value: 0, NumType: ast.NumInt},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for CLS statement, got: %v", errs)
	}
}

func TestResolveStatementViewStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ViewStatement{
				BasePos:     ast.Position{Line: 1, Column: 1},
				X1:          &ast.NumberLiteral{Value: 0, NumType: ast.NumInt},
				Y1:          &ast.NumberLiteral{Value: 0, NumType: ast.NumInt},
				X2:          &ast.NumberLiteral{Value: 319, NumType: ast.NumInt},
				Y2:          &ast.NumberLiteral{Value: 199, NumType: ast.NumInt},
				FillColor:   nil,
				BorderColor: nil,
				Top:         nil,
				Bottom:      nil,
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for VIEW statement, got: %v", errs)
	}
}

func TestResolveStatementFieldStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.FieldStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				FileNum: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				Fields: []ast.FieldDef{
					{
						Length:  &ast.NumberLiteral{Value: 30, NumType: ast.NumInt},
						VarName: "name$",
					},
				},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for FIELD statement, got: %v", errs)
	}
}

func TestResolveStatementLsetStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LsetStatement{
				BasePos:  ast.Position{Line: 1, Column: 1},
				Variable: "name$",
				Value:    &ast.StringLiteral{Value: "hello"},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for LSET statement, got: %v", errs)
	}
}

func TestResolveStatementRsetStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.RsetStatement{
				BasePos:  ast.Position{Line: 1, Column: 1},
				Variable: "name$",
				Value:    &ast.StringLiteral{Value: "hello"},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for RSET statement, got: %v", errs)
	}
}

func TestResolveStatementKillStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.KillStatement{
				BasePos:  ast.Position{Line: 1, Column: 1},
				Filename: &ast.StringLiteral{Value: "temp.dat"},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for KILL statement, got: %v", errs)
	}
}

func TestResolveStatementNameStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.NameStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				OldName: &ast.StringLiteral{Value: "old.dat"},
				NewName: &ast.StringLiteral{Value: "new.dat"},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for NAME statement, got: %v", errs)
	}
}

func TestResolveStatementChdirStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ChdirStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Path:    &ast.StringLiteral{Value: "/tmp"},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for CHDIR statement, got: %v", errs)
	}
}

func TestResolveStatementMkdirStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.MkdirStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Path:    &ast.StringLiteral{Value: "/tmp/newdir"},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for MKDIR statement, got: %v", errs)
	}
}

func TestResolveStatementRmdirStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.RmdirStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Path:    &ast.StringLiteral{Value: "/tmp/olddir"},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for RMDIR statement, got: %v", errs)
	}
}

// defineGlobal – same type no-op and different type error (42.9% → covered)

func TestDefineGlobalSameTypeNoOp(t *testing.T) {
	// Two SUB declarations with the same name and same type → no duplicate error
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.SubDeclaration{
				BasePos:   ast.Position{Line: 1, Column: 1},
				Name:      "MySub",
				IsForward: true,
			},
			&ast.SubDeclaration{
				BasePos:   ast.Position{Line: 5, Column: 1},
				Name:      "MySub",
				IsForward: false,
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for forward then actual sub declaration, got: %v", errs)
	}
}

func TestDefineGlobalDifferentTypeError(t *testing.T) {
	// Register a SUB and a FUNCTION with same name → type conflict
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.SubDeclaration{
				BasePos:   ast.Position{Line: 1, Column: 1},
				Name:      "Conflict",
				IsForward: false,
			},
			&ast.FunctionDeclaration{
				BasePos:   ast.Position{Line: 5, Column: 1},
				Name:      "Conflict",
				IsForward: false,
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) == 0 {
		t.Fatal("expected error for SUB and FUNCTION with same name")
	}
	found := false
	for _, e := range errs {
		if strings.Contains(e, "duplicate") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected duplicate definition error, got: %v", errs)
	}
}

// resolveDefType – all variants (50% → covered)

func TestResolveDefTypeDefInt(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.DefTypeStatement{
				BasePos:      ast.Position{Line: 1, Column: 1},
				Type:         "DEFINT",
				LetterRanges: []ast.LetterRange{{Start: 'I', End: 'N'}},
			},
		},
	}
	resolver := NewResolver(prog)
	table, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for DEFINT, got: %v", errs)
	}
	if got := table.ResolveType("ivar"); got != TypeInteger {
		t.Errorf("expected TypeInteger for ivar after DEFINT I-N, got %v", got)
	}
}

func TestResolveDefTypeDefLng(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.DefTypeStatement{
				BasePos:      ast.Position{Line: 1, Column: 1},
				Type:         "DEFLNG",
				LetterRanges: []ast.LetterRange{{Start: 'A', End: 'A'}},
			},
		},
	}
	resolver := NewResolver(prog)
	table, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for DEFLNG, got: %v", errs)
	}
	if got := table.ResolveType("avar"); got != TypeLong {
		t.Errorf("expected TypeLong for avar after DEFLNG A-A, got %v", got)
	}
}

func TestResolveDefTypeDefSng(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.DefTypeStatement{
				BasePos:      ast.Position{Line: 1, Column: 1},
				Type:         "DEFSNG",
				LetterRanges: []ast.LetterRange{{Start: 'S', End: 'S'}},
			},
		},
	}
	resolver := NewResolver(prog)
	table, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for DEFSNG, got: %v", errs)
	}
	if got := table.ResolveType("svar"); got != TypeSingle {
		t.Errorf("expected TypeSingle for svar after DEFSNG S-S, got %v", got)
	}
}

func TestResolveDefTypeDefDbl(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.DefTypeStatement{
				BasePos:      ast.Position{Line: 1, Column: 1},
				Type:         "DEFDBL",
				LetterRanges: []ast.LetterRange{{Start: 'D', End: 'D'}},
			},
		},
	}
	resolver := NewResolver(prog)
	table, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for DEFDBL, got: %v", errs)
	}
	if got := table.ResolveType("dvar"); got != TypeDouble {
		t.Errorf("expected TypeDouble for dvar after DEFDBL D-D, got %v", got)
	}
}

func TestResolveDefTypeDefStr(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.DefTypeStatement{
				BasePos:      ast.Position{Line: 1, Column: 1},
				Type:         "DEFSTR",
				LetterRanges: []ast.LetterRange{{Start: 'T', End: 'T'}},
			},
		},
	}
	resolver := NewResolver(prog)
	table, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for DEFSTR, got: %v", errs)
	}
	if got := table.ResolveType("tvar"); got != TypeString {
		t.Errorf("expected TypeString for tvar after DEFSTR T-T, got %v", got)
	}
}

func TestResolveDefTypeUnknown(t *testing.T) {
	// Unknown DEFTYPE → no-op (default case returns early)
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.DefTypeStatement{
				BasePos:      ast.Position{Line: 1, Column: 1},
				Type:         "DEFXXX",
				LetterRanges: []ast.LetterRange{{Start: 'X', End: 'X'}},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for unknown DEFTYPE, got: %v", errs)
	}
}

// resolveScopeStmt – SHARED, STATIC, LOCAL, COMMON (55.6% → covered)

func TestResolveScopeStmtSharedWithExistingGlobal(t *testing.T) {
	// SHARED variable that already exists in global scope → marks it IsShared
	prog := &ast.Program{
		Statements: []ast.Statement{
			// Define global x first
			&ast.LetStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Name:    &ast.Identifier{Name: "x"},
				Value:   &ast.NumberLiteral{Value: 0, NumType: ast.NumInt},
			},
			// Then SHARED x
			&ast.ScopeStatement{
				BasePos:   ast.Position{Line: 2, Column: 1},
				Modifier:  "SHARED",
				Variables: []string{"x"},
			},
		},
	}
	resolver := NewResolver(prog)
	table, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for SHARED with existing global, got: %v", errs)
	}
	sym := table.GlobalScope.Symbols["X"]
	if sym == nil {
		t.Fatal("expected x to exist in global scope")
	}
	if !sym.IsShared {
		t.Error("expected x to be marked IsShared")
	}
}

func TestResolveScopeStmtSharedWithNoExistingGlobal(t *testing.T) {
	// SHARED with a variable that doesn't yet exist → implicitly creates global
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ScopeStatement{
				BasePos:   ast.Position{Line: 1, Column: 1},
				Modifier:  "SHARED",
				Variables: []string{"newSharedVar"},
			},
		},
	}
	resolver := NewResolver(prog)
	table, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for SHARED creating new global, got: %v", errs)
	}
	sym := table.GlobalScope.Symbols["NEWSHAREDVAR"]
	if sym == nil {
		t.Fatal("expected newSharedVar to be implicitly created in global scope")
	}
	if !sym.IsShared {
		t.Error("expected newSharedVar to be marked IsShared")
	}
}

func TestResolveScopeStmtStaticWithExistingLocal(t *testing.T) {
	// STATIC variable that already exists in local scope → marks it IsStatic
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.SubDeclaration{
				BasePos: ast.Position{Line: 1, Column: 1},
				Name:    "TestSub",
				Body: []ast.Statement{
					// Define local variable first
					&ast.LetStatement{
						BasePos: ast.Position{Line: 2, Column: 1},
						Name:    &ast.Identifier{Name: "counter"},
						Value:   &ast.NumberLiteral{Value: 0, NumType: ast.NumInt},
					},
					// Then STATIC counter
					&ast.ScopeStatement{
						BasePos:   ast.Position{Line: 3, Column: 1},
						Modifier:  "STATIC",
						Variables: []string{"counter"},
					},
				},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for STATIC with existing local, got: %v", errs)
	}
}

func TestResolveScopeStmtStaticWithNoExistingLocal(t *testing.T) {
	// STATIC with variable not yet in local scope → creates it as static
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.SubDeclaration{
				BasePos: ast.Position{Line: 1, Column: 1},
				Name:    "TestSub2",
				Body: []ast.Statement{
					&ast.ScopeStatement{
						BasePos:   ast.Position{Line: 2, Column: 1},
						Modifier:  "STATIC",
						Variables: []string{"staticVar"},
					},
				},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for STATIC creating new local, got: %v", errs)
	}
}

func TestResolveScopeStmtLocalWithNoExistingLocal(t *testing.T) {
	// LOCAL with a variable not yet existing → creates it
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ScopeStatement{
				BasePos:   ast.Position{Line: 1, Column: 1},
				Modifier:  "LOCAL",
				Variables: []string{"localVar"},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for LOCAL statement, got: %v", errs)
	}
}

func TestResolveScopeStmtLocalWithExistingLocal(t *testing.T) {
	// LOCAL with a variable that already exists locally → no-op (LookupLocal returns non-nil)
	prog := &ast.Program{
		Statements: []ast.Statement{
			// Define the variable first
			&ast.LetStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Name:    &ast.Identifier{Name: "existVar"},
				Value:   &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
			},
			// Then LOCAL it
			&ast.ScopeStatement{
				BasePos:   ast.Position{Line: 2, Column: 1},
				Modifier:  "LOCAL",
				Variables: []string{"existVar"},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for LOCAL with existing variable, got: %v", errs)
	}
}

func TestResolveScopeStmtCommon(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ScopeStatement{
				BasePos:   ast.Position{Line: 1, Column: 1},
				Modifier:  "COMMON",
				Variables: []string{"commonVar"},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for COMMON statement, got: %v", errs)
	}
}

// resolveFor with Step != nil (72.7% → covered)

func TestResolveForWithStep(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ForStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Counter: &ast.Identifier{Name: "i"},
				Start:   &ast.NumberLiteral{Value: 0, NumType: ast.NumInt},
				End:     &ast.NumberLiteral{Value: 100, NumType: ast.NumInt},
				Step:    &ast.NumberLiteral{Value: 5, NumType: ast.NumInt},
				Body:    []ast.Statement{},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for FOR with STEP, got: %v", errs)
	}
}

// resolveIf with ElseIf clauses (66.7% → covered)

func TestResolveIfWithElseIf(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.IfStatement{
				BasePos:   ast.Position{Line: 1, Column: 1},
				Condition: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				ThenBlock: []ast.Statement{},
				ElseIfClauses: []ast.ElseIfClause{
					{
						BasePos:   ast.Position{Line: 3, Column: 1},
						Condition: &ast.NumberLiteral{Value: 0, NumType: ast.NumInt},
						Body: []ast.Statement{
							&ast.LetStatement{
								BasePos: ast.Position{Line: 4, Column: 1},
								Name:    &ast.Identifier{Name: "x"},
								Value:   &ast.NumberLiteral{Value: 2, NumType: ast.NumInt},
							},
						},
					},
				},
				ElseBlock: []ast.Statement{},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for IF with ELSEIF, got: %v", errs)
	}
}

func TestResolveIfElseBlock(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.IfStatement{
				BasePos:   ast.Position{Line: 1, Column: 1},
				Condition: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				ThenBlock: []ast.Statement{},
				ElseBlock: []ast.Statement{
					&ast.LetStatement{
						BasePos: ast.Position{Line: 3, Column: 1},
						Name:    &ast.Identifier{Name: "y"},
						Value:   &ast.NumberLiteral{Value: 99, NumType: ast.NumInt},
					},
				},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for IF with ELSE block, got: %v", errs)
	}
}

// ===========================================================================
// references.go gap coverage
// ===========================================================================

// collectTargets – SubDeclaration/FunctionDeclaration body recursion

func TestCollectTargetsSubDeclarationBody(t *testing.T) {
	// Label inside a SUB body should be collected
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.SubDeclaration{
				BasePos: ast.Position{Line: 1, Column: 1},
				Name:    "MySub",
				Body: []ast.Statement{
					&ast.LabelStatement{
						BasePos: ast.Position{Line: 2, Column: 1},
						Name:    "innerSubLabel",
					},
				},
			},
			&ast.GotoStatement{
				BasePos: ast.Position{Line: 5, Column: 1},
				Target:  "innerSubLabel",
			},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors (label inside SUB body collected), got: %v", errs)
	}
}

func TestCollectTargetsFunctionDeclarationBody(t *testing.T) {
	// Label inside a FUNCTION body should be collected
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.FunctionDeclaration{
				BasePos: ast.Position{Line: 1, Column: 1},
				Name:    "MyFunc",
				Body: []ast.Statement{
					&ast.LabelStatement{
						BasePos: ast.Position{Line: 2, Column: 1},
						Name:    "innerFuncLabel",
					},
				},
			},
			&ast.GotoStatement{
				BasePos: ast.Position{Line: 5, Column: 1},
				Target:  "innerFuncLabel",
			},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors (label inside FUNCTION body collected), got: %v", errs)
	}
}

func TestCollectTargetsIfStatementAllBlocks(t *testing.T) {
	// Labels in ThenBlock, ElseIf body, and ElseBlock should all be collected
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.IfStatement{
				BasePos:   ast.Position{Line: 1, Column: 1},
				Condition: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				ThenBlock: []ast.Statement{
					&ast.LabelStatement{BasePos: ast.Position{Line: 2, Column: 1}, Name: "thenLabel"},
				},
				ElseIfClauses: []ast.ElseIfClause{
					{
						BasePos:   ast.Position{Line: 3, Column: 1},
						Condition: &ast.NumberLiteral{Value: 0, NumType: ast.NumInt},
						Body: []ast.Statement{
							&ast.LabelStatement{BasePos: ast.Position{Line: 4, Column: 1}, Name: "elseIfLabel"},
						},
					},
				},
				ElseBlock: []ast.Statement{
					&ast.LabelStatement{BasePos: ast.Position{Line: 5, Column: 1}, Name: "elseLabel"},
				},
			},
			&ast.GotoStatement{BasePos: ast.Position{Line: 6, Column: 1}, Target: "thenLabel"},
			&ast.GotoStatement{BasePos: ast.Position{Line: 7, Column: 1}, Target: "elseIfLabel"},
			&ast.GotoStatement{BasePos: ast.Position{Line: 8, Column: 1}, Target: "elseLabel"},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors (labels in all IF blocks collected), got: %v", errs)
	}
}

func TestCollectTargetsWhileStatementBody(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.WhileStatement{
				BasePos:   ast.Position{Line: 1, Column: 1},
				Condition: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				Body: []ast.Statement{
					&ast.LabelStatement{BasePos: ast.Position{Line: 2, Column: 1}, Name: "whileLabel"},
				},
			},
			&ast.GotoStatement{BasePos: ast.Position{Line: 3, Column: 1}, Target: "whileLabel"},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors (label in WHILE body collected), got: %v", errs)
	}
}

func TestCollectTargetsDoLoopStatementBody(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.DoLoopStatement{
				BasePos:   ast.Position{Line: 1, Column: 1},
				Condition: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				Body: []ast.Statement{
					&ast.LabelStatement{BasePos: ast.Position{Line: 2, Column: 1}, Name: "doLabel"},
				},
			},
			&ast.GotoStatement{BasePos: ast.Position{Line: 3, Column: 1}, Target: "doLabel"},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors (label in DO LOOP body collected), got: %v", errs)
	}
}

func TestCollectTargetsSelectCaseAllBlocks(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.SelectCaseStatement{
				BasePos:  ast.Position{Line: 1, Column: 1},
				TestExpr: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				Cases: []ast.CaseClause{
					{
						Values: []ast.CaseValue{
							{Value: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt}},
						},
						Body: []ast.Statement{
							&ast.LabelStatement{BasePos: ast.Position{Line: 2, Column: 1}, Name: "caseLabel"},
						},
					},
				},
				ElseBlock: []ast.Statement{
					&ast.LabelStatement{BasePos: ast.Position{Line: 3, Column: 1}, Name: "caseElseLabel"},
				},
			},
			&ast.GotoStatement{BasePos: ast.Position{Line: 4, Column: 1}, Target: "caseLabel"},
			&ast.GotoStatement{BasePos: ast.Position{Line: 5, Column: 1}, Target: "caseElseLabel"},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors (labels in SELECT CASE blocks collected), got: %v", errs)
	}
}

func TestCollectTargetsDataStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.DataStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Values: []ast.Expression{
					&ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
					&ast.StringLiteral{Value: "hello"},
				},
			},
		},
	}
	rr := NewReferenceResolver(prog)
	rr.Resolve()
	pool := rr.GetDataPool()
	if len(pool) != 2 {
		t.Fatalf("expected 2 items in data pool, got %d", len(pool))
	}
}

// validateReferences – nested blocks

func TestValidateReferencesOnEventGosubStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LabelStatement{BasePos: ast.Position{Line: 1, Column: 1}, Name: "timerHandler"},
			&ast.OnEventGosubStatement{
				BasePos:   ast.Position{Line: 2, Column: 1},
				EventType: "TIMER",
				Target:    "timerHandler",
			},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for ON TIMER GOSUB with valid target, got: %v", errs)
	}
}

func TestValidateReferencesRestoreWithNonEmptyTarget(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LabelStatement{BasePos: ast.Position{Line: 1, Column: 1}, Name: "dataSection"},
			&ast.RestoreStatement{
				BasePos: ast.Position{Line: 2, Column: 1},
				Target:  "dataSection",
			},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for RESTORE with valid target, got: %v", errs)
	}
}

func TestValidateReferencesRestoreMissingTarget(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.RestoreStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Target:  "missing",
			},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) == 0 {
		t.Fatal("expected error for RESTORE with undefined target")
	}
}

func TestValidateReferencesResumeWithNonNextLabel(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LabelStatement{BasePos: ast.Position{Line: 1, Column: 1}, Name: "errRecovery"},
			&ast.ResumeStatement{
				BasePos: ast.Position{Line: 2, Column: 1},
				Type:    "errRecovery",
			},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for RESUME <label>, got: %v", errs)
	}
}

func TestValidateReferencesResumeMissingLabel(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ResumeStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Type:    "missingLabel",
			},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) == 0 {
		t.Fatal("expected error for RESUME with undefined label")
	}
}

func TestValidateReferencesSubDeclarationBody(t *testing.T) {
	// GOTO inside SUB body to a label inside the same SUB
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.SubDeclaration{
				BasePos: ast.Position{Line: 1, Column: 1},
				Name:    "MySub",
				Body: []ast.Statement{
					&ast.LabelStatement{BasePos: ast.Position{Line: 2, Column: 1}, Name: "subTarget"},
					&ast.GotoStatement{BasePos: ast.Position{Line: 3, Column: 1}, Target: "subTarget"},
				},
			},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for GOTO inside SUB body, got: %v", errs)
	}
}

func TestValidateReferencesFunctionDeclarationBody(t *testing.T) {
	// GOTO inside FUNCTION body to a label inside the same FUNCTION
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.FunctionDeclaration{
				BasePos: ast.Position{Line: 1, Column: 1},
				Name:    "MyFunc",
				Body: []ast.Statement{
					&ast.LabelStatement{BasePos: ast.Position{Line: 2, Column: 1}, Name: "funcTarget"},
					&ast.GotoStatement{BasePos: ast.Position{Line: 3, Column: 1}, Target: "funcTarget"},
				},
			},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for GOTO inside FUNCTION body, got: %v", errs)
	}
}

func TestValidateReferencesIfAllBlocks(t *testing.T) {
	// GOTO in ThenBlock, ElseIfClause, and ElseBlock validated recursively
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LabelStatement{BasePos: ast.Position{Line: 1, Column: 1}, Name: "validTarget"},
			&ast.IfStatement{
				BasePos:   ast.Position{Line: 2, Column: 1},
				Condition: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				ThenBlock: []ast.Statement{
					&ast.GotoStatement{BasePos: ast.Position{Line: 3, Column: 1}, Target: "validTarget"},
				},
				ElseIfClauses: []ast.ElseIfClause{
					{
						BasePos:   ast.Position{Line: 4, Column: 1},
						Condition: &ast.NumberLiteral{Value: 0, NumType: ast.NumInt},
						Body: []ast.Statement{
							&ast.GotoStatement{BasePos: ast.Position{Line: 5, Column: 1}, Target: "validTarget"},
						},
					},
				},
				ElseBlock: []ast.Statement{
					&ast.GotoStatement{BasePos: ast.Position{Line: 6, Column: 1}, Target: "validTarget"},
				},
			},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for GOTO in all IF blocks, got: %v", errs)
	}
}

func TestValidateReferencesIfBlockMissingLabel(t *testing.T) {
	// GOTO inside ThenBlock to missing label → error
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.IfStatement{
				BasePos:   ast.Position{Line: 1, Column: 1},
				Condition: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				ThenBlock: []ast.Statement{
					&ast.GotoStatement{BasePos: ast.Position{Line: 2, Column: 1}, Target: "noSuchLabel"},
				},
			},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) == 0 {
		t.Fatal("expected error for GOTO in ThenBlock to missing label")
	}
}

func TestValidateReferencesForStatementBody(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LabelStatement{BasePos: ast.Position{Line: 1, Column: 1}, Name: "forTarget"},
			&ast.ForStatement{
				BasePos: ast.Position{Line: 2, Column: 1},
				Counter: &ast.Identifier{Name: "i"},
				Start:   &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				End:     &ast.NumberLiteral{Value: 10, NumType: ast.NumInt},
				Body: []ast.Statement{
					&ast.GotoStatement{BasePos: ast.Position{Line: 3, Column: 1}, Target: "forTarget"},
				},
			},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for GOTO in FOR body, got: %v", errs)
	}
}

func TestValidateReferencesWhileStatementBody(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LabelStatement{BasePos: ast.Position{Line: 1, Column: 1}, Name: "whileTarget"},
			&ast.WhileStatement{
				BasePos:   ast.Position{Line: 2, Column: 1},
				Condition: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				Body: []ast.Statement{
					&ast.GotoStatement{BasePos: ast.Position{Line: 3, Column: 1}, Target: "whileTarget"},
				},
			},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for GOTO in WHILE body, got: %v", errs)
	}
}

func TestValidateReferencesDoLoopStatementBody(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LabelStatement{BasePos: ast.Position{Line: 1, Column: 1}, Name: "doTarget"},
			&ast.DoLoopStatement{
				BasePos:   ast.Position{Line: 2, Column: 1},
				Condition: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				Body: []ast.Statement{
					&ast.GotoStatement{BasePos: ast.Position{Line: 3, Column: 1}, Target: "doTarget"},
				},
			},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for GOTO in DO LOOP body, got: %v", errs)
	}
}

func TestValidateReferencesSelectCaseAllBlocks(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LabelStatement{BasePos: ast.Position{Line: 1, Column: 1}, Name: "selectTarget"},
			&ast.SelectCaseStatement{
				BasePos:  ast.Position{Line: 2, Column: 1},
				TestExpr: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				Cases: []ast.CaseClause{
					{
						Values: []ast.CaseValue{
							{Value: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt}},
						},
						Body: []ast.Statement{
							&ast.GotoStatement{BasePos: ast.Position{Line: 3, Column: 1}, Target: "selectTarget"},
						},
					},
				},
				ElseBlock: []ast.Statement{
					&ast.GotoStatement{BasePos: ast.Position{Line: 4, Column: 1}, Target: "selectTarget"},
				},
			},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for GOTO in SELECT CASE blocks, got: %v", errs)
	}
}

func TestValidateReferencesSelectCaseMissingTarget(t *testing.T) {
	// GOTO inside a CASE body to a missing label → error
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.SelectCaseStatement{
				BasePos:  ast.Position{Line: 1, Column: 1},
				TestExpr: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				Cases: []ast.CaseClause{
					{
						Values: []ast.CaseValue{
							{Value: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt}},
						},
						Body: []ast.Statement{
							&ast.GotoStatement{BasePos: ast.Position{Line: 2, Column: 1}, Target: "noSuchLabel"},
						},
					},
				},
			},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) == 0 {
		t.Fatal("expected error for GOTO in CASE body to missing label")
	}
}
