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
