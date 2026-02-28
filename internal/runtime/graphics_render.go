package runtime

import (
	"fmt"
	"image/color"
	"os"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// graphicsWindow implements ebiten.Game to render the BASIC framebuffer.
type graphicsWindow struct {
	width  int
	height int
}

var (
	// gfxMu protects framebuffer reads/writes between the BASIC goroutine
	// and the Ebitengine render goroutine.
	gfxMu sync.Mutex

	// gfxRunning tracks whether the Ebitengine window is currently active.
	gfxRunning bool

	// gfxReady is closed once the Ebitengine window loop has started, so
	// the BASIC program can wait for the window to be ready before drawing.
	gfxReady chan struct{}

	// gfxQuit signals the Ebitengine loop to exit.
	gfxQuit bool

	// gfxKeyBuf is a channel-based key buffer fed by Ebitengine's Update().
	// The BASIC goroutine reads from it via InkeyFromGraphics().
	gfxKeyBuf = make(chan string, 64)
)

// ebitenSpecialKeys maps Ebitengine key codes to QBasic two-byte scan codes.
var ebitenSpecialKeys = map[ebiten.Key]string{
	ebiten.KeyUp:       "\x00\x48", // scan 72
	ebiten.KeyDown:     "\x00\x50", // scan 80
	ebiten.KeyRight:    "\x00\x4D", // scan 77
	ebiten.KeyLeft:     "\x00\x4B", // scan 75
	ebiten.KeyHome:     "\x00\x47", // scan 71
	ebiten.KeyEnd:      "\x00\x4F", // scan 79
	ebiten.KeyPageUp:   "\x00\x49", // scan 73
	ebiten.KeyPageDown: "\x00\x51", // scan 81
	ebiten.KeyInsert:   "\x00\x52", // scan 82
	ebiten.KeyDelete:   "\x00\x53", // scan 83
	ebiten.KeyF1:       "\x00\x3B", // scan 59
	ebiten.KeyF2:       "\x00\x3C", // scan 60
	ebiten.KeyF3:       "\x00\x3D", // scan 61
	ebiten.KeyF4:       "\x00\x3E", // scan 62
	ebiten.KeyF5:       "\x00\x3F", // scan 63
	ebiten.KeyF6:       "\x00\x40", // scan 64
	ebiten.KeyF7:       "\x00\x41", // scan 65
	ebiten.KeyF8:       "\x00\x42", // scan 66
	ebiten.KeyF9:       "\x00\x43", // scan 67
	ebiten.KeyF10:      "\x00\x44", // scan 68
}

func (g *graphicsWindow) Update() error {
	if gfxQuit {
		return ebiten.Termination
	}

	// Capture printable characters typed this frame.
	runes := ebiten.AppendInputChars(nil)
	for _, r := range runes {
		select {
		case gfxKeyBuf <- string(r):
		default: // buffer full, drop
		}
	}

	// Capture special keys (arrows, function keys, etc.).
	// Only fire on the frame the key is first pressed (duration == 1 tick).
	for key, scanCode := range ebitenSpecialKeys {
		if inpututil.IsKeyJustPressed(key) {
			select {
			case gfxKeyBuf <- scanCode:
			default:
			}
		}
	}

	// Map Enter and Backspace to their BASIC equivalents.
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter) {
		select {
		case gfxKeyBuf <- "\r":
		default:
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		select {
		case gfxKeyBuf <- "\x08":
		default:
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		select {
		case gfxKeyBuf <- "\x1b":
		default:
		}
	}

	return nil
}

func (g *graphicsWindow) Draw(screen *ebiten.Image) {
	gfxMu.Lock()
	s := CurrentScreen
	fb := s.Framebuffer
	pal := s.Palette
	gfxMu.Unlock()

	if fb == nil || len(pal) == 0 {
		return
	}

	h := len(fb)
	if h == 0 {
		return
	}
	w := len(fb[0])

	pixels := make([]byte, w*h*4)
	for y := 0; y < h; y++ {
		row := fb[y]
		for x := 0; x < w; x++ {
			idx := int(row[x])
			if idx >= len(pal) {
				idx = 0
			}
			c := pal[idx]
			off := (y*w + x) * 4
			pixels[off] = c.R
			pixels[off+1] = c.G
			pixels[off+2] = c.B
			pixels[off+3] = 0xFF
		}
	}
	screen.WritePixels(pixels)
}

func (g *graphicsWindow) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.width, g.height
}

// StartGraphicsWindow launches the Ebitengine window for the given screen mode.
// It blocks, so it must be called from a goroutine. The ready channel is closed
// once the window event loop starts.
func StartGraphicsWindow(width, height, mode int, ready chan struct{}) {
	ebiten.SetWindowSize(width*2, height*2)
	ebiten.SetWindowTitle(fmt.Sprintf("BASIC Graphics - SCREEN %d", mode))
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetScreenFilterEnabled(false) // nearest-neighbor for pixel art look

	game := &graphicsWindow{width: width, height: height}

	gfxRunning = true
	close(ready)

	if err := ebiten.RunGame(game); err != nil {
		// Termination is normal when gfxQuit is set.
	}
	gfxRunning = false
}

// GraphicsHeadless disables window creation. Set to true in tests or CI.
var GraphicsHeadless bool

func init() {
	// Detect test/CI environments and disable window creation.
	if os.Getenv("GOBASIC_HEADLESS") == "1" {
		GraphicsHeadless = true
		return
	}
	// Auto-detect go test: the test binary sets -test.v or -test.run flags.
	for _, arg := range os.Args {
		if len(arg) > 5 && arg[:6] == "-test." {
			GraphicsHeadless = true
			return
		}
	}
}

// OpenGraphicsWindow opens the Ebitengine window for the current screen mode.
// Safe to call multiple times — closes the previous window first.
// No-op in headless mode (tests, CI) or when GOBASIC_HEADLESS=1 is set.
func OpenGraphicsWindow() {
	if GraphicsHeadless {
		return
	}
	s := CurrentScreen
	if s.Mode == 0 || s.Framebuffer == nil {
		return
	}
	if gfxRunning {
		// Already running with correct size — nothing to do.
		return
	}

	gfxQuit = false
	gfxReady = make(chan struct{})
	go StartGraphicsWindow(s.Width, s.Height, s.Mode, gfxReady)
	<-gfxReady
}

// CloseGraphicsWindow signals the Ebitengine window to close.
func CloseGraphicsWindow() {
	if gfxRunning {
		gfxQuit = true
	}
}

// LockFramebuffer acquires the framebuffer mutex. Call before bulk drawing.
func LockFramebuffer() {
	gfxMu.Lock()
}

// UnlockFramebuffer releases the framebuffer mutex.
func UnlockFramebuffer() {
	gfxMu.Unlock()
}

// SetPaletteEntry remaps a palette index to a custom RGB color (PALETTE statement).
func SetPaletteEntry(index int, r, g, b uint8) {
	gfxMu.Lock()
	defer gfxMu.Unlock()
	s := CurrentScreen
	if index >= 0 && index < len(s.Palette) {
		s.Palette[index] = Color{r, g, b}
	}
}

// IsGraphicsMode returns true when the Ebitengine window is running and should
// handle keyboard input instead of the terminal.
func IsGraphicsMode() bool {
	return gfxRunning
}

// InkeyFromGraphics reads a key from the Ebitengine key buffer (non-blocking).
// Returns "" if no key is available.
func InkeyFromGraphics() string {
	select {
	case k := <-gfxKeyBuf:
		return k
	default:
		return ""
	}
}

// GetPixelColor returns the color.RGBA for a framebuffer pixel (used by Point).
func GetPixelColor(x, y int) color.RGBA {
	gfxMu.Lock()
	defer gfxMu.Unlock()
	s := CurrentScreen
	fb := s.Framebuffer
	if fb == nil {
		return color.RGBA{}
	}
	if y < 0 || y >= len(fb) || x < 0 || x >= len(fb[y]) {
		return color.RGBA{}
	}
	idx := int(fb[y][x])
	if idx >= len(s.Palette) {
		return color.RGBA{}
	}
	c := s.Palette[idx]
	return color.RGBA{c.R, c.G, c.B, 0xFF}
}
