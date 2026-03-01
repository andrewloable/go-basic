package codegen

import (
	"strings"
	"testing"

	"github.com/loabletech/go-basic/internal/ast"
)

// ---------------------------------------------------------------------------
// Tests for emit_statements.go — PRINT, LET, SWAP, INPUT, READ, DATA,
// RESTORE, POKE, SHELL, BEEP, CLS, END, STOP, SYSTEM, etc.
// ---------------------------------------------------------------------------

func TestPrintNoArgs(t *testing.T) {
	stmts := []ast.Statement{
		&ast.PrintStatement{},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "fmt.Println()") {
		t.Errorf("expected fmt.Println(), got:\n%s", out)
	}
}

func TestPrintWithString(t *testing.T) {
	stmts := []ast.Statement{
		&ast.PrintStatement{
			Expressions: []ast.Expression{
				&ast.StringLiteral{Value: "Hello, World!"},
			},
			Separators: []string{""},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, `fmt.Println("Hello, World!")`) {
		t.Errorf("expected fmt.Println with string, got:\n%s", out)
	}
}

func TestPrintTrailingSemicolon(t *testing.T) {
	stmts := []ast.Statement{
		&ast.PrintStatement{
			Expressions: []ast.Expression{
				&ast.StringLiteral{Value: "no newline"},
			},
			Separators:     []string{";"},
			HasTrailingSep: true,
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "fmt.Print(") {
		t.Errorf("expected fmt.Print (no newline), got:\n%s", out)
	}
}

func TestLetAssignment(t *testing.T) {
	stmts := []ast.Statement{
		&ast.LetStatement{
			Name:  &ast.Identifier{Name: "x"},
			Value: &ast.NumberLiteral{Value: 42, NumType: ast.NumInt},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "var x") {
		t.Errorf("expected variable declaration for x, got:\n%s", out)
	}
	if !strings.Contains(out, "42") {
		t.Errorf("expected value 42, got:\n%s", out)
	}
}

func TestLetStringVariable(t *testing.T) {
	stmts := []ast.Statement{
		&ast.LetStatement{
			Name:  &ast.Identifier{Name: "name", TypeSuffix: "$"},
			Value: &ast.StringLiteral{Value: "Alice"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "name_str") {
		t.Errorf("expected mangled name name_str, got:\n%s", out)
	}
	if !strings.Contains(out, "string") {
		t.Errorf("expected string type, got:\n%s", out)
	}
}

func TestSwapStatement(t *testing.T) {
	stmts := []ast.Statement{
		&ast.SwapStatement{
			Var1: &ast.Identifier{Name: "a"},
			Var2: &ast.Identifier{Name: "b"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "a, b = b, a") {
		t.Errorf("expected Go swap idiom, got:\n%s", out)
	}
}

func TestIncrDecr(t *testing.T) {
	stmts := []ast.Statement{
		&ast.IncrStatement{
			Variable: &ast.Identifier{Name: "x"},
		},
		&ast.DecrStatement{
			Variable: &ast.Identifier{Name: "y"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "x++") {
		t.Errorf("expected x++, got:\n%s", out)
	}
	if !strings.Contains(out, "y--") {
		t.Errorf("expected y--, got:\n%s", out)
	}
}

func TestEmitIncrWithAmount(t *testing.T) {
	stmts := []ast.Statement{
		&ast.IncrStatement{
			Variable: &ast.Identifier{Name: "counter"},
			Amount:   &ast.NumberLiteral{Value: 5, OriginalText: "5"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "+= 5") {
		t.Errorf("INCR with amount: expected '+= 5' in output, got:\n%s", out)
	}
}

func TestEmitDecrWithAmount(t *testing.T) {
	stmts := []ast.Statement{
		&ast.DecrStatement{
			Variable: &ast.Identifier{Name: "counter"},
			Amount:   &ast.NumberLiteral{Value: 3, OriginalText: "3"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "-= 3") {
		t.Errorf("DECR with amount: expected '-= 3' in output, got:\n%s", out)
	}
}

func TestEndStatement(t *testing.T) {
	stmts := []ast.Statement{
		&ast.EndStatement{},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "os.Exit(0)") {
		t.Errorf("expected os.Exit(0), got:\n%s", out)
	}
}

func TestRemComment(t *testing.T) {
	stmts := []ast.Statement{
		&ast.RemStatement{Text: "this is a comment"},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "// this is a comment") {
		t.Errorf("expected Go comment, got:\n%s", out)
	}
}

func TestDataRead(t *testing.T) {
	stmts := []ast.Statement{
		&ast.DataStatement{
			Values: []ast.Expression{
				&ast.NumberLiteral{Value: 10, NumType: ast.NumInt},
				&ast.NumberLiteral{Value: 20, NumType: ast.NumInt},
			},
		},
		&ast.ReadStatement{
			Variables: []ast.Expression{
				&ast.Identifier{Name: "a"},
			},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "dataPool") {
		t.Errorf("expected dataPool, got:\n%s", out)
	}
	if !strings.Contains(out, "dataIdx") {
		t.Errorf("expected dataIdx, got:\n%s", out)
	}
}

func TestBeepStatement(t *testing.T) {
	stmts := []ast.Statement{
		&ast.BeepStatement{},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, `"\a"`) {
		t.Errorf("expected bell character for BEEP, got:\n%s", out)
	}
}

func TestEmitRestore(t *testing.T) {
	stmts := []ast.Statement{
		&ast.RestoreStatement{},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "dataIdx") {
		t.Errorf("expected 'dataIdx' in emitted RESTORE, got:\n%s", out)
	}
}

func TestEmitErase(t *testing.T) {
	stmts := []ast.Statement{
		&ast.EraseStatement{
			Names: []string{"arr1", "arr2"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "arr1") {
		t.Errorf("expected 'arr1' in erase output, got:\n%s", out)
	}
}

func TestEmitRandomizeNoSeed(t *testing.T) {
	stmts := []ast.Statement{
		&ast.RandomizeStatement{},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "Randomize") {
		t.Errorf("expected 'Randomize' in output, got:\n%s", out)
	}
}

func TestEmitRandomizeWithSeed(t *testing.T) {
	stmts := []ast.Statement{
		&ast.RandomizeStatement{
			Seed: &ast.NumberLiteral{Value: 42, OriginalText: "42"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "Randomize") {
		t.Errorf("expected 'Randomize' in output, got:\n%s", out)
	}
}

func TestEmitInputFromStdin(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ReadStatement{
			IsInput: true,
			Prompt:  "Enter value: ",
			Variables: []ast.Expression{
				&ast.Identifier{Name: "x"},
			},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "Scan") || !strings.Contains(out, "Enter value") {
		t.Errorf("expected Scan and prompt in output, got:\n%s", out)
	}
}

func TestEmitInputFromStdinNoPrompt(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ReadStatement{
			IsInput:   true,
			Prompt:    "", // no prompt
			Variables: []ast.Expression{&ast.Identifier{Name: "x"}},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "scanner_") {
		t.Errorf("INPUT no prompt: expected scanner_ in output, got:\n%s", out)
	}
	if strings.Contains(out, `fmt.Print("`) {
		t.Errorf("INPUT no prompt: should NOT have fmt.Print prompt, got:\n%s", out)
	}
}

func TestEmitInputFromStdinStringVar(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ReadStatement{
			IsInput:   true,
			Prompt:    "Enter name: ",
			Variables: []ast.Expression{&ast.Identifier{Name: "name", TypeSuffix: "$"}},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "scanner_.Text()") {
		t.Errorf("INPUT string var: expected scanner_.Text() in output, got:\n%s", out)
	}
}

func TestEmitInputFromStdinMultipleVars(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ReadStatement{
			IsInput: true,
			Prompt:  "Enter a, b: ",
			Variables: []ast.Expression{
				&ast.Identifier{Name: "a"},
				&ast.Identifier{Name: "b"},
			},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "InputSplitLine") {
		t.Errorf("INPUT multiple vars: expected InputSplitLine in output, got:\n%s", out)
	}
	if !strings.Contains(out, "parts_") {
		t.Errorf("INPUT multiple vars: expected parts_ in output, got:\n%s", out)
	}
}

func TestEmitInputFromStdinLineInput(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ReadStatement{
			IsInput:     true,
			IsLineInput: true,
			Prompt:      "",
			Variables:   []ast.Expression{&ast.Identifier{Name: "line", TypeSuffix: "$"}},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "scanner_") {
		t.Errorf("LINE INPUT: expected scanner_ in output, got:\n%s", out)
	}
	if !strings.Contains(out, "scanner_.Text()") {
		t.Errorf("LINE INPUT: expected scanner_.Text() in output, got:\n%s", out)
	}
}

func TestEmitPrintUsing(t *testing.T) {
	stmts := []ast.Statement{
		&ast.PrintStatement{
			Format: &ast.StringLiteral{Value: "##.##"},
			Expressions: []ast.Expression{
				&ast.NumberLiteral{Value: 3, OriginalText: "3"},
			},
			Separators: []string{""},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "PrintUsing") {
		t.Errorf("expected 'PrintUsing' in output, got:\n%s", out)
	}
}

func TestEmitPrintUsingTrailingSep(t *testing.T) {
	stmts := []ast.Statement{
		&ast.PrintStatement{
			Format: &ast.StringLiteral{Value: "##.##"},
			Expressions: []ast.Expression{
				&ast.NumberLiteral{Value: 3, OriginalText: "3"},
			},
			Separators:     []string{""},
			HasTrailingSep: true,
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "PrintUsing") {
		t.Errorf("PRINT USING trailing sep: expected 'PrintUsing' in output, got:\n%s", out)
	}
}

func TestEmitPrintComma(t *testing.T) {
	// Comma separator triggers zone printing
	stmts := []ast.Statement{
		&ast.PrintStatement{
			Expressions: []ast.Expression{
				&ast.NumberLiteral{Value: 1, OriginalText: "1"},
				&ast.NumberLiteral{Value: 2, OriginalText: "2"},
			},
			Separators: []string{",", ""},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "PrintZone") {
		t.Errorf("PRINT with comma: expected 'PrintZone' in output, got:\n%s", out)
	}
}

func TestEmitPrintCommaTrailingSep(t *testing.T) {
	// Comma separator with trailing sep (no newline at end)
	stmts := []ast.Statement{
		&ast.PrintStatement{
			Expressions: []ast.Expression{
				&ast.NumberLiteral{Value: 42, OriginalText: "42"},
			},
			Separators:     []string{","},
			HasTrailingSep: true,
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "PrintZone") {
		t.Errorf("PRINT comma trailing sep: expected 'PrintZone' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "fmt.Print(out_)") {
		t.Errorf("PRINT comma trailing sep: expected 'fmt.Print(out_)' in output, got:\n%s", out)
	}
}

func TestEmitPrintMultipleNoSep(t *testing.T) {
	stmts := []ast.Statement{
		&ast.PrintStatement{
			Expressions: []ast.Expression{
				&ast.NumberLiteral{Value: 1, OriginalText: "1"},
				&ast.NumberLiteral{Value: 2, OriginalText: "2"},
			},
			Separators:     []string{";", ""},
			HasTrailingSep: false,
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "fmt.Println(") {
		t.Errorf("PRINT multiple no trailing sep: expected 'fmt.Println(' in output, got:\n%s", out)
	}
}

func TestEmitPrintMultipleTrailingSep(t *testing.T) {
	stmts := []ast.Statement{
		&ast.PrintStatement{
			Expressions: []ast.Expression{
				&ast.NumberLiteral{Value: 1, OriginalText: "1"},
				&ast.NumberLiteral{Value: 2, OriginalText: "2"},
			},
			Separators:     []string{";", ";"},
			HasTrailingSep: true,
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "fmt.Print(") {
		t.Errorf("PRINT multiple trailing sep: expected 'fmt.Print(' in output, got:\n%s", out)
	}
}

func TestEmitStatementEnd(t *testing.T) {
	stmts := []ast.Statement{&ast.EndStatement{}}
	out := generate(t, stmts)
	if !strings.Contains(out, "os.Exit(0)") {
		t.Errorf("END: expected os.Exit(0) in output, got:\n%s", out)
	}
}

func TestEmitStatementStop(t *testing.T) {
	stmts := []ast.Statement{&ast.StopStatement{}}
	out := generate(t, stmts)
	if !strings.Contains(out, "os.Exit(0)") {
		t.Errorf("STOP: expected os.Exit(0) in output, got:\n%s", out)
	}
}

func TestEmitStatementSystem(t *testing.T) {
	stmts := []ast.Statement{&ast.SystemStatement{}}
	out := generate(t, stmts)
	if !strings.Contains(out, "os.Exit(0)") {
		t.Errorf("SYSTEM: expected os.Exit(0) in output, got:\n%s", out)
	}
}

func TestEmitStatementBeep(t *testing.T) {
	stmts := []ast.Statement{&ast.BeepStatement{}}
	out := generate(t, stmts)
	if !strings.Contains(out, "BEEP") {
		t.Errorf("BEEP: expected BEEP in output, got:\n%s", out)
	}
}

func TestEmitStatementCls(t *testing.T) {
	stmts := []ast.Statement{&ast.ClsStatement{}}
	out := generate(t, stmts)
	if !strings.Contains(out, "rt.Cls(0)") {
		t.Errorf("CLS: expected rt.Cls(0) in output, got:\n%s", out)
	}
}

func TestEmitStatementData(t *testing.T) {
	stmts := []ast.Statement{
		&ast.DataStatement{
			Values: []ast.Expression{
				&ast.NumberLiteral{Value: 1, OriginalText: "1"},
			},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "DATA") {
		t.Errorf("DATA: expected 'DATA' comment in output, got:\n%s", out)
	}
}

func TestEmitStatementOptionBase(t *testing.T) {
	stmts := []ast.Statement{
		&ast.OptionBaseStatement{Value: 1},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "OPTION BASE") {
		t.Errorf("OPTION BASE: expected 'OPTION BASE' comment in output, got:\n%s", out)
	}
}

func TestEmitStatementDefType(t *testing.T) {
	stmts := []ast.Statement{
		&ast.DefTypeStatement{Type: "DEFINT"},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "DEFINT") {
		t.Errorf("DEFTYPE: expected 'DEFINT' in output, got:\n%s", out)
	}
}

func TestEmitStatementScope(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ScopeStatement{
			Modifier:  "SHARED",
			Variables: []string{"x", "y"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "SHARED") {
		t.Errorf("SCOPE: expected 'SHARED' in output, got:\n%s", out)
	}
}

func TestEmitStatementPoke(t *testing.T) {
	stmts := []ast.Statement{
		&ast.PokeStatement{
			Address: &ast.NumberLiteral{Value: 1000, OriginalText: "1000"},
			Value:   &ast.NumberLiteral{Value: 42, OriginalText: "42"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "POKE") {
		t.Errorf("POKE: expected 'POKE' comment in output, got:\n%s", out)
	}
}

func TestEmitStatementConst(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ConstStatement{
			Name:  "PI",
			Value: &ast.NumberLiteral{Value: 3, OriginalText: "3"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "CONST") {
		t.Errorf("CONST: expected 'CONST' in output, got:\n%s", out)
	}
}

func TestEmitStatementClear(t *testing.T) {
	stmts := []ast.Statement{&ast.ClearStatement{}}
	out := generate(t, stmts)
	if !strings.Contains(out, "CLEAR") {
		t.Errorf("CLEAR: expected 'CLEAR' in output, got:\n%s", out)
	}
}

func TestEmitStatementOnErrorGoto(t *testing.T) {
	stmts := []ast.Statement{
		&ast.OnErrorGotoStatement{Target: "errHandler"},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "errState.SetHandler") {
		t.Errorf("ON ERROR GOTO: expected 'errState.SetHandler' in output, got:\n%s", out)
	}
}

func TestEmitStatementOnErrorGotoDisable(t *testing.T) {
	stmts := []ast.Statement{
		&ast.OnErrorGotoStatement{Target: "0"},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "ON ERROR GOTO 0") {
		t.Errorf("ON ERROR GOTO 0: expected 'ON ERROR GOTO 0' comment in output, got:\n%s", out)
	}
}

func TestEmitStatementPlay(t *testing.T) {
	stmts := []ast.Statement{
		&ast.PlayStatement{
			CommandString: &ast.StringLiteral{Value: "T2L4C"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "rt.Play(") {
		t.Errorf("PLAY: expected 'rt.Play(' in output, got:\n%s", out)
	}
}

func TestEmitStatementSound(t *testing.T) {
	stmts := []ast.Statement{
		&ast.SoundStatement{
			Frequency: &ast.NumberLiteral{Value: 440, OriginalText: "440"},
			Duration:  &ast.NumberLiteral{Value: 18, OriginalText: "18"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "rt.Sound(") {
		t.Errorf("SOUND: expected 'rt.Sound(' in output, got:\n%s", out)
	}
}

func TestEmitStatementError(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ErrorStatement{
			Code: &ast.NumberLiteral{Value: 5, OriginalText: "5"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "TriggerError") {
		t.Errorf("ERROR: expected 'TriggerError' in output, got:\n%s", out)
	}
}

func TestEmitStatementResume(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ResumeStatement{Type: "NEXT"},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "RESUME NEXT") {
		t.Errorf("RESUME NEXT: expected 'RESUME NEXT' in output, got:\n%s", out)
	}
}

func TestEmitStatementResumeEmpty(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ResumeStatement{Type: ""},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "RESUME") {
		t.Errorf("RESUME: expected 'RESUME' in output, got:\n%s", out)
	}
}

func TestEmitStatementResumeLabel(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ResumeStatement{Type: "recoverPoint"},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "goto") {
		t.Errorf("RESUME label: expected 'goto' in output, got:\n%s", out)
	}
}

func TestEmitStatementFieldAssign(t *testing.T) {
	stmts := []ast.Statement{
		&ast.FieldAssignStatement{
			Object: &ast.Identifier{Name: "obj"},
			Field:  "x",
			Value:  &ast.NumberLiteral{Value: 42, OriginalText: "42"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "obj.x") {
		t.Errorf("FIELD ASSIGN: expected 'obj.x' in output, got:\n%s", out)
	}
}

func TestEmitStatementFnAssign(t *testing.T) {
	stmts := []ast.Statement{
		&ast.FnAssignStatement{
			Name:  "FNtest",
			Value: &ast.NumberLiteral{Value: 99, OriginalText: "99"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "_fn_") {
		t.Errorf("FN ASSIGN: expected '_fn_' in output, got:\n%s", out)
	}
}
