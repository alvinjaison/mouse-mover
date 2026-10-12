package engine

import (
	"context"
	"sync"
	"time"
)

// Engine drives periodic mouse movement on a cancellable goroutine.
// It auto-stops after the configured duration.
type Engine struct {
	mu        sync.Mutex
	running   bool
	cancel    context.CancelFunc
	done      chan struct{}
	ctrl      CursorController
	logger    Logger
	lastMove  time.Time
	startedAt time.Time
	moveCount uint64
	duration  time.Duration // total active duration, 0 = indefinite

	// onStopped is called (if non-nil) when the engine auto-stops after duration.
	// Called from the engine goroutine, so it must be safe for cross-goroutine use.
	onStopped func()

	// Injected for testing.
	now       func() time.Time
	newTicker func(time.Duration) *time.Ticker
	newTimer  func(time.Duration) *time.Timer
}

// Status is a snapshot of the engine state.
type Status struct {
	Running   bool
	LastMove  time.Time
	StartedAt time.Time
	Elapsed   time.Duration
	Remaining time.Duration // time left before auto-stop, 0 if indefinite or stopped
	MoveCount uint64
}

// New creates an engine that uses ctrl for cursor operations.
func New(ctrl CursorController, logger Logger) *Engine {
	if logger == nil {
		logger = noopLogger{}
	}
	return &Engine{
		ctrl:      ctrl,
		logger:    logger,
		now:       time.Now,
		newTicker: time.NewTicker,
		newTimer:  time.NewTimer,
	}
}

// SetOnStopped sets a callback fired when the engine auto-stops after duration expires.
func (e *Engine) SetOnStopped(fn func()) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.onStopped = fn
}

// Start begins periodic mouse movement for the given duration. If duration is
// 0 it runs indefinitely. moveInterval is the pause between each movement.
// If the engine is already running the call is a no-op.
func (e *Engine) Start(duration time.Duration, moveInterval time.Duration, pattern Pattern, delta int32) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.running {
		return
	}
	if moveInterval < time.Second {
		moveInterval = 30 * time.Second
	}
	if delta < 1 {
		delta = 2
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	e.running = true
	e.cancel = cancel
	e.done = done
	e.startedAt = e.now()
	e.duration = duration

	go e.loop(ctx, done, duration, moveInterval, pattern, delta)
}

// Stop cancels the movement loop and blocks until it exits.
func (e *Engine) Stop() {
	e.mu.Lock()
	if !e.running {
		e.mu.Unlock()
		return
	}
	e.cancel()
	e.running = false
	e.startedAt = time.Time{}
	e.duration = 0
	done := e.done
	e.mu.Unlock()

	<-done
}

// IsRunning returns whether the engine is actively running.
func (e *Engine) IsRunning() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.running
}

// Status returns a point-in-time snapshot.
func (e *Engine) Status() Status {
	e.mu.Lock()
	defer e.mu.Unlock()
	s := Status{
		Running:   e.running,
		LastMove:  e.lastMove,
		StartedAt: e.startedAt,
		MoveCount: e.moveCount,
	}
	if e.running {
		elapsed := e.now().Sub(e.startedAt)
		s.Elapsed = elapsed
		if e.duration > 0 {
			rem := e.duration - elapsed
			if rem < 0 {
				rem = 0
			}
			s.Remaining = rem
		}
	}
	return s
}

func (e *Engine) loop(ctx context.Context, done chan struct{}, duration, moveInterval time.Duration, pattern Pattern, delta int32) {
	defer close(done)

	ticker := e.newTicker(moveInterval)
	defer ticker.Stop()

	// Duration timer: if duration > 0, auto-stop after it expires.
	var deadlineCh <-chan time.Time
	if duration > 0 {
		t := e.newTimer(duration)
		defer t.Stop()
		deadlineCh = t.C
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-deadlineCh:
			// Duration expired — auto-stop.
			e.mu.Lock()
			e.running = false
			e.startedAt = time.Time{}
			e.duration = 0
			cb := e.onStopped
			e.mu.Unlock()
			if cb != nil {
				cb()
			}
			return
		case <-ticker.C:
			e.tick(pattern, delta)
		}
	}
}

func (e *Engine) tick(pattern Pattern, delta int32) {
	screen, err := e.ctrl.VirtualScreen()
	if err != nil {
		e.logger.Warnf("VirtualScreen: %v", err)
		return
	}

	origin, err := e.ctrl.GetPos()
	if err != nil {
		e.logger.Warnf("GetPos: %v", err)
		return
	}

	steps := pattern.offsets(delta)
	for _, off := range steps {
		target := Point{X: origin.X + off.X, Y: origin.Y + off.Y}
		target = screen.Clamp(target)
		if err := e.ctrl.MoveTo(target); err != nil {
			e.logger.Warnf("MoveTo(%v): %v", target, err)
			_ = e.ctrl.MoveTo(screen.Clamp(origin))
			return
		}
	}

	if err := e.ctrl.MoveTo(screen.Clamp(origin)); err != nil {
		e.logger.Warnf("MoveTo(restore %v): %v", origin, err)
	}

	e.mu.Lock()
	e.lastMove = e.now()
	e.moveCount++
	e.mu.Unlock()
}
