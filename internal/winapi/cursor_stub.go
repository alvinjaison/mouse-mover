//go:build !windows

package winapi

import "github.com/jarvis/mouse-mover/internal/engine"

type stubController struct{}

// NewCursorController returns a stub that always returns ErrUnsupported.
func NewCursorController() engine.CursorController {
	return &stubController{}
}

func (s *stubController) GetPos() (engine.Point, error)        { return engine.Point{}, ErrUnsupported }
func (s *stubController) MoveTo(engine.Point) error             { return ErrUnsupported }
func (s *stubController) VirtualScreen() (engine.Rect, error)   { return engine.Rect{}, ErrUnsupported }
