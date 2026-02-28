package codegen

import (
	"strings"
	"testing"

	"github.com/loabletech/go-basic/internal/ast"
)

// ---------------------------------------------------------------------------
// Tests for emit_fileio.go — OPEN, CLOSE, PRINT#, INPUT#, GET, PUT,
// LSET, RSET, FIELD, etc.
// ---------------------------------------------------------------------------

func TestFilePrintStatement(t *testing.T) {
	stmts := []ast.Statement{
		&ast.FilePrintStatement{
			FileNum:     &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
			Expressions: []ast.Expression{&ast.StringLiteral{Value: "hello"}},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "fm.FilePrint(") {
		t.Errorf("expected fm.FilePrint call, got:\n%s", out)
	}
	if !strings.Contains(out, "fm := rt.NewFileManager()") {
		t.Errorf("expected FileManager declaration, got:\n%s", out)
	}
}

func TestFileInputStatement(t *testing.T) {
	stmts := []ast.Statement{
		&ast.FileInputStatement{
			FileNum:   &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
			Variables: []ast.Expression{&ast.Identifier{Name: "a", TypeSuffix: "$"}},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "fm.FileInput(") {
		t.Errorf("expected fm.FileInput call, got:\n%s", out)
	}
}

func TestFileWriteStatement(t *testing.T) {
	stmts := []ast.Statement{
		&ast.FileWriteStatement{
			FileNum:     &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
			Expressions: []ast.Expression{&ast.StringLiteral{Value: "test"}, &ast.NumberLiteral{Value: 100, NumType: ast.NumInt}},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "fm.FileWrite(") {
		t.Errorf("expected fm.FileWrite call, got:\n%s", out)
	}
}

func TestFieldStatement(t *testing.T) {
	stmts := []ast.Statement{
		&ast.FieldStatement{
			FileNum: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
			Fields: []ast.FieldDef{
				{Length: &ast.NumberLiteral{Value: 20, NumType: ast.NumInt}, VarName: "A$"},
				{Length: &ast.NumberLiteral{Value: 10, NumType: ast.NumInt}, VarName: "B$"},
			},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "fm.Field(") {
		t.Errorf("expected fm.Field call, got:\n%s", out)
	}
	if !strings.Contains(out, "rt.FieldDef") {
		t.Errorf("expected rt.FieldDef in output, got:\n%s", out)
	}
}

func TestLsetStatementEmission(t *testing.T) {
	stmts := []ast.Statement{
		&ast.LsetStatement{
			Variable: "A$",
			Value:    &ast.StringLiteral{Value: "hello"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "fm.Lset(") {
		t.Errorf("expected fm.Lset call, got:\n%s", out)
	}
}

func TestPutStatement(t *testing.T) {
	stmts := []ast.Statement{
		&ast.PutStatement{
			FileNum:     &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
			RecordOrPos: &ast.NumberLiteral{Value: 5, NumType: ast.NumInt},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "fm.RandomPut(") {
		t.Errorf("expected fm.RandomPut call, got:\n%s", out)
	}
}

func TestGetStatementEmission(t *testing.T) {
	stmts := []ast.Statement{
		&ast.GetStatement{
			FileNum:     &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
			RecordOrPos: &ast.NumberLiteral{Value: 3, NumType: ast.NumInt},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "fm.RandomGet(") {
		t.Errorf("expected fm.RandomGet call, got:\n%s", out)
	}
}

func TestSeekStatementEmission(t *testing.T) {
	stmts := []ast.Statement{
		&ast.SeekStatement{
			FileNum:  &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
			Position: &ast.NumberLiteral{Value: 10, NumType: ast.NumInt},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "fm.FileSeek(") {
		t.Errorf("expected fm.FileSeek call, got:\n%s", out)
	}
}

func TestEmitOpen(t *testing.T) {
	stmts := []ast.Statement{
		&ast.OpenStatement{
			Filename: &ast.StringLiteral{Value: "test.txt"},
			FileNum:  &ast.NumberLiteral{Value: 1, OriginalText: "1"},
			Mode:     "INPUT",
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "FileOpen") {
		t.Errorf("expected 'FileOpen' in output, got:\n%s", out)
	}
}

func TestEmitClose(t *testing.T) {
	stmts := []ast.Statement{
		&ast.CloseStatement{
			FileNums: []ast.Expression{
				&ast.NumberLiteral{Value: 1, OriginalText: "1"},
			},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "FileClose") {
		t.Errorf("expected 'FileClose' in output, got:\n%s", out)
	}
}

func TestEmitCloseAll(t *testing.T) {
	stmts := []ast.Statement{
		&ast.CloseStatement{
			FileNums: nil, // no args = close all
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "FileCloseAll") {
		t.Errorf("expected 'FileCloseAll' in output, got:\n%s", out)
	}
}

func TestEmitOpenOutputMode(t *testing.T) {
	stmts := []ast.Statement{
		&ast.OpenStatement{
			Filename: &ast.StringLiteral{Value: "out.txt"},
			FileNum:  &ast.NumberLiteral{Value: 1, OriginalText: "1"},
			Mode:     "OUTPUT",
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "FileModeOutput") {
		t.Errorf("expected 'FileModeOutput' in output, got:\n%s", out)
	}
}

func TestEmitOpenAppendMode(t *testing.T) {
	stmts := []ast.Statement{
		&ast.OpenStatement{
			Filename: &ast.StringLiteral{Value: "app.txt"},
			FileNum:  &ast.NumberLiteral{Value: 1, OriginalText: "1"},
			Mode:     "APPEND",
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "FileModeAppend") {
		t.Errorf("expected 'FileModeAppend' in output, got:\n%s", out)
	}
}

func TestEmitOpenRandomMode(t *testing.T) {
	recLen := ast.Expression(&ast.NumberLiteral{Value: 64, OriginalText: "64"})
	stmts := []ast.Statement{
		&ast.OpenStatement{
			Filename: &ast.StringLiteral{Value: "rnd.dat"},
			FileNum:  &ast.NumberLiteral{Value: 1, OriginalText: "1"},
			Mode:     "RANDOM",
			RecLen:   recLen,
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "FileModeRandom") {
		t.Errorf("expected 'FileModeRandom' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "64") {
		t.Errorf("expected RecLen '64' in output, got:\n%s", out)
	}
}

func TestEmitOpenBinaryMode(t *testing.T) {
	stmts := []ast.Statement{
		&ast.OpenStatement{
			Filename: &ast.StringLiteral{Value: "bin.dat"},
			FileNum:  &ast.NumberLiteral{Value: 2, OriginalText: "2"},
			Mode:     "BINARY",
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "FileModeBinary") {
		t.Errorf("expected 'FileModeBinary' in output, got:\n%s", out)
	}
}

func TestEmitRset(t *testing.T) {
	stmts := []ast.Statement{
		&ast.RsetStatement{
			Variable: "field1",
			Value:    &ast.StringLiteral{Value: "hello"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "Rset") {
		t.Errorf("expected 'Rset' in output, got:\n%s", out)
	}
}

func TestEmitFilePrintNoExpressions(t *testing.T) {
	stmts := []ast.Statement{
		&ast.FilePrintStatement{
			FileNum:     &ast.NumberLiteral{Value: 1, OriginalText: "1"},
			Expressions: []ast.Expression{},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "FilePrint(") {
		t.Errorf("FILE PRINT no exprs: expected 'FilePrint(' in output, got:\n%s", out)
	}
}

func TestEmitFileWriteNoExpressions(t *testing.T) {
	stmts := []ast.Statement{
		&ast.FileWriteStatement{
			FileNum:     &ast.NumberLiteral{Value: 1, OriginalText: "1"},
			Expressions: []ast.Expression{},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "FileWrite(") {
		t.Errorf("FILE WRITE no exprs: expected 'FileWrite(' in output, got:\n%s", out)
	}
}

func TestEmitFileInputLineInput(t *testing.T) {
	stmts := []ast.Statement{
		&ast.FileInputStatement{
			FileNum:     &ast.NumberLiteral{Value: 1, OriginalText: "1"},
			IsLineInput: true,
			Variables:   []ast.Expression{&ast.Identifier{Name: "line", TypeSuffix: "$"}},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "FileLineInput") {
		t.Errorf("FILE LINE INPUT: expected FileLineInput in output, got:\n%s", out)
	}
}

func TestEmitFileInputRegularString(t *testing.T) {
	stmts := []ast.Statement{
		&ast.FileInputStatement{
			FileNum:     &ast.NumberLiteral{Value: 1, OriginalText: "1"},
			IsLineInput: false,
			Variables:   []ast.Expression{&ast.Identifier{Name: "s", TypeSuffix: "$"}},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "FileInput") {
		t.Errorf("FILE INPUT string: expected FileInput in output, got:\n%s", out)
	}
}

func TestEmitFileInputRegularNumeric(t *testing.T) {
	stmts := []ast.Statement{
		&ast.FileInputStatement{
			FileNum:     &ast.NumberLiteral{Value: 2, OriginalText: "2"},
			IsLineInput: false,
			Variables:   []ast.Expression{&ast.Identifier{Name: "n"}},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "FileInput") {
		t.Errorf("FILE INPUT numeric: expected FileInput in output, got:\n%s", out)
	}
	if !strings.Contains(out, "ParseFloat") {
		t.Errorf("FILE INPUT numeric: expected ParseFloat in output, got:\n%s", out)
	}
}

func TestEmitFileInputLineInputNumeric(t *testing.T) {
	stmts := []ast.Statement{
		&ast.FileInputStatement{
			FileNum:     &ast.NumberLiteral{Value: 1, OriginalText: "1"},
			IsLineInput: true,
			Variables:   []ast.Expression{&ast.Identifier{Name: "n"}},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "FileLineInput") {
		t.Errorf("FILE LINE INPUT numeric: expected FileLineInput in output, got:\n%s", out)
	}
	if !strings.Contains(out, "ParseFloat") {
		t.Errorf("FILE LINE INPUT numeric: expected ParseFloat in output, got:\n%s", out)
	}
}

func TestEmitStatementGet(t *testing.T) {
	stmts := []ast.Statement{
		&ast.GetStatement{
			FileNum: &ast.NumberLiteral{Value: 1, OriginalText: "1"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "RandomGet") {
		t.Errorf("GET: expected 'RandomGet' in output, got:\n%s", out)
	}
}

func TestEmitStatementPut(t *testing.T) {
	stmts := []ast.Statement{
		&ast.PutStatement{
			FileNum: &ast.NumberLiteral{Value: 1, OriginalText: "1"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "RandomPut") {
		t.Errorf("PUT: expected 'RandomPut' in output, got:\n%s", out)
	}
}

func TestEmitStatementSeek(t *testing.T) {
	stmts := []ast.Statement{
		&ast.SeekStatement{
			FileNum:  &ast.NumberLiteral{Value: 1, OriginalText: "1"},
			Position: &ast.NumberLiteral{Value: 100, OriginalText: "100"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "FileSeek") {
		t.Errorf("SEEK: expected 'FileSeek' in output, got:\n%s", out)
	}
}

func TestEmitStatementField(t *testing.T) {
	stmts := []ast.Statement{
		&ast.FieldStatement{
			FileNum: &ast.NumberLiteral{Value: 1, OriginalText: "1"},
			Fields: []ast.FieldDef{
				{VarName: "name$", Length: &ast.NumberLiteral{Value: 20, OriginalText: "20"}},
			},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "fm.Field(") {
		t.Errorf("FIELD: expected 'fm.Field(' in output, got:\n%s", out)
	}
}

func TestEmitStatementLset(t *testing.T) {
	stmts := []ast.Statement{
		&ast.LsetStatement{
			Variable: "field1",
			Value:    &ast.StringLiteral{Value: "hello"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "Lset") {
		t.Errorf("LSET: expected 'Lset' in output, got:\n%s", out)
	}
}
