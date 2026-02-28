package codegen

import (
	"strings"
	"testing"

	"github.com/loabletech/go-basic/internal/ast"
)

// ---------------------------------------------------------------------------
// Tests for emit_graphics.go — LINE, CIRCLE, PSET, SCREEN, PAINT, VIEW
// and also LOCATE and COLOR (console positioning)
// ---------------------------------------------------------------------------

func TestScreenStatement(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ScreenStatement{
			Mode: &ast.NumberLiteral{Value: 13, NumType: ast.NumInt},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "rt.ScreenMode(int(13))") {
		t.Errorf("expected rt.ScreenMode(int(13)), got:\n%s", out)
	}
}

func TestEmitLocate(t *testing.T) {
	stmts := []ast.Statement{
		&ast.LocateStatement{
			Row: &ast.NumberLiteral{Value: 5, OriginalText: "5"},
			Col: &ast.NumberLiteral{Value: 10, OriginalText: "10"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "AnsiLocate") {
		t.Errorf("expected AnsiLocate in output, got:\n%s", out)
	}
}

func TestEmitLocateNoArgs(t *testing.T) {
	stmts := []ast.Statement{
		&ast.LocateStatement{},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "AnsiLocate") {
		t.Errorf("expected AnsiLocate in output, got:\n%s", out)
	}
}

func TestEmitColor(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ColorStatement{
			Foreground: &ast.NumberLiteral{Value: 7, OriginalText: "7"},
			Background: &ast.NumberLiteral{Value: 0, OriginalText: "0"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "AnsiColor") {
		t.Errorf("expected AnsiColor in output, got:\n%s", out)
	}
}

func TestEmitColorNoBackground(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ColorStatement{
			Foreground: &ast.NumberLiteral{Value: 15, OriginalText: "15"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "AnsiColor") {
		t.Errorf("expected AnsiColor in output, got:\n%s", out)
	}
}

func TestEmitCircle(t *testing.T) {
	stmts := []ast.Statement{
		&ast.CircleStmt{
			X:      &ast.NumberLiteral{Value: 100, OriginalText: "100"},
			Y:      &ast.NumberLiteral{Value: 100, OriginalText: "100"},
			Radius: &ast.NumberLiteral{Value: 50, OriginalText: "50"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "Circle") {
		t.Errorf("expected 'Circle' in output, got:\n%s", out)
	}
}

func TestEmitCircleWithColor(t *testing.T) {
	stmts := []ast.Statement{
		&ast.CircleStmt{
			X:      &ast.NumberLiteral{Value: 160, OriginalText: "160"},
			Y:      &ast.NumberLiteral{Value: 100, OriginalText: "100"},
			Radius: &ast.NumberLiteral{Value: 40, OriginalText: "40"},
			Color:  &ast.NumberLiteral{Value: 4, OriginalText: "4"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "Circle") {
		t.Errorf("expected 'Circle' in output, got:\n%s", out)
	}
}

func TestEmitCircleWithStartEnd(t *testing.T) {
	stmts := []ast.Statement{
		&ast.CircleStmt{
			X:      &ast.NumberLiteral{Value: 100, OriginalText: "100"},
			Y:      &ast.NumberLiteral{Value: 100, OriginalText: "100"},
			Radius: &ast.NumberLiteral{Value: 50, OriginalText: "50"},
			Color:  &ast.NumberLiteral{Value: 4, OriginalText: "4"},
			Start:  &ast.NumberLiteral{Value: 0, OriginalText: "0"},
			End:    &ast.NumberLiteral{Value: 6, OriginalText: "6"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "Circle") {
		t.Errorf("CIRCLE with Start/End: expected 'Circle' in output, got:\n%s", out)
	}
}

func TestEmitCircleWithAspect(t *testing.T) {
	stmts := []ast.Statement{
		&ast.CircleStmt{
			X:      &ast.NumberLiteral{Value: 160, OriginalText: "160"},
			Y:      &ast.NumberLiteral{Value: 100, OriginalText: "100"},
			Radius: &ast.NumberLiteral{Value: 50, OriginalText: "50"},
			Aspect: &ast.NumberLiteral{Value: 2, OriginalText: "2"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "Circle") {
		t.Errorf("CIRCLE with Aspect: expected 'Circle' in output, got:\n%s", out)
	}
}

func TestEmitLine(t *testing.T) {
	stmts := []ast.Statement{
		&ast.LineStmt{
			X2: &ast.NumberLiteral{Value: 200, OriginalText: "200"},
			Y2: &ast.NumberLiteral{Value: 100, OriginalText: "100"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "DrawLine") {
		t.Errorf("expected 'DrawLine' in output, got:\n%s", out)
	}
}

func TestEmitLineWithCoords(t *testing.T) {
	stmts := []ast.Statement{
		&ast.LineStmt{
			X1:    &ast.NumberLiteral{Value: 10, OriginalText: "10"},
			Y1:    &ast.NumberLiteral{Value: 20, OriginalText: "20"},
			X2:    &ast.NumberLiteral{Value: 100, OriginalText: "100"},
			Y2:    &ast.NumberLiteral{Value: 200, OriginalText: "200"},
			Color: &ast.NumberLiteral{Value: 14, OriginalText: "14"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "DrawLine") {
		t.Errorf("expected 'DrawLine' in output, got:\n%s", out)
	}
}

func TestEmitPset(t *testing.T) {
	stmts := []ast.Statement{
		&ast.PsetStatement{
			X: &ast.NumberLiteral{Value: 50, OriginalText: "50"},
			Y: &ast.NumberLiteral{Value: 75, OriginalText: "75"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "Pset") {
		t.Errorf("expected 'Pset' in output, got:\n%s", out)
	}
}

func TestEmitPaint(t *testing.T) {
	stmts := []ast.Statement{
		&ast.PaintStmt{
			X: &ast.NumberLiteral{Value: 100, OriginalText: "100"},
			Y: &ast.NumberLiteral{Value: 50, OriginalText: "50"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "Paint") {
		t.Errorf("expected 'Paint' in output, got:\n%s", out)
	}
}

func TestEmitPaintWithColors(t *testing.T) {
	stmts := []ast.Statement{
		&ast.PaintStmt{
			X:           &ast.NumberLiteral{Value: 100, OriginalText: "100"},
			Y:           &ast.NumberLiteral{Value: 50, OriginalText: "50"},
			FillColor:   &ast.NumberLiteral{Value: 2, OriginalText: "2"},
			BorderColor: &ast.NumberLiteral{Value: 4, OriginalText: "4"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "Paint") {
		t.Errorf("expected 'Paint' in output, got:\n%s", out)
	}
}

func TestEmitViewPrint(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ViewStatement{
			IsPrint: true,
			Top:     &ast.NumberLiteral{Value: 1, OriginalText: "1"},
			Bottom:  &ast.NumberLiteral{Value: 24, OriginalText: "24"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "ViewPrint") {
		t.Errorf("expected 'ViewPrint' in output, got:\n%s", out)
	}
}

func TestEmitViewPrintNoArgs(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ViewStatement{
			IsPrint: true,
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "ViewPrint") {
		t.Errorf("expected 'ViewPrint' in output, got:\n%s", out)
	}
}

func TestEmitViewport(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ViewStatement{
			IsPrint: false,
			X1:      &ast.NumberLiteral{Value: 0, OriginalText: "0"},
			Y1:      &ast.NumberLiteral{Value: 0, OriginalText: "0"},
			X2:      &ast.NumberLiteral{Value: 319, OriginalText: "319"},
			Y2:      &ast.NumberLiteral{Value: 199, OriginalText: "199"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "ViewPort") {
		t.Errorf("VIEW (viewport): expected 'ViewPort' in output, got:\n%s", out)
	}
}

func TestEmitViewportWithColors(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ViewStatement{
			IsPrint:     false,
			X1:          &ast.NumberLiteral{Value: 10, OriginalText: "10"},
			Y1:          &ast.NumberLiteral{Value: 10, OriginalText: "10"},
			X2:          &ast.NumberLiteral{Value: 200, OriginalText: "200"},
			Y2:          &ast.NumberLiteral{Value: 150, OriginalText: "150"},
			FillColor:   &ast.NumberLiteral{Value: 1, OriginalText: "1"},
			BorderColor: &ast.NumberLiteral{Value: 14, OriginalText: "14"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "ViewPort") {
		t.Errorf("VIEW with colors: expected 'ViewPort' in output, got:\n%s", out)
	}
}

func TestEmitViewportNoCoords(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ViewStatement{
			IsPrint: false,
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "ViewPort") {
		t.Errorf("VIEW no coords: expected 'ViewPort' in output, got:\n%s", out)
	}
}

func TestEmitStatementScreen(t *testing.T) {
	stmts := []ast.Statement{
		&ast.ScreenStatement{
			Mode: &ast.NumberLiteral{Value: 12, OriginalText: "12"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "ScreenMode") {
		t.Errorf("SCREEN: expected 'ScreenMode' in output, got:\n%s", out)
	}
}

func TestEmitStatementDraw(t *testing.T) {
	stmts := []ast.Statement{
		&ast.DrawStmt{
			CommandString: &ast.StringLiteral{Value: "BM100,100"},
		},
	}
	out := generate(t, stmts)
	if !strings.Contains(out, "rt.Draw(") {
		t.Errorf("DRAW: expected 'rt.Draw(' in output, got:\n%s", out)
	}
}
