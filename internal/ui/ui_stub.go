//go:build !windows && !darwin

package ui

import (
	"errors"

	"github.com/jarvis/mouse-mover/internal/config"
	"github.com/jarvis/mouse-mover/internal/engine"
)

// ErrUnsupported is returned on non-Windows platforms.
var ErrUnsupported = errors.New("ui: not supported on this platform")

// Run starts the GUI. On non-Windows it returns an error immediately.
func Run(_ config.Config, _ *engine.Engine, _ func(config.Config)) error {
	return ErrUnsupported
}
