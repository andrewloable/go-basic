package codegen

import (
	"strings"
	"testing"

	"github.com/loabletech/go-basic/internal/ast"
)

// ---------------------------------------------------------------------------
// Tests for emit_control.go — IF, FOR, WHILE, DO, SELECT CASE,
// GOTO, GOSUB, EXIT, ON...GOTO/GOSUB
// ---------------------------------------------------------------------------

func TestIfThenElse(t *testing.T) {
	stmts := []ast.Statement{
		&ast.IfStatement{
			Condition: &ast.BinaryExpr{
				Left:     &ast.Identifier{Name: "x"},
				Operator: ">",
				Right:    &ast.NumberLiteral{Value: 10, NumType: ast.NumInt},
			},
			ThenBlock: []ast.Statement{
				&ast.PrintStatement{
					Expressions: []ast.Expression{
						&ast.StringLiteral{Value: "big"},
					},
					Separators: []string{""},
				},
			},
			ElseBlock: []ast.Statement{
				&ast.PrintStatement{
					Expressions: []ast.Expression{
						&ast.StringLiteral{Value: "small"},
					},
					Separators: []string{""},
				},
			},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "if") {
		t.Errorf("expected 'if' statement, got:\n%s", out)
	}
	if !strings.Contains(out, "} else {") {
		t.Errorf("expected '} else {', got:\n%s", out)
	}
	if !strings.Contains(out, `"big"`) {
		t.Errorf("expected then-branch string, got:\n%s", out)
	}
	if !strings.Contains(out, `"small"`) {
		t.Errorf("expected else-branch string, got:\n%s", out)
	}
}

func TestIfElseIf(t *testing.T) {
	stmts := []ast.Statement{
		&ast.IfStatement{
			Condition: &ast.BinaryExpr{
				Left:     &ast.Identifier{Name: "x"},
				Operator: "=",
				Right:    &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
			},
			ThenBlock: []ast.Statement{
				&ast.PrintStatement{
					Expressions: []ast.Expression{&ast.StringLiteral{Value: "one"}},
					Separators:  []string{""},
				},
			},
			ElseIfClauses: []ast.ElseIfClause{
				{
					Condition: &ast.BinaryExpr{
						Left:     &ast.Identifier{Name: "x"},
						Operator: "=",
						Right:    &ast.NumberLiteral{Value: 2, NumType: ast.NumInt},
					},
					Body: []ast.Statement{
						&ast.PrintStatement{
							Expressions: []ast.Expression{&ast.StringLiteral{Value: "two"}},
							Separators:  []string{""},
						},
					},
				},
			},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "} else if") {
		t.Errorf("expected '} else if', got:\n%s", out)
	}
}

func TestForLoop(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ForStatement{
			Counter: &ast.Identifier{Name: "i"},
			Start:   &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
			End:     &ast.NumberLiteral{Value: 10, NumType: ast.NumInt},
			Body: []ast.Statement{
				&ast.PrintStatement{
					Expressions: []ast.Expression{
						&ast.Identifier{Name: "i"},
					},
					Separators: []string{""},
				},
			},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "for i =") {
		t.Errorf("expected for loop with counter i, got:\n%s", out)
	}
}

func TestForLoopWithStep(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ForStatement{
			Counter: &ast.Identifier{Name: "i"},
			Start:   &ast.NumberLiteral{Value: 0, NumType: ast.NumInt},
			End:     &ast.NumberLiteral{Value: 20, NumType: ast.NumInt},
			Step:    &ast.NumberLiteral{Value: 2, NumType: ast.NumInt},
			Body:    []ast.Statement{},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "step_") {
		t.Errorf("expected step variable, got:\n%s", out)
	}
}

func TestWhileLoop(t *testing.T) {
	stmts := []ast.Statement{
		&ast.WhileStatement{
			Condition: &ast.BinaryExpr{
				Left:     &ast.Identifier{Name: "x"},
				Operator: "<",
				Right:    &ast.NumberLiteral{Value: 100, NumType: ast.NumInt},
			},
			Body: []ast.Statement{
				&ast.IncrStatement{
					Variable: &ast.Identifier{Name: "x"},
				},
			},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "for (x < ") {
		t.Errorf("expected while loop as for loop, got:\n%s", out)
	}
}

func TestDoLoopTopWhile(t *testing.T) {
	stmts := []ast.Statement{
		&ast.DoLoopStatement{
			Condition: &ast.BinaryExpr{
				Left:     &ast.Identifier{Name: "x"},
				Operator: ">",
				Right:    &ast.NumberLiteral{Value: 0, NumType: ast.NumInt},
			},
			TestAtTop: true,
			IsUntil:   false,
			Body:      []ast.Statement{},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "for (x >") {
		t.Errorf("expected DO WHILE as for loop, got:\n%s", out)
	}
}

func TestDoLoopBottomUntil(t *testing.T) {
	stmts := []ast.Statement{
		&ast.DoLoopStatement{
			Condition: &ast.BinaryExpr{
				Left:     &ast.Identifier{Name: "done"},
				Operator: "=",
				Right:    &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
			},
			TestAtTop: false,
			IsUntil:   true,
			Body: []ast.Statement{
				&ast.PrintStatement{
					Expressions: []ast.Expression{&ast.StringLiteral{Value: "loop"}},
					Separators:  []string{""},
				},
			},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "for {") {
		t.Errorf("expected DO ... LOOP UNTIL as for { ... break }, got:\n%s", out)
	}
	if !strings.Contains(out, "break") {
		t.Errorf("expected break for UNTIL, got:\n%s", out)
	}
}

func TestEmitDoLoopNoCondition(t *testing.T) {
	stmts := []ast.Statement{
		&ast.DoLoopStatement{
			Condition: nil,
			Body: []ast.Statement{
				&ast.ExitStatement{ExitType: "DO"},
			},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "for {") {
		t.Errorf("expected infinite loop 'for {' in output, got:\n%s", out)
	}
}

func TestEmitDoLoopWhileAtTop(t *testing.T) {
	stmts := []ast.Statement{
		&ast.DoLoopStatement{
			Condition: &ast.NumberLiteral{Value: 1, OriginalText: "1"},
			TestAtTop: true,
			IsUntil:   false,
			Body:      []ast.Statement{},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "for") {
		t.Errorf("expected 'for' loop in output, got:\n%s", out)
	}
}

func TestEmitDoLoopUntilAtTop(t *testing.T) {
	stmts := []ast.Statement{
		&ast.DoLoopStatement{
			Condition: &ast.NumberLiteral{Value: 0, OriginalText: "0"},
			TestAtTop: true,
			IsUntil:   true,
			Body:      []ast.Statement{},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "!(") {
		t.Errorf("expected negated condition '!(' in output, got:\n%s", out)
	}
}

func TestEmitDoLoopWhileAtBottom(t *testing.T) {
	stmts := []ast.Statement{
		&ast.DoLoopStatement{
			Condition: &ast.NumberLiteral{Value: 1, OriginalText: "1"},
			TestAtTop: false,
			IsUntil:   false,
			Body:      []ast.Statement{},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "break") {
		t.Errorf("expected break in output for WHILE at bottom, got:\n%s", out)
	}
	if !strings.Contains(out, "!(") {
		t.Errorf("expected negated condition '!(' in output, got:\n%s", out)
	}
}

func TestEmitDoLoopUntilAtBottom(t *testing.T) {
	stmts := []ast.Statement{
		&ast.DoLoopStatement{
			Condition: &ast.NumberLiteral{Value: 1, OriginalText: "1"},
			TestAtTop: false,
			IsUntil:   true,
			Body:      []ast.Statement{},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "break") {
		t.Errorf("expected break in output for UNTIL at bottom, got:\n%s", out)
	}
}

func TestSelectCase(t *testing.T) {
	stmts := []ast.Statement{
		&ast.SelectCaseStatement{
			TestExpr: &ast.Identifier{Name: "x"},
			Cases: []ast.CaseClause{
				{
					Values: []ast.CaseValue{
						{Value: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt}},
					},
					Body: []ast.Statement{
						&ast.PrintStatement{
							Expressions: []ast.Expression{&ast.StringLiteral{Value: "one"}},
							Separators:  []string{""},
						},
					},
				},
				{
					Values: []ast.CaseValue{
						{Value: &ast.NumberLiteral{Value: 2, NumType: ast.NumInt}},
					},
					Body: []ast.Statement{
						&ast.PrintStatement{
							Expressions: []ast.Expression{&ast.StringLiteral{Value: "two"}},
							Separators:  []string{""},
						},
					},
				},
			},
			ElseBlock: []ast.Statement{
				&ast.PrintStatement{
					Expressions: []ast.Expression{&ast.StringLiteral{Value: "other"}},
					Separators:  []string{""},
				},
			},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "sel_") {
		t.Errorf("expected select temp variable, got:\n%s", out)
	}
	if !strings.Contains(out, "default:") {
		t.Errorf("expected default: block for CASE ELSE, got:\n%s", out)
	}
}

func TestEmitSelectCaseIsComparison(t *testing.T) {
	stmts := []ast.Statement{
		&ast.SelectCaseStatement{
			TestExpr: &ast.NumberLiteral{Value: 5, OriginalText: "5"},
			Cases: []ast.CaseClause{
				{
					Values: []ast.CaseValue{
						{
							Value:        &ast.NumberLiteral{Value: 3, OriginalText: "3"},
							IsComparison: true,
							Comparison:   ">",
						},
					},
					Body: []ast.Statement{
						&ast.PrintStatement{
							Expressions: []ast.Expression{&ast.StringLiteral{Value: "gt3"}},
							Separators:  []string{""},
						},
					},
				},
			},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "switch {") {
		t.Errorf("SELECT CASE IS: expected 'switch {' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "> 3") {
		t.Errorf("SELECT CASE IS: expected '> 3' in output, got:\n%s", out)
	}
}

func TestEmitSelectCaseRange(t *testing.T) {
	stmts := []ast.Statement{
		&ast.SelectCaseStatement{
			TestExpr: &ast.NumberLiteral{Value: 5, OriginalText: "5"},
			Cases: []ast.CaseClause{
				{
					Values: []ast.CaseValue{
						{
							Value:    &ast.NumberLiteral{Value: 1, OriginalText: "1"},
							EndValue: &ast.NumberLiteral{Value: 10, OriginalText: "10"},
							IsRange:  true,
						},
					},
					Body: []ast.Statement{},
				},
			},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, ">=") {
		t.Errorf("SELECT CASE range: expected '>=' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "<=") {
		t.Errorf("SELECT CASE range: expected '<=' in output, got:\n%s", out)
	}
}

func TestEmitSelectCaseElseBlock(t *testing.T) {
	stmts := []ast.Statement{
		&ast.SelectCaseStatement{
			TestExpr: &ast.NumberLiteral{Value: 99, OriginalText: "99"},
			Cases:    []ast.CaseClause{},
			ElseBlock: []ast.Statement{
				&ast.PrintStatement{
					Expressions: []ast.Expression{&ast.StringLiteral{Value: "else"}},
					Separators:  []string{""},
				},
			},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "default:") {
		t.Errorf("SELECT CASE else: expected 'default:' in output, got:\n%s", out)
	}
}

func TestGoto(t *testing.T) {
	stmts := []ast.Statement{
		&ast.LineNumberStatement{Number: 100},
		&ast.GotoStatement{Target: "100"},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "line_100:") {
		t.Errorf("expected label line_100:, got:\n%s", out)
	}
	if !strings.Contains(out, "goto line_100") {
		t.Errorf("expected goto line_100, got:\n%s", out)
	}
}

func TestLabelStatement(t *testing.T) {
	// A label must be referenced by a GOTO to be emitted (unreferenced labels are pruned).
	stmts := []ast.Statement{
		&ast.GotoStatement{BasePos: ast.Position{Line: 1}, Target: "myLabel"},
		&ast.LabelStatement{Name: "myLabel"},
		&ast.PrintStatement{
			Expressions: []ast.Expression{&ast.StringLiteral{Value: "at label"}},
			Separators:  []string{""},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "label_myLabel:") {
		t.Errorf("expected label_myLabel:, got:\n%s", out)
	}
}

func TestEmitGosub(t *testing.T) {
	stmts := []ast.Statement{
		&ast.GosubStatement{Target: "mySub"},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "goto") {
		t.Errorf("expected 'goto' in emitted GOSUB, got:\n%s", out)
	}
}

func TestEmitReturn(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ReturnStatement{},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "return") {
		t.Errorf("expected 'return' in emitted RETURN, got:\n%s", out)
	}
}

func TestEmitOnComputedGoto(t *testing.T) {
	stmts := []ast.Statement{
		&ast.OnComputedGotoStatement{
			Expr:    &ast.NumberLiteral{Value: 1, OriginalText: "1"},
			Targets: []string{"label1", "label2"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "_on_idx") {
		t.Errorf("expected _on_idx in output, got:\n%s", out)
	}
	if !strings.Contains(out, "goto") {
		t.Errorf("expected goto in output, got:\n%s", out)
	}
}

func TestEmitOnComputedGosub(t *testing.T) {
	stmts := []ast.Statement{
		&ast.OnComputedGosubStatement{
			Expr:    &ast.NumberLiteral{Value: 2, OriginalText: "2"},
			Targets: []string{"sub1", "sub2", "sub3"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "_on_idx") {
		t.Errorf("expected _on_idx in output, got:\n%s", out)
	}
}

func TestExitFor(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ExitStatement{ExitType: "FOR"},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "break") {
		t.Errorf("expected break for EXIT FOR, got:\n%s", out)
	}
}

func TestEmitExitFor(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ExitStatement{ExitType: "FOR"},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "break") {
		t.Errorf("EXIT FOR: expected 'break' in output, got:\n%s", out)
	}
}

func TestEmitExitDo(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ExitStatement{ExitType: "DO"},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "break") {
		t.Errorf("EXIT DO: expected 'break' in output, got:\n%s", out)
	}
}

func TestEmitExitWhile(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ExitStatement{ExitType: "WHILE"},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "break") {
		t.Errorf("EXIT WHILE: expected 'break' in output, got:\n%s", out)
	}
}

func TestEmitExitLoop(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ExitStatement{ExitType: "LOOP"},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "break") {
		t.Errorf("EXIT LOOP: expected 'break' in output, got:\n%s", out)
	}
}

func TestEmitExitSub(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ExitStatement{ExitType: "SUB"},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "return") {
		t.Errorf("EXIT SUB: expected 'return' in output, got:\n%s", out)
	}
}

func TestEmitExitFunction(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ExitStatement{ExitType: "FUNCTION"},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "return") {
		t.Errorf("EXIT FUNCTION: expected 'return' in output, got:\n%s", out)
	}
}

func TestEmitExitDef(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ExitStatement{ExitType: "DEF"},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "return") {
		t.Errorf("EXIT DEF (not in defFn): expected 'return' in output, got:\n%s", out)
	}
}

func TestEmitExitUnknown(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ExitStatement{ExitType: "UNKNOWN_TYPE"},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "EXIT") {
		t.Errorf("EXIT UNKNOWN: expected 'EXIT' in output comment, got:\n%s", out)
	}
}

// TestCollectMainVariablesWithGoto exercises collectMainVariables via GOTO hoisting.
// It includes ForStatement, DimStatement, ReadStatement (IsInput), WhileStatement,
// DoLoopStatement, SelectCaseStatement, IfStatement all in one program.
func TestCollectMainVariablesWithGoto(t *testing.T) {
	stmts := []ast.Statement{
		&ast.GotoStatement{Target: "99"},
		&ast.LineNumberStatement{Number: 10},
		&ast.ForStatement{
			Counter: &ast.Identifier{Name: "i"},
			Start:   &ast.NumberLiteral{Value: 1},
			End:     &ast.NumberLiteral{Value: 5},
			Body: []ast.Statement{
				&ast.LetStatement{
					Name:  &ast.Identifier{Name: "inner"},
					Value: &ast.NumberLiteral{Value: 0},
				},
			},
		},
		&ast.DimStatement{
			Declarations: []ast.DimDecl{
				{Name: "scalar"},
			},
		},
		&ast.DimStatement{
			Declarations: []ast.DimDecl{
				{
					Name: "arr1d",
					Dimensions: []ast.DimRange{
						{Upper: &ast.NumberLiteral{Value: 10}},
					},
				},
			},
		},
		&ast.DimStatement{
			Declarations: []ast.DimDecl{
				{
					Name: "arr2d",
					Dimensions: []ast.DimRange{
						{Upper: &ast.NumberLiteral{Value: 5}},
						{Upper: &ast.NumberLiteral{Value: 5}},
					},
				},
			},
		},
		&ast.ReadStatement{
			IsInput: true,
			Variables: []ast.Expression{
				&ast.Identifier{Name: "inputVar"},
			},
		},
		&ast.WhileStatement{
			Condition: &ast.NumberLiteral{Value: 0},
			Body: []ast.Statement{
				&ast.LetStatement{
					Name:  &ast.Identifier{Name: "whileInner"},
					Value: &ast.NumberLiteral{Value: 0},
				},
			},
		},
		&ast.DoLoopStatement{
			Condition: &ast.NumberLiteral{Value: 0},
			Body: []ast.Statement{
				&ast.LetStatement{
					Name:  &ast.Identifier{Name: "doInner"},
					Value: &ast.NumberLiteral{Value: 0},
				},
			},
		},
		&ast.SelectCaseStatement{
			TestExpr: &ast.NumberLiteral{Value: 1},
			Cases: []ast.CaseClause{
				{
					Values: []ast.CaseValue{{Value: &ast.NumberLiteral{Value: 1}}},
					Body: []ast.Statement{
						&ast.LetStatement{
							Name:  &ast.Identifier{Name: "caseVar"},
							Value: &ast.NumberLiteral{Value: 0},
						},
					},
				},
			},
		},
		&ast.IfStatement{
			Condition: &ast.NumberLiteral{Value: 1},
			ThenBlock: []ast.Statement{
				&ast.LetStatement{
					Name:  &ast.Identifier{Name: "thenVar"},
					Value: &ast.NumberLiteral{Value: 0},
				},
			},
			ElseIfClauses: []ast.ElseIfClause{
				{
					Condition: &ast.NumberLiteral{Value: 0},
					Body: []ast.Statement{
						&ast.LetStatement{
							Name:  &ast.Identifier{Name: "elseIfVar"},
							Value: &ast.NumberLiteral{Value: 0},
						},
					},
				},
			},
			ElseBlock: []ast.Statement{
				&ast.LetStatement{
					Name:  &ast.Identifier{Name: "elseVar"},
					Value: &ast.NumberLiteral{Value: 0},
				},
			},
		},
		&ast.LineNumberStatement{Number: 99},
		&ast.PrintStatement{
			Expressions: []ast.Expression{&ast.StringLiteral{Value: "done"}},
			Separators:  []string{""},
		},
	}

	out := generate(t, stmts)
	if !strings.Contains(out, "// Hoisted variable declarations") {
		t.Errorf("Expected hoisted variable declarations comment in output:\n%s", out)
	}
}

func TestCollectMainVariablesWithDefFn(t *testing.T) {
	stmts := []ast.Statement{
		&ast.GotoStatement{Target: "end"},
		&ast.DefFnDeclaration{
			Name: "FNtest",
			Params: []ast.Parameter{
				{Name: "x"},
			},
			Body: []ast.Statement{
				&ast.LetStatement{
					Name:  &ast.Identifier{Name: "FNtest"},
					Value: &ast.Identifier{Name: "x"},
				},
			},
		},
		&ast.LabelStatement{Name: "end"},
	}

	out := generate(t, stmts)
	if !strings.Contains(out, "FNtest") || !strings.Contains(out, "func") {
		t.Errorf("Expected FNtest function in output, got:\n%s", out)
	}
}
