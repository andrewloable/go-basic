// graphics.go — Pixel graphics built-ins for the BASIC runtime.
//
// # Compiler Design Note: BASIC's Graphics Model
//
// Turbo BASIC ran on DOS-era PCs and had direct hardware access to the video
// card's framebuffer — a region of memory where each byte (or pair of bits)
// controlled one pixel on the screen. The SCREEN statement switched the video
// card into different display modes with different resolutions and colour depths.
//
// Common SCREEN modes and their hardware origins:
//
//	SCREEN 0  — Text mode (80x25 or 40x25); no pixel framebuffer.
//	SCREEN 1  — CGA 320x200, 4 colours; the classic early PC game mode.
//	SCREEN 2  — CGA 640x200, 2 colours (mono).
//	SCREEN 7  — EGA 320x200, 16 colours.
//	SCREEN 9  — EGA 640x350, 16 colours.
//	SCREEN 12 — VGA 640x480, 16 colours; common in later DOS programs.
//	SCREEN 13 — VGA/MCGA 320x200, 256 colours; beloved by DOS game developers.
//
// # Framebuffer Simulation
//
// Since we are running on modern hardware with no direct video memory access,
// this runtime simulates the framebuffer as a 2-D byte slice ([][]byte) where
// each element is a colour palette index. The palette maps indices to RGB values.
//
// This is the same approach used by emulators and DOS-compatibility layers: the
// in-memory framebuffer captures exactly what the program draws, and a separate
// rendering layer (not implemented here) would display it.
//
// # Drawing Primitives
//
//   - PSET (x,y), c  — set one pixel.
//   - LINE (x1,y1)-(x2,y2), c  — draw a line using Bresenham's algorithm.
//   - LINE …, c, B   — draw a rectangle outline.
//   - LINE …, c, BF  — draw a filled rectangle.
//   - CIRCLE (x,y), r, c  — draw a circle or ellipse arc.
//   - PAINT (x,y), fill, border  — flood-fill using BFS.
//   - DRAW "cmdstring"  — execute GML (Graphics Macro Language) turtle commands.
//
// # GML (Graphics Macro Language) — the DRAW Command
//
// DRAW accepts a string of single-letter commands that move a virtual pen:
//
//	U n — up n pixels     D n — down     L n — left     R n — right
//	E n — diagonal up-right              F n — down-right
//	G n — down-left                      H n — up-left
//	M x,y — move/draw to absolute or relative coordinates
//	B — prefix: move without drawing (blank move)
//	N — prefix: draw but return pen to starting position
//	C n — set drawing colour index
//	S n — set scale factor
//	A n — set rotation (0-3 = 0°/90°/180°/270°)

package runtime

import (
	"math"
	"strconv"
	"strings"
	"unicode"
)

// Color represents an RGB color in the screen palette.
type Color struct {
	R, G, B uint8
}

// Screen represents the BASIC SCREEN mode state, including a virtual framebuffer.
type Screen struct {
	Mode        int
	Width       int
	Height      int
	Colors      int
	CursorX     int      // current graphics drawing position X
	CursorY     int      // current graphics drawing position Y
	Framebuffer [][]byte // [y][x] = palette index; nil for text mode 0
	Palette     []Color  // color palette for this mode
	// Graphics viewport (0,0,0,0 = full screen)
	ViewX1, ViewY1, ViewX2, ViewY2 int
	ViewportActive                 bool // true if viewport is set
	// Text scrolling viewport (0,0 = full screen)
	TextTop, TextBottom int
	// Draw GML state
	DrawColor byte // current draw color (set by Draw "Cn")
	DrawScale int  // draw scale (default 1)
	DrawAngle int  // draw angle 0-3 (0=0°,1=90°,2=180°,3=270°)
}

// defaultPalette returns the default color palette for the given number of colors.
// Palettes are based on standard CGA/EGA/VGA defaults.
func defaultPalette(nColors int) []Color {
	// Standard CGA 16-color palette (also used as base for EGA modes).
	cga16 := []Color{
		{0, 0, 0},       // 0  Black
		{0, 0, 170},     // 1  Blue
		{0, 170, 0},     // 2  Green
		{0, 170, 170},   // 3  Cyan
		{170, 0, 0},     // 4  Red
		{170, 0, 170},   // 5  Magenta
		{170, 85, 0},    // 6  Brown
		{170, 170, 170}, // 7  Light Gray
		{85, 85, 85},    // 8  Dark Gray
		{85, 85, 255},   // 9  Light Blue
		{85, 255, 85},   // 10 Light Green
		{85, 255, 255},  // 11 Light Cyan
		{255, 85, 85},   // 12 Light Red
		{255, 85, 255},  // 13 Light Magenta
		{255, 255, 85},  // 14 Yellow
		{255, 255, 255}, // 15 White
	}
	switch nColors {
	case 2:
		return cga16[:2]
	case 4:
		// CGA mode 1 default palette (cyan/magenta/white on black)
		return []Color{
			{0, 0, 0},       // 0 Black
			{0, 170, 170},   // 1 Cyan
			{170, 0, 170},   // 2 Magenta
			{170, 170, 170}, // 3 White
		}
	case 16:
		return append([]Color(nil), cga16...)
	case 256:
		// VGA 256-color: first 16 match CGA, rest are gradients (simplified).
		pal := make([]Color, 256)
		copy(pal, cga16)
		for i := 16; i < 256; i++ {
			v := uint8(i)
			pal[i] = Color{v, v, v}
		}
		return pal
	}
	return append([]Color(nil), cga16...)
}

// allocFramebuffer allocates a Width×Height framebuffer filled with color index 0.
func allocFramebuffer(w, h int) [][]byte {
	fb := make([][]byte, h)
	for y := range fb {
		fb[y] = make([]byte, w)
	}
	return fb
}

// CurrentScreen holds the active screen mode. Default is text mode 0 (no framebuffer).
var CurrentScreen = &Screen{Mode: 0, Width: 80, Height: 25, Colors: 16, DrawScale: 1}

// ScreenMode activates the given BASIC SCREEN mode.
//
// Each mode specifies a pixel width, height, and colour depth. The function
// allocates a fresh framebuffer for graphic modes (mode != 0) and sets the
// default colour palette. Unknown mode numbers are silently ignored, matching
// Turbo BASIC's behaviour (the current screen remains unchanged).
//
// SCREEN 0 is special: it is text mode with no framebuffer. Drawing operations
// on a nil framebuffer are no-ops.
func ScreenMode(mode int) {
	type modeSpec struct{ w, h, c int }
	specs := map[int]modeSpec{
		0:  {80, 25, 16},    // Text mode — no framebuffer
		1:  {320, 200, 4},   // CGA
		2:  {640, 200, 2},   // CGA mono
		7:  {320, 200, 16},  // EGA
		8:  {640, 200, 16},  // EGA
		9:  {640, 350, 16},  // EGA
		10: {640, 350, 4},   // EGA mono
		11: {640, 480, 2},   // VGA mono
		12: {640, 480, 16},  // VGA
		13: {320, 200, 256}, // MCGA/VGA 256
	}
	spec, ok := specs[mode]
	if !ok {
		return // unknown mode — leave current screen unchanged
	}
	s := &Screen{
		Mode:      mode,
		Width:     spec.w,
		Height:    spec.h,
		Colors:    spec.c,
		CursorX:   0,
		CursorY:   0,
		Palette:   defaultPalette(spec.c),
		DrawScale: 1,
	}
	if mode != 0 {
		s.Framebuffer = allocFramebuffer(spec.w, spec.h)
	}
	CurrentScreen = s
}

// ClearScreen fills the framebuffer with color index 0 (background).
func ClearScreen() {
	if CurrentScreen.Framebuffer == nil {
		return
	}
	for y := range CurrentScreen.Framebuffer {
		for x := range CurrentScreen.Framebuffer[y] {
			CurrentScreen.Framebuffer[y][x] = 0
		}
	}
}

// pset is an internal helper that sets a pixel with bounds and viewport checking.
func pset(x, y int, color byte) {
	s := CurrentScreen
	fb := s.Framebuffer
	if fb == nil {
		return
	}
	// Viewport clipping
	if s.ViewportActive {
		if x < s.ViewX1 || x > s.ViewX2 || y < s.ViewY1 || y > s.ViewY2 {
			return
		}
	}
	// Framebuffer bounds check
	if y < 0 || y >= len(fb) {
		return
	}
	if x < 0 || x >= len(fb[y]) {
		return
	}
	fb[y][x] = color
	s.CursorX = x
	s.CursorY = y
}

// bresenhamLine draws a straight line from (x0,y0) to (x1,y1) using Bresenham's
// integer line algorithm.
//
// Bresenham's algorithm (1965) draws lines using only integer addition and
// comparison — no floating-point arithmetic required. It maintains an error
// accumulator that tracks how far the actual line has deviated from the ideal
// mathematical line, and steps the minor axis whenever the error exceeds 0.5
// of a pixel. This made it extremely fast on the integer-only hardware of the
// 1980s and it is still the standard algorithm for rasterising lines.
func bresenhamLine(x0, y0, x1, y1 int, color byte) {
	dx := x1 - x0
	if dx < 0 {
		dx = -dx
	}
	dy := y1 - y0
	if dy < 0 {
		dy = -dy
	}
	sx := 1
	if x0 > x1 {
		sx = -1
	}
	sy := 1
	if y0 > y1 {
		sy = -1
	}
	err := dx - dy
	for {
		pset(x0, y0, color)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

// Pset sets a pixel at (x, y) to the given color index.
func Pset(x, y, color float64) {
	pset(int(x), int(y), byte(color))
}

// DrawLine draws a line or box. boxMode is "", "B", or "BF".
// Named DrawLine to avoid conflict with Go's bufio.Scanner line reading.
func DrawLine(x1, y1, x2, y2, color float64, boxMode string) {
	ix1, iy1 := int(x1), int(y1)
	ix2, iy2 := int(x2), int(y2)
	c := byte(color)

	switch boxMode {
	case "":
		// Bresenham line
		bresenhamLine(ix1, iy1, ix2, iy2, c)

	case "B":
		// Rectangle outline: draw 4 sides
		bresenhamLine(ix1, iy1, ix2, iy1, c) // top
		bresenhamLine(ix2, iy1, ix2, iy2, c) // right
		bresenhamLine(ix2, iy2, ix1, iy2, c) // bottom
		bresenhamLine(ix1, iy2, ix1, iy1, c) // left

	case "BF":
		// Filled rectangle
		// Ensure correct ordering
		startY, endY := iy1, iy2
		if startY > endY {
			startY, endY = endY, startY
		}
		startX, endX := ix1, ix2
		if startX > endX {
			startX, endX = endX, startX
		}
		for row := startY; row <= endY; row++ {
			for col := startX; col <= endX; col++ {
				pset(col, row, c)
			}
		}
	}
}

// Circle draws an ellipse or arc.
// aspect: 1.0 = circle, <1 = wide ellipse, >1 = tall ellipse.
// start/end are in radians; 0,0 means full circle.
func Circle(x, y, radius, color, start, end, aspect float64) {
	cx, cy := int(x), int(y)
	c := byte(color)

	rx := radius
	ry := radius * aspect

	if start == 0 && end == 0 {
		// Full ellipse using midpoint algorithm
		// We'll use a parametric approach for generality with aspect ratio
		// For integer radius, use midpoint circle; for aspect != 1 use parametric
		if aspect == 1.0 {
			// Midpoint circle algorithm
			r := int(radius)
			ddx := 0
			ddy := r
			d := 1 - r
			plotCirclePoints(cx, cy, ddx, ddy, c)
			for ddx < ddy {
				if d < 0 {
					d += 2*ddx + 3
				} else {
					d += 2*(ddx-ddy) + 5
					ddy--
				}
				ddx++
				plotCirclePoints(cx, cy, ddx, ddy, c)
			}
		} else {
			// Parametric ellipse
			steps := int(2 * math.Pi * math.Max(rx, ry))
			if steps < 1 {
				steps = 1
			}
			for i := 0; i <= steps; i++ {
				angle := 2 * math.Pi * float64(i) / float64(steps)
				px := cx + int(math.Round(rx*math.Cos(angle)))
				py := cy + int(math.Round(ry*math.Sin(angle)))
				pset(px, py, c)
			}
		}
	} else {
		// Arc: step from start to end
		startAngle := start
		endAngle := end
		if endAngle < startAngle {
			endAngle += 2 * math.Pi
		}
		step := 1.0 / math.Max(rx, ry)
		if step <= 0 {
			step = 0.01
		}
		for angle := startAngle; angle <= endAngle; angle += step {
			px := cx + int(math.Round(rx*math.Cos(angle)))
			py := cy + int(math.Round(ry*math.Sin(angle)))
			pset(px, py, c)
		}
		// Make sure endpoint is drawn
		px := cx + int(math.Round(rx*math.Cos(endAngle)))
		py := cy + int(math.Round(ry*math.Sin(endAngle)))
		pset(px, py, c)
	}
}

// plotCirclePoints plots all 8 symmetric points of a circle.
func plotCirclePoints(cx, cy, x, y int, color byte) {
	pset(cx+x, cy+y, color)
	pset(cx-x, cy+y, color)
	pset(cx+x, cy-y, color)
	pset(cx-x, cy-y, color)
	pset(cx+y, cy+x, color)
	pset(cx-y, cy+x, color)
	pset(cx+y, cy-x, color)
	pset(cx-y, cy-x, color)
}

// Paint flood-fills the area reachable from (x,y) with fillColor, stopping at
// pixels that already have borderColor.
//
// BASIC: PAINT (x,y), fillColor, borderColor
//
// The implementation uses a BFS (breadth-first search) queue. This is the
// standard "bucket fill" algorithm found in paint programs. Alternatives include
// DFS (recursive, risks stack overflow on large areas) and scan-line fill
// (faster but more complex to implement). BFS is chosen here for correctness
// and simplicity.
//
// 4-directional connectivity means only up/down/left/right neighbours are
// examined, not diagonals — matching Turbo BASIC's original behaviour.
func Paint(x, y, fillColor, borderColor float64) {
	s := CurrentScreen
	fb := s.Framebuffer
	if fb == nil {
		return
	}
	startX, startY := int(x), int(y)
	fc := byte(fillColor)
	bc := byte(borderColor)

	// Bounds check start position
	if startY < 0 || startY >= len(fb) || startX < 0 || startX >= len(fb[startY]) {
		return
	}

	// Don't fill if starting pixel is already the fill color or is the border color
	startColor := fb[startY][startX]
	if startColor == fc || startColor == bc {
		return
	}

	// BFS flood fill
	type point struct{ x, y int }
	queue := []point{{startX, startY}}
	fb[startY][startX] = fc

	dirs := [4]point{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		for _, d := range dirs {
			nx, ny := curr.x+d.x, curr.y+d.y
			// Bounds check
			if ny < 0 || ny >= len(fb) || nx < 0 || nx >= len(fb[ny]) {
				continue
			}
			// Viewport check
			if s.ViewportActive {
				if nx < s.ViewX1 || nx > s.ViewX2 || ny < s.ViewY1 || ny > s.ViewY2 {
					continue
				}
			}
			pixel := fb[ny][nx]
			if pixel == bc || pixel == fc {
				continue
			}
			fb[ny][nx] = fc
			queue = append(queue, point{nx, ny})
		}
	}
}

// rotateDirection applies the current DrawAngle rotation to a direction vector.
// angle: 0=0°, 1=90°, 2=180°, 3=270°
func rotateDirection(dx, dy, angle int) (int, int) {
	for i := 0; i < angle; i++ {
		// Rotate 90° clockwise: (dx,dy) -> (dy, -dx) but in screen coords (y increases down)
		// 90° CCW in screen space (matching BASIC DRAW spec): (dx,dy) -> (-dy, dx)
		dx, dy = -dy, dx
	}
	return dx, dy
}

// Draw executes a GML (Graphics Macro Language) command string.
func Draw(cmd string) {
	s := CurrentScreen
	if s.Framebuffer == nil {
		return
	}

	i := 0
	upper := strings.ToUpper(cmd)
	n := len(upper)

	for i < n {
		// Skip whitespace
		if upper[i] == ' ' || upper[i] == ';' || upper[i] == ',' {
			i++
			continue
		}

		// Check for B (blank/no-draw) or N (no-advance) prefix
		noMove := false   // N prefix: draw but return to start
		noDraw := false   // B prefix: move but don't draw
		startX := s.CursorX
		startY := s.CursorY

		if i < n && upper[i] == 'B' {
			// peek ahead: if next char is a movement command, treat as prefix
			if i+1 < n && isDrawMovementCmd(upper[i+1]) {
				noDraw = true
				i++
			}
		} else if i < n && upper[i] == 'N' {
			if i+1 < n && isDrawMovementCmd(upper[i+1]) {
				noMove = true
				i++
			}
		}

		if i >= n {
			break
		}

		ch := upper[i]
		i++

		// Parse optional number argument.
		// For 'M', consume digits/sign/comma to capture "x,y".
		numStr := ""
		if ch == 'M' {
			// Consume optional leading sign + digits, then comma, then digits
			for i < n && (upper[i] == '-' || upper[i] == '+' || (upper[i] >= '0' && upper[i] <= '9') || upper[i] == ',') {
				numStr += string(upper[i])
				i++
			}
		} else {
			for i < n && (upper[i] == '-' || upper[i] == '+' || (upper[i] >= '0' && upper[i] <= '9')) {
				numStr += string(upper[i])
				i++
			}
		}

		switch ch {
		case 'C':
			// Set color
			if numStr != "" {
				v, err := strconv.Atoi(numStr)
				if err == nil {
					s.DrawColor = byte(v)
				}
			}

		case 'S':
			// Set scale
			if numStr != "" {
				v, err := strconv.Atoi(numStr)
				if err == nil && v > 0 {
					s.DrawScale = v
				}
			}

		case 'A':
			// Set angle
			if numStr != "" {
				v, err := strconv.Atoi(numStr)
				if err == nil {
					s.DrawAngle = v & 3 // clamp to 0-3
				}
			}

		case 'M':
			// Move to position: "Mx,y" absolute or "+x,y" / "-x,y" relative
			// numStr contains "x,y" possibly with leading +/-
			parts := strings.SplitN(numStr, ",", 2)
			if len(parts) == 2 {
				xStr := parts[0]
				yStr := parts[1]
				relative := false
				if len(xStr) > 0 && (xStr[0] == '+' || xStr[0] == '-') {
					relative = true
				}
				xv, errX := strconv.Atoi(xStr)
				yv, errY := strconv.Atoi(yStr)
				if errX == nil && errY == nil {
					oldX, oldY := s.CursorX, s.CursorY
					var newX, newY int
					if relative {
						newX = s.CursorX + xv*s.DrawScale
						newY = s.CursorY + yv*s.DrawScale
					} else {
						newX = xv
						newY = yv
					}
					if !noDraw {
						bresenhamLine(oldX, oldY, newX, newY, s.DrawColor)
					}
					if noMove {
						s.CursorX = oldX
						s.CursorY = oldY
					} else {
						s.CursorX = newX
						s.CursorY = newY
					}
				}
			}

		default:
			// Directional commands: U, D, L, R, E, F, G, H
			steps := 1
			if numStr != "" {
				v, err := strconv.Atoi(numStr)
				if err == nil {
					steps = v
				}
			}
			steps *= s.DrawScale

			// Base direction vectors (before angle rotation)
			// U=up, D=down, L=left, R=right
			// E=up-right, F=down-right, G=down-left, H=up-left
			var dx, dy int
			switch ch {
			case 'U':
				dx, dy = 0, -1
			case 'D':
				dx, dy = 0, 1
			case 'L':
				dx, dy = -1, 0
			case 'R':
				dx, dy = 1, 0
			case 'E':
				dx, dy = 1, -1
			case 'F':
				dx, dy = 1, 1
			case 'G':
				dx, dy = -1, 1
			case 'H':
				dx, dy = -1, -1
			default:
				// Unknown command, skip
				continue
			}

			// Apply angle rotation
			dx, dy = rotateDirection(dx, dy, s.DrawAngle)

			newX := s.CursorX + dx*steps
			newY := s.CursorY + dy*steps

			oldX, oldY := s.CursorX, s.CursorY
			if !noDraw {
				bresenhamLine(oldX, oldY, newX, newY, s.DrawColor)
			}
			if noMove {
				s.CursorX = startX
				s.CursorY = startY
			} else {
				s.CursorX = newX
				s.CursorY = newY
			}
		}

		_ = noMove
		_ = noDraw
		_ = startX
		_ = startY
	}
}

// isDrawMovementCmd returns true if ch is a valid Draw movement command letter.
func isDrawMovementCmd(ch byte) bool {
	switch unicode.ToUpper(rune(ch)) {
	case 'U', 'D', 'L', 'R', 'E', 'F', 'G', 'H', 'M':
		return true
	}
	return false
}

// ViewPort sets the graphics viewport.
// fillColor >= 0: fill the viewport area with that color.
// borderColor >= 0: draw a rectangle border around the viewport.
func ViewPort(x1, y1, x2, y2, fillColor, borderColor float64) {
	s := CurrentScreen
	ix1, iy1 := int(x1), int(y1)
	ix2, iy2 := int(x2), int(y2)

	s.ViewX1 = ix1
	s.ViewY1 = iy1
	s.ViewX2 = ix2
	s.ViewY2 = iy2
	s.ViewportActive = true

	if fillColor >= 0 {
		fc := byte(fillColor)
		// Fill the viewport area directly (bypass viewport clipping since we are setting it)
		fb := s.Framebuffer
		if fb != nil {
			startY, endY := iy1, iy2
			if startY > endY {
				startY, endY = endY, startY
			}
			startX, endX := ix1, ix2
			if startX > endX {
				startX, endX = endX, startX
			}
			for row := startY; row <= endY; row++ {
				if row < 0 || row >= len(fb) {
					continue
				}
				for col := startX; col <= endX; col++ {
					if col < 0 || col >= len(fb[row]) {
						continue
					}
					fb[row][col] = fc
				}
			}
		}
	}

	if borderColor >= 0 {
		bc := byte(borderColor)
		bresenhamLine(ix1, iy1, ix2, iy1, bc) // top
		bresenhamLine(ix2, iy1, ix2, iy2, bc) // right
		bresenhamLine(ix2, iy2, ix1, iy2, bc) // bottom
		bresenhamLine(ix1, iy2, ix1, iy1, bc) // left
	}
}

// ViewPrint sets the text scrolling viewport (top/bottom row numbers).
func ViewPrint(top, bottom float64) {
	CurrentScreen.TextTop = int(top)
	CurrentScreen.TextBottom = int(bottom)
}
