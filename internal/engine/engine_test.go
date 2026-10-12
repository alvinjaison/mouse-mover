package engine

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// ---------- fake controller ----------

type call struct {
	op    string // "get", "move", "screen"
	point Point
}

type fakeCtrl struct {
	mu       sync.Mutex
	pos      Point
	screen   Rect
	calls    []call
	failNext error
}

func newFakeCtrl(pos Point, screen Rect) *fakeCtrl {
	return &fakeCtrl{pos: pos, screen: screen}
}

func (f *fakeCtrl) GetPos() (Point, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, call{op: "get", point: f.pos})
	if f.failNext != nil {
		err := f.failNext
		f.failNext = nil
		return Point{}, err
	}
	return f.pos, nil
}

func (f *fakeCtrl) MoveTo(p Point) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, call{op: "move", point: p})
	if f.failNext != nil {
		err := f.failNext
		f.failNext = nil
		return err
	}
	f.pos = p
	return nil
}

func (f *fakeCtrl) VirtualScreen() (Rect, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, call{op: "screen"})
	if f.failNext != nil {
		err := f.failNext
		f.failNext = nil
		return Rect{}, err
	}
	return f.screen, nil
}

func (f *fakeCtrl) getCalls() []call {
	f.mu.Lock()
	defer f.mu.Unlock()
	c := make([]call, len(f.calls))
	copy(c, f.calls)
	return c
}

// ---------- fake ticker ----------

type fakeTicker struct {
	C chan time.Time
}

func newFakeTicker() *fakeTicker {
	return &fakeTicker{C: make(chan time.Time, 1)}
}

func (ft *fakeTicker) tick() { ft.C <- time.Now() }

func (ft *fakeTicker) toTimeTicker() *time.Ticker {
	real := time.NewTicker(time.Hour)
	real.Stop()
	real.C = ft.C
	return real
}

// ---------- fake timer (for duration) ----------

type fakeTimer struct {
	C chan time.Time
}

func newFakeTimer() *fakeTimer {
	return &fakeTimer{C: make(chan time.Time, 1)}
}

func (ft *fakeTimer) fire() { ft.C <- time.Now() }

func (ft *fakeTimer) toTimeTimer() *time.Timer {
	real := time.NewTimer(time.Hour)
	real.Stop()
	real.C = ft.C
	return real
}

// ---------- tests ----------

func TestDuplicateStartIsNoop(t *testing.T) {
	ctrl := newFakeCtrl(Point{100, 100}, Rect{0, 0, 1920, 1080})
	eng := New(ctrl, nil)
	ft := newFakeTicker()
	eng.newTicker = func(d time.Duration) *time.Ticker { return ft.toTimeTicker() }
	eng.newTimer = func(d time.Duration) *time.Timer { return newFakeTimer().toTimeTimer() }

	eng.Start(time.Hour, time.Second, PatternNudge, 2)
	eng.Start(time.Hour, time.Second, PatternNudge, 2) // no-op

	s := eng.Status()
	if !s.Running {
		t.Fatal("expected running")
	}

	eng.Stop()
	if eng.IsRunning() {
		t.Fatal("expected stopped")
	}
}

func TestStopCancelsPromptly(t *testing.T) {
	ctrl := newFakeCtrl(Point{100, 100}, Rect{0, 0, 1920, 1080})
	eng := New(ctrl, nil)
	ft := newFakeTicker()
	eng.newTicker = func(d time.Duration) *time.Ticker { return ft.toTimeTicker() }
	eng.newTimer = func(d time.Duration) *time.Timer { return newFakeTimer().toTimeTimer() }

	eng.Start(time.Hour, time.Second, PatternNudge, 2)

	done := make(chan struct{})
	go func() {
		eng.Stop()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Stop did not return within 2s")
	}
}

func TestTickCallOrdering(t *testing.T) {
	ctrl := newFakeCtrl(Point{500, 500}, Rect{0, 0, 1920, 1080})
	eng := New(ctrl, nil)
	ft := newFakeTicker()
	eng.newTicker = func(d time.Duration) *time.Ticker { return ft.toTimeTicker() }
	eng.newTimer = func(d time.Duration) *time.Timer { return newFakeTimer().toTimeTimer() }

	eng.Start(time.Hour, time.Second, PatternNudge, 3)
	ft.tick()
	time.Sleep(50 * time.Millisecond)
	eng.Stop()

	calls := ctrl.getCalls()
	if len(calls) < 4 {
		t.Fatalf("expected at least 4 calls, got %d: %+v", len(calls), calls)
	}
	if calls[0].op != "screen" {
		t.Errorf("call[0] expected screen, got %s", calls[0].op)
	}
	if calls[1].op != "get" {
		t.Errorf("call[1] expected get, got %s", calls[1].op)
	}
	if calls[2].op != "move" || calls[2].point.X != 503 || calls[2].point.Y != 503 {
		t.Errorf("call[2] expected move(503,503), got %s%v", calls[2].op, calls[2].point)
	}
	last := calls[len(calls)-1]
	if last.op != "move" || last.point.X != 500 || last.point.Y != 500 {
		t.Errorf("last expected move(500,500), got %s%v", last.op, last.point)
	}
}

func TestClampToBounds(t *testing.T) {
	ctrl := newFakeCtrl(Point{1918, 1078}, Rect{0, 0, 1920, 1080})
	eng := New(ctrl, nil)
	ft := newFakeTicker()
	eng.newTicker = func(d time.Duration) *time.Ticker { return ft.toTimeTicker() }
	eng.newTimer = func(d time.Duration) *time.Timer { return newFakeTimer().toTimeTimer() }

	eng.Start(time.Hour, time.Second, PatternNudge, 5)
	ft.tick()
	time.Sleep(50 * time.Millisecond)
	eng.Stop()

	for _, c := range ctrl.getCalls() {
		if c.op == "move" {
			if c.point.X < 0 || c.point.X >= 1920 {
				t.Errorf("X out of bounds: %d", c.point.X)
			}
			if c.point.Y < 0 || c.point.Y >= 1080 {
				t.Errorf("Y out of bounds: %d", c.point.Y)
			}
		}
	}
}

func TestClampNegativeOrigin(t *testing.T) {
	screen := Rect{Left: -1920, Top: -500, Right: 1920, Bottom: 1080}
	ctrl := newFakeCtrl(Point{-1920, -500}, screen)
	eng := New(ctrl, nil)
	ft := newFakeTicker()
	eng.newTicker = func(d time.Duration) *time.Ticker { return ft.toTimeTicker() }
	eng.newTimer = func(d time.Duration) *time.Timer { return newFakeTimer().toTimeTimer() }

	eng.Start(time.Hour, time.Second, PatternNudge, 5)
	ft.tick()
	time.Sleep(50 * time.Millisecond)
	eng.Stop()

	for _, c := range ctrl.getCalls() {
		if c.op == "move" {
			if c.point.X < screen.Left || c.point.X > screen.Right-1 {
				t.Errorf("X out of virtual bounds: %d", c.point.X)
			}
			if c.point.Y < screen.Top || c.point.Y > screen.Bottom-1 {
				t.Errorf("Y out of virtual bounds: %d", c.point.Y)
			}
		}
	}
}

func TestNoMovesAfterStop(t *testing.T) {
	ctrl := newFakeCtrl(Point{100, 100}, Rect{0, 0, 1920, 1080})
	eng := New(ctrl, nil)
	ft := newFakeTicker()
	eng.newTicker = func(d time.Duration) *time.Ticker { return ft.toTimeTicker() }
	eng.newTimer = func(d time.Duration) *time.Timer { return newFakeTimer().toTimeTimer() }

	eng.Start(time.Hour, time.Second, PatternNudge, 2)
	ft.tick()
	time.Sleep(50 * time.Millisecond)
	eng.Stop()

	countBefore := len(ctrl.getCalls())
	select {
	case ft.C <- time.Now():
	default:
	}
	time.Sleep(100 * time.Millisecond)

	if len(ctrl.getCalls()) != countBefore {
		t.Errorf("expected no new calls after stop")
	}
}

func TestMoveCountAndLastMove(t *testing.T) {
	ctrl := newFakeCtrl(Point{100, 100}, Rect{0, 0, 1920, 1080})
	eng := New(ctrl, nil)
	ft := newFakeTicker()
	eng.newTicker = func(d time.Duration) *time.Ticker { return ft.toTimeTicker() }
	eng.newTimer = func(d time.Duration) *time.Timer { return newFakeTimer().toTimeTimer() }

	now := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	eng.now = func() time.Time { return now }

	s := eng.Status()
	if s.MoveCount != 0 || !s.LastMove.IsZero() {
		t.Fatalf("expected zero state, got %+v", s)
	}

	eng.Start(time.Hour, time.Second, PatternNudge, 2)
	ft.tick()
	time.Sleep(50 * time.Millisecond)
	ft.tick()
	time.Sleep(50 * time.Millisecond)

	s = eng.Status()
	if s.MoveCount != 2 {
		t.Errorf("expected MoveCount=2, got %d", s.MoveCount)
	}
	if s.LastMove != now {
		t.Errorf("expected LastMove=%v, got %v", now, s.LastMove)
	}

	eng.Stop()
	s = eng.Status()
	if s.Running {
		t.Error("expected stopped")
	}
}

func TestAutoStopAfterDuration(t *testing.T) {
	ctrl := newFakeCtrl(Point{100, 100}, Rect{0, 0, 1920, 1080})
	eng := New(ctrl, nil)
	ft := newFakeTicker()
	fd := newFakeTimer()
	eng.newTicker = func(d time.Duration) *time.Ticker { return ft.toTimeTicker() }
	eng.newTimer = func(d time.Duration) *time.Timer { return fd.toTimeTimer() }

	var stopped atomic.Bool
	eng.SetOnStopped(func() { stopped.Store(true) })

	eng.Start(time.Hour, time.Second, PatternNudge, 2)

	// Fire a tick, confirm it works.
	ft.tick()
	time.Sleep(50 * time.Millisecond)
	if !eng.IsRunning() {
		t.Fatal("should still be running")
	}

	// Fire the duration timer — engine should auto-stop.
	fd.fire()
	time.Sleep(100 * time.Millisecond)

	if eng.IsRunning() {
		t.Error("should have auto-stopped")
	}
	if !stopped.Load() {
		t.Error("onStopped callback not called")
	}
}

func TestStatusRemaining(t *testing.T) {
	ctrl := newFakeCtrl(Point{100, 100}, Rect{0, 0, 1920, 1080})
	eng := New(ctrl, nil)
	ft := newFakeTicker()
	eng.newTicker = func(d time.Duration) *time.Ticker { return ft.toTimeTicker() }
	eng.newTimer = func(d time.Duration) *time.Timer { return newFakeTimer().toTimeTimer() }

	start := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	current := start
	eng.now = func() time.Time { return current }

	eng.Start(time.Hour, time.Second, PatternNudge, 2)

	// Advance time by 20 minutes.
	current = start.Add(20 * time.Minute)
	s := eng.Status()
	if s.Remaining < 39*time.Minute || s.Remaining > 41*time.Minute {
		t.Errorf("expected ~40m remaining, got %v", s.Remaining)
	}

	eng.Stop()
}

func TestPatternSquare(t *testing.T) {
	ctrl := newFakeCtrl(Point{500, 500}, Rect{0, 0, 1920, 1080})
	eng := New(ctrl, nil)
	ft := newFakeTicker()
	eng.newTicker = func(d time.Duration) *time.Ticker { return ft.toTimeTicker() }
	eng.newTimer = func(d time.Duration) *time.Timer { return newFakeTimer().toTimeTimer() }

	eng.Start(time.Hour, time.Second, PatternSquare, 4)
	ft.tick()
	time.Sleep(50 * time.Millisecond)
	eng.Stop()

	var moves []call
	for _, c := range ctrl.getCalls() {
		if c.op == "move" {
			moves = append(moves, c)
		}
	}

	if len(moves) != 5 {
		t.Fatalf("expected 5 moves for square, got %d: %+v", len(moves), moves)
	}
	expected := []Point{{504, 500}, {504, 504}, {500, 504}, {500, 500}, {500, 500}}
	for i, exp := range expected {
		if moves[i].point != exp {
			t.Errorf("move[%d] expected %v, got %v", i, exp, moves[i].point)
		}
	}
}

func TestIndefiniteDuration(t *testing.T) {
	ctrl := newFakeCtrl(Point{100, 100}, Rect{0, 0, 1920, 1080})
	eng := New(ctrl, nil)
	ft := newFakeTicker()
	eng.newTicker = func(d time.Duration) *time.Ticker { return ft.toTimeTicker() }
	// duration=0 means indefinite, newTimer should not be called
	eng.newTimer = func(d time.Duration) *time.Timer {
		t.Error("newTimer should not be called for indefinite duration")
		return time.NewTimer(time.Hour)
	}

	eng.Start(0, time.Second, PatternNudge, 2)
	ft.tick()
	time.Sleep(50 * time.Millisecond)

	s := eng.Status()
	if s.Remaining != 0 {
		t.Errorf("expected 0 remaining for indefinite, got %v", s.Remaining)
	}

	eng.Stop()
}
