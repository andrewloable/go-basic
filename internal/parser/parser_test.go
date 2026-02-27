package parser

import (
	"testing"

	"github.com/loabletech/go-basic/internal/ast"
	"github.com/loabletech/go-basic/internal/lexer"
)

func parse(input string) (*ast.Program, []string) {
	l := lexer.New(input)
	p := New(l)
	prog := p.ParseProgram()
	return prog, p.Errors()
}

func expectNoErrors(t *testing.T, errors []string) {
	t.Helper()
	if len(errors) > 0 {
		t.Fatalf("unexpected parse errors: %v", errors)
	}
}

func TestParseEmpty(t *testing.T) {
	prog, errs := parse("")
	expectNoErrors(t, errs)
	if len(prog.Statements) != 0 {
		t.Errorf("expected 0 statements, got %d", len(prog.Statements))
	}
}

func TestParsePrint(t *testing.T) {
	prog, errs := parse(`PRINT "Hello World"`)
	expectNoErrors(t, errs)
	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}
	ps, ok := prog.Statements[0].(*ast.PrintStatement)
	if !ok {
		t.Fatalf("expected PrintStatement, got %T", prog.Statements[0])
	}
	if len(ps.Expressions) != 1 {
		t.Fatalf("expected 1 expression, got %d", len(ps.Expressions))
	}
	sl, ok := ps.Expressions[0].(*ast.StringLiteral)
	if !ok {
		t.Fatalf("expected StringLiteral, got %T", ps.Expressions[0])
	}
	if sl.Value != "Hello World" {
		t.Errorf("expected 'Hello World', got %q", sl.Value)
	}
}

func TestParsePrintMultipleExpressions(t *testing.T) {
	prog, errs := parse(`PRINT "x="; x; "y="; y`)
	expectNoErrors(t, errs)
	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}
	ps, ok := prog.Statements[0].(*ast.PrintStatement)
	if !ok {
		t.Fatalf("expected PrintStatement, got %T", prog.Statements[0])
	}
	if len(ps.Expressions) != 4 {
		t.Errorf("expected 4 expressions, got %d", len(ps.Expressions))
	}
}

func TestParseLetAssignment(t *testing.T) {
	prog, errs := parse("LET x = 42")
	expectNoErrors(t, errs)
	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}
	ls, ok := prog.Statements[0].(*ast.LetStatement)
	if !ok {
		t.Fatalf("expected LetStatement, got %T", prog.Statements[0])
	}
	if ls.Name.Name != "x" {
		t.Errorf("expected variable 'x', got %q", ls.Name.Name)
	}
}

func TestParseImplicitAssignment(t *testing.T) {
	prog, errs := parse("x = 42")
	expectNoErrors(t, errs)
	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}
	_, ok := prog.Statements[0].(*ast.LetStatement)
	if !ok {
		t.Fatalf("expected LetStatement for implicit assignment, got %T", prog.Statements[0])
	}
}

func TestParseIfThenElse(t *testing.T) {
	input := `IF x > 0 THEN
  PRINT "positive"
ELSE
  PRINT "non-positive"
END IF`
	prog, errs := parse(input)
	expectNoErrors(t, errs)
	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}
	ifs, ok := prog.Statements[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("expected IfStatement, got %T", prog.Statements[0])
	}
	if len(ifs.ThenBlock) != 1 {
		t.Errorf("expected 1 then statement, got %d", len(ifs.ThenBlock))
	}
	if len(ifs.ElseBlock) != 1 {
		t.Errorf("expected 1 else statement, got %d", len(ifs.ElseBlock))
	}
}

func TestParseSingleLineIf(t *testing.T) {
	prog, errs := parse(`IF x = 1 THEN PRINT "one"`)
	expectNoErrors(t, errs)
	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}
	ifs, ok := prog.Statements[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("expected IfStatement, got %T", prog.Statements[0])
	}
	if !ifs.IsSingleLine {
		t.Error("expected single-line IF")
	}
}

func TestParseForNext(t *testing.T) {
	input := `FOR i = 1 TO 10
  PRINT i
NEXT i`
	prog, errs := parse(input)
	expectNoErrors(t, errs)
	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}
	fs, ok := prog.Statements[0].(*ast.ForStatement)
	if !ok {
		t.Fatalf("expected ForStatement, got %T", prog.Statements[0])
	}
	if fs.Counter.Name != "i" {
		t.Errorf("expected counter 'i', got %q", fs.Counter.Name)
	}
	if fs.Step != nil {
		t.Error("expected nil Step")
	}
	if len(fs.Body) != 1 {
		t.Errorf("expected 1 body statement, got %d", len(fs.Body))
	}
}

func TestParseForNextWithStep(t *testing.T) {
	input := `FOR i = 10 TO 1 STEP -1
  PRINT i
NEXT i`
	prog, errs := parse(input)
	expectNoErrors(t, errs)
	fs := prog.Statements[0].(*ast.ForStatement)
	if fs.Step == nil {
		t.Fatal("expected non-nil Step")
	}
}

func TestParseWhileWend(t *testing.T) {
	input := `WHILE x < 10
  x = x + 1
WEND`
	prog, errs := parse(input)
	expectNoErrors(t, errs)
	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}
	_, ok := prog.Statements[0].(*ast.WhileStatement)
	if !ok {
		t.Fatalf("expected WhileStatement, got %T", prog.Statements[0])
	}
}

func TestParseDoLoop(t *testing.T) {
	input := `DO WHILE x < 10
  x = x + 1
LOOP`
	prog, errs := parse(input)
	expectNoErrors(t, errs)
	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}
	dl, ok := prog.Statements[0].(*ast.DoLoopStatement)
	if !ok {
		t.Fatalf("expected DoLoopStatement, got %T", prog.Statements[0])
	}
	if !dl.TestAtTop {
		t.Error("expected TestAtTop=true for DO WHILE")
	}
	if dl.IsUntil {
		t.Error("expected IsUntil=false for DO WHILE")
	}
}

func TestParseSelectCase(t *testing.T) {
	input := `SELECT CASE x
  CASE 1
    PRINT "one"
  CASE 2, 3
    PRINT "two or three"
  CASE ELSE
    PRINT "other"
END SELECT`
	prog, errs := parse(input)
	expectNoErrors(t, errs)
	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}
	sc, ok := prog.Statements[0].(*ast.SelectCaseStatement)
	if !ok {
		t.Fatalf("expected SelectCaseStatement, got %T", prog.Statements[0])
	}
	if len(sc.Cases) != 2 {
		t.Errorf("expected 2 cases, got %d", len(sc.Cases))
	}
	if len(sc.ElseBlock) != 1 {
		t.Errorf("expected 1 else block statement, got %d", len(sc.ElseBlock))
	}
}

func TestParseGotoGosub(t *testing.T) {
	prog, errs := parse("GOTO 100\nGOSUB myLabel\nRETURN")
	expectNoErrors(t, errs)
	if len(prog.Statements) != 3 {
		t.Fatalf("expected 3 statements, got %d", len(prog.Statements))
	}
	gt, ok := prog.Statements[0].(*ast.GotoStatement)
	if !ok {
		t.Fatalf("expected GotoStatement, got %T", prog.Statements[0])
	}
	if gt.Target != "100" {
		t.Errorf("expected GOTO target '100', got %q", gt.Target)
	}
	gs, ok := prog.Statements[1].(*ast.GosubStatement)
	if !ok {
		t.Fatalf("expected GosubStatement, got %T", prog.Statements[1])
	}
	if gs.Target != "myLabel" {
		t.Errorf("expected GOSUB target 'myLabel', got %q", gs.Target)
	}
	_, ok = prog.Statements[2].(*ast.ReturnStatement)
	if !ok {
		t.Fatalf("expected ReturnStatement, got %T", prog.Statements[2])
	}
}

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

func TestParseLabel(t *testing.T) {
	prog, errs := parse("myLabel:\n  PRINT \"here\"")
	expectNoErrors(t, errs)
	if len(prog.Statements) < 1 {
		t.Fatal("expected at least 1 statement")
	}
	ls, ok := prog.Statements[0].(*ast.LabelStatement)
	if !ok {
		t.Fatalf("expected LabelStatement, got %T", prog.Statements[0])
	}
	if ls.Name != "myLabel" {
		t.Errorf("expected label 'myLabel', got %q", ls.Name)
	}
}

func TestParseLineNumber(t *testing.T) {
	prog, errs := parse("100 PRINT \"hello\"")
	expectNoErrors(t, errs)
	if len(prog.Statements) < 2 {
		t.Fatalf("expected at least 2 statements, got %d", len(prog.Statements))
	}
	ln, ok := prog.Statements[0].(*ast.LineNumberStatement)
	if !ok {
		t.Fatalf("expected LineNumberStatement, got %T", prog.Statements[0])
	}
	if ln.Number != 100 {
		t.Errorf("expected line number 100, got %d", ln.Number)
	}
}

func TestParseDataRead(t *testing.T) {
	prog, errs := parse("DATA 1, 2, 3\nREAD a, b, c")
	expectNoErrors(t, errs)
	if len(prog.Statements) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(prog.Statements))
	}
	ds, ok := prog.Statements[0].(*ast.DataStatement)
	if !ok {
		t.Fatalf("expected DataStatement, got %T", prog.Statements[0])
	}
	if len(ds.Values) != 3 {
		t.Errorf("expected 3 DATA values, got %d", len(ds.Values))
	}
	rs, ok := prog.Statements[1].(*ast.ReadStatement)
	if !ok {
		t.Fatalf("expected ReadStatement, got %T", prog.Statements[1])
	}
	if len(rs.Variables) != 3 {
		t.Errorf("expected 3 READ variables, got %d", len(rs.Variables))
	}
}

func TestParseEnd(t *testing.T) {
	prog, errs := parse("END")
	expectNoErrors(t, errs)
	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}
	_, ok := prog.Statements[0].(*ast.EndStatement)
	if !ok {
		t.Fatalf("expected EndStatement, got %T", prog.Statements[0])
	}
}

func TestParseMultiStatementLine(t *testing.T) {
	prog, errs := parse("x = 1 : y = 2 : PRINT x + y")
	expectNoErrors(t, errs)
	if len(prog.Statements) != 3 {
		t.Fatalf("expected 3 statements, got %d", len(prog.Statements))
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

func TestParseSwap(t *testing.T) {
	prog, errs := parse("SWAP a, b")
	expectNoErrors(t, errs)
	_, ok := prog.Statements[0].(*ast.SwapStatement)
	if !ok {
		t.Fatalf("expected SwapStatement, got %T", prog.Statements[0])
	}
}

func TestParseOnErrorGoto(t *testing.T) {
	prog, errs := parse("ON ERROR GOTO handler")
	expectNoErrors(t, errs)
	oe, ok := prog.Statements[0].(*ast.OnErrorGotoStatement)
	if !ok {
		t.Fatalf("expected OnErrorGotoStatement, got %T", prog.Statements[0])
	}
	if oe.Target != "handler" {
		t.Errorf("expected target 'handler', got %q", oe.Target)
	}
}

func TestParseCls(t *testing.T) {
	prog, errs := parse("CLS")
	expectNoErrors(t, errs)
	_, ok := prog.Statements[0].(*ast.ClsStatement)
	if !ok {
		t.Fatalf("expected ClsStatement, got %T", prog.Statements[0])
	}
}

func TestParseComplexProgram(t *testing.T) {
	input := `10 DIM a(10)
20 FOR i = 1 TO 10
30   a(i) = i * i
40 NEXT i
50 PRINT "Squares:"
60 FOR i = 1 TO 10
70   PRINT a(i)
80 NEXT i
90 END`
	prog, errs := parse(input)
	expectNoErrors(t, errs)
	if len(prog.Statements) < 5 {
		t.Errorf("expected at least 5 statements, got %d", len(prog.Statements))
	}
}

func TestParseErrors(t *testing.T) {
	// This should produce parse errors but not panic
	_, errs := parse("IF THEN")
	if len(errs) == 0 {
		t.Error("expected parse errors for 'IF THEN'")
	}
}

// ---------------------------------------------------------------------------
// Regression tests — each test documents a parser bug found while running
// the example programs. Keep these to prevent regressions.
// ---------------------------------------------------------------------------

// TestRegressionCLEAR verifies that the CLEAR statement (which resets
// variable memory) parses correctly. CLEAR was not a keyword before.
func TestRegressionCLEAR(t *testing.T) {
	_, errs := parse("CLEAR")
	expectNoErrors(t, errs)
	_, errs = parse("CLEAR 500")
	expectNoErrors(t, errs)
}

// TestRegressionOnComputedGoto verifies multi-target ON expr GOTO parsing.
// Before the fix, only ON EVENT GOSUB (KEY, TIMER, etc.) was handled.
func TestRegressionOnComputedGoto(t *testing.T) {
	_, errs := parse("ON choice GOTO Label1, Label2, Label3")
	expectNoErrors(t, errs)
}

// TestRegressionOnComputedGosub verifies multi-target ON expr GOSUB parsing.
func TestRegressionOnComputedGosub(t *testing.T) {
	_, errs := parse("ON n GOSUB Sub1, Sub2, Sub3")
	expectNoErrors(t, errs)
}

// TestRegressionPrintFileOutput verifies PRINT #n, expr for file output.
// Previously only console PRINT was handled.
func TestRegressionPrintFileOutput(t *testing.T) {
	_, errs := parse(`OPEN "out.txt" FOR OUTPUT AS #1
PRINT #1, "hello"
CLOSE #1`)
	expectNoErrors(t, errs)
}

// TestRegressionPrintImplicitConcat verifies PRINT with implicit concatenation
// (adjacent items without a separator). Before the fix the parser stopped at
// the first item and generated errors for subsequent tokens.
func TestRegressionPrintImplicitConcat(t *testing.T) {
	_, errs := parse(`PRINT a$ " world"`)
	expectNoErrors(t, errs)
}

// TestRegressionPrintSemicolonComment verifies that a ; followed by a comment
// does not cause the parser to expect another expression after the comment.
func TestRegressionPrintSemicolonComment(t *testing.T) {
	_, errs := parse("PRINT x; ' end of line comment")
	expectNoErrors(t, errs)
}

// TestRegressionDimShared verifies that DIM SHARED (and LOCAL, STATIC, COMMON)
// modifiers are consumed without being treated as the variable name.
func TestRegressionDimShared(t *testing.T) {
	_, errs := parse("DIM SHARED arr(1 TO 10) AS INTEGER")
	expectNoErrors(t, errs)
	_, errs = parse("DIM STATIC count AS LONG")
	expectNoErrors(t, errs)
}

// TestRegressionDataUnquoted verifies that unquoted DATA values (strings
// containing spaces or special characters) parse correctly.
func TestRegressionDataUnquoted(t *testing.T) {
	_, errs := parse("DATA Hello World, 42, Foo Bar")
	expectNoErrors(t, errs)
}

// TestRegressionInputPrompt verifies that INPUT with a prompt string and
// semicolon separator parses correctly (INPUT "Prompt: "; var).
func TestRegressionInputPrompt(t *testing.T) {
	_, errs := parse(`INPUT "Enter name: "; name$`)
	expectNoErrors(t, errs)
	_, errs = parse(`INPUT "Choose (1-3): ", choice`)
	expectNoErrors(t, errs)
}

// TestRegressionLineInputPrompt verifies LINE INPUT with a prompt string.
func TestRegressionLineInputPrompt(t *testing.T) {
	_, errs := parse(`LINE INPUT "Enter line: "; text$`)
	expectNoErrors(t, errs)
}

// TestRegressionDefSeg verifies DEF SEG with the optional = segment argument.
func TestRegressionDefSeg(t *testing.T) {
	_, errs := parse("DEF SEG = 0")
	expectNoErrors(t, errs)
	_, errs = parse("DEF SEG")
	expectNoErrors(t, errs)
}

// TestRegressionPoke verifies POKE address, value parsing.
// TOKEN_POKE existed in the lexer but had no parser handler before the fix.
func TestRegressionPoke(t *testing.T) {
	_, errs := parse("POKE 1047, 7")
	expectNoErrors(t, errs)
}

// TestRegressionWidth verifies WIDTH statement parsing.
func TestRegressionWidth(t *testing.T) {
	_, errs := parse("WIDTH 80, 25")
	expectNoErrors(t, errs)
	_, errs = parse("WIDTH 40")
	expectNoErrors(t, errs)
}

// TestRegressionConst verifies CONST name = expr parsing.
func TestRegressionConst(t *testing.T) {
	_, errs := parse("CONST PI = 3.14159")
	expectNoErrors(t, errs)
	_, errs = parse(`CONST APPNAME = "MyApp"`)
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

// TestRegressionMetacompilerDirectives verifies that $IF/$ENDIF/$DYNAMIC and
// other metacompiler directives are consumed without error.
// Turbo BASIC uses $ENDIF (not $END IF) as the block-closing directive.
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

// TestRegressionGraphicsPut verifies the graphics PUT (x,y), array, mode form.
// Before the fix, the parser tried to handle it as file PUT.
func TestRegressionGraphicsPut(t *testing.T) {
	_, errs := parse("PUT (x, y), imgData, PSET")
	expectNoErrors(t, errs)
	_, errs = parse("PUT (x, y), imgData, PRESET")
	expectNoErrors(t, errs)
	_, errs = parse("PUT (x, y), imgData, XOR")
	expectNoErrors(t, errs)
}

// TestRegressionGraphicsGet verifies the graphics GET (x1,y1)-(x2,y2), array form.
func TestRegressionGraphicsGet(t *testing.T) {
	_, errs := parse("GET (0, 0)-(100, 100), screen$")
	expectNoErrors(t, errs)
}

// TestRegressionEndInsideIfBlock verifies that a bare END (program terminator)
// inside an IF block is not mistaken for END IF and does not close the block early.
func TestRegressionEndInsideIfBlock(t *testing.T) {
	input := `IF x = 1 THEN
  PRINT "one"
  END
END IF`
	_, errs := parse(input)
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
// The "." character is TOKEN_ILLEGAL in BASIC but used for TYPE field access.
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

// TestRegressionWriteHash verifies WRITE# n, expr... for file output.
// The lexer fuses WRITE and # into the identifier "WRITE#".
func TestRegressionWriteHash(t *testing.T) {
	_, errs := parse(`WRITE# 1, "hello", 42`)
	expectNoErrors(t, errs)
}
