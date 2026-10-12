//go:build windows

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jarvis/mouse-mover/internal/config"
	"github.com/jarvis/mouse-mover/internal/engine"
	"github.com/jarvis/mouse-mover/internal/ui"
	"github.com/jarvis/mouse-mover/internal/winapi"
)

func main() {
	cfg, _ := config.Load()

	ctrl := winapi.NewCursorController()
	eng := engine.New(ctrl, &stdLogger{})
	defer eng.Stop()

	if cfg.AutoStart {
		pat, _ := engine.ParsePattern(cfg.Pattern)
		eng.Start(cfg.Duration(), cfg.MoveInterval(), pat, int32(cfg.DeltaPixels))
	}

	saveCfg := func(c config.Config) {
		if err := config.Save(c); err != nil {
			log.Printf("failed to save config: %v", err)
		}
	}

	if err := ui.Run(cfg, eng, saveCfg); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

type stdLogger struct{}

func (l *stdLogger) Warnf(format string, args ...any) {
	log.Printf("[WARN] "+format, args...)
}
