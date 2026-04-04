package runtime

import "testing"

func TestScreenModeFramebuffer(t *testing.T) {
	tests := []struct {
		mode    int
		w, h, c int
		hasFB   bool
	}{
		{0, 80, 25, 16, false},
		{1, 320, 200, 4, true},
		{2, 640, 200, 2, true},
		{7, 320, 200, 16, true},
		{8, 640, 200, 16, true},
		{9, 640, 350, 16, true},
		{10, 640, 350, 4, true},
		{11, 640, 480, 2, true},
		{12, 640, 480, 16, true},
		{13, 320, 200, 256, true},
	}
	for _, tt := range tests {
		ScreenMode(tt.mode)
		s := CurrentScreen
		if s.Mode != tt.mode {
			t.Errorf("mode %d: Mode=%d want %d", tt.mode, s.Mode, tt.mode)
		}
		if s.Width != tt.w || s.Height != tt.h {
			t.Errorf("mode %d: size=%dx%d want %dx%d", tt.mode, s.Width, s.Height, tt.w, tt.h)
		}
		if s.Colors != tt.c {
			t.Errorf("mode %d: Colors=%d want %d", tt.mode, s.Colors, tt.c)
		}
		if tt.hasFB && s.Framebuffer == nil {
			t.Errorf("mode %d: expected non-nil framebuffer", tt.mode)
		}
		if !tt.hasFB && s.Framebuffer != nil {
			t.Errorf("mode %d: expected nil framebuffer (text mode)", tt.mode)
		}
		if tt.hasFB {
			if len(s.Framebuffer) != tt.h {
				t.Errorf("mode %d: fb height=%d want %d", tt.mode, len(s.Framebuffer), tt.h)
			}
			if len(s.Framebuffer[0]) != tt.w {
				t.Errorf("mode %d: fb width=%d want %d", tt.mode, len(s.Framebuffer[0]), tt.w)
			}
		}
		if len(s.Palette) != tt.c {
			t.Errorf("mode %d: palette size=%d want %d", tt.mode, len(s.Palette), tt.c)
		}
		if s.CursorX != 0 || s.CursorY != 0 {
			t.Errorf("mode %d: cursor not reset to (0,0)", tt.mode)
		}
	}
}

func TestScreenModeCursorReset(t *testing.T) {
	ScreenMode(13)
	CurrentScreen.CursorX = 100
	CurrentScreen.CursorY = 50
	ScreenMode(13) // re-enter same mode
	if CurrentScreen.CursorX != 0 || CurrentScreen.CursorY != 0 {
		t.Errorf("cursor not reset on ScreenMode: got (%d,%d)", CurrentScreen.CursorX, CurrentScreen.CursorY)
	}
}

func TestClearScreen(t *testing.T) {
	ScreenMode(13)
	CurrentScreen.Framebuffer[10][20] = 5
	ClearScreen()
	for y := range CurrentScreen.Framebuffer {
		for x, v := range CurrentScreen.Framebuffer[y] {
			if v != 0 {
				t.Errorf("ClearScreen: pixel [%d][%d] = %d, want 0", y, x, v)
			}
		}
	}
}

func TestClearScreenTextMode(t *testing.T) {
	ScreenMode(0) // text mode — no framebuffer
	ClearScreen() // must not panic
}

func TestScreenModeUnknown(t *testing.T) {
	ScreenMode(13)
	before := CurrentScreen
	ScreenMode(99) // unknown mode — should leave screen unchanged
	if CurrentScreen != before {
		t.Error("unknown ScreenMode should leave CurrentScreen unchanged")
	}
}

// TestPsetBasic verifies that Pset sets the correct framebuffer pixel.
func TestPsetBasic(t *testing.T) {
	ScreenMode(13)
	Pset(10, 10, 1)
	got := CurrentScreen.Framebuffer[10][10]
	if got != 1 {
		t.Errorf("Pset(10,10,1): fb[10][10] = %d, want 1", got)
	}
}

// TestPsetOutOfBounds verifies that Pset does not panic on out-of-bounds coordinates.
func TestPsetOutOfBounds(t *testing.T) {
	ScreenMode(13)
	// These must not panic
	Pset(-1, -1, 1)
	Pset(999, 999, 1)
}

// TestDrawLineHorizontal verifies that a horizontal line sets pixels in a row.
func TestDrawLineHorizontal(t *testing.T) {
	ScreenMode(13)
	DrawLine(0, 0, 9, 0, 1, "")
	for x := 0; x <= 9; x++ {
		if CurrentScreen.Framebuffer[0][x] != 1 {
			t.Errorf("DrawLineHorizontal: fb[0][%d] = %d, want 1", x, CurrentScreen.Framebuffer[0][x])
		}
	}
}

// TestDrawLineVertical verifies that a vertical line sets pixels in a column.
func TestDrawLineVertical(t *testing.T) {
	ScreenMode(13)
	DrawLine(0, 0, 0, 9, 1, "")
	for y := 0; y <= 9; y++ {
		if CurrentScreen.Framebuffer[y][0] != 1 {
			t.Errorf("DrawLineVertical: fb[%d][0] = %d, want 1", y, CurrentScreen.Framebuffer[y][0])
		}
	}
}

// TestDrawLineBox verifies that boxMode "B" draws a rectangle outline.
func TestDrawLineBox(t *testing.T) {
	ScreenMode(13)
	DrawLine(1, 1, 3, 3, 1, "B")

	// Top edge
	for x := 1; x <= 3; x++ {
		if CurrentScreen.Framebuffer[1][x] != 1 {
			t.Errorf("BoxTop: fb[1][%d] = %d, want 1", x, CurrentScreen.Framebuffer[1][x])
		}
	}
	// Bottom edge
	for x := 1; x <= 3; x++ {
		if CurrentScreen.Framebuffer[3][x] != 1 {
			t.Errorf("BoxBottom: fb[3][%d] = %d, want 1", x, CurrentScreen.Framebuffer[3][x])
		}
	}
	// Left edge
	for y := 1; y <= 3; y++ {
		if CurrentScreen.Framebuffer[y][1] != 1 {
			t.Errorf("BoxLeft: fb[%d][1] = %d, want 1", y, CurrentScreen.Framebuffer[y][1])
		}
	}
	// Right edge
	for y := 1; y <= 3; y++ {
		if CurrentScreen.Framebuffer[y][3] != 1 {
			t.Errorf("BoxRight: fb[%d][3] = %d, want 1", y, CurrentScreen.Framebuffer[y][3])
		}
	}
}

// TestPaintFill verifies that Paint flood-fills an enclosed area.
func TestPaintFill(t *testing.T) {
	ScreenMode(13)
	// Draw a rectangle border from (5,5) to (9,9) with color 1 using boxMode "B"
	DrawLine(5, 5, 9, 9, 1, "B")
	// Verify interior pixels are 0 before paint
	for y := 6; y <= 8; y++ {
		for x := 6; x <= 8; x++ {
			if CurrentScreen.Framebuffer[y][x] != 0 {
				t.Fatalf("before Paint: fb[%d][%d] = %d, want 0", y, x, CurrentScreen.Framebuffer[y][x])
			}
		}
	}
	// Paint interior starting at (7,7) with color 2, border color 1
	Paint(7, 7, 2, 1)
	// Verify interior is now filled with color 2
	for y := 6; y <= 8; y++ {
		for x := 6; x <= 8; x++ {
			if CurrentScreen.Framebuffer[y][x] != 2 {
				t.Errorf("after Paint: fb[%d][%d] = %d, want 2", y, x, CurrentScreen.Framebuffer[y][x])
			}
		}
	}
}

// TestCircleFullNoError verifies that Circle does not panic.
func TestCircleFullNoError(t *testing.T) {
	ScreenMode(13)
	// Must not panic
	Circle(50, 50, 20, 1, 0, 0, 1)
}

// TestViewPort verifies viewport clipping: pixels outside viewport are not drawn,
// pixels inside viewport are drawn.
func TestViewPort(t *testing.T) {
	ScreenMode(13)
	// Set viewport (5,5)-(15,15), no fill or border
	ViewPort(5, 5, 15, 15, -1, -1)

	// Pset outside viewport: (3,3) should be clipped
	Pset(3, 3, 1)
	if CurrentScreen.Framebuffer[3][3] != 0 {
		t.Errorf("ViewPort: pset at (3,3) outside viewport was not clipped; fb[3][3] = %d", CurrentScreen.Framebuffer[3][3])
	}

	// Pset inside viewport: (10,10) should succeed
	Pset(10, 10, 1)
	if CurrentScreen.Framebuffer[10][10] != 1 {
		t.Errorf("ViewPort: pset at (10,10) inside viewport not drawn; fb[10][10] = %d", CurrentScreen.Framebuffer[10][10])
	}
}

// TestViewPrint verifies that ViewPrint sets TextTop and TextBottom correctly.
func TestViewPrint(t *testing.T) {
	ScreenMode(13)
	ViewPrint(3, 22)
	if CurrentScreen.TextTop != 3 {
		t.Errorf("ViewPrint: TextTop = %d, want 3", CurrentScreen.TextTop)
	}
	if CurrentScreen.TextBottom != 22 {
		t.Errorf("ViewPrint: TextBottom = %d, want 22", CurrentScreen.TextBottom)
	}
}

// TestDrawGML verifies that Draw "U10" moves the cursor up by 10 pixels.
func TestDrawGML(t *testing.T) {
	ScreenMode(13)
	// Position cursor at center
	CurrentScreen.CursorX = 50
	CurrentScreen.CursorY = 50
	CurrentScreen.DrawColor = 1
	CurrentScreen.DrawScale = 1
	CurrentScreen.DrawAngle = 0

	Draw("U10")

	// CursorY should have decreased by 10 (up = y decreases)
	if CurrentScreen.CursorY != 40 {
		t.Errorf("Draw U10: CursorY = %d, want 40", CurrentScreen.CursorY)
	}
	if CurrentScreen.CursorX != 50 {
		t.Errorf("Draw U10: CursorX = %d, want 50", CurrentScreen.CursorX)
	}
}

// ---------------------------------------------------------------------------
// DrawLine — BF (filled rectangle)
// ---------------------------------------------------------------------------

func TestDrawLineFilledRect(t *testing.T) {
	ScreenMode(13)
	DrawLine(5, 5, 15, 15, 3, "BF")
	// Interior should be filled with color 3.
	if CurrentScreen.Framebuffer[10][10] != 3 {
		t.Errorf("DrawLine BF: fb[10][10] = %d, want 3", CurrentScreen.Framebuffer[10][10])
	}
}

func TestDrawLineFilledRectReversed(t *testing.T) {
	// Reversed coordinates (y2 < y1, x2 < x1) should still fill.
	ScreenMode(13)
	DrawLine(15, 15, 5, 5, 2, "BF")
	if CurrentScreen.Framebuffer[10][10] != 2 {
		t.Errorf("DrawLine BF reversed: fb[10][10] = %d, want 2", CurrentScreen.Framebuffer[10][10])
	}
}

// ---------------------------------------------------------------------------
// Circle — parametric ellipse and arc cases
// ---------------------------------------------------------------------------

func TestCircleParametricEllipse(t *testing.T) {
	// aspect != 1.0 → parametric ellipse path.
	ScreenMode(13)
	Circle(50, 50, 20, 1, 0, 0, 0.5) // wide ellipse, should not panic
}

func TestCircleArc(t *testing.T) {
	// start != 0, end != 0 → arc path.
	ScreenMode(13)
	Circle(50, 50, 20, 1, 0, 1.5707963, 1.0) // quarter arc, should not panic
}

func TestCircleArcWrapped(t *testing.T) {
	// end < start → wrapping (endAngle += 2*pi).
	ScreenMode(13)
	Circle(50, 50, 15, 2, 5.0, 1.0, 1.0) // end < start, should wrap
}

func TestCircleSmallRadius(t *testing.T) {
	// Very small radius to exercise steps < 1 path.
	ScreenMode(13)
	Circle(50, 50, 0.1, 1, 0, 0, 2.0) // tiny radius with aspect != 1
}

// ---------------------------------------------------------------------------
// ViewPort — fill and border branches
// ---------------------------------------------------------------------------

func TestViewPortWithFillAndBorder(t *testing.T) {
	ScreenMode(13)
	// fillColor >= 0 and borderColor >= 0 to exercise both branches.
	ViewPort(10, 10, 20, 20, 5, 7)
	// Interior should be filled with color 5.
	if CurrentScreen.Framebuffer[15][15] != 5 {
		t.Errorf("ViewPort fill: fb[15][15] = %d, want 5", CurrentScreen.Framebuffer[15][15])
	}
}

func TestViewPortFillOutOfBounds(t *testing.T) {
	// Viewport that partially extends outside framebuffer.
	ScreenMode(13)
	ViewPort(-5, -5, 5, 5, 3, -1) // negative coords → out-of-bounds rows/cols
}

func TestViewPortNilFramebuffer(t *testing.T) {
	// When Framebuffer is nil, fillColor branch skips filling.
	ScreenMode(0) // mode 0 has no framebuffer
	ViewPort(0, 0, 10, 10, 1, -1) // should not panic
}

// ---------------------------------------------------------------------------
// isDrawMovementCmd — all movement commands
// ---------------------------------------------------------------------------

func TestIsDrawMovementCmd(t *testing.T) {
	valid := []byte{'U', 'D', 'L', 'R', 'E', 'F', 'G', 'H', 'M',
		'u', 'd', 'l', 'r', 'e', 'f', 'g', 'h', 'm'}
	for _, ch := range valid {
		if !isDrawMovementCmd(ch) {
			t.Errorf("isDrawMovementCmd(%c) = false, want true", ch)
		}
	}
	// Non-movement chars should return false.
	for _, ch := range []byte{'C', 'S', 'A', 'X', 'Z', '1', ' '} {
		if isDrawMovementCmd(ch) {
			t.Errorf("isDrawMovementCmd(%c) = true, want false", ch)
		}
	}
}

// ---------------------------------------------------------------------------
// Draw — additional GML commands
// ---------------------------------------------------------------------------

func TestDrawDirectionsAll(t *testing.T) {
	ScreenMode(13)
	commands := []struct {
		cmd     string
		expectX int
		expectY int
		name    string
	}{
		{"D10", 50, 60, "Down"},
		{"L10", 40, 50, "Left"},
		{"R10", 60, 50, "Right"},
		{"E10", 60, 40, "Up-right"},
		{"F10", 60, 60, "Down-right"},
		{"G10", 40, 60, "Down-left"},
		{"H10", 40, 40, "Up-left"},
	}
	for _, tc := range commands {
		t.Run(tc.name, func(t *testing.T) {
			ScreenMode(13)
			CurrentScreen.CursorX = 50
			CurrentScreen.CursorY = 50
			CurrentScreen.DrawAngle = 0
			CurrentScreen.DrawScale = 1
			CurrentScreen.DrawColor = 1
			Draw(tc.cmd)
			if CurrentScreen.CursorX != tc.expectX || CurrentScreen.CursorY != tc.expectY {
				t.Errorf("Draw %s: cursor=(%d,%d), want (%d,%d)",
					tc.cmd, CurrentScreen.CursorX, CurrentScreen.CursorY, tc.expectX, tc.expectY)
			}
		})
	}
}

func TestDrawColorAndScale(t *testing.T) {
	ScreenMode(13)
	CurrentScreen.CursorX = 50
	CurrentScreen.CursorY = 50
	CurrentScreen.DrawAngle = 0
	CurrentScreen.DrawScale = 1
	// C command sets color, S command sets scale.
	Draw("C3 S2 U5")
	if CurrentScreen.DrawColor != 3 {
		t.Errorf("Draw C3: DrawColor = %d, want 3", CurrentScreen.DrawColor)
	}
	if CurrentScreen.DrawScale != 2 {
		t.Errorf("Draw S2: DrawScale = %d, want 2", CurrentScreen.DrawScale)
	}
	// U5 with scale 2 = 10 pixels up.
	if CurrentScreen.CursorY != 40 {
		t.Errorf("Draw S2 U5: CursorY = %d, want 40", CurrentScreen.CursorY)
	}
}

func TestDrawAngle(t *testing.T) {
	ScreenMode(13)
	CurrentScreen.CursorX = 50
	CurrentScreen.CursorY = 50
	CurrentScreen.DrawScale = 1
	// A1 = 90° rotation, so U becomes R.
	Draw("A1 U10")
	if CurrentScreen.CursorX != 60 {
		t.Errorf("Draw A1 U10: CursorX = %d, want 60", CurrentScreen.CursorX)
	}
}

func TestDrawMoveAbsolute(t *testing.T) {
	ScreenMode(13)
	CurrentScreen.CursorX = 10
	CurrentScreen.CursorY = 10
	CurrentScreen.DrawAngle = 0
	CurrentScreen.DrawScale = 1
	CurrentScreen.DrawColor = 1
	Draw("M30,40")
	if CurrentScreen.CursorX != 30 || CurrentScreen.CursorY != 40 {
		t.Errorf("Draw M30,40: cursor=(%d,%d), want (30,40)",
			CurrentScreen.CursorX, CurrentScreen.CursorY)
	}
}

func TestDrawMoveRelative(t *testing.T) {
	ScreenMode(13)
	CurrentScreen.CursorX = 50
	CurrentScreen.CursorY = 50
	CurrentScreen.DrawAngle = 0
	CurrentScreen.DrawScale = 1
	CurrentScreen.DrawColor = 1
	Draw("M+10,-5")
	if CurrentScreen.CursorX != 60 || CurrentScreen.CursorY != 45 {
		t.Errorf("Draw M+10,-5: cursor=(%d,%d), want (60,45)",
			CurrentScreen.CursorX, CurrentScreen.CursorY)
	}
}

func TestDrawBPrefix(t *testing.T) {
	// B prefix: move without drawing.
	ScreenMode(13)
	CurrentScreen.CursorX = 50
	CurrentScreen.CursorY = 50
	CurrentScreen.DrawAngle = 0
	CurrentScreen.DrawScale = 1
	CurrentScreen.DrawColor = 1
	origFB := CurrentScreen.Framebuffer[50][50]
	Draw("BU10")
	// CursorY should move but no pixels should be set.
	if CurrentScreen.CursorY != 40 {
		t.Errorf("Draw BU10: CursorY = %d, want 40", CurrentScreen.CursorY)
	}
	_ = origFB
}

func TestDrawNPrefix(t *testing.T) {
	// N prefix: draw but return to start position.
	ScreenMode(13)
	CurrentScreen.CursorX = 50
	CurrentScreen.CursorY = 50
	CurrentScreen.DrawAngle = 0
	CurrentScreen.DrawScale = 1
	CurrentScreen.DrawColor = 1
	Draw("NU10")
	// Cursor should return to original position.
	if CurrentScreen.CursorX != 50 || CurrentScreen.CursorY != 50 {
		t.Errorf("Draw NU10: cursor=(%d,%d), want (50,50)",
			CurrentScreen.CursorX, CurrentScreen.CursorY)
	}
}

func TestDrawNilFramebuffer(t *testing.T) {
	// Draw with no screen mode set should not panic.
	ScreenMode(0)
	Draw("U10") // should return immediately
}

func TestDrawUnknownCommand(t *testing.T) {
	ScreenMode(13)
	CurrentScreen.CursorX = 50
	CurrentScreen.CursorY = 50
	Draw("Z5") // unknown command, skip
	// Cursor should be unchanged.
	if CurrentScreen.CursorX != 50 || CurrentScreen.CursorY != 50 {
		t.Errorf("Draw Z5 (unknown): cursor moved unexpectedly to (%d,%d)",
			CurrentScreen.CursorX, CurrentScreen.CursorY)
	}
}

func TestDrawWhitespaceAndSemicolon(t *testing.T) {
	ScreenMode(13)
	CurrentScreen.CursorX = 50
	CurrentScreen.CursorY = 50
	CurrentScreen.DrawAngle = 0
	CurrentScreen.DrawScale = 1
	CurrentScreen.DrawColor = 1
	// Whitespace and semicolons should be skipped.
	Draw(" ; U5 ; D5")
	// Net displacement is 0.
	if CurrentScreen.CursorY != 50 {
		t.Errorf("Draw with spaces/semicolons: CursorY = %d, want 50", CurrentScreen.CursorY)
	}
}

// ---------------------------------------------------------------------------
// rotateDirection — all 4 rotation steps
// ---------------------------------------------------------------------------

func TestRotateDirectionAllAngles(t *testing.T) {
	// U direction (0,-1) rotated 0,1,2,3 times.
	cases := []struct {
		angle int
		wantX int
		wantY int
	}{
		{0, 0, -1},  // no rotation: up
		{1, 1, 0},   // 90° CCW in screen: right
		{2, 0, 1},   // 180°: down
		{3, -1, 0},  // 270°: left
	}
	for _, tc := range cases {
		gotX, gotY := rotateDirection(0, -1, tc.angle)
		if gotX != tc.wantX || gotY != tc.wantY {
			t.Errorf("rotateDirection(0,-1,%d) = (%d,%d), want (%d,%d)",
				tc.angle, gotX, gotY, tc.wantX, tc.wantY)
		}
	}
}

func TestScreenMode(t *testing.T) {
	cases := []int{0, 1, 2, 7, 8, 9, 10, 11, 12, 13, 99}
	for _, mode := range cases {
		ScreenMode(mode)
	}
	ScreenMode(0)
	if CurrentScreen.Width != 80 {
		t.Errorf("ScreenMode(0) width = %d, want 80", CurrentScreen.Width)
	}
}

func TestCircleStub(t *testing.T) {
	Circle(100, 100, 50, 15, 0, 0, 1) // should not panic
}

func TestDrawLineStub(t *testing.T) {
	DrawLine(0, 0, 100, 100, 15, "B") // should not panic
}

func TestPsetStub(t *testing.T) {
	Pset(50, 50, 7) // should not panic
}

func TestPaintStub(t *testing.T) {
	Paint(50, 50, 4, 15) // should not panic
}

func TestDrawStub(t *testing.T) {
	Draw("U10 R10 D10 L10") // should not panic
}

func TestViewPortStub(t *testing.T) {
	ViewPort(0, 0, 640, 480, 0, 0) // should not panic
}

func TestViewPrintStub(t *testing.T) {
	ViewPrint(1, 24) // should not panic
}

// ---------------------------------------------------------------------------
// clampByte — table-driven
// ---------------------------------------------------------------------------

func TestClampByte(t *testing.T) {
	tests := []struct {
		name string
		v    float64
		want byte
	}{
		{"negative", -10, 0},
		{"zero", 0.0, 0},
		{"normal", 127.0, 127},
		{"boundary 255", 255.0, 255},
		{"over 255", 300.0, 255},
		{"fractional", 127.7, 127}, // byte truncates fractional part
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := clampByte(tt.v)
			if got != tt.want {
				t.Errorf("clampByte(%v) = %d, want %d", tt.v, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Pset — color clamping
// ---------------------------------------------------------------------------

func TestPsetColorClamping(t *testing.T) {
	ScreenMode(13)

	// Negative color should clamp to 0
	Pset(10, 10, -5)
	got := CurrentScreen.Framebuffer[10][10]
	if got != 0 {
		t.Errorf("Pset(10,10,-5): fb[10][10] = %d, want 0", got)
	}

	// Color > 255 should clamp to 255
	Pset(10, 10, 300)
	got = CurrentScreen.Framebuffer[10][10]
	if got != 255 {
		t.Errorf("Pset(10,10,300): fb[10][10] = %d, want 255", got)
	}
}

// ---------------------------------------------------------------------------
// Circle — verify actual pixel output
// ---------------------------------------------------------------------------

func TestCirclePixelPlotted(t *testing.T) {
	ScreenMode(13)
	Circle(50, 50, 10, 5, 0, 0, 1.0)

	fb := CurrentScreen.Framebuffer
	// The right-most point of the circle (50+10, 50) should be color 5
	if fb[50][60] != 5 {
		t.Errorf("Circle pixel at (60,50): fb[50][60] = %d, want 5", fb[50][60])
	}

	// Count pixels at roughly radius distance that are set to color 5
	count := 0
	for y := 0; y < 200; y++ {
		for x := 0; x < 320; x++ {
			if fb[y][x] == 5 {
				count++
			}
		}
	}
	if count == 0 {
		t.Error("Circle drew zero pixels with color 5")
	}
}

// ---------------------------------------------------------------------------
// Paint — edge cases
// ---------------------------------------------------------------------------

func TestPaintOnBorderColor(t *testing.T) {
	ScreenMode(13)
	// Draw a box with border color 1
	DrawLine(0, 0, 10, 10, 1, "B")

	// (5, 0) is ON the border (top edge)
	Paint(5, 0, 2, 1)

	// Interior pixel (5, 5) should still be 0 — paint started on border, so no fill
	if CurrentScreen.Framebuffer[5][5] != 0 {
		t.Errorf("PaintOnBorder: fb[5][5] = %d, want 0 (should not fill when starting on border)",
			CurrentScreen.Framebuffer[5][5])
	}
}

func TestPaintAlreadyFilled(t *testing.T) {
	ScreenMode(13)
	// Pre-set the start pixel to the fill color
	CurrentScreen.Framebuffer[50][50] = 2

	// Paint with fillColor=2 starting on a pixel that is already 2
	Paint(50, 50, 2, 1)

	// Neighboring pixels should remain 0 — no flood fill should occur
	if CurrentScreen.Framebuffer[50][51] != 0 {
		t.Errorf("PaintAlreadyFilled: fb[50][51] = %d, want 0 (should not flood fill)",
			CurrentScreen.Framebuffer[50][51])
	}
}

func TestPaintOutOfBounds(t *testing.T) {
	ScreenMode(13)
	// These must not panic
	Paint(-1, -1, 2, 1)
	Paint(999, 999, 2, 1)
}

// ---------------------------------------------------------------------------
// ViewPort — reversed coords and border-only
// ---------------------------------------------------------------------------

func TestViewPortReversedCoords(t *testing.T) {
	ScreenMode(13)
	// Reversed: x1>x2, y1>y2 — fill color 5, no border
	ViewPort(20, 20, 10, 10, 5, -1)

	// Interior pixel should be filled with color 5
	fb := CurrentScreen.Framebuffer
	if fb[15][15] != 5 {
		t.Errorf("ViewPortReversed: fb[15][15] = %d, want 5", fb[15][15])
	}
}

func TestViewPortBorderOnly(t *testing.T) {
	ScreenMode(13)
	// No fill (-1), border color 3
	ViewPort(10, 10, 20, 20, -1, 3)

	fb := CurrentScreen.Framebuffer
	// A pixel on the top border should be color 3
	if fb[10][15] != 3 {
		t.Errorf("ViewPortBorderOnly: fb[10][15] = %d, want 3", fb[10][15])
	}
	// Interior should remain 0 (no fill)
	if fb[15][15] != 0 {
		t.Errorf("ViewPortBorderOnly: fb[15][15] = %d, want 0 (no fill)", fb[15][15])
	}
}

// ---------------------------------------------------------------------------
// Draw — M with absolute coords to origin
// ---------------------------------------------------------------------------

func TestDrawMoveNegative(t *testing.T) {
	ScreenMode(13)
	CurrentScreen.CursorX = 50
	CurrentScreen.CursorY = 50
	CurrentScreen.DrawAngle = 0
	CurrentScreen.DrawScale = 1
	CurrentScreen.DrawColor = 1

	Draw("M0,0") // absolute move to origin
	if CurrentScreen.CursorX != 0 || CurrentScreen.CursorY != 0 {
		t.Errorf("Draw M0,0: cursor=(%d,%d), want (0,0)",
			CurrentScreen.CursorX, CurrentScreen.CursorY)
	}
}

// ---------------------------------------------------------------------------
// ScreenMode — FgColor / BgColor initialization
// ---------------------------------------------------------------------------

func TestScreenModeInitColors(t *testing.T) {
	// Mode 13: 256 colors, FgColor should be 255
	ScreenMode(13)
	if CurrentScreen.FgColor != 255 {
		t.Errorf("ScreenMode(13): FgColor = %d, want 255", CurrentScreen.FgColor)
	}
	if CurrentScreen.BgColor != 0 {
		t.Errorf("ScreenMode(13): BgColor = %d, want 0", CurrentScreen.BgColor)
	}

	// Mode 1: 4 colors, FgColor should be 3
	ScreenMode(1)
	if CurrentScreen.FgColor != 3 {
		t.Errorf("ScreenMode(1): FgColor = %d, want 3", CurrentScreen.FgColor)
	}
	if CurrentScreen.BgColor != 0 {
		t.Errorf("ScreenMode(1): BgColor = %d, want 0", CurrentScreen.BgColor)
	}
}
