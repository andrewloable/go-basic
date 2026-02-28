package parser

import (
	"testing"

	"github.com/loabletech/go-basic/internal/ast"
)

// ===========================================================================
// Tests for parse_declarations.go: SUB, FUNCTION, TYPE, DIM, CONST, DECLARE,
// DEFTYPE, OPTION BASE
// ===========================================================================

func TestParseDim(t *testing.T) {
	prog, errs := parse("DIM a(10), b$(5, 5)")
	expectNoErrors(t, errs)
	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}
	ds, ok := prog.Statements[0].(*ast.DimStatement)
	if !ok {
		t.Fatalf("expected DimStatement, got %T", prog.Statements[0])
	}
	if len(ds.Declarations) != 2 {
		t.Fatalf("expected 2 declarations, got %d", len(ds.Declarations))
	}
	if ds.Declarations[0].Name != "a" {
		t.Errorf("expected first decl name 'a', got %q", ds.Declarations[0].Name)
	}
	if len(ds.Declarations[0].Dimensions) != 1 {
		t.Errorf("expected 1 dimension for 'a', got %d", len(ds.Declarations[0].Dimensions))
	}
	if len(ds.Declarations[1].Dimensions) != 2 {
		t.Errorf("expected 2 dimensions for 'b$', got %d", len(ds.Declarations[1].Dimensions))
	}
}

func TestParseSubDeclaration(t *testing.T) {
	input := `SUB MySub (x, y)
  PRINT x + y
END SUB`
	prog, errs := parse(input)
	expectNoErrors(t, errs)
	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}
	sd, ok := prog.Statements[0].(*ast.SubDeclaration)
	if !ok {
		t.Fatalf("expected SubDeclaration, got %T", prog.Statements[0])
	}
	if sd.Name != "MySub" {
		t.Errorf("expected SUB name 'MySub', got %q", sd.Name)
	}
	if len(sd.Params) != 2 {
		t.Errorf("expected 2 params, got %d", len(sd.Params))
	}
}

func TestParseFunctionDeclaration(t *testing.T) {
	input := `FUNCTION Add(a, b)
  Add = a + b
END FUNCTION`
	prog, errs := parse(input)
	expectNoErrors(t, errs)
	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}
	fd, ok := prog.Statements[0].(*ast.FunctionDeclaration)
	if !ok {
		t.Fatalf("expected FunctionDeclaration, got %T", prog.Statements[0])
	}
	if fd.Name != "Add" {
		t.Errorf("expected FUNCTION name 'Add', got %q", fd.Name)
	}
}

// TestRegressionDimShared verifies that DIM SHARED (and LOCAL, STATIC, COMMON)
// modifiers are consumed without being treated as the variable name.
func TestRegressionDimShared(t *testing.T) {
	_, errs := parse("DIM SHARED arr(1 TO 10) AS INTEGER")
	expectNoErrors(t, errs)
	_, errs = parse("DIM STATIC count AS LONG")
	expectNoErrors(t, errs)
}

// TestRegressionTypeBlock verifies TYPE name ... END TYPE parsing.
func TestRegressionTypeBlock(t *testing.T) {
	input := `TYPE Point
  X AS INTEGER
  Y AS INTEGER
END TYPE`
	_, errs := parse(input)
	expectNoErrors(t, errs)
}

// TestRegressionDefFn verifies multi-line DEF FN and FN call in expressions.
func TestRegressionDefFn(t *testing.T) {
	input := `DEF FN Double(x) = x * 2
y = FN Double(5)`
	_, errs := parse(input)
	expectNoErrors(t, errs)
}

// TestRegressionConst verifies CONST name = expr parsing.
func TestRegressionConst(t *testing.T) {
	_, errs := parse("CONST PI = 3.14159")
	expectNoErrors(t, errs)
	_, errs = parse(`CONST APPNAME = "MyApp"`)
	expectNoErrors(t, errs)
}

// TestRegressionSubArrayParam verifies that SUB/FUNCTION declarations
// accept array parameters with () notation and optional AS TypeName.
func TestRegressionSubArrayParam(t *testing.T) {
	input := `SUB FillArray (arr() AS INTEGER, size AS INTEGER)
  FOR i = 1 TO size
    arr(i) = 0
  NEXT i
END SUB`
	_, errs := parse(input)
	expectNoErrors(t, errs)
}

// TestRegressionTypeFieldAccess verifies struct/TYPE field access (dot notation).
func TestRegressionTypeFieldAccess(t *testing.T) {
	input := `TYPE Point
  X AS INTEGER
  Y AS INTEGER
END TYPE
DIM p AS Point
p.X = 10
p.Y = 20
PRINT p.X`
	_, errs := parse(input)
	expectNoErrors(t, errs)
}

// TestRegressionTypeFieldArrayAccess verifies struct field access on an array element.
func TestRegressionTypeFieldArrayAccess(t *testing.T) {
	input := `DIM pts(10) AS Point
pts(1).X = 5
pts(1).Y = 3
total = pts(1).X + pts(1).Y`
	_, errs := parse(input)
	expectNoErrors(t, errs)
}

// ===========================================================================
// OPTION BASE statement
// ===========================================================================

func TestParseOptionBase0(t *testing.T) {
	prog, errs := parse("OPTION BASE 0")
	expectNoErrors(t, errs)
	s := getStmt[*ast.OptionBaseStatement](t, prog, 0)
	if s.Value != 0 {
		t.Errorf("expected OPTION BASE 0, got %d", s.Value)
	}
}

func TestParseOptionBase1(t *testing.T) {
	prog, errs := parse("OPTION BASE 1")
	expectNoErrors(t, errs)
	s := getStmt[*ast.OptionBaseStatement](t, prog, 0)
	if s.Value != 1 {
		t.Errorf("expected OPTION BASE 1, got %d", s.Value)
	}
}

func TestParseOptionBaseZero(t *testing.T) {
	prog, errs := parse("OPTION BASE 0")
	expectNoErrors(t, errs)
	s := getStmt[*ast.OptionBaseStatement](t, prog, 0)
	if s.Value != 0 {
		t.Errorf("OPTION BASE 0: Value = %d, want 0", s.Value)
	}
}

func TestParseOptionBaseOne(t *testing.T) {
	prog, errs := parse("OPTION BASE 1")
	expectNoErrors(t, errs)
	s := getStmt[*ast.OptionBaseStatement](t, prog, 0)
	if s.Value != 1 {
		t.Errorf("OPTION BASE 1: Value = %d, want 1", s.Value)
	}
}

func TestParseOptionMissingBase(t *testing.T) {
	// OPTION without BASE → error.
	_, errs := parse("OPTION 1")
	if len(errs) == 0 {
		t.Error("expected error for OPTION without BASE")
	}
}

// ===========================================================================
// DEFTYPE statements
// ===========================================================================

func TestParseDefInt(t *testing.T) {
	prog, errs := parse("DEFINT A-Z")
	expectNoErrors(t, errs)
	s := getStmt[*ast.DefTypeStatement](t, prog, 0)
	if s.Type != "DEFINT" {
		t.Errorf("expected DEFINT, got %q", s.Type)
	}
	if len(s.LetterRanges) == 0 {
		t.Fatal("expected at least 1 letter range")
	}
	if s.LetterRanges[0].Start != 'A' || s.LetterRanges[0].End != 'Z' {
		t.Errorf("expected range A-Z, got %c-%c", s.LetterRanges[0].Start, s.LetterRanges[0].End)
	}
}

func TestParseDefDbl(t *testing.T) {
	prog, errs := parse("DEFDBL X-Z")
	expectNoErrors(t, errs)
	s := getStmt[*ast.DefTypeStatement](t, prog, 0)
	if s.Type != "DEFDBL" {
		t.Errorf("expected DEFDBL, got %q", s.Type)
	}
}

func TestParseDefStr(t *testing.T) {
	prog, errs := parse("DEFSTR N")
	expectNoErrors(t, errs)
	s := getStmt[*ast.DefTypeStatement](t, prog, 0)
	if s.Type != "DEFSTR" {
		t.Errorf("expected DEFSTR, got %q", s.Type)
	}
}

// ===========================================================================
// DECLARE statement
// ===========================================================================

func TestParseDeclare(t *testing.T) {
	prog, errs := parse("DECLARE SUB MySub(x AS INTEGER, y AS STRING)")
	expectNoErrors(t, errs)
	found := false
	for _, stmt := range prog.Statements {
		if d, ok := stmt.(*ast.SubDeclaration); ok && d.IsForward {
			found = true
			if d.Name != "MySub" {
				t.Errorf("expected MySub, got %q", d.Name)
			}
			if len(d.Params) != 2 {
				t.Errorf("expected 2 params, got %d", len(d.Params))
			}
			break
		}
	}
	if !found {
		t.Error("expected forward-declared SubDeclaration")
	}
}

func TestParseDeclareError(t *testing.T) {
	// DECLARE without SUB or FUNCTION triggers the error path
	_, errs := parse("DECLARE DOUBLE")
	if len(errs) == 0 {
		t.Error("expected parse error for DECLARE without SUB/FUNCTION")
	}
}

func TestDeclareFunctionWithReturnType(t *testing.T) {
	prog, errs := parse("DECLARE FUNCTION MyFunc() AS SINGLE")
	expectNoErrors(t, errs)
	found := false
	for _, stmt := range prog.Statements {
		if d, ok := stmt.(*ast.FunctionDeclaration); ok && d.IsForward {
			found = true
			if d.Name != "MyFunc" {
				t.Errorf("expected MyFunc, got %q", d.Name)
			}
			if d.ReturnType != "SINGLE" {
				t.Errorf("expected SINGLE return type, got %q", d.ReturnType)
			}
		}
	}
	if !found {
		t.Error("expected forward FunctionDeclaration")
	}
}

func TestDeclareFunctionNoReturnType(t *testing.T) {
	prog, errs := parse("DECLARE FUNCTION Calculate(a%, b%)")
	expectNoErrors(t, errs)
	found := false
	for _, stmt := range prog.Statements {
		if d, ok := stmt.(*ast.FunctionDeclaration); ok && d.IsForward {
			found = true
			if d.Name != "Calculate" {
				t.Errorf("expected Calculate, got %q", d.Name)
			}
		}
	}
	if !found {
		t.Error("expected forward FunctionDeclaration")
	}
}

// ===========================================================================
// TYPE block
// ===========================================================================

func TestParseTypeBlock(t *testing.T) {
	prog, errs := parse(`TYPE Point
x AS INTEGER
y AS INTEGER
END TYPE`)
	expectNoErrors(t, errs)
	found := false
	for _, stmt := range prog.Statements {
		if tb, ok := stmt.(*ast.TypeBlockStatement); ok {
			found = true
			if tb.Name != "Point" {
				t.Errorf("expected Point, got %q", tb.Name)
			}
			if len(tb.Fields) != 2 {
				t.Errorf("expected 2 fields, got %d", len(tb.Fields))
			}
		}
	}
	if !found {
		t.Error("expected TypeBlockStatement")
	}
}

// ===========================================================================
// CONST statement
// ===========================================================================

func TestParseConstStatement(t *testing.T) {
	prog, errs := parse("CONST PI = 3.14159")
	expectNoErrors(t, errs)
	s := getStmt[*ast.ConstStatement](t, prog, 0)
	if s.Name == "" {
		t.Error("expected Name")
	}
	if s.Value == nil {
		t.Error("expected Value")
	}
}

// ===========================================================================
// DEF FN (multi-line)
// ===========================================================================

func TestParseDefFnMultiLine(t *testing.T) {
	prog, errs := parse(`DEF FNDouble(x!)
  FNDouble = x! * 2
END DEF`)
	expectNoErrors(t, errs)
	found := false
	for _, stmt := range prog.Statements {
		if d, ok := stmt.(*ast.DefFnDeclaration); ok {
			found = true
			if d.Name != "FNDouble" {
				t.Errorf("expected FNDouble, got %q", d.Name)
			}
			break
		}
	}
	if !found {
		t.Error("expected DefFnDeclaration")
	}
}

// ===========================================================================
// SHARED/LOCAL/STATIC scope statements (inside SUB/FUNCTION)
// ===========================================================================

func TestParseSharedStatement(t *testing.T) {
	prog, errs := parse(`SUB MySub
SHARED x%, y$
END SUB`)
	expectNoErrors(t, errs)
	sub := getStmt[*ast.SubDeclaration](t, prog, 0)
	if len(sub.Body) == 0 {
		t.Fatal("expected at least 1 statement in sub body")
	}
	scope, ok := sub.Body[0].(*ast.ScopeStatement)
	if !ok {
		t.Fatalf("expected ScopeStatement, got %T", sub.Body[0])
	}
	if scope.Modifier != "SHARED" {
		t.Errorf("expected SHARED modifier, got %q", scope.Modifier)
	}
	if len(scope.Variables) != 2 {
		t.Errorf("expected 2 variables, got %d", len(scope.Variables))
	}
}

func TestParseStaticStatement(t *testing.T) {
	prog, errs := parse(`SUB Counter
STATIC count%
count% = count% + 1
END SUB`)
	expectNoErrors(t, errs)
	sub := getStmt[*ast.SubDeclaration](t, prog, 0)
	scope, ok := sub.Body[0].(*ast.ScopeStatement)
	if !ok {
		t.Fatalf("expected ScopeStatement, got %T", sub.Body[0])
	}
	if scope.Modifier != "STATIC" {
		t.Errorf("expected STATIC modifier, got %q", scope.Modifier)
	}
}

// ===========================================================================
// DEF SEG statement
// ===========================================================================

// TestRegressionDefSeg verifies DEF SEG with the optional = segment argument.
func TestRegressionDefSeg(t *testing.T) {
	_, errs := parse("DEF SEG = 0")
	expectNoErrors(t, errs)
	_, errs = parse("DEF SEG")
	expectNoErrors(t, errs)
}

func TestParseDefSeg(t *testing.T) {
	prog, errs := parse("DEF SEG = &HB800")
	expectNoErrors(t, errs)
	s := getStmt[*ast.RemStatement](t, prog, 0)
	if s.Text != "DEF SEG" {
		t.Errorf("expected 'DEF SEG', got %q", s.Text)
	}
}

func TestParseDefSegNoValue(t *testing.T) {
	prog, errs := parse("DEF SEG")
	expectNoErrors(t, errs)
	s := getStmt[*ast.RemStatement](t, prog, 0)
	if s.Text != "DEF SEG" {
		t.Errorf("expected 'DEF SEG', got %q", s.Text)
	}
}

// ===========================================================================
// Metacompiler directives ($IF, $ENDIF, $DYNAMIC, $STATIC)
// ===========================================================================

// TestRegressionMetacompilerDirectives verifies that $IF/$ENDIF/$DYNAMIC and
// other metacompiler directives are consumed without error.
func TestRegressionMetacompilerDirectives(t *testing.T) {
	input := `$DYNAMIC
$IF %Debug THEN
PRINT "debug"
$ENDIF`
	_, errs := parse(input)
	expectNoErrors(t, errs)
	_, errs = parse("$STATIC")
	expectNoErrors(t, errs)
}
