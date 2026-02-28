package parser

import (
	"testing"

	"github.com/loabletech/go-basic/internal/ast"
)

// ===========================================================================
// Tests for parse_expressions.go: expression parsing, precedence, literals,
// operators, function calls, builtins
// ===========================================================================

func TestParseExpressionPrecedence(t *testing.T) {
	prog, errs := parse("x = 2 + 3 * 4")
	expectNoErrors(t, errs)
	ls := prog.Statements[0].(*ast.LetStatement)
	// Should parse as 2 + (3 * 4), i.e. top-level is +
	be, ok := ls.Value.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", ls.Value)
	}
	if be.Operator != "+" {
		t.Errorf("expected top operator '+', got %q", be.Operator)
	}
	// Right side should be 3 * 4
	rbe, ok := be.Right.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected right to be BinaryExpr, got %T", be.Right)
	}
	if rbe.Operator != "*" {
		t.Errorf("expected right operator '*', got %q", rbe.Operator)
	}
}

func TestParseUnaryMinus(t *testing.T) {
	prog, errs := parse("x = -5")
	expectNoErrors(t, errs)
	ls := prog.Statements[0].(*ast.LetStatement)
	ue, ok := ls.Value.(*ast.UnaryExpr)
	if !ok {
		t.Fatalf("expected UnaryExpr, got %T", ls.Value)
	}
	if ue.Operator != "-" {
		t.Errorf("expected operator '-', got %q", ue.Operator)
	}
}

func TestParseFunctionCall(t *testing.T) {
	prog, errs := parse(`x = ABS(-5)`)
	expectNoErrors(t, errs)
	ls := prog.Statements[0].(*ast.LetStatement)
	fc, ok := ls.Value.(*ast.FunctionCall)
	if !ok {
		t.Fatalf("expected FunctionCall, got %T", ls.Value)
	}
	if fc.Name != "ABS" {
		t.Errorf("expected function name 'ABS', got %q", fc.Name)
	}
	if len(fc.Args) != 1 {
		t.Errorf("expected 1 arg, got %d", len(fc.Args))
	}
}

func TestParseHexLiteral(t *testing.T) {
	prog, errs := parse("x% = &HFF")
	expectNoErrors(t, errs)
	s := getStmt[*ast.LetStatement](t, prog, 0)
	nl, ok := s.Value.(*ast.NumberLiteral)
	if !ok {
		t.Fatalf("expected NumberLiteral, got %T", s.Value)
	}
	if nl.Value != 255 {
		t.Errorf("expected 255 (&HFF), got %v", nl.Value)
	}
}

func TestParseOctalLiteral(t *testing.T) {
	prog, errs := parse("x% = &O17")
	expectNoErrors(t, errs)
	s := getStmt[*ast.LetStatement](t, prog, 0)
	nl, ok := s.Value.(*ast.NumberLiteral)
	if !ok {
		t.Fatalf("expected NumberLiteral, got %T", s.Value)
	}
	if nl.Value != 15 { // &O17 = 15 in decimal
		t.Errorf("expected 15 (&O17), got %v", nl.Value)
	}
}

func TestParseBuiltinFunction(t *testing.T) {
	// LEFT$ is a builtin function
	prog, errs := parse(`x$ = LEFT$("hello", 3)`)
	expectNoErrors(t, errs)
	s := getStmt[*ast.LetStatement](t, prog, 0)
	if s.Value == nil {
		t.Fatal("expected Value to be set")
	}
}

func TestParseMidFunction(t *testing.T) {
	prog, errs := parse(`x$ = MID$("hello", 2, 3)`)
	expectNoErrors(t, errs)
	s := getStmt[*ast.LetStatement](t, prog, 0)
	if s.Value == nil {
		t.Fatal("expected Value to be set")
	}
}

func TestParseOperatorPrecedences(t *testing.T) {
	prog, errs := parse("x% = 10 MOD 3")
	expectNoErrors(t, errs)
	s := getStmt[*ast.LetStatement](t, prog, 0)
	be, ok := s.Value.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", s.Value)
	}
	if be.Operator != "MOD" {
		t.Errorf("expected MOD, got %q", be.Operator)
	}
}

func TestParseAndOrXorOperators(t *testing.T) {
	prog, errs := parse("x% = a% AND b% OR c% XOR d%")
	expectNoErrors(t, errs)
	s := getStmt[*ast.LetStatement](t, prog, 0)
	if s.Value == nil {
		t.Fatal("expected Value")
	}
}

// ===========================================================================
// IMP, EQV, exponentiation, integer division, relational operators
// ===========================================================================

func TestParseImpOperator(t *testing.T) {
	prog, errs := parse("x% = a% IMP b%")
	expectNoErrors(t, errs)
	s := getStmt[*ast.LetStatement](t, prog, 0)
	be, ok := s.Value.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", s.Value)
	}
	if be.Operator != "IMP" {
		t.Errorf("expected IMP, got %q", be.Operator)
	}
}

func TestParseEqvOperator(t *testing.T) {
	prog, errs := parse("x% = a% EQV b%")
	expectNoErrors(t, errs)
	s := getStmt[*ast.LetStatement](t, prog, 0)
	be, ok := s.Value.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", s.Value)
	}
	if be.Operator != "EQV" {
		t.Errorf("expected EQV, got %q", be.Operator)
	}
}

func TestParseExponentiationOperator(t *testing.T) {
	prog, errs := parse("x = 2 ^ 3")
	expectNoErrors(t, errs)
	s := getStmt[*ast.LetStatement](t, prog, 0)
	be, ok := s.Value.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", s.Value)
	}
	if be.Operator != "^" {
		t.Errorf("expected ^, got %q", be.Operator)
	}
}

func TestParseIntegerDivisionOperator(t *testing.T) {
	prog, errs := parse(`x% = 10 \ 3`)
	expectNoErrors(t, errs)
	s := getStmt[*ast.LetStatement](t, prog, 0)
	be, ok := s.Value.(*ast.BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", s.Value)
	}
	if be.Operator != `\` {
		t.Errorf("expected \\, got %q", be.Operator)
	}
}

func TestParseRelationalOperators(t *testing.T) {
	tests := []struct {
		input string
		op    string
	}{
		{"x% = a% <> b%", "<>"},
		{"x% = a% <= b%", "<="},
		{"x% = a% >= b%", ">="},
		{"x% = a% < b%", "<"},
		{"x% = a% > b%", ">"},
	}
	for _, tt := range tests {
		prog, errs := parse(tt.input)
		expectNoErrors(t, errs)
		s := getStmt[*ast.LetStatement](t, prog, 0)
		be, ok := s.Value.(*ast.BinaryExpr)
		if !ok {
			t.Fatalf("expected BinaryExpr for %q, got %T", tt.input, s.Value)
		}
		if be.Operator != tt.op {
			t.Errorf("expected %q, got %q", tt.op, be.Operator)
		}
	}
}

func TestTokenPrecedenceAllOperators(t *testing.T) {
	tests := []string{
		"x% = a% IMP b% EQV c%",      // IMP, EQV
		"x% = a% XOR b% OR c% AND d%", // XOR, OR, AND
		"x% = a% + b% - c%",           // ADD (PLUS, MINUS)
		"x% = a% MOD b%",              // MOD
		`x% = a% \ b%`,                // IDIV (BACKSLASH)
		"x% = a% * b% / c%",           // MUL (STAR, SLASH)
		"x% = a% ^ b%",                // POWER (CARET)
	}
	for _, input := range tests {
		prog, errs := parse(input)
		if len(errs) > 0 {
			t.Errorf("unexpected errors for %q: %v", input, errs)
			continue
		}
		if len(prog.Statements) == 0 {
			t.Errorf("expected statements for %q", input)
		}
	}
}

// ===========================================================================
// Number literals: types, D-notation, radix (hex, octal, binary)
// ===========================================================================

func TestParseNumberLiteralInteger(t *testing.T) {
	prog, errs := parse("x% = 42")
	expectNoErrors(t, errs)
	s := getStmt[*ast.LetStatement](t, prog, 0)
	nl, ok := s.Value.(*ast.NumberLiteral)
	if !ok {
		t.Fatalf("expected NumberLiteral, got %T", s.Value)
	}
	if nl.NumType != ast.NumInt {
		t.Errorf("expected NumInt, got %d", nl.NumType)
	}
	if nl.Value != 42 {
		t.Errorf("expected 42, got %v", nl.Value)
	}
}

func TestParseNumberLiteralLong(t *testing.T) {
	prog, errs := parse("x& = 100000&")
	expectNoErrors(t, errs)
	s := getStmt[*ast.LetStatement](t, prog, 0)
	nl, ok := s.Value.(*ast.NumberLiteral)
	if !ok {
		t.Fatalf("expected NumberLiteral, got %T", s.Value)
	}
	if nl.NumType != ast.NumLong {
		t.Errorf("expected NumLong, got %d", nl.NumType)
	}
}

func TestParseNumberLiteralSingle(t *testing.T) {
	prog, errs := parse("x! = 3.14!")
	expectNoErrors(t, errs)
	s := getStmt[*ast.LetStatement](t, prog, 0)
	nl, ok := s.Value.(*ast.NumberLiteral)
	if !ok {
		t.Fatalf("expected NumberLiteral, got %T", s.Value)
	}
	if nl.NumType != ast.NumSingle {
		t.Errorf("expected NumSingle, got %d", nl.NumType)
	}
}

func TestParseNumberLiteralDouble(t *testing.T) {
	prog, errs := parse("x# = 3.14#")
	expectNoErrors(t, errs)
	s := getStmt[*ast.LetStatement](t, prog, 0)
	nl, ok := s.Value.(*ast.NumberLiteral)
	if !ok {
		t.Fatalf("expected NumberLiteral, got %T", s.Value)
	}
	if nl.NumType != ast.NumDouble {
		t.Errorf("expected NumDouble, got %d", nl.NumType)
	}
}

func TestParseNumberLiteralDNotation(t *testing.T) {
	prog, errs := parse("x# = 1.5D2")
	expectNoErrors(t, errs)
	s := getStmt[*ast.LetStatement](t, prog, 0)
	nl, ok := s.Value.(*ast.NumberLiteral)
	if !ok {
		t.Fatalf("expected NumberLiteral, got %T", s.Value)
	}
	if nl.Value != 150 {
		t.Errorf("expected 150, got %v", nl.Value)
	}
}

func TestParseBinaryLiteral(t *testing.T) {
	prog, errs := parse("x% = &B1010")
	expectNoErrors(t, errs)
	s := getStmt[*ast.LetStatement](t, prog, 0)
	nl, ok := s.Value.(*ast.NumberLiteral)
	if !ok {
		t.Fatalf("expected NumberLiteral, got %T", s.Value)
	}
	if nl.Value != 10 { // 1010 binary = 10 decimal
		t.Errorf("expected 10 (&B1010), got %v", nl.Value)
	}
}

func TestParseHexLiteralUpperCase(t *testing.T) {
	prog, errs := parse("x% = &H1A")
	expectNoErrors(t, errs)
	s := getStmt[*ast.LetStatement](t, prog, 0)
	nl, ok := s.Value.(*ast.NumberLiteral)
	if !ok {
		t.Fatalf("expected NumberLiteral, got %T", s.Value)
	}
	if nl.Value != 26 { // 0x1A = 26
		t.Errorf("expected 26, got %v", nl.Value)
	}
}

func TestParseOctalLiteralValue(t *testing.T) {
	prog, errs := parse("x = &O17")
	expectNoErrors(t, errs)
	s := getStmt[*ast.LetStatement](t, prog, 0)
	nl := s.Value.(*ast.NumberLiteral)
	if nl.Value != 15 { // &O17 = 15 decimal
		t.Errorf("expected 15 (&O17), got %v", nl.Value)
	}
}

func TestParseHexLiteralValue(t *testing.T) {
	prog, errs := parse("x = &HFF")
	expectNoErrors(t, errs)
	s := getStmt[*ast.LetStatement](t, prog, 0)
	nl := s.Value.(*ast.NumberLiteral)
	if nl.Value != 255 {
		t.Errorf("expected 255 (&HFF), got %v", nl.Value)
	}
}

// ===========================================================================
// Grouped expressions
// ===========================================================================

func TestParseGroupExpressionMissingParen(t *testing.T) {
	// Missing ) triggers addError in parseGroupExpression
	_, errs := parse("x% = (1 + 2")
	_ = errs // Just verify it doesn't panic
}

func TestParseGroupExpressionValid(t *testing.T) {
	prog, errs := parse("x% = (5 + 3) * 2")
	expectNoErrors(t, errs)
	s := getStmt[*ast.LetStatement](t, prog, 0)
	if s.Value == nil {
		t.Error("expected Value")
	}
}

// ===========================================================================
// Unary NOT
// ===========================================================================

func TestParseNotExpression(t *testing.T) {
	prog, errs := parse("x% = NOT a%")
	expectNoErrors(t, errs)
	s := getStmt[*ast.LetStatement](t, prog, 0)
	ue, ok := s.Value.(*ast.UnaryExpr)
	if !ok {
		t.Fatalf("expected UnaryExpr, got %T", s.Value)
	}
	if ue.Operator != "NOT" {
		t.Errorf("expected NOT, got %q", ue.Operator)
	}
}

// ===========================================================================
// Builtin keyword functions (without parentheses or with)
// ===========================================================================

func TestParseBuiltinTimer(t *testing.T) {
	prog, errs := parse("x = TIMER")
	expectNoErrors(t, errs)
	s := getStmt[*ast.LetStatement](t, prog, 0)
	fc, ok := s.Value.(*ast.FunctionCall)
	if !ok {
		t.Fatalf("expected FunctionCall, got %T", s.Value)
	}
	if fc.Name != "TIMER" {
		t.Errorf("expected TIMER, got %q", fc.Name)
	}
}

func TestParseBuiltinInstat(t *testing.T) {
	prog, errs := parse("x% = INSTAT")
	expectNoErrors(t, errs)
	s := getStmt[*ast.LetStatement](t, prog, 0)
	if s.Value == nil {
		t.Error("expected Value")
	}
}

func TestParseBuiltinLenWithParens(t *testing.T) {
	prog, errs := parse(`n% = LEN("hello")`)
	expectNoErrors(t, errs)
	s := getStmt[*ast.LetStatement](t, prog, 0)
	fc, ok := s.Value.(*ast.FunctionCall)
	if !ok {
		t.Fatalf("expected FunctionCall, got %T", s.Value)
	}
	if fc.Name != "LEN" {
		t.Errorf("expected LEN, got %q", fc.Name)
	}
	if len(fc.Args) != 1 {
		t.Errorf("expected 1 arg, got %d", len(fc.Args))
	}
}

func TestParseBuiltinEofWithParens(t *testing.T) {
	prog, errs := parse("x% = EOF(1)")
	expectNoErrors(t, errs)
	s := getStmt[*ast.LetStatement](t, prog, 0)
	fc, ok := s.Value.(*ast.FunctionCall)
	if !ok {
		t.Fatalf("expected FunctionCall, got %T", s.Value)
	}
	if fc.Name != "EOF" {
		t.Errorf("expected EOF, got %q", fc.Name)
	}
}

func TestParseBuiltinLboundUbound(t *testing.T) {
	prog, errs := parse("n% = LBOUND(arr, 1)")
	expectNoErrors(t, errs)
	s := getStmt[*ast.LetStatement](t, prog, 0)
	fc, ok := s.Value.(*ast.FunctionCall)
	if !ok {
		t.Fatalf("expected FunctionCall, got %T", s.Value)
	}
	if fc.Name != "LBOUND" {
		t.Errorf("expected LBOUND, got %q", fc.Name)
	}

	prog2, errs2 := parse("n% = UBOUND(arr)")
	expectNoErrors(t, errs2)
	s2 := getStmt[*ast.LetStatement](t, prog2, 0)
	fc2, ok2 := s2.Value.(*ast.FunctionCall)
	if !ok2 {
		t.Fatalf("expected FunctionCall, got %T", s2.Value)
	}
	if fc2.Name != "UBOUND" {
		t.Errorf("expected UBOUND, got %q", fc2.Name)
	}
}

func TestParseBuiltinVarptr(t *testing.T) {
	prog, errs := parse("addr% = VARPTR(x%)")
	expectNoErrors(t, errs)
	s := getStmt[*ast.LetStatement](t, prog, 0)
	fc, ok := s.Value.(*ast.FunctionCall)
	if !ok {
		t.Fatalf("expected FunctionCall, got %T", s.Value)
	}
	if fc.Name != "VARPTR" {
		t.Errorf("expected VARPTR, got %q", fc.Name)
	}
}

func TestParseBuiltinScreenFunction(t *testing.T) {
	// SCREEN(row, col) — builtin function returning char at position
	prog, errs := parse("c% = SCREEN(5, 10)")
	expectNoErrors(t, errs)
	s := getStmt[*ast.LetStatement](t, prog, 0)
	if s.Value == nil {
		t.Error("expected Value")
	}
}

func TestParseBuiltinPeek(t *testing.T) {
	prog, errs := parse("v% = PEEK(100)")
	expectNoErrors(t, errs)
	s := getStmt[*ast.LetStatement](t, prog, 0)
	fc, ok := s.Value.(*ast.FunctionCall)
	if !ok {
		t.Fatalf("expected FunctionCall, got %T", s.Value)
	}
	if fc.Name != "PEEK" {
		t.Errorf("expected PEEK, got %q", fc.Name)
	}
}

func TestParseBuiltinInp(t *testing.T) {
	prog, errs := parse("v% = INP(60)")
	expectNoErrors(t, errs)
	s := getStmt[*ast.LetStatement](t, prog, 0)
	fc, ok := s.Value.(*ast.FunctionCall)
	if !ok {
		t.Fatalf("expected FunctionCall, got %T", s.Value)
	}
	if fc.Name != "INP" {
		t.Errorf("expected INP, got %q", fc.Name)
	}
}

// ===========================================================================
// Dollar-suffix builtin functions
// ===========================================================================

func TestParseStringDollarFunction(t *testing.T) {
	prog, errs := parse(`x$ = STRING$(5, 65)`)
	expectNoErrors(t, errs)
	s := getStmt[*ast.LetStatement](t, prog, 0)
	fc, ok := s.Value.(*ast.FunctionCall)
	if !ok {
		t.Fatalf("expected FunctionCall, got %T", s.Value)
	}
	if fc.Name != "STRING$" {
		t.Errorf("expected STRING$, got %q", fc.Name)
	}
}

func TestParseSpaceDollarFunction(t *testing.T) {
	prog, errs := parse(`x$ = SPACE$(10)`)
	expectNoErrors(t, errs)
	s := getStmt[*ast.LetStatement](t, prog, 0)
	fc, ok := s.Value.(*ast.FunctionCall)
	if !ok {
		t.Fatalf("expected FunctionCall, got %T", s.Value)
	}
	if fc.Name != "SPACE$" {
		t.Errorf("expected SPACE$, got %q", fc.Name)
	}
}

func TestParseRightDollarFunction(t *testing.T) {
	prog, errs := parse(`x$ = RIGHT$("hello", 3)`)
	expectNoErrors(t, errs)
	s := getStmt[*ast.LetStatement](t, prog, 0)
	fc, ok := s.Value.(*ast.FunctionCall)
	if !ok {
		t.Fatalf("expected FunctionCall, got %T", s.Value)
	}
	if fc.Name != "RIGHT$" {
		t.Errorf("expected RIGHT$, got %q", fc.Name)
	}
}

func TestParseChrDollarFunction(t *testing.T) {
	prog, errs := parse(`x$ = CHR$(65)`)
	expectNoErrors(t, errs)
	s := getStmt[*ast.LetStatement](t, prog, 0)
	fc, ok := s.Value.(*ast.FunctionCall)
	if !ok {
		t.Fatalf("expected FunctionCall, got %T", s.Value)
	}
	if fc.Name != "CHR$" {
		t.Errorf("expected CHR$, got %q", fc.Name)
	}
}

func TestParseAscFunction(t *testing.T) {
	prog, errs := parse(`x% = ASC("A")`)
	expectNoErrors(t, errs)
	s := getStmt[*ast.LetStatement](t, prog, 0)
	fc, ok := s.Value.(*ast.FunctionCall)
	if !ok {
		t.Fatalf("expected FunctionCall, got %T", s.Value)
	}
	if fc.Name != "ASC" {
		t.Errorf("expected ASC, got %q", fc.Name)
	}
}

func TestParseValFunction(t *testing.T) {
	prog, errs := parse(`x = VAL("42")`)
	expectNoErrors(t, errs)
	s := getStmt[*ast.LetStatement](t, prog, 0)
	fc, ok := s.Value.(*ast.FunctionCall)
	if !ok {
		t.Fatalf("expected FunctionCall, got %T", s.Value)
	}
	if fc.Name != "VAL" {
		t.Errorf("expected VAL, got %q", fc.Name)
	}
}

func TestParseRndFunction(t *testing.T) {
	prog, errs := parse("x = RND")
	expectNoErrors(t, errs)
	s := getStmt[*ast.LetStatement](t, prog, 0)
	if s.Value == nil {
		t.Error("expected Value")
	}
}

func TestParseAbsFunction(t *testing.T) {
	prog, errs := parse("x = ABS(-5)")
	expectNoErrors(t, errs)
	s := getStmt[*ast.LetStatement](t, prog, 0)
	fc, ok := s.Value.(*ast.FunctionCall)
	if !ok {
		t.Fatalf("expected FunctionCall, got %T", s.Value)
	}
	if fc.Name != "ABS" {
		t.Errorf("expected ABS, got %q", fc.Name)
	}
}

func TestParseInkeyFunction(t *testing.T) {
	prog, errs := parse("k$ = INKEY$")
	expectNoErrors(t, errs)
	s := getStmt[*ast.LetStatement](t, prog, 0)
	if s.Value == nil {
		t.Error("expected Value")
	}
}

// ===========================================================================
// Identifier expression: type suffixes, FN prefix, field access, user calls
// ===========================================================================

func TestParseIdentifierWithTypeSuffix(t *testing.T) {
	prog, errs := parse("x% = 5")
	expectNoErrors(t, errs)
	s := getStmt[*ast.LetStatement](t, prog, 0)
	if s == nil {
		t.Fatal("expected LetStatement")
	}
}

func TestParseUserFunctionCall(t *testing.T) {
	// Non-builtin function call → ArrayAccess (semantic analysis distinguishes)
	prog, errs := parse("x = MyFunc(5, 10)")
	expectNoErrors(t, errs)
	s := getStmt[*ast.LetStatement](t, prog, 0)
	if s.Value == nil {
		t.Error("expected value")
	}
}

func TestParseFieldAccessWithSubscript(t *testing.T) {
	// Spaced dot notation to produce FieldAccess with subscript.
	prog, errs := parse("x = arr . item(3)")
	// This may produce errors or not, but shouldn't panic.
	_ = prog
	_ = errs
}

func TestParseExpressionNilPrefix(t *testing.T) {
	// A token that is not a valid expression start.
	prog, errs := parse("PRINT TO")
	// Should produce an error but not panic.
	_ = prog
	_ = errs
}

// ===========================================================================
// FN assign and call
// ===========================================================================

func TestParseFnAssignStatement(t *testing.T) {
	prog, errs := parse("FN double = 42")
	expectNoErrors(t, errs)
	found := false
	for _, stmt := range prog.Statements {
		if fa, ok := stmt.(*ast.FnAssignStatement); ok {
			found = true
			if fa.Name != "double" {
				t.Errorf("expected 'double', got %q", fa.Name)
			}
			if fa.Value == nil {
				t.Error("expected Value")
			}
		}
	}
	if !found {
		t.Error("expected FnAssignStatement")
	}
}

func TestParseFnAssignWithTypeSuffix(t *testing.T) {
	prog, errs := parse(`FN greet$ = "Hello"`)
	expectNoErrors(t, errs)
	found := false
	for _, stmt := range prog.Statements {
		if fa, ok := stmt.(*ast.FnAssignStatement); ok {
			found = true
			if fa.Name != "greet" {
				t.Errorf("expected 'greet', got %q", fa.Name)
			}
		}
	}
	if !found {
		t.Error("expected FnAssignStatement")
	}
}

func TestParseFnCallStatement(t *testing.T) {
	prog, errs := parse("FN calc(5)")
	expectNoErrors(t, errs)
	found := false
	for _, stmt := range prog.Statements {
		if ls, ok := stmt.(*ast.LetStatement); ok {
			if ls.Name != nil && len(ls.Name.Name) >= 4 && ls.Name.Name[:4] == "_fn_" {
				found = true
				if ls.Value == nil {
					t.Error("expected Value")
				}
			}
		}
	}
	if !found {
		t.Error("expected LetStatement wrapping FnCallExpression for standalone FN call")
	}
}

func TestParseFnCallExpression(t *testing.T) {
	prog, errs := parse(`DEF FNSquare(x!)
FNSquare = x! * x!
END DEF
y = FNSquare(5)`)
	expectNoErrors(t, errs)
	found := false
	for _, stmt := range prog.Statements {
		if ls, ok := stmt.(*ast.LetStatement); ok {
			if ls.Name != nil && ls.Name.Name == "y" {
				found = true
			}
		}
	}
	if !found {
		t.Error("expected assignment to y using FN call")
	}
}

// ===========================================================================
// TAB and SPC
// ===========================================================================

func TestParseBuiltinTab(t *testing.T) {
	prog, errs := parse("PRINT TAB(20); x$")
	expectNoErrors(t, errs)
	if len(prog.Statements) == 0 {
		t.Fatal("expected at least 1 statement")
	}
}

func TestParseBuiltinSpc(t *testing.T) {
	prog, errs := parse("PRINT SPC(5)")
	expectNoErrors(t, errs)
	if len(prog.Statements) == 0 {
		t.Fatal("expected at least 1 statement")
	}
}
