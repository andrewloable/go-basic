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

	// gfxWindowCreated is true once ebiten.RunGame has been called.
	// It stays true forever since RunGame can only be called once per process.
	gfxWindowCreated bool

	// gfxActive is true when a graphics SCREEN mode (>0) is active.
	// When false (text mode), the window shows black but RunGame keeps running.
	gfxActive bool

	// gfxTextMode is true when SCREEN 0 is being rendered in the Ebitengine window.
	// In this mode, the TextBuffer is rendered instead of the framebuffer.
	gfxTextMode bool

	// gfxExiting is set when the BASIC program finishes execution,
	// causing Update() to return ebiten.Termination.
	gfxExiting bool

	// gfxNewWidth/Height/Mode hold pending resize requests from the BASIC
	// goroutine. Update() applies them on the main thread.
	gfxNewWidth  int
	gfxNewHeight int
	gfxNewMode   int
	gfxResizeReq bool

	// gfxKeyBuf is a channel-based key buffer fed by Ebitengine's Update().
	// The BASIC goroutine reads from it via InkeyFromGraphics().
	gfxKeyBuf = make(chan string, 64)

	// gfxStartReq is used by OpenGraphicsWindow to signal the main thread
	// (which runs RunMain) to start the Ebitengine window the first time.
	gfxStartReq = make(chan gfxStartRequest, 1)

	// mainDone is closed when the BASIC program finishes execution.
	mainDone = make(chan struct{})
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
	if gfxExiting {
		return ebiten.Termination
	}

	// Apply pending resize from the BASIC goroutine.
	if gfxResizeReq {
		g.width = gfxNewWidth
		g.height = gfxNewHeight
		ebiten.SetWindowSize(g.width*2, g.height*2)
		ebiten.SetWindowTitle(fmt.Sprintf("BASIC Graphics - SCREEN %d", gfxNewMode))
		gfxResizeReq = false
	}

	// Capture keyboard input when in an active graphics or text mode.
	if !gfxActive {
		return nil
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
	if !gfxActive {
		return // inactive — show black
	}

	if gfxTextMode {
		g.drawTextMode(screen)
		return
	}

	gfxMu.Lock()
	s := CurrentScreen
	fb := s.Framebuffer
	pal := s.Palette
	gfxMu.Unlock()

	if fb == nil || len(pal) == 0 {
		return
	}

	fbH := len(fb)
	if fbH == 0 {
		return
	}
	fbW := len(fb[0])

	// Use the screen's actual bounds to determine pixel buffer size.
	// This avoids panics when Layout() and the framebuffer are temporarily
	// out of sync during mode switches (e.g. SCREEN 9 → SCREEN 0 → SCREEN 9).
	bounds := screen.Bounds()
	screenW := bounds.Dx()
	screenH := bounds.Dy()
	if screenW <= 0 || screenH <= 0 {
		return
	}

	pixels := make([]byte, screenW*screenH*4)

	// Copy framebuffer into the pixel buffer, clamping to the smaller dimensions.
	copyW := fbW
	if copyW > screenW {
		copyW = screenW
	}
	copyH := fbH
	if copyH > screenH {
		copyH = screenH
	}

	for y := 0; y < copyH; y++ {
		row := fb[y]
		for x := 0; x < copyW; x++ {
			idx := int(row[x])
			if idx >= len(pal) {
				idx = 0
			}
			c := pal[idx]
			off := (y*screenW + x) * 4
			pixels[off] = c.R
			pixels[off+1] = c.G
			pixels[off+2] = c.B
			pixels[off+3] = 0xFF
		}
	}
	screen.WritePixels(pixels)
}

// drawTextMode renders the TextBuffer contents using the CP437 bitmap font.
func (g *graphicsWindow) drawTextMode(screen *ebiten.Image) {
	tb := TextBuf
	if tb == nil {
		return
	}

	// Use the standard 16-color CGA/EGA palette.
	pal := defaultPalette(16)

	const charW = 8
	const charH = 16

	tb.mu.Lock()
	cols := tb.Cols
	rows := tb.Rows
	// Take a snapshot of cells to minimize lock time.
	cellsCopy := make([][]TextCell, rows)
	for r := 0; r < rows; r++ {
		cellsCopy[r] = make([]TextCell, cols)
		copy(cellsCopy[r], tb.Cells[r])
	}
	tb.mu.Unlock()

	// Use the screen's actual bounds to size the pixel buffer.
	bounds := screen.Bounds()
	screenW := bounds.Dx()
	screenH := bounds.Dy()
	if screenW <= 0 || screenH <= 0 {
		return
	}

	pixW := screenW
	pixH := screenH
	pixels := make([]byte, pixW*pixH*4)

	// Render cells that fit within the screen bounds.
	for r := 0; r < rows; r++ {
		if r*charH >= pixH {
			break
		}
		for c := 0; c < cols; c++ {
			if c*charW >= pixW {
				break
			}
			cell := cellsCopy[r][c]
			fg := pal[cell.Fg%16]
			bg := pal[cell.Bg%16]
			DrawCharToPixels(pixels, pixW, c*charW, r*charH, cell.Char, fg, bg)
		}
	}

	screen.WritePixels(pixels)
}

func (g *graphicsWindow) Layout(outsideWidth, outsideHeight int) (int, int) {
	if gfxTextMode {
		return 640, 400 // 80×8 cols, 25×16 rows
	}
	return g.width, g.height
}

// gfxStartRequest carries the parameters for starting a graphics window.
type gfxStartRequest struct {
	width, height, mode int
	ready               chan struct{}
}

// startGraphicsWindow launches the Ebitengine window on the main OS thread.
// Called once and blocks for the entire program lifetime.
func startGraphicsWindow(req gfxStartRequest) {
	ebiten.SetWindowSize(req.width*2, req.height*2)
	ebiten.SetWindowTitle(fmt.Sprintf("BASIC Graphics - SCREEN %d", req.mode))
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetScreenFilterEnabled(false) // nearest-neighbor for pixel art look

	game := &graphicsWindow{width: req.width, height: req.height}

	gfxWindowCreated = true
	gfxActive = true
	close(req.ready)

	if err := ebiten.RunGame(game); err != nil {
		// Termination is normal when gfxExiting is set.
	}
}

// RunMain runs the BASIC program in a goroutine while keeping the main OS
// thread available for Ebitengine window creation (required by macOS).
// The generated main() should call rt.RunMain(basicMain).
func RunMain(basicProgram func()) {
	if GraphicsHeadless {
		// In headless mode, no window will ever be created. Run directly.
		basicProgram()
		return
	}

	go func() {
		basicProgram()
		// Clean up text writer if active before exiting.
		UninstallTextWriter()
		// Signal the Ebitengine window to exit and the main thread to stop.
		gfxExiting = true
		close(mainDone)
	}()

	// Main thread: wait for graphics window request or program exit.
	select {
	case req := <-gfxStartReq:
		// Start the Ebitengine window (blocks for program lifetime).
		startGraphicsWindow(req)
	case <-mainDone:
		// Program finished without ever using graphics.
		return
	}
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

// OpenGraphicsWindow activates graphics rendering for the current screen mode.
// On the first call, it starts the Ebitengine window on the main thread.
// On subsequent calls, it resizes the existing window.
// No-op in headless mode (tests, CI) or when GOBASIC_HEADLESS=1 is set.
func OpenGraphicsWindow() {
	if GraphicsHeadless {
		return
	}
	s := CurrentScreen
	if s.Mode == 0 || s.Framebuffer == nil {
		return
	}

	if gfxWindowCreated {
		// Window already exists — resize and reactivate.
		gfxNewWidth = s.Width
		gfxNewHeight = s.Height
		gfxNewMode = s.Mode
		gfxResizeReq = true
		gfxActive = true
		return
	}

	// First time: send request to main thread to start RunGame.
	ready := make(chan struct{})
	gfxStartReq <- gfxStartRequest{
		width:  s.Width,
		height: s.Height,
		mode:   s.Mode,
		ready:  ready,
	}
	<-ready // Wait for the window to be ready before returning.
}

// CloseGraphicsWindow deactivates graphics rendering (for SCREEN 0 / text mode).
// The Ebitengine window stays open but shows black. RunGame is NOT terminated.
func CloseGraphicsWindow() {
	gfxActive = false
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

// PaletteRemap remaps a palette index using an EGA color number (0-63).
// This is the standard QBasic PALETTE statement: PALETTE index, egaColor.
// EGA colors encode 2-bit RGB values in the format: xxRGBrgb
// where R,G,B are high bits and r,g,b are low bits.
func PaletteRemap(index, egaColor int) {
	if egaColor < 0 || egaColor > 63 {
		return
	}
	r, g, b := egaToRGB(egaColor)
	SetPaletteEntry(index, r, g, b)
}

// egaToRGB converts a 6-bit EGA color number (0-63) to 8-bit RGB.
// EGA color encoding: bits 5-3 = high RGB, bits 2-0 = low rgb
// Format: bit5=R' bit4=G' bit3=B' bit2=r bit1=g bit0=b
// Each channel = (high_bit * 2 + low_bit) * 85 → 0, 85, 170, 255
func egaToRGB(c int) (r, g, b uint8) {
	// Extract 2-bit color channels from the EGA color number.
	rBit := ((c >> 5) & 1) << 1 | ((c >> 2) & 1) // R' and r
	gBit := ((c >> 4) & 1) << 1 | ((c >> 1) & 1) // G' and g
	bBit := ((c >> 3) & 1) << 1 | (c & 1)         // B' and b
	return uint8(rBit * 85), uint8(gBit * 85), uint8(bBit * 85)
}

// IsGraphicsMode returns true when a graphics SCREEN mode is active and the
// Ebitengine window should handle keyboard input instead of the terminal.
func IsGraphicsMode() bool {
	return gfxActive || gfxTextMode
}

// IsTextMode returns true when SCREEN 0 is being rendered in the Ebitengine window.
func IsTextMode() bool {
	return gfxTextMode
}

// IsGraphicsWindowOpen returns true when the Ebitengine window has been created,
// regardless of whether we're in a graphics SCREEN mode or text mode (SCREEN 0).
// When the window exists, keyboard input goes to Ebitengine, so INPUT/LINE INPUT
// must read from the Ebitengine key buffer instead of stdin.
func IsGraphicsWindowOpen() bool {
	return gfxWindowCreated
}

// GraphicsReadLine reads a full line from the Ebitengine key buffer (blocking).
// Used by INPUT/LINE INPUT when the graphics window is open and stdin is unavailable.
// Supports basic line editing: printable chars, backspace, enter.
func GraphicsReadLine() string {
	var buf []byte
	for {
		select {
		case k := <-gfxKeyBuf:
			if k == "\r" || k == "\n" {
				fmt.Println() // echo newline
				return string(buf)
			}
			if k == "\x08" || k == "\x7f" {
				// Backspace
				if len(buf) > 0 {
					buf = buf[:len(buf)-1]
					if TextBuf != nil && IsTextWriterActive() {
						// Text writer active: move cursor back, overwrite with space, move back.
						TextBuf.PutChar('\b')
						TextBuf.PutChar(' ')
						TextBuf.PutChar('\b')
					} else {
						fmt.Print("\x08 \x08") // erase character on terminal
					}
				}
				continue
			}
			// Skip special keys (two-byte scan codes starting with \x00 or \x1b)
			if len(k) > 0 && (k[0] == 0x00 || k[0] == 0x1b) {
				continue
			}
			// Printable character
			buf = append(buf, k...)
			fmt.Print(k) // echo (goes to text buffer via pipe when text writer is active)
		}
	}
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
