package parser

import (
	"testing"

	"github.com/loabletech/go-basic/internal/ast"
)

// ===========================================================================
// Tests for parse_fileio.go: OPEN, CLOSE, GET, PUT, FIELD, SEEK, LSET, RSET,
// NAME, WRITE#, FILE INPUT/PRINT
// ===========================================================================

// TestRegressionPrintFileOutput verifies PRINT #n, expr for file output.
func TestRegressionPrintFileOutput(t *testing.T) {
	_, errs := parse(`OPEN "out.txt" FOR OUTPUT AS #1
PRINT #1, "hello"
CLOSE #1`)
	expectNoErrors(t, errs)
}

// TestRegressionWriteHash verifies WRITE# n, expr... for file output.
func TestRegressionWriteHash(t *testing.T) {
	_, errs := parse(`WRITE# 1, "hello", 42`)
	expectNoErrors(t, errs)
}

// ===========================================================================
// OPEN and CLOSE
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
// GET statement: file and graphics forms
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
	s := getStmt[*ast.GraphicsGetStatement](t, prog, 0)
	if s.X1 == nil || s.Y1 == nil || s.X2 == nil || s.Y2 == nil {
		t.Error("expected all coordinates to be set")
	}
	if s.ArrayVar == nil {
		t.Error("expected ArrayVar to be set")
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
// PUT statement: file and graphics forms
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
	// PUT (x,y), array — graphics PUT (default XOR)
	prog, errs := parse("PUT (50, 50), sprite()")
	expectNoErrors(t, errs)
	s := getStmt[*ast.GraphicsPutStatement](t, prog, 0)
	if s.X == nil || s.Y == nil {
		t.Error("expected coordinates to be set")
	}
	if s.ArrayVar == nil {
		t.Error("expected ArrayVar to be set")
	}
	if s.Action != "XOR" {
		t.Errorf("expected default action 'XOR', got %q", s.Action)
	}
}

func TestParsePutGraphicsWithMode(t *testing.T) {
	// PUT (x,y), array, XOR — graphics PUT with mode keyword
	prog, errs := parse("PUT (50, 50), sprite(), XOR")
	expectNoErrors(t, errs)
	s := getStmt[*ast.GraphicsPutStatement](t, prog, 0)
	if s.Action != "XOR" {
		t.Errorf("expected action 'XOR', got %q", s.Action)
	}
}

func TestParsePutGraphicsWithPsetMode(t *testing.T) {
	prog, errs := parse("PUT (50, 50), sprite(), PSET")
	expectNoErrors(t, errs)
	s := getStmt[*ast.GraphicsPutStatement](t, prog, 0)
	if s.Action != "PSET" {
		t.Errorf("expected action 'PSET', got %q", s.Action)
	}
}

func TestParsePutGraphicsWithOrMode(t *testing.T) {
	prog, errs := parse("PUT (50, 50), sprite(), OR")
	expectNoErrors(t, errs)
	s := getStmt[*ast.GraphicsPutStatement](t, prog, 0)
	if s.Action != "OR" {
		t.Errorf("expected action 'OR', got %q", s.Action)
	}
}

// ===========================================================================
// FIELD statement
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
// SEEK statement
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
// LSET / RSET statements
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
// NAME statement
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
// WRITE# (file write) statement
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

// ===========================================================================
// PUT$ / GET$ binary file I/O via identifier path
// ===========================================================================

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
// FILE INPUT statement
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

// ===========================================================================
// PRINT to file
// ===========================================================================

func TestParsePrintToFile(t *testing.T) {
	prog, errs := parse(`PRINT #1, "hello"`)
	expectNoErrors(t, errs)
	getStmt[*ast.FilePrintStatement](t, prog, 0)
}

// ===========================================================================
// KILL, CHDIR, MKDIR, RMDIR (single-expression file-related statements)
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
