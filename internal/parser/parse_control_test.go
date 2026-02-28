package parser

import (
	"testing"

	"github.com/loabletech/go-basic/internal/ast"
)

// ===========================================================================
// Tests for parse_control.go: IF, FOR, WHILE, DO, SELECT CASE, GOTO, GOSUB,
// ON, EXIT, RESUME
// ===========================================================================

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

// TestRegressionOnComputedGoto verifies multi-target ON expr GOTO parsing.
func TestRegressionOnComputedGoto(t *testing.T) {
	_, errs := parse("ON choice GOTO Label1, Label2, Label3")
	expectNoErrors(t, errs)
}

// TestRegressionOnComputedGosub verifies multi-target ON expr GOSUB parsing.
func TestRegressionOnComputedGosub(t *testing.T) {
	_, errs := parse("ON n GOSUB Sub1, Sub2, Sub3")
	expectNoErrors(t, errs)
}

// TestRegressionEndInsideIfBlock verifies that a bare END inside an IF block
// is not mistaken for END IF and does not close the block early.
func TestRegressionEndInsideIfBlock(t *testing.T) {
	input := `IF x = 1 THEN
  PRINT "one"
  END
END IF`
	_, errs := parse(input)
	expectNoErrors(t, errs)
}

// TestRegressionGraphicsPut verifies the graphics PUT (x,y), array, mode form.
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

// ===========================================================================
// EXIT statement
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
// RESUME statement
// ===========================================================================

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

// ===========================================================================
// DO LOOP variants
// ===========================================================================

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

func TestParseDoLoopWithWhileAtTop(t *testing.T) {
	prog, errs := parse(`DO WHILE x% < 10
x% = x% + 1
LOOP`)
	expectNoErrors(t, errs)
	s := getStmt[*ast.DoLoopStatement](t, prog, 0)
	if s.Condition == nil {
		t.Error("expected Condition")
	}
}

func TestParseDoUntilAtTop(t *testing.T) {
	prog, errs := parse(`DO UNTIL x% >= 10
x% = x% + 1
LOOP`)
	expectNoErrors(t, errs)
	s := getStmt[*ast.DoLoopStatement](t, prog, 0)
	if s.Condition == nil {
		t.Error("expected Condition")
	}
	if !s.IsUntil {
		t.Error("expected IsUntil=true")
	}
}

// ===========================================================================
// FOR statement variants
// ===========================================================================

func TestParseForWithStep(t *testing.T) {
	prog, errs := parse("FOR i% = 1 TO 10 STEP 2\nNEXT i%")
	expectNoErrors(t, errs)
	s := getStmt[*ast.ForStatement](t, prog, 0)
	if s.Step == nil {
		t.Error("expected Step expression")
	}
}

func TestParseForMissingIdentifier(t *testing.T) {
	_, errs := parse("FOR = 1 TO 10\nNEXT")
	if len(errs) == 0 {
		t.Error("expected error for FOR without identifier")
	}
}

func TestParseForMissingEquals(t *testing.T) {
	_, errs := parse("FOR x 1 TO 10\nNEXT")
	if len(errs) == 0 {
		t.Error("expected error for FOR missing =")
	}
}

func TestParseForMissingTO(t *testing.T) {
	_, errs := parse("FOR x = 1 10\nNEXT")
	if len(errs) == 0 {
		t.Error("expected error for FOR missing TO")
	}
}

func TestParseForWithNextCounter(t *testing.T) {
	prog, errs := parse("FOR i = 1 TO 5\nNEXT i")
	expectNoErrors(t, errs)
	getStmt[*ast.ForStatement](t, prog, 0)
}

// ===========================================================================
// ON statement: computed GOTO/GOSUB and event handlers
// ===========================================================================

func TestParseOnComputedGoto(t *testing.T) {
	prog, errs := parse("10 ON x% GOTO 100, 200, 300\n100 PRINT \"a\"\n200 PRINT \"b\"\n300 PRINT \"c\"")
	expectNoErrors(t, errs)
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

func TestParseOnKeyGosub(t *testing.T) {
	prog, errs := parse("ON KEY(1) GOSUB handler\nhandler:")
	expectNoErrors(t, errs)
	found := false
	for _, stmt := range prog.Statements {
		if ev, ok := stmt.(*ast.OnEventGosubStatement); ok {
			found = true
			if ev.EventType != "KEY" {
				t.Errorf("expected KEY event, got %q", ev.EventType)
			}
			if ev.Target != "handler" {
				t.Errorf("expected target 'handler', got %q", ev.Target)
			}
		}
	}
	if !found {
		t.Error("expected OnEventGosubStatement")
	}
}

func TestParseOnTimerGosub(t *testing.T) {
	prog, errs := parse("ON TIMER(5) GOSUB timerSub\ntimerSub:")
	expectNoErrors(t, errs)
	found := false
	for _, stmt := range prog.Statements {
		if ev, ok := stmt.(*ast.OnEventGosubStatement); ok {
			found = true
			if ev.EventType != "TIMER" {
				t.Errorf("expected TIMER event, got %q", ev.EventType)
			}
		}
	}
	if !found {
		t.Error("expected OnEventGosubStatement")
	}
}

func TestParseOnPlayGosub(t *testing.T) {
	prog, errs := parse("ON PLAY(32) GOSUB musicSub\nmusicSub:")
	expectNoErrors(t, errs)
	found := false
	for _, stmt := range prog.Statements {
		if ev, ok := stmt.(*ast.OnEventGosubStatement); ok {
			found = true
			if ev.EventType != "PLAY" {
				t.Errorf("expected PLAY event, got %q", ev.EventType)
			}
		}
	}
	if !found {
		t.Error("expected OnEventGosubStatement")
	}
}

func TestParseOnStrigGosub(t *testing.T) {
	prog, errs := parse("ON STRIG(0) GOSUB strigSub\nstrigSub:")
	expectNoErrors(t, errs)
	found := false
	for _, stmt := range prog.Statements {
		if ev, ok := stmt.(*ast.OnEventGosubStatement); ok {
			found = true
			if ev.EventType != "STRIG" {
				t.Errorf("expected STRIG event, got %q", ev.EventType)
			}
		}
	}
	if !found {
		t.Error("expected OnEventGosubStatement")
	}
}

func TestParseOnNoGotoGosub(t *testing.T) {
	// ON expr without GOTO or GOSUB triggers addError path
	_, errs := parse("ON x% PRINT")
	if len(errs) == 0 {
		t.Error("expected parse error for ON expr without GOTO/GOSUB")
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

func TestParseOnErrorGotoWithLabel(t *testing.T) {
	prog, errs := parse("ON ERROR GOTO errHandler\nerrHandler:")
	expectNoErrors(t, errs)
	found := false
	for _, stmt := range prog.Statements {
		if e, ok := stmt.(*ast.OnErrorGotoStatement); ok {
			found = true
			if e.Target != "errHandler" {
				t.Errorf("expected errHandler, got %q", e.Target)
			}
		}
	}
	if !found {
		t.Error("expected OnErrorGotoStatement")
	}
}

func TestParseOnErrorGoto0(t *testing.T) {
	// ON ERROR GOTO 0 — disable error handler
	prog, errs := parse("ON ERROR GOTO 0")
	expectNoErrors(t, errs)
	found := false
	for _, stmt := range prog.Statements {
		if e, ok := stmt.(*ast.OnErrorGotoStatement); ok {
			found = true
			if e.Target != "0" {
				t.Errorf("expected '0', got %q", e.Target)
			}
		}
	}
	if !found {
		t.Error("expected OnErrorGotoStatement")
	}
}

// ===========================================================================
// SELECT CASE variants
// ===========================================================================

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

func TestParseSelectCaseWithElse(t *testing.T) {
	prog, errs := parse(`SELECT CASE x%
CASE 1
  PRINT "one"
CASE ELSE
  PRINT "other"
END SELECT`)
	expectNoErrors(t, errs)
	s := getStmt[*ast.SelectCaseStatement](t, prog, 0)
	if s.ElseBlock == nil {
		t.Error("expected ElseBlock")
	}
}

// ===========================================================================
// IF variants
// ===========================================================================

func TestParseElseIfChain(t *testing.T) {
	prog, errs := parse(`IF x% = 1 THEN
PRINT "one"
ELSEIF x% = 2 THEN
PRINT "two"
ELSEIF x% = 3 THEN
PRINT "three"
ELSE
PRINT "other"
END IF`)
	expectNoErrors(t, errs)
	s := getStmt[*ast.IfStatement](t, prog, 0)
	if len(s.ElseIfClauses) != 2 {
		t.Errorf("expected 2 ELSEIF clauses, got %d", len(s.ElseIfClauses))
	}
	if s.ElseBlock == nil {
		t.Error("expected ElseBlock")
	}
}

func TestParseSingleLineIfElse(t *testing.T) {
	prog, errs := parse(`IF x% > 0 THEN PRINT "pos" ELSE PRINT "non-pos"`)
	expectNoErrors(t, errs)
	s := getStmt[*ast.IfStatement](t, prog, 0)
	if !s.IsSingleLine {
		t.Error("expected IsSingleLine=true")
	}
	if len(s.ElseBlock) == 0 {
		t.Error("expected ElseBlock")
	}
}

// ===========================================================================
// END statement variants
// ===========================================================================

func TestParseEndBareOnLine(t *testing.T) {
	prog, errs := parse("END\n")
	expectNoErrors(t, errs)
	getStmt[*ast.EndStatement](t, prog, 0)
}

func TestParseEndWithColon(t *testing.T) {
	// END followed by colon (inline statement terminator).
	prog, errs := parse("END : PRINT 1")
	_ = errs
	_ = prog
}
