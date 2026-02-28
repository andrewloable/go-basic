package semantic

import (
	"testing"

	"github.com/loabletech/go-basic/internal/ast"
)

// ===========================================================================
// Statement resolver tests (resolve_statements.go)
// ===========================================================================

// resolveStatement – graphics/sound/file statements

func TestResolveStatementGetStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.GetStatement{
				BasePos:     ast.Position{Line: 1, Column: 1},
				FileNum:     &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				RecordOrPos: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				Variable:    nil,
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for GET statement, got: %v", errs)
	}
}

func TestResolveStatementPutStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.PutStatement{
				BasePos:     ast.Position{Line: 1, Column: 1},
				FileNum:     &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				RecordOrPos: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				Variable:    nil,
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for PUT statement, got: %v", errs)
	}
}

func TestResolveStatementSeekStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.SeekStatement{
				BasePos:  ast.Position{Line: 1, Column: 1},
				FileNum:  &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				Position: &ast.NumberLiteral{Value: 100, NumType: ast.NumInt},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for SEEK statement, got: %v", errs)
	}
}

func TestResolveStatementErrorStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ErrorStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Code:    &ast.NumberLiteral{Value: 5, NumType: ast.NumInt},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for ERROR statement, got: %v", errs)
	}
}

func TestResolveStatementLabelStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LabelStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Name:    "myLabel",
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for label statement, got: %v", errs)
	}
}

func TestResolveStatementEndStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.EndStatement{BasePos: ast.Position{Line: 1, Column: 1}},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for END statement, got: %v", errs)
	}
}

func TestResolveStatementBeepStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.BeepStatement{BasePos: ast.Position{Line: 1, Column: 1}},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for BEEP statement, got: %v", errs)
	}
}

func TestResolveStatementEraseStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.EraseStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Names:   []string{"myArr"},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for ERASE statement, got: %v", errs)
	}
}

func TestResolveStatementScreenStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ScreenStatement{
				BasePos:     ast.Position{Line: 1, Column: 1},
				Mode:        &ast.NumberLiteral{Value: 13, NumType: ast.NumInt},
				ColorSwitch: nil,
				ActivePage:  nil,
				VisualPage:  nil,
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for SCREEN statement, got: %v", errs)
	}
}

func TestResolveStatementColorStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ColorStatement{
				BasePos:    ast.Position{Line: 1, Column: 1},
				Foreground: &ast.NumberLiteral{Value: 7, NumType: ast.NumInt},
				Background: &ast.NumberLiteral{Value: 0, NumType: ast.NumInt},
				Border:     nil,
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for COLOR statement, got: %v", errs)
	}
}

func TestResolveStatementPsetStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.PsetStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				X:       &ast.NumberLiteral{Value: 100, NumType: ast.NumInt},
				Y:       &ast.NumberLiteral{Value: 100, NumType: ast.NumInt},
				Color:   &ast.NumberLiteral{Value: 15, NumType: ast.NumInt},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for PSET statement, got: %v", errs)
	}
}

func TestResolveStatementLineStmt(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LineStmt{
				BasePos: ast.Position{Line: 1, Column: 1},
				X1:      &ast.NumberLiteral{Value: 0, NumType: ast.NumInt},
				Y1:      &ast.NumberLiteral{Value: 0, NumType: ast.NumInt},
				X2:      &ast.NumberLiteral{Value: 100, NumType: ast.NumInt},
				Y2:      &ast.NumberLiteral{Value: 100, NumType: ast.NumInt},
				Color:   &ast.NumberLiteral{Value: 15, NumType: ast.NumInt},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for LINE statement, got: %v", errs)
	}
}

func TestResolveStatementCircleStmt(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.CircleStmt{
				BasePos: ast.Position{Line: 1, Column: 1},
				X:       &ast.NumberLiteral{Value: 160, NumType: ast.NumInt},
				Y:       &ast.NumberLiteral{Value: 100, NumType: ast.NumInt},
				Radius:  &ast.NumberLiteral{Value: 50, NumType: ast.NumInt},
				Color:   &ast.NumberLiteral{Value: 15, NumType: ast.NumInt},
				Start:   nil,
				End:     nil,
				Aspect:  nil,
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for CIRCLE statement, got: %v", errs)
	}
}

func TestResolveStatementPaintStmt(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.PaintStmt{
				BasePos:     ast.Position{Line: 1, Column: 1},
				X:           &ast.NumberLiteral{Value: 100, NumType: ast.NumInt},
				Y:           &ast.NumberLiteral{Value: 100, NumType: ast.NumInt},
				FillColor:   &ast.NumberLiteral{Value: 4, NumType: ast.NumInt},
				BorderColor: &ast.NumberLiteral{Value: 15, NumType: ast.NumInt},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for PAINT statement, got: %v", errs)
	}
}

func TestResolveStatementDrawStmt(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.DrawStmt{
				BasePos:       ast.Position{Line: 1, Column: 1},
				CommandString: &ast.StringLiteral{Value: "U10 R10 D10 L10"},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for DRAW statement, got: %v", errs)
	}
}

func TestResolveStatementSoundStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.SoundStatement{
				BasePos:   ast.Position{Line: 1, Column: 1},
				Frequency: &ast.NumberLiteral{Value: 440, NumType: ast.NumSingle},
				Duration:  &ast.NumberLiteral{Value: 18, NumType: ast.NumSingle},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for SOUND statement, got: %v", errs)
	}
}

func TestResolveStatementPlayStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.PlayStatement{
				BasePos:       ast.Position{Line: 1, Column: 1},
				CommandString: &ast.StringLiteral{Value: "CDE"},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for PLAY statement, got: %v", errs)
	}
}

func TestResolveStatementLocateStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LocateStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Row:     &ast.NumberLiteral{Value: 10, NumType: ast.NumInt},
				Col:     &ast.NumberLiteral{Value: 20, NumType: ast.NumInt},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for LOCATE statement, got: %v", errs)
	}
}

func TestResolveStatementClsStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ClsStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Mode:    &ast.NumberLiteral{Value: 0, NumType: ast.NumInt},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for CLS statement, got: %v", errs)
	}
}

func TestResolveStatementViewStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ViewStatement{
				BasePos:     ast.Position{Line: 1, Column: 1},
				X1:          &ast.NumberLiteral{Value: 0, NumType: ast.NumInt},
				Y1:          &ast.NumberLiteral{Value: 0, NumType: ast.NumInt},
				X2:          &ast.NumberLiteral{Value: 319, NumType: ast.NumInt},
				Y2:          &ast.NumberLiteral{Value: 199, NumType: ast.NumInt},
				FillColor:   nil,
				BorderColor: nil,
				Top:         nil,
				Bottom:      nil,
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for VIEW statement, got: %v", errs)
	}
}

func TestResolveStatementFieldStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.FieldStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				FileNum: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				Fields: []ast.FieldDef{
					{
						Length:  &ast.NumberLiteral{Value: 30, NumType: ast.NumInt},
						VarName: "name$",
					},
				},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for FIELD statement, got: %v", errs)
	}
}

func TestResolveStatementLsetStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.LsetStatement{
				BasePos:  ast.Position{Line: 1, Column: 1},
				Variable: "name$",
				Value:    &ast.StringLiteral{Value: "hello"},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for LSET statement, got: %v", errs)
	}
}

func TestResolveStatementRsetStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.RsetStatement{
				BasePos:  ast.Position{Line: 1, Column: 1},
				Variable: "name$",
				Value:    &ast.StringLiteral{Value: "hello"},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for RSET statement, got: %v", errs)
	}
}

func TestResolveStatementKillStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.KillStatement{
				BasePos:  ast.Position{Line: 1, Column: 1},
				Filename: &ast.StringLiteral{Value: "temp.dat"},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for KILL statement, got: %v", errs)
	}
}

func TestResolveStatementNameStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.NameStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				OldName: &ast.StringLiteral{Value: "old.dat"},
				NewName: &ast.StringLiteral{Value: "new.dat"},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for NAME statement, got: %v", errs)
	}
}

func TestResolveStatementChdirStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ChdirStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Path:    &ast.StringLiteral{Value: "/tmp"},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for CHDIR statement, got: %v", errs)
	}
}

func TestResolveStatementMkdirStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.MkdirStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Path:    &ast.StringLiteral{Value: "/tmp/newdir"},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for MKDIR statement, got: %v", errs)
	}
}

func TestResolveStatementRmdirStatement(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.RmdirStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Path:    &ast.StringLiteral{Value: "/tmp/olddir"},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for RMDIR statement, got: %v", errs)
	}
}

// ===========================================================================
// resolveDefType – all variants
// ===========================================================================

func TestResolveDefTypeDefInt(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.DefTypeStatement{
				BasePos:      ast.Position{Line: 1, Column: 1},
				Type:         "DEFINT",
				LetterRanges: []ast.LetterRange{{Start: 'I', End: 'N'}},
			},
		},
	}
	resolver := NewResolver(prog)
	table, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for DEFINT, got: %v", errs)
	}
	if got := table.ResolveType("ivar"); got != TypeInteger {
		t.Errorf("expected TypeInteger for ivar after DEFINT I-N, got %v", got)
	}
}

func TestResolveDefTypeDefLng(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.DefTypeStatement{
				BasePos:      ast.Position{Line: 1, Column: 1},
				Type:         "DEFLNG",
				LetterRanges: []ast.LetterRange{{Start: 'A', End: 'A'}},
			},
		},
	}
	resolver := NewResolver(prog)
	table, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for DEFLNG, got: %v", errs)
	}
	if got := table.ResolveType("avar"); got != TypeLong {
		t.Errorf("expected TypeLong for avar after DEFLNG A-A, got %v", got)
	}
}

func TestResolveDefTypeDefSng(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.DefTypeStatement{
				BasePos:      ast.Position{Line: 1, Column: 1},
				Type:         "DEFSNG",
				LetterRanges: []ast.LetterRange{{Start: 'S', End: 'S'}},
			},
		},
	}
	resolver := NewResolver(prog)
	table, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for DEFSNG, got: %v", errs)
	}
	if got := table.ResolveType("svar"); got != TypeSingle {
		t.Errorf("expected TypeSingle for svar after DEFSNG S-S, got %v", got)
	}
}

func TestResolveDefTypeDefDbl(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.DefTypeStatement{
				BasePos:      ast.Position{Line: 1, Column: 1},
				Type:         "DEFDBL",
				LetterRanges: []ast.LetterRange{{Start: 'D', End: 'D'}},
			},
		},
	}
	resolver := NewResolver(prog)
	table, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for DEFDBL, got: %v", errs)
	}
	if got := table.ResolveType("dvar"); got != TypeDouble {
		t.Errorf("expected TypeDouble for dvar after DEFDBL D-D, got %v", got)
	}
}

func TestResolveDefTypeDefStr(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.DefTypeStatement{
				BasePos:      ast.Position{Line: 1, Column: 1},
				Type:         "DEFSTR",
				LetterRanges: []ast.LetterRange{{Start: 'T', End: 'T'}},
			},
		},
	}
	resolver := NewResolver(prog)
	table, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for DEFSTR, got: %v", errs)
	}
	if got := table.ResolveType("tvar"); got != TypeString {
		t.Errorf("expected TypeString for tvar after DEFSTR T-T, got %v", got)
	}
}

func TestResolveDefTypeUnknown(t *testing.T) {
	// Unknown DEFTYPE → no-op (default case returns early)
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.DefTypeStatement{
				BasePos:      ast.Position{Line: 1, Column: 1},
				Type:         "DEFXXX",
				LetterRanges: []ast.LetterRange{{Start: 'X', End: 'X'}},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for unknown DEFTYPE, got: %v", errs)
	}
}

// ===========================================================================
// resolveScopeStmt – SHARED, STATIC, LOCAL, COMMON
// ===========================================================================

func TestResolveScopeStmtSharedWithExistingGlobal(t *testing.T) {
	// SHARED variable that already exists in global scope → marks it IsShared
	prog := &ast.Program{
		Statements: []ast.Statement{
			// Define global x first
			&ast.LetStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Name:    &ast.Identifier{Name: "x"},
				Value:   &ast.NumberLiteral{Value: 0, NumType: ast.NumInt},
			},
			// Then SHARED x
			&ast.ScopeStatement{
				BasePos:   ast.Position{Line: 2, Column: 1},
				Modifier:  "SHARED",
				Variables: []string{"x"},
			},
		},
	}
	resolver := NewResolver(prog)
	table, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for SHARED with existing global, got: %v", errs)
	}
	sym := table.GlobalScope.Symbols["X"]
	if sym == nil {
		t.Fatal("expected x to exist in global scope")
	}
	if !sym.IsShared {
		t.Error("expected x to be marked IsShared")
	}
}

func TestResolveScopeStmtSharedWithNoExistingGlobal(t *testing.T) {
	// SHARED with a variable that doesn't yet exist → implicitly creates global
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ScopeStatement{
				BasePos:   ast.Position{Line: 1, Column: 1},
				Modifier:  "SHARED",
				Variables: []string{"newSharedVar"},
			},
		},
	}
	resolver := NewResolver(prog)
	table, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for SHARED creating new global, got: %v", errs)
	}
	sym := table.GlobalScope.Symbols["NEWSHAREDVAR"]
	if sym == nil {
		t.Fatal("expected newSharedVar to be implicitly created in global scope")
	}
	if !sym.IsShared {
		t.Error("expected newSharedVar to be marked IsShared")
	}
}

func TestResolveScopeStmtStaticWithExistingLocal(t *testing.T) {
	// STATIC variable that already exists in local scope → marks it IsStatic
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.SubDeclaration{
				BasePos: ast.Position{Line: 1, Column: 1},
				Name:    "TestSub",
				Body: []ast.Statement{
					// Define local variable first
					&ast.LetStatement{
						BasePos: ast.Position{Line: 2, Column: 1},
						Name:    &ast.Identifier{Name: "counter"},
						Value:   &ast.NumberLiteral{Value: 0, NumType: ast.NumInt},
					},
					// Then STATIC counter
					&ast.ScopeStatement{
						BasePos:   ast.Position{Line: 3, Column: 1},
						Modifier:  "STATIC",
						Variables: []string{"counter"},
					},
				},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for STATIC with existing local, got: %v", errs)
	}
}

func TestResolveScopeStmtStaticWithNoExistingLocal(t *testing.T) {
	// STATIC with variable not yet in local scope → creates it as static
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.SubDeclaration{
				BasePos: ast.Position{Line: 1, Column: 1},
				Name:    "TestSub2",
				Body: []ast.Statement{
					&ast.ScopeStatement{
						BasePos:   ast.Position{Line: 2, Column: 1},
						Modifier:  "STATIC",
						Variables: []string{"staticVar"},
					},
				},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for STATIC creating new local, got: %v", errs)
	}
}

func TestResolveScopeStmtLocalWithNoExistingLocal(t *testing.T) {
	// LOCAL with a variable not yet existing → creates it
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ScopeStatement{
				BasePos:   ast.Position{Line: 1, Column: 1},
				Modifier:  "LOCAL",
				Variables: []string{"localVar"},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for LOCAL statement, got: %v", errs)
	}
}

func TestResolveScopeStmtLocalWithExistingLocal(t *testing.T) {
	// LOCAL with a variable that already exists locally → no-op (LookupLocal returns non-nil)
	prog := &ast.Program{
		Statements: []ast.Statement{
			// Define the variable first
			&ast.LetStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Name:    &ast.Identifier{Name: "existVar"},
				Value:   &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
			},
			// Then LOCAL it
			&ast.ScopeStatement{
				BasePos:   ast.Position{Line: 2, Column: 1},
				Modifier:  "LOCAL",
				Variables: []string{"existVar"},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for LOCAL with existing variable, got: %v", errs)
	}
}

func TestResolveScopeStmtCommon(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ScopeStatement{
				BasePos:   ast.Position{Line: 1, Column: 1},
				Modifier:  "COMMON",
				Variables: []string{"commonVar"},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for COMMON statement, got: %v", errs)
	}
}

// ===========================================================================
// resolveFor with Step != nil
// ===========================================================================

func TestResolveForWithStep(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.ForStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Counter: &ast.Identifier{Name: "i"},
				Start:   &ast.NumberLiteral{Value: 0, NumType: ast.NumInt},
				End:     &ast.NumberLiteral{Value: 100, NumType: ast.NumInt},
				Step:    &ast.NumberLiteral{Value: 5, NumType: ast.NumInt},
				Body:    []ast.Statement{},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for FOR with STEP, got: %v", errs)
	}
}

// resolveFor — counter already defined (sym != nil branch)

func TestResolveForCounterAlreadyDefined(t *testing.T) {
	// Pre-define the counter variable (via LET) so the sym != nil branch is taken.
	prog := &ast.Program{
		Statements: []ast.Statement{
			// Assign i first so it is registered in the symbol table.
			&ast.LetStatement{
				BasePos: ast.Position{Line: 1, Column: 1},
				Name:    &ast.Identifier{Name: "i"},
				Value:   &ast.NumberLiteral{Value: 0, NumType: ast.NumInt},
			},
			&ast.ForStatement{
				BasePos: ast.Position{Line: 2, Column: 1},
				Counter: &ast.Identifier{Name: "i"},
				Start:   &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				End:     &ast.NumberLiteral{Value: 5, NumType: ast.NumInt},
				Body:    []ast.Statement{},
			},
		},
	}
	rr := NewResolver(prog)
	_, errs := rr.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for FOR with pre-declared counter, got: %v", errs)
	}
}

// ===========================================================================
// resolveIf with ElseIf clauses
// ===========================================================================

func TestResolveIfWithElseIf(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.IfStatement{
				BasePos:   ast.Position{Line: 1, Column: 1},
				Condition: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				ThenBlock: []ast.Statement{},
				ElseIfClauses: []ast.ElseIfClause{
					{
						BasePos:   ast.Position{Line: 3, Column: 1},
						Condition: &ast.NumberLiteral{Value: 0, NumType: ast.NumInt},
						Body: []ast.Statement{
							&ast.LetStatement{
								BasePos: ast.Position{Line: 4, Column: 1},
								Name:    &ast.Identifier{Name: "x"},
								Value:   &ast.NumberLiteral{Value: 2, NumType: ast.NumInt},
							},
						},
					},
				},
				ElseBlock: []ast.Statement{},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for IF with ELSEIF, got: %v", errs)
	}
}

func TestResolveIfElseBlock(t *testing.T) {
	prog := &ast.Program{
		Statements: []ast.Statement{
			&ast.IfStatement{
				BasePos:   ast.Position{Line: 1, Column: 1},
				Condition: &ast.NumberLiteral{Value: 1, NumType: ast.NumInt},
				ThenBlock: []ast.Statement{},
				ElseBlock: []ast.Statement{
					&ast.LetStatement{
						BasePos: ast.Position{Line: 3, Column: 1},
						Name:    &ast.Identifier{Name: "y"},
						Value:   &ast.NumberLiteral{Value: 99, NumType: ast.NumInt},
					},
				},
			},
		},
	}
	resolver := NewResolver(prog)
	_, errs := resolver.Resolve()
	if len(errs) != 0 {
		t.Fatalf("expected no errors for IF with ELSE block, got: %v", errs)
	}
}
