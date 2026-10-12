package engine

// Point is a cursor coordinate in the Windows virtual-screen space.
type Point struct {
	X, Y int32
}

// Rect describes a rectangle in virtual-screen coordinates. Right and Bottom
// are exclusive (half-open), matching how Windows virtual-screen metrics are
// consumed here: a point is inside when Left <= X < Right and Top <= Y < Bottom.
type Rect struct {
	Left, Top, Right, Bottom int32
}

// Width returns the rectangle width.
func (r Rect) Width() int32 { return r.Right - r.Left }

// Height returns the rectangle height.
func (r Rect) Height() int32 { return r.Bottom - r.Top }

// Clamp returns p constrained to lie within r. Because Right/Bottom are
// exclusive, the maximum valid coordinate is Right-1 / Bottom-1.
func (r Rect) Clamp(p Point) Point {
	x, y := p.X, p.Y
	if r.Width() > 0 {
		if x < r.Left {
			x = r.Left
		}
		if x > r.Right-1 {
			x = r.Right - 1
		}
	}
	if r.Height() > 0 {
		if y < r.Top {
			y = r.Top
		}
		if y > r.Bottom-1 {
			y = r.Bottom - 1
		}
	}
	return Point{X: x, Y: y}
}

// CursorController abstracts the platform cursor operations the engine needs.
// The real implementation lives in internal/winapi (Windows only); tests inject
// a fake. Defining the interface here keeps the dependency one-directional
// (winapi imports engine, never the reverse).
type CursorController interface {
	// GetPos returns the current cursor position.
	GetPos() (Point, error)
	// MoveTo places the cursor at an absolute virtual-screen coordinate.
	MoveTo(p Point) error
	// VirtualScreen returns the current virtual-screen bounds (half-open).
	VirtualScreen() (Rect, error)
}

// Logger is the minimal logging surface the engine uses. A nil logger is
// tolerated via the noopLogger default.
type Logger interface {
	Warnf(format string, args ...any)
}

type noopLogger struct{}

func (noopLogger) Warnf(string, ...any) {}
