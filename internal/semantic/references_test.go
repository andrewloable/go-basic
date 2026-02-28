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
