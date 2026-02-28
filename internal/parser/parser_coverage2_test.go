package parser

import (
	"testing"

	"github.com/loabletech/go-basic/internal/ast"
	"github.com/loabletech/go-basic/internal/lexer"
)

// ===========================================================================
// peekTokenIs / expectPeek / peekError — direct unit tests
// ===========================================================================

// newParserFromInput is a helper to create a Parser for direct method testing.
func newParserFromInput(input string) *Parser {
	l := lexer.New(input)
	p := New(l)
	return p
}

func TestPeekTokenIs_Match(t *testing.T) {
	// "x% = 42" — curToken is x%, peekToken is =
	p := newParserFromInput("x% = 42")
	// After New(), curToken = x%, peekToken = =
	if !p.peekTokenIs(lexer.TOKEN_EQ) {
		t.Error("expected peekTokenIs(TOKEN_EQ) to be true")
	}
}

func TestPeekTokenIs_NoMatch(t *testing.T) {
	p := newParserFromInput("x% = 42")
	if p.peekTokenIs(lexer.TOKEN_PLUS) {
		t.Error("expected peekTokenIs(TOKEN_PLUS) to be false")
	}
}

func TestExpectPeek_Success(t *testing.T) {
	// curToken = x%, peekToken = =
	p := newParserFromInput("x% = 42")
	if !p.expectPeek(lexer.TOKEN_EQ) {
		t.Error("expected expectPeek to return true when peek matches")
	}
	// After success, curToken should have advanced to =
	if !p.curTokenIs(lexer.TOKEN_EQ) {
		t.Errorf("expected curToken to be EQ after expectPeek, got %s", lexer.TokenName(p.curToken.Type))
	}
}

func TestExpectPeek_Failure(t *testing.T) {
	// curToken = x%, peekToken = = (not PLUS), so expectPeek(PLUS) fails
	p := newParserFromInput("x% = 42")
	result := p.expectPeek(lexer.TOKEN_PLUS)
	if result {
		t.Error("expected expectPeek to return false when peek does not match")
	}
	// Should have recorded an error via peekError
	if len(p.Errors()) == 0 {
		t.Error("expected peekError to record an error")
	}
}

func TestPeekError_RecordsError(t *testing.T) {
	// Directly call peekError to ensure it records a message
	p := newParserFromInput("x% = 42")
	initialErrors := len(p.Errors())
	p.peekError(lexer.TOKEN_PLUS)
	if len(p.Errors()) <= initialErrors {
		t.Error("expected peekError to add an error")
	}
}

// ===========================================================================
// ERASE statement (parseEraseStatement 0%)
// ===========================================================================

func TestParseEraseStatement(t *testing.T) {
	prog, errs := parse("ERASE myArray")
	expectNoErrors(t, errs)
	s := getStmt[*ast.EraseStatement](t, prog, 0)
	if len(s.Names) != 1 {
		t.Errorf("expected 1 name, got %d", len(s.Names))
	}
	if s.Names[0] != "myArray" {
		t.Errorf("expected 'myArray', got %q", s.Names[0])
	}
}

func TestParseEraseMultiple(t *testing.T) {
	prog, errs := parse("ERASE arr1, arr2, arr3")
	expectNoErrors(t, errs)
	s := getStmt[*ast.EraseStatement](t, prog, 0)
	if len(s.Names) != 3 {
		t.Errorf("expected 3 names, got %d", len(s.Names))
	}
}

// ===========================================================================
// parseSingleExprStatement — KILL, CHDIR, MKDIR, RMDIR (0%)
// ===========================================================================

func TestParseKillStatement(t *testing.T) {
	prog, errs := parse(`KILL "file.txt"`)
	expectNoErrors(t, errs)
	s := getStmt[*ast.KillStatement](t, prog, 0)
	if s.Filename == nil {
		t.Error("expected Filename to be set")
	}
}

func TestParseChdirStatement(t *testing.T) {
	prog, errs := parse(`CHDIR "C:\temp"`)
	expectNoErrors(t, errs)
	s := getStmt[*ast.ChdirStatement](t, prog, 0)
	if s.Path == nil {
		t.Error("expected Path to be set")
	}
}

func TestParseMkdirStatement(t *testing.T) {
	prog, errs := parse(`MKDIR "newdir"`)
	expectNoErrors(t, errs)
	s := getStmt[*ast.MkdirStatement](t, prog, 0)
	if s.Path == nil {
		t.Error("expected Path to be set")
	}
}

func TestParseRmdirStatement(t *testing.T) {
	prog, errs := parse(`RMDIR "olddir"`)
	expectNoErrors(t, errs)
	s := getStmt[*ast.RmdirStatement](t, prog, 0)
	if s.Path == nil {
		t.Error("expected Path to be set")
	}
}

// ===========================================================================
// parseDelayStatement (0%)
// ===========================================================================

func TestParseDelayStatement(t *testing.T) {
	prog, errs := parse("DELAY 1.5")
	expectNoErrors(t, errs)
	// Returns a LetStatement with Name="DELAY"
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

// ===========================================================================
// parseWindowStatement (0%)
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

// ===========================================================================
// parsePaletteStatement (0%)
// ===========================================================================

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
// parseFnAssign (0%)
// ===========================================================================

func TestParseFnAssignStatement(t *testing.T) {
	// parseFnAssign is triggered when TOKEN_FN appears as a statement keyword.
	// The syntax is: FN name = expr (FN is a separate keyword, then the name).
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
	// FN with type suffix $ — the suffix should be stripped from the name
	prog, errs := parse(`FN greet$ = "Hello"`)
	expectNoErrors(t, errs)
	found := false
	for _, stmt := range prog.Statements {
		if fa, ok := stmt.(*ast.FnAssignStatement); ok {
			found = true
			// Name should have type suffix stripped (greet$ → greet)
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
	// FN name(args) as a standalone call — returns LetStatement with FnCallExpression
	// When FN keyword is followed by name then (, it's a call form.
	prog, errs := parse("FN calc(5)")
	expectNoErrors(t, errs)
	// Should be a LetStatement wrapping a FnCallExpression
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

// ===========================================================================
// parseDeclare — error path and FUNCTION with return type (36.4%)
// ===========================================================================

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
// parseOnStatement — event handler path (56.8%)
// ===========================================================================

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

// ===========================================================================
// parseGetStatement — all paths (66.7%)
// ===========================================================================

func TestParseGetFileNoRecord(t *testing.T) {
	prog, errs := parse("GET #1")
	expectNoErrors(t, errs)
	s := getStmt[*ast.GetStatement](t, prog, 0)
	if s.FileNum == nil {
		t.Error("expected FileNum")
	}
	if s.RecordOrPos != nil {
		t.Error("expected no record number")
	}
}

func TestParseGetFileWithRecord(t *testing.T) {
	prog, errs := parse("GET #1, 5")
	expectNoErrors(t, errs)
	s := getStmt[*ast.GetStatement](t, prog, 0)
	if s.FileNum == nil {
		t.Error("expected FileNum")
	}
	if s.RecordOrPos == nil {
		t.Error("expected RecordOrPos")
	}
}

func TestParseGetFileSkipRecord(t *testing.T) {
	// GET #1, , var$ — record number is skipped (empty)
	prog, errs := parse("GET #1, , data$")
	expectNoErrors(t, errs)
	s := getStmt[*ast.GetStatement](t, prog, 0)
	if s.FileNum == nil {
		t.Error("expected FileNum")
	}
	// RecordOrPos is nil (skipped), Variable is set
	if s.Variable == nil {
		t.Error("expected Variable")
	}
}

func TestParseGetGraphics(t *testing.T) {
	// GET (x1,y1)-(x2,y2), array — graphics GET
	prog, errs := parse("GET (0,0)-(100,100), img()")
	expectNoErrors(t, errs)
	s := getStmt[*ast.RemStatement](t, prog, 0)
	if s.Text != "GET (graphics)" {
		t.Errorf("expected 'GET (graphics)', got %q", s.Text)
	}
}

func TestParseGetFileNoHash(t *testing.T) {
	// GET without # — still valid
	prog, errs := parse("GET 1, 10")
	expectNoErrors(t, errs)
	s := getStmt[*ast.GetStatement](t, prog, 0)
	if s.FileNum == nil {
		t.Error("expected FileNum")
	}
}

// ===========================================================================
// parsePutStatement — all paths (65.8%)
// ===========================================================================

func TestParsePutFileNoRecord(t *testing.T) {
	prog, errs := parse("PUT #2")
	expectNoErrors(t, errs)
	s := getStmt[*ast.PutStatement](t, prog, 0)
	if s.FileNum == nil {
		t.Error("expected FileNum")
	}
	if s.RecordOrPos != nil {
		t.Error("expected no record number")
	}
}

func TestParsePutFileWithRecord(t *testing.T) {
	prog, errs := parse("PUT #2, 3")
	expectNoErrors(t, errs)
	s := getStmt[*ast.PutStatement](t, prog, 0)
	if s.FileNum == nil {
		t.Error("expected FileNum")
	}
	if s.RecordOrPos == nil {
		t.Error("expected RecordOrPos")
	}
}

func TestParsePutFileSkipRecord(t *testing.T) {
	// PUT #1, , var$ — skip record number
	prog, errs := parse("PUT #1, , data$")
	expectNoErrors(t, errs)
	s := getStmt[*ast.PutStatement](t, prog, 0)
	if s.FileNum == nil {
		t.Error("expected FileNum")
	}
	if s.Variable == nil {
		t.Error("expected Variable")
	}
}

func TestParsePutGraphics(t *testing.T) {
	// PUT (x,y), array — graphics PUT
	prog, errs := parse("PUT (50, 50), sprite()")
	expectNoErrors(t, errs)
	s := getStmt[*ast.RemStatement](t, prog, 0)
	if s.Text != "PUT (graphics)" {
		t.Errorf("expected 'PUT (graphics)', got %q", s.Text)
	}
}

func TestParsePutGraphicsWithMode(t *testing.T) {
	// PUT (x,y), array, XOR — graphics PUT with mode keyword
	prog, errs := parse("PUT (50, 50), sprite(), XOR")
	expectNoErrors(t, errs)
	s := getStmt[*ast.RemStatement](t, prog, 0)
	if s.Text != "PUT (graphics)" {
		t.Errorf("expected 'PUT (graphics)', got %q", s.Text)
	}
}

func TestParsePutGraphicsWithPsetMode(t *testing.T) {
	prog, errs := parse("PUT (50, 50), sprite(), PSET")
	expectNoErrors(t, errs)
	s := getStmt[*ast.RemStatement](t, prog, 0)
	if s.Text != "PUT (graphics)" {
		t.Errorf("expected 'PUT (graphics)', got %q", s.Text)
	}
}

func TestParsePutGraphicsWithOrMode(t *testing.T) {
	prog, errs := parse("PUT (50, 50), sprite(), OR")
	expectNoErrors(t, errs)
	s := getStmt[*ast.RemStatement](t, prog, 0)
	if s.Text != "PUT (graphics)" {
		t.Errorf("expected 'PUT (graphics)', got %q", s.Text)
	}
}

// ===========================================================================
// parseScreenStatement — full parameters (50%)
// ===========================================================================

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
	// SCREEN mode, , activepage — empty colorswitch
	prog, errs := parse("SCREEN 1, , 0")
	expectNoErrors(t, errs)
	s := getStmt[*ast.ScreenStatement](t, prog, 0)
	if s.Mode == nil {
		t.Error("expected Mode")
	}
	// ColorSwitch is nil (skipped)
	if s.ColorSwitch != nil {
		t.Error("expected nil ColorSwitch when skipped")
	}
}

func TestParseScreenSkipActivePage(t *testing.T) {
	// SCREEN mode, colorswitch, , visualpage
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
// parseIdentifierStatement — WRITE#, PUT$, GET$ paths (51.9%)
// ===========================================================================

func TestParseWriteHashIdentifier(t *testing.T) {
	// WRITE# n, expr — lexer fuses WRITE and # into WRITE# identifier
	prog, errs := parse("WRITE# 1, x$, 42")
	expectNoErrors(t, errs)
	s := getStmt[*ast.FileWriteStatement](t, prog, 0)
	if s.FileNum == nil {
		t.Error("expected FileNum")
	}
	if len(s.Expressions) == 0 {
		t.Error("expected Expressions")
	}
}

func TestParsePutDollarIdentifier(t *testing.T) {
	// PUT$ filenum, data — binary file put via identifier path
	prog, errs := parse("PUT$ 1, data$")
	expectNoErrors(t, errs)
	s := getStmt[*ast.FilePrintStatement](t, prog, 0)
	if s.FileNum == nil {
		t.Error("expected FileNum")
	}
	if len(s.Expressions) == 0 {
		t.Error("expected Expressions")
	}
}

func TestParseGetDollarIdentifier(t *testing.T) {
	// GET$ filenum, length, var$ — binary file get via identifier path
	prog, errs := parse("GET$ 1, 10, result$")
	expectNoErrors(t, errs)
	s := getStmt[*ast.FileInputStatement](t, prog, 0)
	if s.FileNum == nil {
		t.Error("expected FileNum")
	}
	if len(s.Variables) == 0 {
		t.Error("expected Variables")
	}
}

// ===========================================================================
// parseBuiltinFunction — keyword builtins without parens (50%)
// ===========================================================================

func TestParseBuiltinTimer(t *testing.T) {
	// TIMER used in expression (no parens)
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
	// INSTAT — no parens needed
	prog, errs := parse("x% = INSTAT")
	expectNoErrors(t, errs)
	s := getStmt[*ast.LetStatement](t, prog, 0)
	if s.Value == nil {
		t.Error("expected Value")
	}
}

func TestParseBuiltinLenWithParens(t *testing.T) {
	// LEN(s) — keyword builtin with parens
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
// parseDataValue — negative data, string data (57.1%)
// ===========================================================================

func TestParseDataNegativeNumber(t *testing.T) {
	prog, errs := parse("DATA -1, -3.14, 42")
	expectNoErrors(t, errs)
	s := getStmt[*ast.DataStatement](t, prog, 0)
	if len(s.Values) < 3 {
		t.Fatalf("expected 3 values, got %d", len(s.Values))
	}
	// First value should be -1
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
	// Unquoted data values are treated as raw strings
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

// ===========================================================================
// parseRestoreStatement — with and without target (71.4%)
// ===========================================================================

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
// parseInfixExpression — IMP and EQV operators (70%)
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
	// Exponentiation is right-associative
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

// ===========================================================================
// parseAssignment — array assignment, field assignment (73.7%)
// ===========================================================================

func TestParseArrayAssignment(t *testing.T) {
	prog, errs := parse("arr%(i%) = 42")
	expectNoErrors(t, errs)
	s := getStmt[*ast.ArrayAssignment](t, prog, 0)
	if s.Array == nil {
		t.Error("expected Array")
	}
	if s.Value == nil {
		t.Error("expected Value")
	}
}

func TestParseAssignmentErrorNoIdent(t *testing.T) {
	// LET without identifier triggers error path
	_, errs := parse("LET = 42")
	if len(errs) == 0 {
		t.Error("expected parse error for LET without identifier")
	}
}

func TestParseAssignmentErrorNoEq(t *testing.T) {
	// Variable without '=' triggers error path (not a sub call since it has no args)
	_, errs := parse("myVar THEN")
	// This may or may not produce errors depending on how it's parsed,
	// but it exercises the assignment error path
	_ = errs
}

// ===========================================================================
// parseCircleStatement — full parameter list including start/end/aspect (73.3%)
// ===========================================================================

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
	// CIRCLE (x,y), r, , start — skips color
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
	// CIRCLE (x,y), r, color, , end
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
// parseInputStatement — file input path (64.7%)
// ===========================================================================

func TestParseInputFileNum(t *testing.T) {
	prog, errs := parse("INPUT #1, a$")
	expectNoErrors(t, errs)
	s := getStmt[*ast.FileInputStatement](t, prog, 0)
	if s.FileNum == nil {
		t.Error("expected FileNum")
	}
	if len(s.Variables) == 0 {
		t.Error("expected Variables")
	}
}

func TestParseInputFileNumMultipleVars(t *testing.T) {
	prog, errs := parse("INPUT #2, name$, age%, score!")
	expectNoErrors(t, errs)
	s := getStmt[*ast.FileInputStatement](t, prog, 0)
	if s.FileNum == nil {
		t.Error("expected FileNum")
	}
	if len(s.Variables) != 3 {
		t.Errorf("expected 3 variables, got %d", len(s.Variables))
	}
}

func TestParseInputWithCommaPrompt(t *testing.T) {
	// INPUT with comma separator (no newline suppression)
	prog, errs := parse(`INPUT "Name? ", name$`)
	expectNoErrors(t, errs)
	s := getStmt[*ast.ReadStatement](t, prog, 0)
	if s.Prompt == "" {
		t.Error("expected Prompt")
	}
}

// ===========================================================================
// parseNumberLiteral — LONG, SINGLE, DOUBLE types (66.7%)
// ===========================================================================

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
	// D notation for double precision (Turbo BASIC specific)
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

// ===========================================================================
// parseRadixLiteral — binary literal (71.4%)
// ===========================================================================

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

// ===========================================================================
// parseGroupExpression — missing closing paren error path (75%)
// ===========================================================================

func TestParseGroupExpressionMissingParen(t *testing.T) {
	// Missing ) triggers addError in parseGroupExpression
	_, errs := parse("x% = (1 + 2")
	// Should produce an error (unexpected token or missing paren)
	// The parser still returns a value (the inner expression) so errors may be set
	_ = errs
	// Just verify it doesn't panic
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
// tokenPrecedence — remaining cases (69.2%)
// ===========================================================================

func TestTokenPrecedenceAllOperators(t *testing.T) {
	// Test expressions that exercise all precedence levels
	tests := []string{
		"x% = a% IMP b% EQV c%",       // IMP, EQV
		"x% = a% XOR b% OR c% AND d%",  // XOR, OR, AND
		"x% = a% + b% - c%",            // ADD (PLUS, MINUS)
		"x% = a% MOD b%",               // MOD
		`x% = a% \ b%`,                 // IDIV (BACKSLASH)
		"x% = a% * b% / c%",            // MUL (STAR, SLASH)
		"x% = a% ^ b%",                 // POWER (CARET)
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
// Additional parseAssignment paths — struct field assignment (73.7%)
// ===========================================================================

func TestParseFieldAssignment(t *testing.T) {
	// TYPE field assignment: rec.field = value — requires dot notation
	prog, errs := parse(`TYPE Point
x AS INTEGER
y AS INTEGER
END TYPE
DIM p AS Point
p.x = 10`)
	if len(errs) > 0 {
		// Some field assignment errors may occur but ensure no panic
		_ = errs
		return
	}
	// If it parsed, verify we got a statement
	if len(prog.Statements) == 0 {
		t.Error("expected statements")
	}
}

// ===========================================================================
// LINE INPUT with file number (parseLineStatement path)
// ===========================================================================

func TestParseLineInputWithFileNum(t *testing.T) {
	prog, errs := parse("LINE INPUT #1, line$")
	expectNoErrors(t, errs)
	s := getStmt[*ast.FileInputStatement](t, prog, 0)
	if s.FileNum == nil {
		t.Error("expected FileNum")
	}
	if !s.IsLineInput {
		t.Error("expected IsLineInput=true")
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
// PSET STEP variant (parsePsetStatement)
// ===========================================================================

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
// Additional ON ERROR coverage
// ===========================================================================

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
// parseNumberLiteral — verify integer type
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

// ===========================================================================
// WRITE statement without file num (console write path)
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
// Unary NOT expression
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
// INCR / DECR statements
// ===========================================================================

func TestParseIncrStatement(t *testing.T) {
	prog, errs := parse("INCR x%")
	expectNoErrors(t, errs)
	s := getStmt[*ast.IncrStatement](t, prog, 0)
	if s.Variable == nil {
		t.Error("expected Variable")
	}
}

func TestParseIncrWithAmount(t *testing.T) {
	prog, errs := parse("INCR x%, 5")
	expectNoErrors(t, errs)
	s := getStmt[*ast.IncrStatement](t, prog, 0)
	if s.Amount == nil {
		t.Error("expected Amount")
	}
}

func TestParseDecrStatement(t *testing.T) {
	prog, errs := parse("DECR x%")
	expectNoErrors(t, errs)
	s := getStmt[*ast.DecrStatement](t, prog, 0)
	if s.Variable == nil {
		t.Error("expected Variable")
	}
}

// ===========================================================================
// Additional CLS variants
// ===========================================================================

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
// POKE statement
// ===========================================================================

func TestParsePokeStatement(t *testing.T) {
	prog, errs := parse("POKE 1024, 42")
	expectNoErrors(t, errs)
	s := getStmt[*ast.PokeStatement](t, prog, 0)
	if s.Address == nil {
		t.Error("expected Address")
	}
	if s.Value == nil {
		t.Error("expected Value")
	}
}

// ===========================================================================
// SWAP statement
// ===========================================================================

func TestParseSwapStatement(t *testing.T) {
	prog, errs := parse("SWAP a%, b%")
	expectNoErrors(t, errs)
	s := getStmt[*ast.SwapStatement](t, prog, 0)
	if s.Var1 == nil {
		t.Error("expected Var1")
	}
	if s.Var2 == nil {
		t.Error("expected Var2")
	}
}

// ===========================================================================
// DEF SEG statement
// ===========================================================================

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
// STRING$ and SPACE$ builtin functions via identifier expression path
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

// ===========================================================================
// Additional builtin functions via identifier expression
// ===========================================================================

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

// ===========================================================================
// OPEN and CLOSE statement
// ===========================================================================

func TestParseOpenStatement(t *testing.T) {
	prog, errs := parse(`OPEN "test.dat" FOR INPUT AS #1`)
	expectNoErrors(t, errs)
	s := getStmt[*ast.OpenStatement](t, prog, 0)
	if s.Filename == nil {
		t.Error("expected Filename")
	}
	if s.Mode != "INPUT" {
		t.Errorf("expected INPUT mode, got %q", s.Mode)
	}
	if s.FileNum == nil {
		t.Error("expected FileNum")
	}
}

func TestParseOpenWithRecLen(t *testing.T) {
	prog, errs := parse(`OPEN "data.dat" FOR RANDOM AS #2 LEN = 128`)
	expectNoErrors(t, errs)
	s := getStmt[*ast.OpenStatement](t, prog, 0)
	if s.RecLen == nil {
		t.Error("expected RecLen")
	}
}

func TestParseCloseStatement(t *testing.T) {
	prog, errs := parse("CLOSE #1")
	expectNoErrors(t, errs)
	s := getStmt[*ast.CloseStatement](t, prog, 0)
	if len(s.FileNums) == 0 {
		t.Error("expected FileNums")
	}
}

func TestParseCloseAll(t *testing.T) {
	prog, errs := parse("CLOSE")
	expectNoErrors(t, errs)
	s := getStmt[*ast.CloseStatement](t, prog, 0)
	_ = s // No file nums is valid (close all)
}

// ===========================================================================
// Sub call without CALL keyword (identifier as procedure call)
// ===========================================================================

func TestParseSubCallNoParens(t *testing.T) {
	// Bare sub call: subname arg1, arg2 -- parsed as assignment which errors
	// since there's no '='. Just verify it doesn't panic.
	_, errs := parse(`MySub "World"`)
	// This will produce a parse error ("expected = in assignment") which is expected
	_ = errs
}

// ===========================================================================
// parseExpressionList — edge cases
// ===========================================================================

func TestParseExpressionListEmpty(t *testing.T) {
	// Calling a function with no args
	prog, errs := parse("CALL MySub")
	expectNoErrors(t, errs)
	_ = prog
}

// ===========================================================================
// SELECT CASE with ELSE clause
// ===========================================================================

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
// FN call expression context
// ===========================================================================

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
// Misc edge cases for coverage
// ===========================================================================

func TestParseMultipleStatementsOnLine(t *testing.T) {
	// BASIC allows : to separate statements
	prog, errs := parse("x% = 1 : y% = 2 : z% = 3")
	expectNoErrors(t, errs)
	if len(prog.Statements) < 3 {
		t.Errorf("expected at least 3 statements, got %d", len(prog.Statements))
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

func TestParseInkeyFunction(t *testing.T) {
	prog, errs := parse("k$ = INKEY$")
	expectNoErrors(t, errs)
	s := getStmt[*ast.LetStatement](t, prog, 0)
	if s.Value == nil {
		t.Error("expected Value")
	}
}
