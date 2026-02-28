package parser

import (
	"testing"

	"github.com/loabletech/go-basic/internal/ast"
)

// ===========================================================================
// Helper to get the Nth statement as type T
// ===========================================================================

func getStmt[T ast.Statement](t *testing.T, prog *ast.Program, idx int) T {
	t.Helper()
	if idx >= len(prog.Statements) {
		t.Fatalf("expected at least %d statements, got %d", idx+1, len(prog.Statements))
	}
	s, ok := prog.Statements[idx].(T)
	if !ok {
		t.Fatalf("statement[%d]: expected %T, got %T", idx, *new(T), prog.Statements[idx])
	}
	return s
}

// ===========================================================================
// EXIT statement (parseExitStatement at 0%)
// ===========================================================================

func TestParseExitFor(t *testing.T) {
	prog, errs := parse("EXIT FOR")
	expectNoErrors(t, errs)
	s := getStmt[*ast.ExitStatement](t, prog, 0)
	if s.ExitType != "FOR" {
		t.Errorf("expected ExitType=FOR, got %q", s.ExitType)
	}
}

func TestParseExitDo(t *testing.T) {
	prog, errs := parse("EXIT DO")
	expectNoErrors(t, errs)
	s := getStmt[*ast.ExitStatement](t, prog, 0)
	if s.ExitType != "DO" {
		t.Errorf("expected ExitType=DO, got %q", s.ExitType)
	}
}

func TestParseExitSub(t *testing.T) {
	prog, errs := parse("EXIT SUB")
	expectNoErrors(t, errs)
	s := getStmt[*ast.ExitStatement](t, prog, 0)
	if s.ExitType != "SUB" {
		t.Errorf("expected ExitType=SUB, got %q", s.ExitType)
	}
}

func TestParseExitFunction(t *testing.T) {
	prog, errs := parse("EXIT FUNCTION")
	expectNoErrors(t, errs)
	s := getStmt[*ast.ExitStatement](t, prog, 0)
	if s.ExitType != "FUNCTION" {
		t.Errorf("expected ExitType=FUNCTION, got %q", s.ExitType)
	}
}

// ===========================================================================
// CALL statement (parseCallStatement at 0%)
// ===========================================================================

func TestParseCallStatement(t *testing.T) {
	prog, errs := parse("CALL MySub(5, 10)")
	expectNoErrors(t, errs)
	// CALL MySub(5, 10) should parse as a FunctionCall expression via LetStatement or directly
	if len(prog.Statements) == 0 {
		t.Fatal("expected at least 1 statement")
	}
	// It could be parsed as ExpressionStatement or FunctionCallStatement depending on implementation
	// Just verify it parses without errors
}

func TestParseCallNoArgs(t *testing.T) {
	prog, errs := parse("CALL PrintHeader")
	expectNoErrors(t, errs)
	if len(prog.Statements) == 0 {
		t.Fatal("expected at least 1 statement")
	}
}

// ===========================================================================
// OPTION BASE statement (parseOptionStatement at 0%)
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

// ===========================================================================
// DEFTYPE statements (parseDefTypeStatement at 0%)
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
// SHARED/LOCAL/STATIC statements (parseScopeStatement at 0%)
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
// ERROR statement (parseErrorStatement at 0%)
// ===========================================================================

func TestParseErrorStatement(t *testing.T) {
	prog, errs := parse("ERROR 53")
	expectNoErrors(t, errs)
	s := getStmt[*ast.ErrorStatement](t, prog, 0)
	if s.Code == nil {
		t.Error("expected Code expression to be set")
	}
}

// ===========================================================================
// RANDOMIZE statement (parseRandomizeStatement at 0%)
// ===========================================================================

func TestParseRandomizeTimer(t *testing.T) {
	prog, errs := parse("RANDOMIZE TIMER")
	expectNoErrors(t, errs)
	s := getStmt[*ast.RandomizeStatement](t, prog, 0)
	if s.Seed == nil {
		t.Error("expected RANDOMIZE TIMER to have a Seed expression")
	}
}

func TestParseRandomizeWithSeed(t *testing.T) {
	prog, errs := parse("RANDOMIZE 42")
	expectNoErrors(t, errs)
	s := getStmt[*ast.RandomizeStatement](t, prog, 0)
	if s.Seed == nil {
		t.Error("expected RANDOMIZE 42 to have a Seed expression")
	}
	nl, ok := s.Seed.(*ast.NumberLiteral)
	if !ok {
		t.Fatalf("expected NumberLiteral seed, got %T", s.Seed)
	}
	if nl.Value != 42 {
		t.Errorf("expected seed 42, got %v", nl.Value)
	}
}

func TestParseRandomizeNoArg(t *testing.T) {
	prog, errs := parse("RANDOMIZE\n")
	expectNoErrors(t, errs)
	s := getStmt[*ast.RandomizeStatement](t, prog, 0)
	_ = s // no seed is ok
}

// ===========================================================================
// Graphics: PSET (parsePsetStatement at 0%)
// ===========================================================================

func TestParsePsetBasic(t *testing.T) {
	prog, errs := parse("PSET (100, 50)")
	expectNoErrors(t, errs)
	s := getStmt[*ast.PsetStatement](t, prog, 0)
	if s.X == nil {
		t.Error("expected X to be set")
	}
	if s.Y == nil {
		t.Error("expected Y to be set")
	}
}

func TestParsePsetWithColor(t *testing.T) {
	prog, errs := parse("PSET (100, 50), 14")
	expectNoErrors(t, errs)
	s := getStmt[*ast.PsetStatement](t, prog, 0)
	if s.Color == nil {
		t.Error("expected Color to be set")
	}
}

// ===========================================================================
// Graphics: CIRCLE (parseCircleStatement at 0%)
// ===========================================================================

func TestParseCircleBasic(t *testing.T) {
	prog, errs := parse("CIRCLE (160, 100), 50")
	expectNoErrors(t, errs)
	s := getStmt[*ast.CircleStmt](t, prog, 0)
	if s.X == nil {
		t.Error("expected X")
	}
	if s.Y == nil {
		t.Error("expected Y")
	}
	if s.Radius == nil {
		t.Error("expected Radius")
	}
}

func TestParseCircleWithColor(t *testing.T) {
	prog, errs := parse("CIRCLE (160, 100), 50, 14")
	expectNoErrors(t, errs)
	s := getStmt[*ast.CircleStmt](t, prog, 0)
	if s.Color == nil {
		t.Error("expected Color to be set")
	}
}

// ===========================================================================
// Graphics: PAINT (parsePaintStatement at 0%)
// ===========================================================================

func TestParsePaintBasic(t *testing.T) {
	prog, errs := parse("PAINT (80, 60), 2")
	expectNoErrors(t, errs)
	s := getStmt[*ast.PaintStmt](t, prog, 0)
	if s.X == nil {
		t.Error("expected X")
	}
	if s.FillColor == nil {
		t.Error("expected FillColor")
	}
}

func TestParsePaintWithBorder(t *testing.T) {
	prog, errs := parse("PAINT (80, 60), 2, 1")
	expectNoErrors(t, errs)
	s := getStmt[*ast.PaintStmt](t, prog, 0)
	if s.BorderColor == nil {
		t.Error("expected BorderColor")
	}
}

// ===========================================================================
// Graphics: SCREEN (parseScreenStatement at 0%)
// ===========================================================================

func TestParseScreen(t *testing.T) {
	prog, errs := parse("SCREEN 9")
	expectNoErrors(t, errs)
	s := getStmt[*ast.ScreenStatement](t, prog, 0)
	if s.Mode == nil {
		t.Error("expected Mode expression")
	}
}

func TestParseScreenZero(t *testing.T) {
	prog, errs := parse("SCREEN 0")
	expectNoErrors(t, errs)
	s := getStmt[*ast.ScreenStatement](t, prog, 0)
	if s.Mode == nil {
		t.Error("expected Mode expression")
	}
}

// ===========================================================================
// Graphics: LINE statement - remaining branches (at 26.5%)
// ===========================================================================

func TestParseLineBasic(t *testing.T) {
	prog, errs := parse("LINE (0,0)-(100,100)")
	expectNoErrors(t, errs)
	s := getStmt[*ast.LineStmt](t, prog, 0)
	if s.X1 == nil || s.Y1 == nil || s.X2 == nil || s.Y2 == nil {
		t.Error("expected all 4 coordinates")
	}
}

func TestParseLineWithColor(t *testing.T) {
	prog, errs := parse("LINE (0,0)-(100,100), 15")
	expectNoErrors(t, errs)
	s := getStmt[*ast.LineStmt](t, prog, 0)
	if s.Color == nil {
		t.Error("expected Color")
	}
}

func TestParseLineBoxFill(t *testing.T) {
	prog, errs := parse("LINE (0,0)-(100,100), 15, BF")
	expectNoErrors(t, errs)
	s := getStmt[*ast.LineStmt](t, prog, 0)
	if s.BoxMode != "BF" {
		t.Errorf("expected BoxMode=BF, got %q", s.BoxMode)
	}
}

func TestParseLineBox(t *testing.T) {
	prog, errs := parse("LINE (0,0)-(100,100), 15, B")
	expectNoErrors(t, errs)
	s := getStmt[*ast.LineStmt](t, prog, 0)
	if s.BoxMode != "B" {
		t.Errorf("expected BoxMode=B, got %q", s.BoxMode)
	}
}

// ===========================================================================
// Graphics: VIEW (parseViewStatement at 0%)
// ===========================================================================

func TestParseViewport(t *testing.T) {
	// VIEW graphics viewport: coordinates are currently skipped by the parser
	// but it should still return a ViewStatement with IsPrint=false
	prog, errs := parse("VIEW (10, 10)-(300, 180)")
	expectNoErrors(t, errs)
	s := getStmt[*ast.ViewStatement](t, prog, 0)
	if s.IsPrint {
		t.Error("expected IsPrint=false for VIEW (not VIEW PRINT)")
	}
}

func TestParseViewPrint(t *testing.T) {
	prog, errs := parse("VIEW PRINT 1 TO 24")
	expectNoErrors(t, errs)
	s := getStmt[*ast.ViewStatement](t, prog, 0)
	if !s.IsPrint {
		t.Error("expected IsPrint=true for VIEW PRINT")
	}
	if s.Top == nil {
		t.Error("expected Top expression")
	}
}

// ===========================================================================
// Display: LOCATE (parseLocateStatement at 0%)
// ===========================================================================

func TestParseLocate(t *testing.T) {
	prog, errs := parse("LOCATE 10, 20")
	expectNoErrors(t, errs)
	s := getStmt[*ast.LocateStatement](t, prog, 0)
	if s.Row == nil {
		t.Error("expected Row")
	}
	if s.Col == nil {
		t.Error("expected Col")
	}
}

// ===========================================================================
// Display: COLOR (parseColorStatement at 0%)
// ===========================================================================

func TestParseColor(t *testing.T) {
	prog, errs := parse("COLOR 14, 1")
	expectNoErrors(t, errs)
	s := getStmt[*ast.ColorStatement](t, prog, 0)
	if s.Foreground == nil {
		t.Error("expected Foreground")
	}
	if s.Background == nil {
		t.Error("expected Background")
	}
}

func TestParseColorForegroundOnly(t *testing.T) {
	prog, errs := parse("COLOR 7")
	expectNoErrors(t, errs)
	s := getStmt[*ast.ColorStatement](t, prog, 0)
	if s.Foreground == nil {
		t.Error("expected Foreground")
	}
}

// ===========================================================================
// Audio: SOUND (parseSoundStatement at 0%)
// ===========================================================================

func TestParseSound(t *testing.T) {
	prog, errs := parse("SOUND 440, 18")
	expectNoErrors(t, errs)
	s := getStmt[*ast.SoundStatement](t, prog, 0)
	if s.Frequency == nil {
		t.Error("expected Frequency")
	}
	if s.Duration == nil {
		t.Error("expected Duration")
	}
}

// ===========================================================================
// Audio: PLAY (parsePlayDrawStatement at 0%)
// ===========================================================================

func TestParsePlay(t *testing.T) {
	prog, errs := parse(`PLAY "CDE"`)
	expectNoErrors(t, errs)
	s := getStmt[*ast.PlayStatement](t, prog, 0)
	if s.CommandString == nil {
		t.Error("expected CommandString")
	}
}

func TestParseDraw(t *testing.T) {
	prog, errs := parse(`DRAW "BM50,50"`)
	expectNoErrors(t, errs)
	s := getStmt[*ast.DrawStmt](t, prog, 0)
	if s.CommandString == nil {
		t.Error("expected CommandString")
	}
}

// ===========================================================================
// File I/O: WRITE# (parseWriteStatement at 0%)
// ===========================================================================

func TestParseWriteStatement(t *testing.T) {
	prog, errs := parse("WRITE #1, x, y$, 42")
	expectNoErrors(t, errs)
	s := getStmt[*ast.FileWriteStatement](t, prog, 0)
	if s.FileNum == nil {
		t.Error("expected FileNum")
	}
	if len(s.Expressions) == 0 {
		t.Error("expected Expressions")
	}
}

// ===========================================================================
// File I/O: SEEK (parseSeekStatement at 0%)
// ===========================================================================

func TestParseSeek(t *testing.T) {
	prog, errs := parse("SEEK #1, 100")
	expectNoErrors(t, errs)
	s := getStmt[*ast.SeekStatement](t, prog, 0)
	if s.FileNum == nil {
		t.Error("expected FileNum")
	}
	if s.Position == nil {
		t.Error("expected Position")
	}
}

// ===========================================================================
// File I/O: FIELD (parseFieldStatement at 0%)
// ===========================================================================

func TestParseField(t *testing.T) {
	prog, errs := parse("FIELD #1, 20 AS name$, 4 AS age$")
	expectNoErrors(t, errs)
	s := getStmt[*ast.FieldStatement](t, prog, 0)
	if s.FileNum == nil {
		t.Error("expected FileNum")
	}
	if len(s.Fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(s.Fields))
	}
}

// ===========================================================================
// File I/O: LSET/RSET (parseLsetRsetStatement at 0%)
// ===========================================================================

func TestParseLset(t *testing.T) {
	prog, errs := parse(`LSET name$ = "John"`)
	expectNoErrors(t, errs)
	s := getStmt[*ast.LsetStatement](t, prog, 0)
	if s.Variable == "" {
		t.Error("expected Variable to be set")
	}
	if s.Value == nil {
		t.Error("expected Value")
	}
}

func TestParseRset(t *testing.T) {
	prog, errs := parse(`RSET code$ = "ABC"`)
	expectNoErrors(t, errs)
	s := getStmt[*ast.RsetStatement](t, prog, 0)
	if s.Variable == "" {
		t.Error("expected Variable to be set")
	}
}

// ===========================================================================
// File I/O: NAME (parseNameStatement at 0%)
// ===========================================================================

func TestParseNameStatement(t *testing.T) {
	prog, errs := parse(`NAME "old.dat" AS "new.dat"`)
	expectNoErrors(t, errs)
	s := getStmt[*ast.NameStatement](t, prog, 0)
	if s.OldName == nil {
		t.Error("expected OldName")
	}
	if s.NewName == nil {
		t.Error("expected NewName")
	}
}

// ===========================================================================
// Parser low-coverage: parseStatement additional dispatches
// ===========================================================================

func TestParseInputWithPrompt(t *testing.T) {
	prog, errs := parse(`INPUT "Enter name: "; name$`)
	expectNoErrors(t, errs)
	s := getStmt[*ast.ReadStatement](t, prog, 0)
	if s.Prompt == "" {
		t.Error("expected Prompt to be set")
	}
	if len(s.Variables) == 0 {
		t.Error("expected at least 1 variable")
	}
}

func TestParseInputMultiVar(t *testing.T) {
	prog, errs := parse("INPUT x%, y%, z$")
	expectNoErrors(t, errs)
	s := getStmt[*ast.ReadStatement](t, prog, 0)
	if len(s.Variables) != 3 {
		t.Errorf("expected 3 variables, got %d", len(s.Variables))
	}
}

func TestParseLineInput(t *testing.T) {
	prog, errs := parse("LINE INPUT x$")
	expectNoErrors(t, errs)
	s := getStmt[*ast.ReadStatement](t, prog, 0)
	if !s.IsLineInput {
		t.Error("expected IsLineInput=true")
	}
}

func TestParseCaseIsComparison(t *testing.T) {
	prog, errs := parse(`SELECT CASE x%
CASE IS > 5
  PRINT "big"
CASE 1 TO 4
  PRINT "small"
END SELECT`)
	expectNoErrors(t, errs)
	s := getStmt[*ast.SelectCaseStatement](t, prog, 0)
	if len(s.Cases) < 2 {
		t.Fatalf("expected at least 2 CASE clauses, got %d", len(s.Cases))
	}
	if !s.Cases[0].Values[0].IsComparison {
		t.Error("first CASE should be IS comparison")
	}
	if !s.Cases[1].Values[0].IsRange {
		t.Error("second CASE should be a range")
	}
}

func TestParseOnComputedGoto(t *testing.T) {
	prog, errs := parse("10 ON x% GOTO 100, 200, 300\n100 PRINT \"a\"\n200 PRINT \"b\"\n300 PRINT \"c\"")
	expectNoErrors(t, errs)
	// Find the OnComputedGotoStatement
	found := false
	for _, stmt := range prog.Statements {
		if cg, ok := stmt.(*ast.OnComputedGotoStatement); ok {
			found = true
			if len(cg.Targets) != 3 {
				t.Errorf("expected 3 targets, got %d", len(cg.Targets))
			}
			break
		}
	}
	if !found {
		t.Error("expected OnComputedGotoStatement")
	}
}

func TestParseOnComputedGosub(t *testing.T) {
	prog, errs := parse("sub1:\nON x% GOSUB sub1, sub2\nsub2:")
	expectNoErrors(t, errs)
	found := false
	for _, stmt := range prog.Statements {
		if cg, ok := stmt.(*ast.OnComputedGosubStatement); ok {
			found = true
			if len(cg.Targets) != 2 {
				t.Errorf("expected 2 targets, got %d", len(cg.Targets))
			}
			break
		}
	}
	if !found {
		t.Error("expected OnComputedGosubStatement")
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

func TestParseOperatorPrecedences(t *testing.T) {
	// Test MOD and AND operators via expression parsing
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
	// Just verify it parses without error - precedence is tested by the tree structure
}

func TestParseResumeNext(t *testing.T) {
	prog, errs := parse("RESUME NEXT")
	expectNoErrors(t, errs)
	s := getStmt[*ast.ResumeStatement](t, prog, 0)
	if s.Type != "NEXT" {
		t.Errorf("expected RESUME NEXT, got Type=%q", s.Type)
	}
}

func TestParseResumeWithLabel(t *testing.T) {
	prog, errs := parse("handler:\nRESUME handler")
	expectNoErrors(t, errs)
	found := false
	for _, stmt := range prog.Statements {
		if r, ok := stmt.(*ast.ResumeStatement); ok {
			found = true
			if r.Type == "" {
				t.Error("expected non-empty resume target")
			}
			break
		}
	}
	if !found {
		t.Error("expected ResumeStatement")
	}
}

func TestParseDoLoopUntil(t *testing.T) {
	prog, errs := parse(`DO
  x% = x% + 1
LOOP UNTIL x% >= 10`)
	expectNoErrors(t, errs)
	s := getStmt[*ast.DoLoopStatement](t, prog, 0)
	if !s.IsUntil {
		t.Error("expected IsUntil=true")
	}
}

func TestParseDoWhileAtBottom(t *testing.T) {
	prog, errs := parse(`DO
  x% = x% + 1
LOOP WHILE x% < 10`)
	expectNoErrors(t, errs)
	s := getStmt[*ast.DoLoopStatement](t, prog, 0)
	if s.IsUntil {
		t.Error("expected IsUntil=false for WHILE")
	}
	if s.Condition == nil {
		t.Error("expected Condition")
	}
}

func TestParseForWithStep(t *testing.T) {
	prog, errs := parse("FOR i% = 1 TO 10 STEP 2\nNEXT i%")
	expectNoErrors(t, errs)
	s := getStmt[*ast.ForStatement](t, prog, 0)
	if s.Step == nil {
		t.Error("expected Step expression")
	}
}

func TestParseDeclare(t *testing.T) {
	prog, errs := parse("DECLARE SUB MySub(x AS INTEGER, y AS STRING)")
	expectNoErrors(t, errs)
	// DECLARE creates a forward-declared SubDeclaration
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
