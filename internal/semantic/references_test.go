package semantic

import (
	"strings"
	"testing"

	"github.com/loabletech/go-basic/internal/ast"
)

// ===========================================================================
// ReferenceResolver tests
// ===========================================================================

func TestNewReferenceResolver(t *testing.T) {
	prog := &ast.Program{}
	rr := NewReferenceResolver(prog)
	if rr == nil {
		t.Fatal("NewReferenceResolver returned nil")
	}
	if rr.Program != prog {
		t.Error("Program not set correctly")
	}
	if rr.LineMap == nil {
		t.Error("LineMap should be initialized")
	}
	if rr.LabelMap == nil {
		t.Error("LabelMap should be initialized")
	}
}

func TestReferenceResolverGotoValidLabel(t *testing.T) {
	// GOTO myLabel → valid because myLabel: exists
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LabelStatement{BasePos: ast.Position{Line: 1, Column: 1}, Name: "myLabel"},
			&ast.GotoStatement{BasePos: ast.Position{Line: 2, Column: 1}, Target: "myLabel"},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got: %v", errs)
	}
}

func TestReferenceResolverGotoUndefinedLabel(t *testing.T) {
	// GOTO missing → error
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.GotoStatement{BasePos: ast.Position{Line: 1, Column: 1}, Target: "missing"},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) == 0 {
		t.Fatal("expected error for missing GOTO label")
	}
	if !strings.Contains(errs[0], "undefined label") {
		t.Fatalf("expected 'undefined label' error, got: %s", errs[0])
	}
}

func TestReferenceResolverGotoLineNumber(t *testing.T) {
	// GOTO 100 → valid because line 100 exists
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LineNumberStatement{BasePos: ast.Position{Line: 1, Column: 1}, Number: 100},
			&ast.GotoStatement{BasePos: ast.Position{Line: 2, Column: 1}, Target: "100"},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got: %v", errs)
	}
}

func TestReferenceResolverGotoMissingLineNumber(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.GotoStatement{BasePos: ast.Position{Line: 1, Column: 1}, Target: "999"},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) == 0 {
		t.Fatal("expected error for missing line number GOTO")
	}
	if !strings.Contains(errs[0], "undefined line number") {
		t.Fatalf("expected 'undefined line number' error, got: %s", errs[0])
	}
}

func TestReferenceResolverGosubValidTarget(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LabelStatement{BasePos: ast.Position{Line: 1, Column: 1}, Name: "handler"},
			&ast.GosubStatement{BasePos: ast.Position{Line: 2, Column: 1}, Target: "handler"},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got: %v", errs)
	}
}

func TestReferenceResolverGosubMissingTarget(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.GosubStatement{BasePos: ast.Position{Line: 1, Column: 1}, Target: "noSub"},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) == 0 {
		t.Fatal("expected error for undefined GOSUB target")
	}
}

func TestReferenceResolverOnErrorGotoValid(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LabelStatement{BasePos: ast.Position{Line: 1, Column: 1}, Name: "errHandler"},
			&ast.OnErrorGotoStatement{BasePos: ast.Position{Line: 2, Column: 1}, Target: "errHandler"},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got: %v", errs)
	}
}

func TestReferenceResolverOnErrorGotoZeroDisables(t *testing.T) {
	// ON ERROR GOTO 0 is a special case that disables error handling (no label check)
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.OnErrorGotoStatement{BasePos: ast.Position{Line: 1, Column: 1}, Target: "0"},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for ON ERROR GOTO 0, got: %v", errs)
	}
}

func TestReferenceResolverOnErrorGotoMissing(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.OnErrorGotoStatement{BasePos: ast.Position{Line: 1, Column: 1}, Target: "missing"},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) == 0 {
		t.Fatal("expected error for undefined ON ERROR GOTO target")
	}
}

func TestReferenceResolverOnEventGosub(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LabelStatement{BasePos: ast.Position{Line: 1, Column: 1}, Name: "keyHandler"},
			&ast.OnEventGosubStatement{
				BasePos:   ast.Position{Line: 2, Column: 1},
				EventType: "KEY",
				EventParam: &ast.NumberLiteral{BasePos: ast.Position{Line: 2, Column: 7}, Value: 1, OriginalText: "1"},
				Target:    "keyHandler",
			},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got: %v", errs)
	}
}

func TestReferenceResolverOnEventGosubMissing(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.OnEventGosubStatement{
				BasePos:   ast.Position{Line: 1, Column: 1},
				EventType: "TIMER",
				Target:    "noHandler",
			},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) == 0 {
		t.Fatal("expected error for undefined ON TIMER GOSUB target")
	}
}

func TestReferenceResolverRestoreWithTarget(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LabelStatement{BasePos: ast.Position{Line: 1, Column: 1}, Name: "dataBlock"},
			&ast.RestoreStatement{BasePos: ast.Position{Line: 2, Column: 1}, Target: "dataBlock"},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got: %v", errs)
	}
}

func TestReferenceResolverRestoreWithoutTarget(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.RestoreStatement{BasePos: ast.Position{Line: 1, Column: 1}, Target: ""},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for RESTORE with no target, got: %v", errs)
	}
}

func TestReferenceResolverResumeWithLabel(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LabelStatement{BasePos: ast.Position{Line: 1, Column: 1}, Name: "resumeHere"},
			&ast.ResumeStatement{BasePos: ast.Position{Line: 2, Column: 1}, Type: "resumeHere"},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors, got: %v", errs)
	}
}

func TestReferenceResolverResumeNext(t *testing.T) {
	// RESUME NEXT doesn't need a label check
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ResumeStatement{BasePos: ast.Position{Line: 1, Column: 1}, Type: "NEXT"},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for RESUME NEXT, got: %v", errs)
	}
}

func TestReferenceResolverDataPool(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.DataStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Values: []ast.Expression{
					&ast.NumberLiteral{BasePos: ast.Position{}, Value: 10, OriginalText: "10"},
					&ast.NumberLiteral{BasePos: ast.Position{}, Value: 20, OriginalText: "20"},
					&ast.StringLiteral{BasePos: ast.Position{}, Value: "hello"},
				},
			},
		},
	}
	rr := NewReferenceResolver(prog)
	rr.Resolve()

	pool := rr.GetDataPool()
	if len(pool) != 3 {
		t.Fatalf("expected 3 data pool entries, got %d", len(pool))
	}
}

func TestReferenceResolverGetLineMap(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LineNumberStatement{BasePos: ast.Position{Line: 1, Column: 1}, Number: 100},
			&ast.LineNumberStatement{BasePos: ast.Position{Line: 2, Column: 1}, Number: 200},
		},
	}
	rr := NewReferenceResolver(prog)
	rr.Resolve()

	lm := rr.GetLineMap()
	if len(lm) != 2 {
		t.Fatalf("expected 2 entries in line map, got %d", len(lm))
	}
	if _, ok := lm[100]; !ok {
		t.Error("expected line 100 in line map")
	}
	if _, ok := lm[200]; !ok {
		t.Error("expected line 200 in line map")
	}
}

func TestReferenceResolverGetLabelMap(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LabelStatement{BasePos: ast.Position{Line: 1, Column: 1}, Name: "alpha"},
			&ast.LabelStatement{BasePos: ast.Position{Line: 2, Column: 1}, Name: "beta"},
		},
	}
	rr := NewReferenceResolver(prog)
	rr.Resolve()

	lm := rr.GetLabelMap()
	if len(lm) != 2 {
		t.Fatalf("expected 2 entries in label map, got %d", len(lm))
	}
	if _, ok := lm["ALPHA"]; !ok {
		t.Error("expected ALPHA in label map (case-insensitive)")
	}
	if _, ok := lm["BETA"]; !ok {
		t.Error("expected BETA in label map")
	}
}

func TestReferenceResolverDuplicateLabel(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LabelStatement{BasePos: ast.Position{Line: 1, Column: 1}, Name: "dupe"},
			&ast.LabelStatement{BasePos: ast.Position{Line: 2, Column: 1}, Name: "dupe"},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) == 0 {
		t.Fatal("expected error for duplicate label")
	}
	if !strings.Contains(errs[0], "duplicate label") {
		t.Fatalf("expected 'duplicate label' error, got: %s", errs[0])
	}
}

func TestReferenceResolverDuplicateLineNumber(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LineNumberStatement{BasePos: ast.Position{Line: 1, Column: 1}, Number: 100},
			&ast.LineNumberStatement{BasePos: ast.Position{Line: 2, Column: 1}, Number: 100},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) == 0 {
		t.Fatal("expected error for duplicate line number")
	}
	if !strings.Contains(errs[0], "duplicate line number") {
		t.Fatalf("expected 'duplicate line number' error, got: %s", errs[0])
	}
}

func TestReferenceResolverNestedBlocksCollectTargets(t *testing.T) {
	// Labels inside IF blocks, FOR loops, etc. should be found
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ForStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Counter: &ast.Identifier{Name: "i"},
				Start:   &ast.NumberLiteral{Value: 1, OriginalText: "1"},
				End:     &ast.NumberLiteral{Value: 10, OriginalText: "10"},
				Body: []ast.Statement{
					&ast.LabelStatement{BasePos: ast.Position{Line: 2, Column: 1}, Name: "innerLabel"},
				},
			},
			&ast.GotoStatement{BasePos: ast.Position{Line: 3, Column: 1}, Target: "innerLabel"},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors (innerLabel found in nested FOR body), got: %v", errs)
	}
}

func TestReferenceResolverCaseInsensitiveLabel(t *testing.T) {
	// Label defined as "MyLabel", referenced as "MYLABEL"
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LabelStatement{BasePos: ast.Position{Line: 1, Column: 1}, Name: "MyLabel"},
			&ast.GotoStatement{BasePos: ast.Position{Line: 2, Column: 1}, Target: "MYLABEL"},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for case-insensitive label match, got: %v", errs)
	}
}

// ===========================================================================
// collectTargets – nested block recursion
// ===========================================================================

func TestCollectTargetsSubDeclarationBody(t *testing.T) {
	// Label inside a SUB body should be collected
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.SubDeclaration{
				BasePos: ast.Position{Line: 1, Column: 1},
				Name:    "MySub",
				Body: []ast.Statement{
					&ast.LabelStatement{
						BasePos: ast.Position{Line: 2, Column: 1},
						Name:    "innerSubLabel",
					},
				},
			},
			&ast.GotoStatement{
				BasePos: ast.Position{Line: 5, Column: 1},
				Target:  "innerSubLabel",
			},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors (label inside SUB body collected), got: %v", errs)
	}
}

func TestCollectTargetsFunctionDeclarationBody(t *testing.T) {
	// Label inside a FUNCTION body should be collected
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.FunctionDeclaration{
				BasePos: ast.Position{Line: 1, Column: 1},
				Name:    "MyFunc",
				Body: []ast.Statement{
					&ast.LabelStatement{
						BasePos: ast.Position{Line: 2, Column: 1},
						Name:    "innerFuncLabel",
					},
				},
			},
			&ast.GotoStatement{
				BasePos: ast.Position{Line: 5, Column: 1},
				Target:  "innerFuncLabel",
			},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors (label inside FUNCTION body collected), got: %v", errs)
	}
}

func TestCollectTargetsIfStatementAllBlocks(t *testing.T) {
	// Labels in ThenBlock, ElseIf body, and ElseBlock should all be collected
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.IfStatement{
				BasePos:   ast.Position{Line: 1, Column: 1},
				Condition: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				ThenBlock: []ast.Statement{
					&ast.LabelStatement{BasePos: ast.Position{Line: 2, Column: 1}, Name: "thenLabel"},
				},
				ElseIfClauses: []ast.ElseIfClause{
					{
						BasePos:   ast.Position{Line: 3, Column: 1},
						Condition: &ast.NumberLiteral{Value: 0, NumType: ast.NumInt},
						Body: []ast.Statement{
							&ast.LabelStatement{BasePos: ast.Position{Line: 4, Column: 1}, Name: "elseIfLabel"},
						},
					},
				},
				ElseBlock: []ast.Statement{
					&ast.LabelStatement{BasePos: ast.Position{Line: 5, Column: 1}, Name: "elseLabel"},
				},
			},
			&ast.GotoStatement{BasePos: ast.Position{Line: 6, Column: 1}, Target: "thenLabel"},
			&ast.GotoStatement{BasePos: ast.Position{Line: 7, Column: 1}, Target: "elseIfLabel"},
			&ast.GotoStatement{BasePos: ast.Position{Line: 8, Column: 1}, Target: "elseLabel"},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors (labels in all IF blocks collected), got: %v", errs)
	}
}

func TestCollectTargetsWhileStatementBody(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.WhileStatement{
				BasePos:   ast.Position{Line: 1, Column: 1},
				Condition: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				Body: []ast.Statement{
					&ast.LabelStatement{BasePos: ast.Position{Line: 2, Column: 1}, Name: "whileLabel"},
				},
			},
			&ast.GotoStatement{BasePos: ast.Position{Line: 3, Column: 1}, Target: "whileLabel"},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors (label in WHILE body collected), got: %v", errs)
	}
}

func TestCollectTargetsDoLoopStatementBody(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.DoLoopStatement{
				BasePos:   ast.Position{Line: 1, Column: 1},
				Condition: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				Body: []ast.Statement{
					&ast.LabelStatement{BasePos: ast.Position{Line: 2, Column: 1}, Name: "doLabel"},
				},
			},
			&ast.GotoStatement{BasePos: ast.Position{Line: 3, Column: 1}, Target: "doLabel"},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors (label in DO LOOP body collected), got: %v", errs)
	}
}

func TestCollectTargetsSelectCaseAllBlocks(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.SelectCaseStatement{
				BasePos:  ast.Position{Line: 1, Column: 1},
				TestExpr: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				Cases: []ast.CaseClause{
					{
						Values: []ast.CaseValue{
							{Value: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt}},
						},
						Body: []ast.Statement{
							&ast.LabelStatement{BasePos: ast.Position{Line: 2, Column: 1}, Name: "caseLabel"},
						},
					},
				},
				ElseBlock: []ast.Statement{
					&ast.LabelStatement{BasePos: ast.Position{Line: 3, Column: 1}, Name: "caseElseLabel"},
				},
			},
			&ast.GotoStatement{BasePos: ast.Position{Line: 4, Column: 1}, Target: "caseLabel"},
			&ast.GotoStatement{BasePos: ast.Position{Line: 5, Column: 1}, Target: "caseElseLabel"},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors (labels in SELECT CASE blocks collected), got: %v", errs)
	}
}

func TestCollectTargetsDataStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.DataStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Values: []ast.Expression{
					&ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
					&ast.StringLiteral{Value: "hello"},
				},
			},
		},
	}
	rr := NewReferenceResolver(prog)
	rr.Resolve()
	pool := rr.GetDataPool()
	if len(pool) != 2 {
		t.Fatalf("expected 2 items in data pool, got %d", len(pool))
	}
}

// ===========================================================================
// validateReferences – nested block recursion
// ===========================================================================

func TestValidateReferencesOnEventGosubStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LabelStatement{BasePos: ast.Position{Line: 1, Column: 1}, Name: "timerHandler"},
			&ast.OnEventGosubStatement{
				BasePos:   ast.Position{Line: 2, Column: 1},
				EventType: "TIMER",
				Target:    "timerHandler",
			},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for ON TIMER GOSUB with valid target, got: %v", errs)
	}
}

func TestValidateReferencesRestoreWithNonEmptyTarget(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LabelStatement{BasePos: ast.Position{Line: 1, Column: 1}, Name: "dataSection"},
			&ast.RestoreStatement{
				BasePos: ast.Position{Line: 2, Column: 1},
				Target:  "dataSection",
			},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for RESTORE with valid target, got: %v", errs)
	}
}

func TestValidateReferencesRestoreMissingTarget(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.RestoreStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Target:  "missing",
			},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) == 0 {
		t.Fatal("expected error for RESTORE with undefined target")
	}
}

func TestValidateReferencesResumeWithNonNextLabel(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LabelStatement{BasePos: ast.Position{Line: 1, Column: 1}, Name: "errRecovery"},
			&ast.ResumeStatement{
				BasePos: ast.Position{Line: 2, Column: 1},
				Type:    "errRecovery",
			},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for RESUME <label>, got: %v", errs)
	}
}

func TestValidateReferencesResumeMissingLabel(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ResumeStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Type:    "missingLabel",
			},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) == 0 {
		t.Fatal("expected error for RESUME with undefined label")
	}
}

func TestValidateReferencesSubDeclarationBody(t *testing.T) {
	// GOTO inside SUB body to a label inside the same SUB
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.SubDeclaration{
				BasePos: ast.Position{Line: 1, Column: 1},
				Name:    "MySub",
				Body: []ast.Statement{
					&ast.LabelStatement{BasePos: ast.Position{Line: 2, Column: 1}, Name: "subTarget"},
					&ast.GotoStatement{BasePos: ast.Position{Line: 3, Column: 1}, Target: "subTarget"},
				},
			},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for GOTO inside SUB body, got: %v", errs)
	}
}

func TestValidateReferencesFunctionDeclarationBody(t *testing.T) {
	// GOTO inside FUNCTION body to a label inside the same FUNCTION
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.FunctionDeclaration{
				BasePos: ast.Position{Line: 1, Column: 1},
				Name:    "MyFunc",
				Body: []ast.Statement{
					&ast.LabelStatement{BasePos: ast.Position{Line: 2, Column: 1}, Name: "funcTarget"},
					&ast.GotoStatement{BasePos: ast.Position{Line: 3, Column: 1}, Target: "funcTarget"},
				},
			},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for GOTO inside FUNCTION body, got: %v", errs)
	}
}

func TestValidateReferencesIfAllBlocks(t *testing.T) {
	// GOTO in ThenBlock, ElseIfClause, and ElseBlock validated recursively
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LabelStatement{BasePos: ast.Position{Line: 1, Column: 1}, Name: "validTarget"},
			&ast.IfStatement{
				BasePos:   ast.Position{Line: 2, Column: 1},
				Condition: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				ThenBlock: []ast.Statement{
					&ast.GotoStatement{BasePos: ast.Position{Line: 3, Column: 1}, Target: "validTarget"},
				},
				ElseIfClauses: []ast.ElseIfClause{
					{
						BasePos:   ast.Position{Line: 4, Column: 1},
						Condition: &ast.NumberLiteral{Value: 0, NumType: ast.NumInt},
						Body: []ast.Statement{
							&ast.GotoStatement{BasePos: ast.Position{Line: 5, Column: 1}, Target: "validTarget"},
						},
					},
				},
				ElseBlock: []ast.Statement{
					&ast.GotoStatement{BasePos: ast.Position{Line: 6, Column: 1}, Target: "validTarget"},
				},
			},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for GOTO in all IF blocks, got: %v", errs)
	}
}

func TestValidateReferencesIfBlockMissingLabel(t *testing.T) {
	// GOTO inside ThenBlock to missing label → error
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.IfStatement{
				BasePos:   ast.Position{Line: 1, Column: 1},
				Condition: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				ThenBlock: []ast.Statement{
					&ast.GotoStatement{BasePos: ast.Position{Line: 2, Column: 1}, Target: "noSuchLabel"},
				},
			},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) == 0 {
		t.Fatal("expected error for GOTO in ThenBlock to missing label")
	}
}

func TestValidateReferencesForStatementBody(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LabelStatement{BasePos: ast.Position{Line: 1, Column: 1}, Name: "forTarget"},
			&ast.ForStatement{
				BasePos: ast.Position{Line: 2, Column: 1},
				Counter: &ast.Identifier{Name: "i"},
				Start:   &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				End:     &ast.NumberLiteral{Value: 10, NumType: ast.NumInt},
				Body: []ast.Statement{
					&ast.GotoStatement{BasePos: ast.Position{Line: 3, Column: 1}, Target: "forTarget"},
				},
			},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for GOTO in FOR body, got: %v", errs)
	}
}

func TestValidateReferencesWhileStatementBody(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LabelStatement{BasePos: ast.Position{Line: 1, Column: 1}, Name: "whileTarget"},
			&ast.WhileStatement{
				BasePos:   ast.Position{Line: 2, Column: 1},
				Condition: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				Body: []ast.Statement{
					&ast.GotoStatement{BasePos: ast.Position{Line: 3, Column: 1}, Target: "whileTarget"},
				},
			},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for GOTO in WHILE body, got: %v", errs)
	}
}

func TestValidateReferencesDoLoopStatementBody(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LabelStatement{BasePos: ast.Position{Line: 1, Column: 1}, Name: "doTarget"},
			&ast.DoLoopStatement{
				BasePos:   ast.Position{Line: 2, Column: 1},
				Condition: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				Body: []ast.Statement{
					&ast.GotoStatement{BasePos: ast.Position{Line: 3, Column: 1}, Target: "doTarget"},
				},
			},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for GOTO in DO LOOP body, got: %v", errs)
	}
}

func TestValidateReferencesSelectCaseAllBlocks(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LabelStatement{BasePos: ast.Position{Line: 1, Column: 1}, Name: "selectTarget"},
			&ast.SelectCaseStatement{
				BasePos:  ast.Position{Line: 2, Column: 1},
				TestExpr: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				Cases: []ast.CaseClause{
					{
						Values: []ast.CaseValue{
							{Value: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt}},
						},
						Body: []ast.Statement{
							&ast.GotoStatement{BasePos: ast.Position{Line: 3, Column: 1}, Target: "selectTarget"},
						},
					},
				},
				ElseBlock: []ast.Statement{
					&ast.GotoStatement{BasePos: ast.Position{Line: 4, Column: 1}, Target: "selectTarget"},
				},
			},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for GOTO in SELECT CASE blocks, got: %v", errs)
	}
}

func TestValidateReferencesSelectCaseMissingTarget(t *testing.T) {
	// GOTO inside a CASE body to a missing label → error
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.SelectCaseStatement{
				BasePos:  ast.Position{Line: 1, Column: 1},
				TestExpr: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				Cases: []ast.CaseClause{
					{
						Values: []ast.CaseValue{
							{Value: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt}},
						},
						Body: []ast.Statement{
							&ast.GotoStatement{BasePos: ast.Position{Line: 2, Column: 1}, Target: "noSuchLabel"},
						},
					},
				},
			},
		},
	}
	rr := NewReferenceResolver(prog)
	errs := rr.Resolve()
	if len(errs) == 0 {
		t.Fatal("expected error for GOTO in CASE body to missing label")
	}
}
