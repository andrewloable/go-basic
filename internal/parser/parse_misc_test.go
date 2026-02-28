package parser

import (
	"testing"

	"github.com/loabletech/go-basic/internal/ast"
)

// ===========================================================================
// Tests for parse_misc.go: PRINT, INPUT, LINE INPUT, REM/label, DATA, RESTORE,
// CLS, COLOR, LOCATE, SCREEN, SOUND, PLAY, DRAW, VIEW, LINE (graphics),
// CIRCLE, PAINT, PSET/PRESET, WINDOW, PALETTE
// ===========================================================================

// ===========================================================================
// PRINT statement variants
// ===========================================================================

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

// TestRegressionPrintImplicitConcat verifies PRINT with implicit concatenation.
func TestRegressionPrintImplicitConcat(t *testing.T) {
	_, errs := parse(`PRINT a$ " world"`)
	expectNoErrors(t, errs)
}

// TestRegressionPrintSemicolonComment verifies that a ; followed by a comment
// does not cause the parser to expect another expression.
func TestRegressionPrintSemicolonComment(t *testing.T) {
	_, errs := parse("PRINT x; ' end of line comment")
	expectNoErrors(t, errs)
}

func TestParsePrintUsing(t *testing.T) {
	prog, errs := parse(`PRINT USING "##.##"; 3.14`)
	expectNoErrors(t, errs)
	s := getStmt[*ast.PrintStatement](t, prog, 0)
	if s.Format == nil {
		t.Error("expected Format expression for PRINT USING")
	}
}

func TestParsePrintTrailingSemicolon(t *testing.T) {
	prog, errs := parse(`PRINT "hello";`)
	expectNoErrors(t, errs)
	s := getStmt[*ast.PrintStatement](t, prog, 0)
	if !s.HasTrailingSep {
		t.Error("expected HasTrailingSep for trailing semicolon")
	}
}

// ===========================================================================
// INPUT statement variants
// ===========================================================================

// TestRegressionInputPrompt verifies that INPUT with a prompt string and
// semicolon separator parses correctly.
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

func TestParseInputWithCommaPrompt(t *testing.T) {
	prog, errs := parse(`INPUT "Name? ", name$`)
	expectNoErrors(t, errs)
	s := getStmt[*ast.ReadStatement](t, prog, 0)
	if s.Prompt == "" {
		t.Error("expected Prompt")
	}
}

func TestParseLineInputWithPrompt(t *testing.T) {
	prog, errs := parse(`LINE INPUT "Enter: "; line$`)
	expectNoErrors(t, errs)
	s := getStmt[*ast.ReadStatement](t, prog, 0)
	if s.Prompt == "" {
		t.Error("expected Prompt")
	}
	if !s.IsLineInput {
		t.Error("expected IsLineInput=true")
	}
}

// ===========================================================================
// DATA and RESTORE
// ===========================================================================

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

// TestRegressionDataUnquoted verifies that unquoted DATA values parse correctly.
func TestRegressionDataUnquoted(t *testing.T) {
	_, errs := parse("DATA Hello World, 42, Foo Bar")
	expectNoErrors(t, errs)
}

func TestParseDataNegativeNumber(t *testing.T) {
	prog, errs := parse("DATA -1, -3.14, 42")
	expectNoErrors(t, errs)
	s := getStmt[*ast.DataStatement](t, prog, 0)
	if len(s.Values) < 3 {
		t.Fatalf("expected 3 values, got %d", len(s.Values))
	}
	nl, ok := s.Values[0].(*ast.NumberLiteral)
	if !ok {
		t.Fatalf("expected NumberLiteral, got %T", s.Values[0])
	}
	if nl.Value != -1 {
		t.Errorf("expected -1, got %v", nl.Value)
	}
}

func TestParseDataStringValue(t *testing.T) {
	prog, errs := parse(`DATA "hello", "world"`)
	expectNoErrors(t, errs)
	s := getStmt[*ast.DataStatement](t, prog, 0)
	if len(s.Values) != 2 {
		t.Fatalf("expected 2 values, got %d", len(s.Values))
	}
	sl, ok := s.Values[0].(*ast.StringLiteral)
	if !ok {
		t.Fatalf("expected StringLiteral, got %T", s.Values[0])
	}
	if sl.Value != "hello" {
		t.Errorf("expected 'hello', got %q", sl.Value)
	}
}

func TestParseDataUnquotedString(t *testing.T) {
	prog, errs := parse("DATA hello, world")
	expectNoErrors(t, errs)
	s := getStmt[*ast.DataStatement](t, prog, 0)
	if len(s.Values) < 1 {
		t.Fatal("expected at least 1 value")
	}
}

func TestParseDataMixedTypes(t *testing.T) {
	prog, errs := parse(`DATA 1, "text", -42, 3.14`)
	expectNoErrors(t, errs)
	s := getStmt[*ast.DataStatement](t, prog, 0)
	if len(s.Values) != 4 {
		t.Errorf("expected 4 values, got %d", len(s.Values))
	}
}

func TestParseDataNegativeFloat(t *testing.T) {
	prog, errs := parse("DATA -3.14")
	expectNoErrors(t, errs)
	s := getStmt[*ast.DataStatement](t, prog, 0)
	if len(s.Values) != 1 {
		t.Fatalf("expected 1 value, got %d", len(s.Values))
	}
	nl, ok := s.Values[0].(*ast.NumberLiteral)
	if !ok {
		t.Fatalf("expected NumberLiteral, got %T", s.Values[0])
	}
	if nl.Value >= 0 {
		t.Errorf("expected negative value, got %v", nl.Value)
	}
}

func TestParseRestoreNoTarget(t *testing.T) {
	prog, errs := parse("RESTORE")
	expectNoErrors(t, errs)
	s := getStmt[*ast.RestoreStatement](t, prog, 0)
	if s.Target != "" {
		t.Errorf("expected empty target, got %q", s.Target)
	}
}

func TestParseRestoreWithLabel(t *testing.T) {
	prog, errs := parse("dataStart:\nRESTORE dataStart")
	expectNoErrors(t, errs)
	found := false
	for _, stmt := range prog.Statements {
		if r, ok := stmt.(*ast.RestoreStatement); ok {
			found = true
			if r.Target != "dataStart" {
				t.Errorf("expected 'dataStart', got %q", r.Target)
			}
		}
	}
	if !found {
		t.Error("expected RestoreStatement")
	}
}

func TestParseRestoreWithLineNumber(t *testing.T) {
	prog, errs := parse("100 DATA 1, 2, 3\nRESTORE 100")
	expectNoErrors(t, errs)
	found := false
	for _, stmt := range prog.Statements {
		if r, ok := stmt.(*ast.RestoreStatement); ok {
			found = true
			if r.Target == "" {
				t.Error("expected non-empty target for line number")
			}
		}
	}
	if !found {
		t.Error("expected RestoreStatement")
	}
}

// ===========================================================================
// CLS statement
// ===========================================================================

func TestParseCls(t *testing.T) {
	prog, errs := parse("CLS")
	expectNoErrors(t, errs)
	_, ok := prog.Statements[0].(*ast.ClsStatement)
	if !ok {
		t.Fatalf("expected ClsStatement, got %T", prog.Statements[0])
	}
}

func TestParseClsWithMode(t *testing.T) {
	prog, errs := parse("CLS 1")
	expectNoErrors(t, errs)
	s := getStmt[*ast.ClsStatement](t, prog, 0)
	if s.Mode == nil {
		t.Error("expected Mode")
	}
}

func TestParseClsNoMode(t *testing.T) {
	prog, errs := parse("CLS")
	expectNoErrors(t, errs)
	s := getStmt[*ast.ClsStatement](t, prog, 0)
	_ = s // Mode can be nil
}

// ===========================================================================
// COLOR statement
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
// LOCATE statement
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
// SCREEN statement
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

func TestParseScreenWithColorSwitch(t *testing.T) {
	prog, errs := parse("SCREEN 1, 0")
	expectNoErrors(t, errs)
	s := getStmt[*ast.ScreenStatement](t, prog, 0)
	if s.Mode == nil {
		t.Error("expected Mode")
	}
	if s.ColorSwitch == nil {
		t.Error("expected ColorSwitch")
	}
}

func TestParseScreenWithAllParams(t *testing.T) {
	prog, errs := parse("SCREEN 1, 0, 0, 0")
	expectNoErrors(t, errs)
	s := getStmt[*ast.ScreenStatement](t, prog, 0)
	if s.Mode == nil {
		t.Error("expected Mode")
	}
	if s.ColorSwitch == nil {
		t.Error("expected ColorSwitch")
	}
	if s.ActivePage == nil {
		t.Error("expected ActivePage")
	}
	if s.VisualPage == nil {
		t.Error("expected VisualPage")
	}
}

func TestParseScreenSkipColorSwitch(t *testing.T) {
	prog, errs := parse("SCREEN 1, , 0")
	expectNoErrors(t, errs)
	s := getStmt[*ast.ScreenStatement](t, prog, 0)
	if s.Mode == nil {
		t.Error("expected Mode")
	}
	if s.ColorSwitch != nil {
		t.Error("expected nil ColorSwitch when skipped")
	}
}

func TestParseScreenSkipActivePage(t *testing.T) {
	prog, errs := parse("SCREEN 1, 0, , 1")
	expectNoErrors(t, errs)
	s := getStmt[*ast.ScreenStatement](t, prog, 0)
	if s.Mode == nil {
		t.Error("expected Mode")
	}
	if s.ActivePage != nil {
		t.Error("expected nil ActivePage when skipped")
	}
	if s.VisualPage == nil {
		t.Error("expected VisualPage")
	}
}

// ===========================================================================
// SOUND statement
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
// PLAY and DRAW (audio/graphics command strings)
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
// Graphics: LINE statement
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
// Graphics: CIRCLE statement
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

func TestParseCircleWithStartEnd(t *testing.T) {
	prog, errs := parse("CIRCLE (160, 100), 50, 14, 0, 3.14")
	expectNoErrors(t, errs)
	s := getStmt[*ast.CircleStmt](t, prog, 0)
	if s.Color == nil {
		t.Error("expected Color")
	}
	if s.Start == nil {
		t.Error("expected Start")
	}
	if s.End == nil {
		t.Error("expected End")
	}
}

func TestParseCircleWithAspect(t *testing.T) {
	prog, errs := parse("CIRCLE (160, 100), 50, 14, 0, 3.14, 1.5")
	expectNoErrors(t, errs)
	s := getStmt[*ast.CircleStmt](t, prog, 0)
	if s.Aspect == nil {
		t.Error("expected Aspect")
	}
}

func TestParseCircleSkipColor(t *testing.T) {
	prog, errs := parse("CIRCLE (80, 80), 40, , 0")
	expectNoErrors(t, errs)
	s := getStmt[*ast.CircleStmt](t, prog, 0)
	if s.Color != nil {
		t.Error("expected nil Color when skipped")
	}
	if s.Start == nil {
		t.Error("expected Start")
	}
}

func TestParseCircleSkipStart(t *testing.T) {
	prog, errs := parse("CIRCLE (80, 80), 40, 14, , 3.14")
	expectNoErrors(t, errs)
	s := getStmt[*ast.CircleStmt](t, prog, 0)
	if s.Color == nil {
		t.Error("expected Color")
	}
	if s.Start != nil {
		t.Error("expected nil Start when skipped")
	}
	if s.End == nil {
		t.Error("expected End")
	}
}

// ===========================================================================
// Graphics: PAINT statement
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
// Graphics: PSET / PRESET statement
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

func TestParsePsetStep(t *testing.T) {
	prog, errs := parse("PSET STEP(10, 5)")
	expectNoErrors(t, errs)
	s := getStmt[*ast.PsetStatement](t, prog, 0)
	if !s.IsStep {
		t.Error("expected IsStep=true")
	}
}

func TestParsePreset(t *testing.T) {
	prog, errs := parse("PRESET (100, 50)")
	expectNoErrors(t, errs)
	s := getStmt[*ast.PsetStatement](t, prog, 0)
	if !s.IsPreset {
		t.Error("expected IsPreset=true")
	}
}

// ===========================================================================
// Graphics: VIEW statement
// ===========================================================================

func TestParseViewport(t *testing.T) {
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
// WINDOW and PALETTE (stub statements)
// ===========================================================================

func TestParseWindowStatement(t *testing.T) {
	prog, errs := parse("WINDOW (0,0)-(640,480)")
	expectNoErrors(t, errs)
	s := getStmt[*ast.RemStatement](t, prog, 0)
	if s.Text != "WINDOW" {
		t.Errorf("expected Text='WINDOW', got %q", s.Text)
	}
}

func TestParseWindowStatementNoArgs(t *testing.T) {
	prog, errs := parse("WINDOW")
	expectNoErrors(t, errs)
	s := getStmt[*ast.RemStatement](t, prog, 0)
	if s.Text != "WINDOW" {
		t.Errorf("expected Text='WINDOW', got %q", s.Text)
	}
}

func TestParsePaletteStatement(t *testing.T) {
	prog, errs := parse("PALETTE 1, 4")
	expectNoErrors(t, errs)
	s := getStmt[*ast.RemStatement](t, prog, 0)
	if s.Text != "PALETTE" {
		t.Errorf("expected Text='PALETTE', got %q", s.Text)
	}
}

func TestParsePaletteStatementNoArgs(t *testing.T) {
	prog, errs := parse("PALETTE")
	expectNoErrors(t, errs)
	s := getStmt[*ast.RemStatement](t, prog, 0)
	if s.Text != "PALETTE" {
		t.Errorf("expected Text='PALETTE', got %q", s.Text)
	}
}

// ===========================================================================
// WRITE (console) statement
// ===========================================================================

func TestParseWriteConsole(t *testing.T) {
	prog, errs := parse("WRITE x%, y$")
	expectNoErrors(t, errs)
	s := getStmt[*ast.PrintStatement](t, prog, 0)
	if len(s.Expressions) == 0 {
		t.Error("expected Expressions")
	}
}

// ===========================================================================
// Label and line number
// ===========================================================================

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

// ===========================================================================
// DELAY statement
// ===========================================================================

func TestParseDelayStatement(t *testing.T) {
	prog, errs := parse("DELAY 1.5")
	expectNoErrors(t, errs)
	s := getStmt[*ast.LetStatement](t, prog, 0)
	if s.Name == nil || s.Name.Name != "DELAY" {
		t.Errorf("expected DELAY name, got %v", s.Name)
	}
	if s.Value == nil {
		t.Error("expected Value expression")
	}
}

func TestParseDelayStatementInt(t *testing.T) {
	prog, errs := parse("DELAY 2")
	expectNoErrors(t, errs)
	s := getStmt[*ast.LetStatement](t, prog, 0)
	if s.Name.Name != "DELAY" {
		t.Errorf("expected DELAY name, got %q", s.Name.Name)
	}
}
