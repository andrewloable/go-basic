package runtime

// Screen represents the BASIC SCREEN mode state.
type Screen struct {
	Mode   int
	Width  int
	Height int
	Colors int
}

// CurrentScreen holds the active screen mode. Default is text mode 0.
var CurrentScreen = &Screen{0, 80, 25, 16}

// ScreenMode sets the display mode. Stub implementation for compatibility.
func ScreenMode(mode int) {
	switch mode {
	case 0:
		CurrentScreen = &Screen{0, 80, 25, 16}
	case 1:
		CurrentScreen = &Screen{1, 320, 200, 4}
	case 2:
		CurrentScreen = &Screen{2, 640, 200, 2}
	case 7:
		CurrentScreen = &Screen{7, 320, 200, 16}
	case 8:
		CurrentScreen = &Screen{8, 640, 200, 16}
	case 9:
		CurrentScreen = &Screen{9, 640, 350, 16}
	case 10:
		CurrentScreen = &Screen{10, 640, 350, 4}
	case 11:
		CurrentScreen = &Screen{11, 640, 480, 2}
	case 12:
		CurrentScreen = &Screen{12, 640, 480, 16}
	case 13:
		CurrentScreen = &Screen{13, 320, 200, 256}
	}
}

// Circle draws an ellipse. Stub implementation for compatibility.
func Circle(x, y, radius, color, start, end, aspect float64) {}

// DrawLine draws a line or box. boxMode is "", "B", or "BF".
// Named DrawLine to avoid conflict with Go's bufio.Scanner line reading.
func DrawLine(x1, y1, x2, y2, color float64, boxMode string) {}

// Pset sets a pixel. Stub implementation for compatibility.
func Pset(x, y, color float64) {}

// Paint flood-fills an area. Stub implementation for compatibility.
func Paint(x, y, fillColor, borderColor float64) {}

// Draw executes a GML (Graphics Macro Language) command string.
func Draw(cmd string) {}

// ViewPort sets the graphics viewport. Stub implementation for compatibility.
func ViewPort(x1, y1, x2, y2, fillColor, borderColor float64) {}

// ViewPrint sets the text viewport. Stub implementation for compatibility.
func ViewPrint(top, bottom float64) {}
