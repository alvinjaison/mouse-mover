//go:build darwin

package ui

/*
#cgo LDFLAGS: -framework Cocoa
#include <stdlib.h>

// Objective-C implementation is in ui_darwin.m
extern void RunCocoaApp(
    int durationMinutes,
    int moveIntervalSeconds,
    int patternIndex,
    int deltaPixels,
    int startMinimized,
    int autoStart
);
extern void CocoaStop(void);
*/
import "C"

import (
	"github.com/jarvis/mouse-mover/internal/config"
	"github.com/jarvis/mouse-mover/internal/engine"
)

var (
	gEng     *engine.Engine
	gCfg     config.Config
	gSaveCfg func(config.Config)
)

// Run starts the macOS menu bar app. Blocks until quit.
func Run(cfg config.Config, eng *engine.Engine, saveCfg func(config.Config)) error {
	gEng = eng
	gCfg = cfg
	gSaveCfg = saveCfg

	patIdx := 0
	switch cfg.Pattern {
	case "jiggle":
		patIdx = 1
	case "square":
		patIdx = 2
	}

	startMin := 0
	if cfg.StartMinimized {
		startMin = 1
	}
	autoStart := 0
	if cfg.AutoStart {
		autoStart = 1
	}

	C.RunCocoaApp(
		C.int(cfg.DurationMinutes),
		C.int(cfg.MoveIntervalSeconds),
		C.int(patIdx),
		C.int(cfg.DeltaPixels),
		C.int(startMin),
		C.int(autoStart),
	)
	return nil
}

//export goStartEngine
func goStartEngine(durationMin, intervalSec, patternIdx, delta C.int) {
	patterns := []engine.Pattern{engine.PatternNudge, engine.PatternJiggle, engine.PatternSquare}
	pat := engine.PatternNudge
	if int(patternIdx) < len(patterns) {
		pat = patterns[int(patternIdx)]
	}

	dur := config.Config{DurationMinutes: int(durationMin), MoveIntervalSeconds: int(intervalSec)}
	gEng.Start(dur.Duration(), dur.MoveInterval(), pat, int32(delta))
}

//export goStopEngine
func goStopEngine() {
	gEng.Stop()
}

//export goIsRunning
func goIsRunning() C.int {
	if gEng.IsRunning() {
		return 1
	}
	return 0
}

//export goSaveConfig
func goSaveConfig(durationMin, intervalSec, patternIdx, delta C.int) {
	patterns := []string{"nudge", "jiggle", "square"}
	pat := "nudge"
	if int(patternIdx) < len(patterns) {
		pat = patterns[int(patternIdx)]
	}

	cfg := config.Config{
		DurationMinutes:     int(durationMin),
		MoveIntervalSeconds: int(intervalSec),
		Pattern:             pat,
		DeltaPixels:         int(delta),
		StartMinimized:      gCfg.StartMinimized,
		AutoStart:           gCfg.AutoStart,
	}
	if cfg.MoveIntervalSeconds == 0 {
		cfg.MoveIntervalSeconds = config.DefaultMoveIntervalS
	}
	if cfg.DeltaPixels == 0 {
		cfg.DeltaPixels = 2
	}

	gCfg = cfg
	if gSaveCfg != nil {
		gSaveCfg(cfg)
	}
}
