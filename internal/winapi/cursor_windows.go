//go:build windows

package winapi

import (
	"fmt"
	"math"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/jarvis/mouse-mover/internal/engine"
)

var (
	user32     = windows.NewLazySystemDLL("user32.dll")
	getCursorP = user32.NewProc("GetCursorPos")
	setCursorP = user32.NewProc("SetCursorPos")
	sendInputP = user32.NewProc("SendInput")
	getSysMetr = user32.NewProc("GetSystemMetrics")
)

// Win32 constants.
const (
	smXVirtualScreen  = 76
	smYVirtualScreen  = 77
	smCXVirtualScreen = 78
	smCYVirtualScreen = 79

	inputMouse             = 0
	mouseeventfMove        = 0x0001
	mouseeventfAbsolute    = 0x8000
	mouseeventfVirtualDesk = 0x4000
)

// _POINT matches the Win32 POINT struct.
type _POINT struct {
	X, Y int32
}

// _MOUSEINPUT matches the MOUSEINPUT part of INPUT.
type _MOUSEINPUT struct {
	Dx          int32
	Dy          int32
	MouseData   uint32
	DwFlags     uint32
	Time        uint32
	DwExtraInfo uintptr
}

// _INPUT matches the Win32 INPUT struct (for type = INPUT_MOUSE).
type _INPUT struct {
	Type uint32
	_    [4]byte // alignment padding on 64-bit
	Mi   _MOUSEINPUT
}

// cursorController implements engine.CursorController using the Win32 API.
type cursorController struct{}

// NewCursorController returns a real Windows CursorController.
func NewCursorController() engine.CursorController {
	return &cursorController{}
}

func (c *cursorController) GetPos() (engine.Point, error) {
	var pt _POINT
	r, _, err := getCursorP.Call(uintptr(unsafe.Pointer(&pt)))
	if r == 0 {
		return engine.Point{}, fmt.Errorf("GetCursorPos: %w", err)
	}
	return engine.Point{X: pt.X, Y: pt.Y}, nil
}

func (c *cursorController) MoveTo(p engine.Point) error {
	// Use SendInput with absolute virtual-desktop coordinates.
	screen, err := c.VirtualScreen()
	if err != nil {
		return err
	}

	w := screen.Width()
	h := screen.Height()
	if w <= 0 || h <= 0 {
		return fmt.Errorf("invalid virtual screen size %dx%d", w, h)
	}

	// Normalise to 0..65535 range relative to the virtual screen.
	nx := int32(math.Round(float64(p.X-screen.Left) * 65535.0 / float64(w-1)))
	ny := int32(math.Round(float64(p.Y-screen.Top) * 65535.0 / float64(h-1)))

	inp := _INPUT{
		Type: inputMouse,
		Mi: _MOUSEINPUT{
			Dx:      nx,
			Dy:      ny,
			DwFlags: mouseeventfMove | mouseeventfAbsolute | mouseeventfVirtualDesk,
		},
	}

	r, _, err := sendInputP.Call(
		1,
		uintptr(unsafe.Pointer(&inp)),
		uintptr(unsafe.Sizeof(inp)),
	)
	if r == 0 {
		// Fallback to SetCursorPos.
		r2, _, err2 := setCursorP.Call(uintptr(p.X), uintptr(p.Y))
		if r2 == 0 {
			return fmt.Errorf("SendInput failed (%w) and SetCursorPos fallback also failed: %v", err, err2)
		}
	}
	return nil
}

func (c *cursorController) VirtualScreen() (engine.Rect, error) {
	x, _, _ := getSysMetr.Call(smXVirtualScreen)
	y, _, _ := getSysMetr.Call(smYVirtualScreen)
	cx, _, _ := getSysMetr.Call(smCXVirtualScreen)
	cy, _, _ := getSysMetr.Call(smCYVirtualScreen)

	if cx == 0 || cy == 0 {
		return engine.Rect{}, fmt.Errorf("GetSystemMetrics returned zero virtual screen size")
	}
	return engine.Rect{
		Left:   int32(x),
		Top:    int32(y),
		Right:  int32(x) + int32(cx),
		Bottom: int32(y) + int32(cy),
	}, nil
}
