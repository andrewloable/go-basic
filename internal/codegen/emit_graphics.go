package codegen

import (
	"strconv"

	"github.com/loabletech/go-basic/internal/ast"
)

// ---------------------------------------------------------------------------
// Graphics drawing commands: CIRCLE, LINE, PSET, PAINT, VIEW
// ---------------------------------------------------------------------------

func (g *CodeGenerator) emitCircle(s *ast.CircleStmt) {
	x := g.emitExpr(s.X)
	y := g.emitExpr(s.Y)
	radius := g.emitExpr(s.Radius)
	color := "rt.GetForegroundColor()"
	if s.Color != nil {
		color = "float64(" + g.emitExpr(s.Color) + ")"
	}
	start := "0"
	if s.Start != nil {
		start = g.emitExpr(s.Start)
	}
	end := "0"
	if s.End != nil {
		end = g.emitExpr(s.End)
	}
	aspect := "1"
	if s.Aspect != nil {
		aspect = g.emitExpr(s.Aspect)
	}
	g.writeLinef("rt.Circle(float64(%s), float64(%s), float64(%s), %s, float64(%s), float64(%s), float64(%s))",
		x, y, radius, color, start, end, aspect)
}

func (g *CodeGenerator) emitLine(s *ast.LineStmt) {
	x1 := "0"
	y1 := "0"
	if s.X1 != nil {
		x1 = g.emitExpr(s.X1)
	}
	if s.Y1 != nil {
		y1 = g.emitExpr(s.Y1)
	}
	x2 := g.emitExpr(s.X2)
	y2 := g.emitExpr(s.Y2)
	color := "rt.GetForegroundColor()"
	if s.Color != nil {
		color = "float64(" + g.emitExpr(s.Color) + ")"
	}
	boxMode := strconv.Quote(s.BoxMode)
	g.writeLinef("rt.DrawLine(float64(%s), float64(%s), float64(%s), float64(%s), %s, %s)",
		x1, y1, x2, y2, color, boxMode)
}

func (g *CodeGenerator) emitPset(s *ast.PsetStatement) {
	x := g.emitExpr(s.X)
	y := g.emitExpr(s.Y)
	color := "rt.GetForegroundColor()"
	if s.IsPreset && s.Color == nil {
		// PRESET with no color uses the background color.
		color = "rt.GetBackgroundColor()"
	} else if s.Color != nil {
		color = "float64(" + g.emitExpr(s.Color) + ")"
	}
	g.writeLinef("rt.Pset(float64(%s), float64(%s), %s)", x, y, color)
}

func (g *CodeGenerator) emitPaint(s *ast.PaintStmt) {
	x := g.emitExpr(s.X)
	y := g.emitExpr(s.Y)
	fillColor := "rt.GetForegroundColor()"
	if s.FillColor != nil {
		fillColor = "float64(" + g.emitExpr(s.FillColor) + ")"
	}
	borderColor := fillColor
	if s.BorderColor != nil {
		borderColor = "float64(" + g.emitExpr(s.BorderColor) + ")"
	}
	g.writeLinef("rt.Paint(float64(%s), float64(%s), %s, %s)", x, y, fillColor, borderColor)
}

func (g *CodeGenerator) emitView(s *ast.ViewStatement) {
	if s.IsPrint {
		top := "1"
		bottom := "25"
		if s.Top != nil {
			top = g.emitExpr(s.Top)
		}
		if s.Bottom != nil {
			bottom = g.emitExpr(s.Bottom)
		}
		g.writeLinef("rt.ViewPrint(float64(%s), float64(%s))", top, bottom)
		return
	}
	x1 := "0"
	y1 := "0"
	x2 := "0"
	y2 := "0"
	if s.X1 != nil {
		x1 = g.emitExpr(s.X1)
	}
	if s.Y1 != nil {
		y1 = g.emitExpr(s.Y1)
	}
	if s.X2 != nil {
		x2 = g.emitExpr(s.X2)
	}
	if s.Y2 != nil {
		y2 = g.emitExpr(s.Y2)
	}
	fillColor := "0"
	if s.FillColor != nil {
		fillColor = g.emitExpr(s.FillColor)
	}
	borderColor := "0"
	if s.BorderColor != nil {
		borderColor = g.emitExpr(s.BorderColor)
	}
	g.writeLinef("rt.ViewPort(float64(%s), float64(%s), float64(%s), float64(%s), float64(%s), float64(%s))",
		x1, y1, x2, y2, fillColor, borderColor)
}

// ---------------------------------------------------------------------------
// LOCATE / COLOR
// ---------------------------------------------------------------------------

func (g *CodeGenerator) emitLocate(s *ast.LocateStatement) {
	row := "1"
	col := "1"
	if s.Row != nil {
		row = g.emitExpr(s.Row)
	}
	if s.Col != nil {
		col = g.emitExpr(s.Col)
	}
	g.writeLinef("fmt.Print(rt.AnsiLocate(int(%s), int(%s)))", row, col)
}

func (g *CodeGenerator) emitColor(s *ast.ColorStatement) {
	fg := "7"
	bg := "0"
	if s.Foreground != nil {
		fg = g.emitExpr(s.Foreground)
	}
	if s.Background != nil {
		bg = g.emitExpr(s.Background)
	}
	// Set ANSI terminal colors (text mode).
	g.writeLinef("fmt.Print(rt.AnsiColor(int(%s), int(%s)))", fg, bg)
	// Also set graphics foreground/background for PSET/LINE/CIRCLE defaults.
	g.writeLinef("rt.SetGraphicsColor(int(%s), int(%s))", fg, bg)
}

// ---------------------------------------------------------------------------
// Graphics GET / PUT (sprite capture and blit)
// ---------------------------------------------------------------------------

func (g *CodeGenerator) emitGraphicsGet(s *ast.GraphicsGetStatement) {
	x1 := g.emitExpr(s.X1)
	y1 := g.emitExpr(s.Y1)
	x2 := "0"
	y2 := "0"
	if s.X2 != nil {
		x2 = g.emitExpr(s.X2)
	}
	if s.Y2 != nil {
		y2 = g.emitExpr(s.Y2)
	}
	arrayName := g.emitExpr(s.ArrayVar)
	g.writeLinef("rt.GraphicsGet(int(%s), int(%s), int(%s), int(%s), &%s)", x1, y1, x2, y2, arrayName)
}

func (g *CodeGenerator) emitGraphicsPut(s *ast.GraphicsPutStatement) {
	x := g.emitExpr(s.X)
	y := g.emitExpr(s.Y)
	arrayName := g.emitExpr(s.ArrayVar)
	action := strconv.Quote(s.Action)
	g.writeLinef("rt.GraphicsPut(int(%s), int(%s), %s[:], %s)", x, y, arrayName, action)
}
