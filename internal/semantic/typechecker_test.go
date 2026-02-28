package semantic

import (
	"strings"
	"testing"

	"github.com/loabletech/go-basic/internal/ast"
)

// ===========================================================================
// Helpers
// ===========================================================================

// buildSymTable creates a symbol table with predefined symbols.
func buildSymTable(syms map[string]*Symbol) *SymbolTable {
	st := NewSymbolTable()
	for name, sym := range syms {
		_ = st.Define(name, sym)
	}
	return st
}

// ===========================================================================
// NewTypeChecker and Check
// ===========================================================================

func TestNewTypeChecker(t *testing.T) {
	st := NewSymbolTable()
	tc := NewTypeChecker(st)
	if tc == nil {
		t.Fatal("NewTypeChecker returned nil")
	}
	if tc.Table != st {
		t.Error("TypeChecker.Table should be the provided symbol table")
	}
	if tc.Errors != nil {
		t.Error("TypeChecker.Errors should start nil")
	}
}

func TestTypeCheckerCheckEmptyProgram(t *testing.T) {
	st := NewSymbolTable()
	tc := NewTypeChecker(st)
	errs := tc.Check(&ast.Program{})
	if len(errs) != 0 {
		t.Fatalf("expected no errors for empty program, got: %v", errs)
	}
}

func TestTypeCheckerCheckSimpleProgram(t *testing.T) {
	st := buildSymTable(map[string]*Symbol{
		"X%": {Name: "X%", Type: SymVariable, DataType: TypeInteger},
	})
	tc := NewTypeChecker(st)
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LetStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Name:    &ast.Identifier{Name: "X", TypeSuffix: "%"},
				Value:   &ast.NumberLiteral{Value: 42, OriginalText: "42", NumType: ast.NumInt},
			},
		},
	}
	errs := tc.Check(prog)
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got: %v", errs)
	}
}

// ===========================================================================
// resolveExprType tests
// ===========================================================================

func TestResolveExprTypeNil(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	got := tc.resolveExprType(nil)
	if got != TypeUnknown {
		t.Errorf("expected TypeUnknown for nil, got %v", got)
	}
}

func TestResolveExprTypeNumberLiterals(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	tests := []struct {
		numType int
		want    DataType
	}{
		{ast.NumInt, TypeInteger},
		{ast.NumLong, TypeLong},
		{ast.NumSingle, TypeSingle},
		{ast.NumDouble, TypeDouble},
	}
	for _, tc2 := range tests {
		expr := &ast.NumberLiteral{Value: 1, OriginalText: "1", NumType: tc2.numType}
		got := tc.resolveExprType(expr)
		if got != tc2.want {
			t.Errorf("NumType=%d: expected %v, got %v", tc2.numType, tc2.want, got)
		}
	}
}

func TestResolveExprTypeNumberLiteralDefault(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	// Unknown NumType should default to Single
	expr := &ast.NumberLiteral{Value: 1, OriginalText: "1", NumType: 99}
	got := tc.resolveExprType(expr)
	if got != TypeSingle {
		t.Errorf("expected TypeSingle for unknown NumType, got %v", got)
	}
}

func TestResolveExprTypeStringLiteral(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	got := tc.resolveExprType(&ast.StringLiteral{Value: "hello"})
	if got != TypeString {
		t.Errorf("expected TypeString for StringLiteral, got %v", got)
	}
}

func TestResolveExprTypeIdentifierFromSymTable(t *testing.T) {
	st := buildSymTable(map[string]*Symbol{
		"X%": {Name: "X%", Type: SymVariable, DataType: TypeInteger},
	})
	tc := NewTypeChecker(st)
	got := tc.resolveExprType(&ast.Identifier{Name: "X%"})
	if got != TypeInteger {
		t.Errorf("expected TypeInteger from sym table, got %v", got)
	}
}

func TestResolveExprTypeIdentifierFallback(t *testing.T) {
	// Identifier not in table - resolves via suffix
	tc := NewTypeChecker(NewSymbolTable())
	got := tc.resolveExprType(&ast.Identifier{Name: "z#"})
	if got != TypeDouble {
		t.Errorf("expected TypeDouble from # suffix, got %v", got)
	}
}

func TestResolveExprTypeBinaryExpr(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	expr := &ast.BinaryExpr{
		Left:     &ast.NumberLiteral{Value: 1, OriginalText: "1", NumType: ast.NumInt},
		Operator: "+",
		Right:    &ast.NumberLiteral{Value: 2, OriginalText: "2", NumType: ast.NumInt},
	}
	got := tc.resolveExprType(expr)
	if got != TypeInteger {
		t.Errorf("expected TypeInteger for int+int, got %v", got)
	}
}

func TestResolveExprTypeUnaryExpr(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	expr := &ast.UnaryExpr{
		Operator: "-",
		Operand:  &ast.NumberLiteral{Value: 5, OriginalText: "5", NumType: ast.NumDouble},
	}
	got := tc.resolveExprType(expr)
	if got != TypeDouble {
		t.Errorf("expected TypeDouble for unary minus on double, got %v", got)
	}
}

func TestResolveExprTypeFunctionCall(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	// SIN returns TypeSingle
	got := tc.resolveExprType(&ast.FunctionCall{Name: "SIN", Args: []ast.Expression{
		&ast.NumberLiteral{Value: 1, OriginalText: "1"},
	}})
	if got != TypeSingle {
		t.Errorf("expected TypeSingle for SIN, got %v", got)
	}
}

func TestResolveExprTypeArrayAccess(t *testing.T) {
	// resolveExprType for ArrayAccess looks up e.Name (not e.Name+suffix)
	st := buildSymTable(map[string]*Symbol{
		"ARR": {Name: "ARR", Type: SymArray, DataType: TypeInteger},
	})
	tc := NewTypeChecker(st)
	expr := &ast.ArrayAccess{
		Name:    "ARR",
		Indices: []ast.Expression{&ast.NumberLiteral{Value: 1, OriginalText: "1"}},
	}
	got := tc.resolveExprType(expr)
	if got != TypeInteger {
		t.Errorf("expected TypeInteger from array symbol, got %v", got)
	}
}

func TestResolveExprTypeArrayAccessFallback(t *testing.T) {
	// Without symbol in table, falls back to ResolveType(e.Name)
	// "data$" has $ suffix → TypeString
	tc := NewTypeChecker(NewSymbolTable())
	expr := &ast.ArrayAccess{
		Name:    "data$",
		Indices: []ast.Expression{&ast.NumberLiteral{Value: 0, OriginalText: "0"}},
	}
	got := tc.resolveExprType(expr)
	if got != TypeString {
		t.Errorf("expected TypeString from $ suffix fallback, got %v", got)
	}
}

func TestResolveExprTypeGroupExpr(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	expr := &ast.GroupExpr{
		Inner: &ast.StringLiteral{Value: "hello"},
	}
	got := tc.resolveExprType(expr)
	if got != TypeString {
		t.Errorf("expected TypeString from grouped string, got %v", got)
	}
}

// ===========================================================================
// resolveBinaryType tests
// ===========================================================================

func TestResolveBinaryTypeRelational(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	ops := []string{"=", "<>", "<", ">", "<=", ">="}
	for _, op := range ops {
		expr := &ast.BinaryExpr{
			Left:     &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
			Operator: op,
			Right:    &ast.NumberLiteral{Value: 2, NumType: ast.NumInt},
		}
		got := tc.resolveBinaryType(expr)
		if got != TypeInteger {
			t.Errorf("relational op %s: expected TypeInteger, got %v", op, got)
		}
	}
}

func TestResolveBinaryTypeStringConcatenation(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	expr := &ast.BinaryExpr{
		Left:     &ast.StringLiteral{Value: "a"},
		Operator: "+",
		Right:    &ast.StringLiteral{Value: "b"},
	}
	got := tc.resolveBinaryType(expr)
	if got != TypeString {
		t.Errorf("expected TypeString for string+string, got %v", got)
	}
}

func TestResolveBinaryTypeStringMismatch(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	expr := &ast.BinaryExpr{
		BasePos:  ast.Position{Line: 1, Column: 1},
		Left:     &ast.StringLiteral{Value: "a"},
		Operator: "+",
		Right:    &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
	}
	got := tc.resolveBinaryType(expr)
	if got != TypeUnknown {
		t.Errorf("expected TypeUnknown for string+int mismatch, got %v", got)
	}
	if len(tc.Errors) == 0 {
		t.Error("expected type mismatch error")
	}
	if !strings.Contains(tc.Errors[0], "type mismatch") {
		t.Errorf("expected 'type mismatch' error, got: %s", tc.Errors[0])
	}
}

func TestResolveBinaryTypeLogical(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	ops := []string{"AND", "OR", "XOR"}
	for _, op := range ops {
		tc.Errors = nil
		expr := &ast.BinaryExpr{
			Left:     &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
			Operator: op,
			Right:    &ast.NumberLiteral{Value: 0, NumType: ast.NumInt},
		}
		got := tc.resolveBinaryType(expr)
		if got != TypeInteger {
			t.Errorf("logical op %s: expected TypeInteger, got %v", op, got)
		}
		if len(tc.Errors) != 0 {
			t.Errorf("logical op %s: unexpected errors: %v", op, tc.Errors)
		}
	}
}

func TestResolveBinaryTypeLogicalWithString(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	expr := &ast.BinaryExpr{
		BasePos:  ast.Position{Line: 1, Column: 1},
		Left:     &ast.StringLiteral{Value: "a"},
		Operator: "AND",
		Right:    &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
	}
	tc.resolveBinaryType(expr)
	if len(tc.Errors) == 0 {
		t.Error("expected error for string AND numeric")
	}
}

func TestResolveBinaryTypeIntegerDivision(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	expr := &ast.BinaryExpr{
		Left:     &ast.NumberLiteral{Value: 10, NumType: ast.NumSingle},
		Operator: `\`,
		Right:    &ast.NumberLiteral{Value: 3, NumType: ast.NumSingle},
	}
	got := tc.resolveBinaryType(expr)
	if got != TypeInteger {
		t.Errorf("integer division: expected TypeInteger, got %v", got)
	}
}

func TestResolveBinaryTypeMOD(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	expr := &ast.BinaryExpr{
		Left:     &ast.NumberLiteral{Value: 10, NumType: ast.NumInt},
		Operator: "MOD",
		Right:    &ast.NumberLiteral{Value: 3, NumType: ast.NumInt},
	}
	got := tc.resolveBinaryType(expr)
	if got != TypeInteger {
		t.Errorf("MOD: expected TypeInteger, got %v", got)
	}
}

func TestResolveBinaryTypeDivisionPromotion(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	// int / int → single (promoted)
	expr := &ast.BinaryExpr{
		Left:     &ast.NumberLiteral{Value: 5, NumType: ast.NumInt},
		Operator: "/",
		Right:    &ast.NumberLiteral{Value: 2, NumType: ast.NumInt},
	}
	got := tc.resolveBinaryType(expr)
	if got != TypeSingle {
		t.Errorf("int/int division: expected TypeSingle, got %v", got)
	}
}

func TestResolveBinaryTypeExponentiation(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	// int ^ int → single
	expr := &ast.BinaryExpr{
		Left:     &ast.NumberLiteral{Value: 2, NumType: ast.NumInt},
		Operator: "^",
		Right:    &ast.NumberLiteral{Value: 3, NumType: ast.NumInt},
	}
	got := tc.resolveBinaryType(expr)
	if got != TypeSingle {
		t.Errorf("int^int: expected TypeSingle, got %v", got)
	}
}

func TestResolveBinaryTypeWidening(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	// int + double → double
	expr := &ast.BinaryExpr{
		Left:     &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
		Operator: "+",
		Right:    &ast.NumberLiteral{Value: 1, NumType: ast.NumDouble},
	}
	got := tc.resolveBinaryType(expr)
	if got != TypeDouble {
		t.Errorf("int+double: expected TypeDouble, got %v", got)
	}
}

func TestResolveBinaryTypeStringCompare(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	// string = string → integer (boolean)
	expr := &ast.BinaryExpr{
		Left:     &ast.StringLiteral{Value: "a"},
		Operator: "=",
		Right:    &ast.StringLiteral{Value: "b"},
	}
	got := tc.resolveBinaryType(expr)
	if got != TypeInteger {
		t.Errorf("string comparison: expected TypeInteger, got %v", got)
	}
}

func TestResolveBinaryTypeRelationalTypeMismatch(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	// string > int → error
	expr := &ast.BinaryExpr{
		BasePos:  ast.Position{Line: 1, Column: 1},
		Left:     &ast.StringLiteral{Value: "a"},
		Operator: ">",
		Right:    &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
	}
	tc.resolveBinaryType(expr)
	if len(tc.Errors) == 0 {
		t.Error("expected type mismatch error for string > int")
	}
}

// ===========================================================================
// resolveFunctionReturnType tests
// ===========================================================================

func TestResolveFunctionReturnTypeDollarSuffix(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	tests := []string{"LEFT$", "RIGHT$", "MID$", "CHR$", "STR$", "HEX$", "OCT$", "INKEY$"}
	for _, name := range tests {
		got := tc.resolveFunctionReturnType(&ast.FunctionCall{Name: name})
		if got != TypeString {
			t.Errorf("%s: expected TypeString, got %v", name, got)
		}
	}
}

func TestResolveFunctionReturnTypeIntegerBuiltins(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	tests := []string{"ASC", "LEN", "INSTR", "EOF", "LOF", "LOC", "CINT", "INT", "FIX", "ERR", "ERL"}
	for _, name := range tests {
		got := tc.resolveFunctionReturnType(&ast.FunctionCall{Name: name})
		if got != TypeInteger {
			t.Errorf("%s: expected TypeInteger, got %v", name, got)
		}
	}
}

func TestResolveFunctionReturnTypeLongBuiltins(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	tests := []string{"CLNG", "VARPTR"}
	for _, name := range tests {
		got := tc.resolveFunctionReturnType(&ast.FunctionCall{Name: name})
		if got != TypeLong {
			t.Errorf("%s: expected TypeLong, got %v", name, got)
		}
	}
}

func TestResolveFunctionReturnTypeSingleBuiltins(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	tests := []string{"SIN", "COS", "TAN", "ATN", "SQR", "LOG", "EXP", "ABS", "SGN", "RND", "TIMER", "CSNG"}
	for _, name := range tests {
		got := tc.resolveFunctionReturnType(&ast.FunctionCall{Name: name})
		if got != TypeSingle {
			t.Errorf("%s: expected TypeSingle, got %v", name, got)
		}
	}
}

func TestResolveFunctionReturnTypeDoubleBuiltins(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	got := tc.resolveFunctionReturnType(&ast.FunctionCall{Name: "CDBL"})
	if got != TypeDouble {
		t.Errorf("CDBL: expected TypeDouble, got %v", got)
	}
}

func TestResolveFunctionReturnTypeUserDefined(t *testing.T) {
	st := buildSymTable(map[string]*Symbol{
		"MYSUM": {Name: "MYSUM", Type: SymFunction, ReturnType: TypeDouble},
	})
	tc := NewTypeChecker(st)
	got := tc.resolveFunctionReturnType(&ast.FunctionCall{Name: "MySum"})
	if got != TypeDouble {
		t.Errorf("user-defined function: expected TypeDouble, got %v", got)
	}
}

func TestResolveFunctionReturnTypeDefault(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	// Unknown function → TypeSingle default
	got := tc.resolveFunctionReturnType(&ast.FunctionCall{Name: "UnknownFn"})
	if got != TypeSingle {
		t.Errorf("unknown function: expected TypeSingle default, got %v", got)
	}
}

// ===========================================================================
// checkStatement dispatch tests
// ===========================================================================

func TestCheckStatementLetCompatible(t *testing.T) {
	st := buildSymTable(map[string]*Symbol{
		"X%": {Name: "X%", Type: SymVariable, DataType: TypeInteger},
	})
	tc := NewTypeChecker(st)
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LetStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Name:    &ast.Identifier{Name: "X", TypeSuffix: "%"},
				Value:   &ast.NumberLiteral{Value: 42, NumType: ast.NumInt},
			},
		},
	}
	errs := tc.Check(prog)
	if len(errs) != 0 {
		t.Fatalf("expected no errors for compatible assignment, got: %v", errs)
	}
}

func TestCheckStatementLetIncompatible(t *testing.T) {
	st := buildSymTable(map[string]*Symbol{
		"X%": {Name: "X%", Type: SymVariable, DataType: TypeInteger},
	})
	tc := NewTypeChecker(st)
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LetStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Name:    &ast.Identifier{Name: "X"},
				Value:   &ast.StringLiteral{Value: "hello"},
			},
		},
	}
	errs := tc.Check(prog)
	if len(errs) == 0 {
		t.Fatal("expected type mismatch error for string to int assignment")
	}
	if !strings.Contains(errs[0], "type mismatch") {
		t.Fatalf("expected type mismatch error, got: %s", errs[0])
	}
}

func TestCheckStatementArrayAssignmentCompatible(t *testing.T) {
	st := buildSymTable(map[string]*Symbol{
		"ARR%": {Name: "ARR%", Type: SymArray, DataType: TypeInteger},
	})
	tc := NewTypeChecker(st)
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ArrayAssignment{
				BasePos: ast.Position{Line: 1, Column: 1},
				Array: &ast.ArrayAccess{
					Name:       "ARR",
					TypeSuffix: "%",
					Indices:    []ast.Expression{&ast.NumberLiteral{Value: 1, NumType: ast.NumInt}},
				},
				Value: &ast.NumberLiteral{Value: 10, NumType: ast.NumInt},
			},
		},
	}
	errs := tc.Check(prog)
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got: %v", errs)
	}
}

func TestCheckStatementArrayAssignmentIncompatible(t *testing.T) {
	st := buildSymTable(map[string]*Symbol{
		"ARR": {Name: "ARR", Type: SymArray, DataType: TypeInteger},
	})
	tc := NewTypeChecker(st)
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ArrayAssignment{
				BasePos: ast.Position{Line: 1, Column: 1},
				Array: &ast.ArrayAccess{
					Name:    "ARR",
					Indices: []ast.Expression{&ast.NumberLiteral{Value: 0, NumType: ast.NumInt}},
				},
				Value: &ast.StringLiteral{Value: "bad"},
			},
		},
	}
	errs := tc.Check(prog)
	if len(errs) == 0 {
		t.Fatal("expected type mismatch error for string to int array")
	}
}

func TestCheckStatementArrayAssignmentStringIndex(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ArrayAssignment{
				BasePos: ast.Position{Line: 1, Column: 1},
				Array: &ast.ArrayAccess{
					Name:    "arr",
					Indices: []ast.Expression{&ast.StringLiteral{Value: "bad_index"}},
				},
				Value: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
			},
		},
	}
	errs := tc.Check(prog)
	if len(errs) == 0 {
		t.Fatal("expected error for string array index")
	}
}

func TestCheckStatementBlockRecursion(t *testing.T) {
	// Test checkBlock via checkStatement for SubDeclaration
	// Assign string to numeric variable inside SUB body → type error
	st := buildSymTable(map[string]*Symbol{
		"NUMVAR": {Name: "NUMVAR", Type: SymVariable, DataType: TypeInteger},
	})
	tc := NewTypeChecker(st)
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.SubDeclaration{
				BasePos: ast.Position{Line: 1},
				Name:    "MySub",
				Body: []ast.Statement{
					&ast.LetStatement{
						BasePos: ast.Position{Line: 2},
						Name:    &ast.Identifier{Name: "NUMVAR"},
						Value:   &ast.StringLiteral{Value: "bad"},
					},
				},
			},
		},
	}
	errs := tc.Check(prog)
	if len(errs) == 0 {
		t.Fatal("expected type mismatch error in sub body")
	}
}

func TestCheckStatementIfBlock(t *testing.T) {
	// IF with string condition → error
	tc := NewTypeChecker(NewSymbolTable())
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.IfStatement{
				BasePos:   ast.Position{Line: 1},
				Condition: &ast.StringLiteral{Value: "yes"},
				ThenBlock: []ast.Statement{},
			},
		},
	}
	errs := tc.Check(prog)
	if len(errs) == 0 {
		t.Fatal("expected error for string IF condition")
	}
	if !strings.Contains(errs[0], "IF condition must be numeric") {
		t.Fatalf("expected IF condition error, got: %s", errs[0])
	}
}

func TestCheckStatementForBlock(t *testing.T) {
	// FOR with string start → error
	tc := NewTypeChecker(NewSymbolTable())
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ForStatement{
				BasePos: ast.Position{Line: 1},
				Counter: &ast.Identifier{Name: "i"},
				Start:   &ast.StringLiteral{Value: "bad"},
				End:     &ast.NumberLiteral{Value: 10, NumType: ast.NumInt},
			},
		},
	}
	errs := tc.Check(prog)
	if len(errs) == 0 {
		t.Fatal("expected error for string FOR start")
	}
}

func TestCheckStatementWhileBlock(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.WhileStatement{
				BasePos:   ast.Position{Line: 1},
				Condition: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				Body:      []ast.Statement{},
			},
		},
	}
	errs := tc.Check(prog)
	if len(errs) != 0 {
		t.Fatalf("expected no errors for WHILE with numeric condition, got: %v", errs)
	}
}

func TestCheckStatementDoLoopBlock(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.DoLoopStatement{
				BasePos:   ast.Position{Line: 1},
				Condition: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				Body:      []ast.Statement{},
			},
		},
	}
	errs := tc.Check(prog)
	if len(errs) != 0 {
		t.Fatalf("expected no errors for DO WHILE with numeric condition, got: %v", errs)
	}
}

func TestCheckStatementDoLoopNilCondition(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.DoLoopStatement{
				BasePos:   ast.Position{Line: 1},
				Condition: nil,
				Body:      []ast.Statement{},
			},
		},
	}
	errs := tc.Check(prog)
	if len(errs) != 0 {
		t.Fatalf("expected no errors for infinite DO LOOP, got: %v", errs)
	}
}

func TestCheckStatementSelectCase(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.SelectCaseStatement{
				BasePos:  ast.Position{Line: 1},
				TestExpr: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				Cases: []ast.CaseClause{
					{
						Values: []ast.CaseValue{
							{Value: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt}},
						},
						Body: []ast.Statement{},
					},
				},
			},
		},
	}
	errs := tc.Check(prog)
	if len(errs) != 0 {
		t.Fatalf("expected no errors for SELECT CASE with matching types, got: %v", errs)
	}
}

func TestCheckStatementSelectCaseMismatch(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	// SELECT CASE numeric / CASE "string" → type mismatch
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.SelectCaseStatement{
				BasePos:  ast.Position{Line: 1},
				TestExpr: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				Cases: []ast.CaseClause{
					{
						Values: []ast.CaseValue{
							{BasePos: ast.Position{Line: 2}, Value: &ast.StringLiteral{Value: "one"}},
						},
						Body: []ast.Statement{},
					},
				},
			},
		},
	}
	errs := tc.Check(prog)
	if len(errs) == 0 {
		t.Fatal("expected type mismatch error for SELECT CASE numeric vs string CASE")
	}
}

func TestCheckStatementFunctionDecl(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.FunctionDeclaration{
				BasePos: ast.Position{Line: 1},
				Name:    "Add",
				Body: []ast.Statement{
					// Assign string to integer variable → error
					&ast.LetStatement{
						BasePos: ast.Position{Line: 2},
						Name:    &ast.Identifier{Name: "result", TypeSuffix: "%"},
						Value:   &ast.StringLiteral{Value: "bad"},
					},
				},
			},
		},
	}
	errs := tc.Check(prog)
	if len(errs) == 0 {
		t.Fatal("expected type error inside FUNCTION body")
	}
}

func TestCheckStatementDefFn(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.DefFnDeclaration{
				BasePos:        ast.Position{Line: 1},
				Name:           "FNDouble",
				SingleLineExpr: &ast.NumberLiteral{Value: 2, NumType: ast.NumSingle},
			},
		},
	}
	errs := tc.Check(prog)
	if len(errs) != 0 {
		t.Fatalf("expected no errors for DEF FN, got: %v", errs)
	}
}

func TestCheckStatementSwap(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	// SWAP numeric with string → error
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.SwapStatement{
				BasePos: ast.Position{Line: 1},
				Var1:    &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				Var2:    &ast.StringLiteral{Value: "bad"},
			},
		},
	}
	errs := tc.Check(prog)
	if len(errs) == 0 {
		t.Fatal("expected error for SWAP numeric with string")
	}
}

func TestCheckStatementIncrStringError(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.IncrStatement{
				BasePos:  ast.Position{Line: 1},
				Variable: &ast.StringLiteral{Value: "bad"},
			},
		},
	}
	errs := tc.Check(prog)
	if len(errs) == 0 {
		t.Fatal("expected error for INCR of string")
	}
	if !strings.Contains(errs[0], "numeric") {
		t.Fatalf("expected numeric error, got: %s", errs[0])
	}
}

func TestCheckStatementDecrStringError(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.DecrStatement{
				BasePos:  ast.Position{Line: 1},
				Variable: &ast.StringLiteral{Value: "bad"},
			},
		},
	}
	errs := tc.Check(prog)
	if len(errs) == 0 {
		t.Fatal("expected error for DECR of string")
	}
}

func TestCheckStatementSoundStatement(t *testing.T) {
	tc := NewTypeChecker(NewSymbolTable())
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.SoundStatement{
				BasePos:   ast.Position{Line: 1},
				Frequency: &ast.NumberLiteral{Value: 440, NumType: ast.NumSingle},
				Duration:  &ast.NumberLiteral{Value: 18, NumType: ast.NumSingle},
			},
		},
	}
	errs := tc.Check(prog)
	if len(errs) != 0 {
		t.Fatalf("expected no errors for SOUND, got: %v", errs)
	}
}
