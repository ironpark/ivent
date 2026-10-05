// Package iventtest simulates keyboard and mouse input for testing code that uses ivent,
// without installing an OS hook or needing permissions.
//
//	h := ivent.New()
//	h.Bind("Ctrl+K Ctrl+C", comment)
//	sim := iventtest.New(h)
//	sim.Type("Ctrl+K Ctrl+C")
//	sim.Flush() // runs the callbacks
package iventtest

import (
	"time"

	"github.com/ironpark/ivent"
	"github.com/ironpark/ivent/internal/hook"
	"github.com/ironpark/ivent/internal/simulate"
	"github.com/ironpark/ivent/key"
)

// focusPID is the pid of simulated events; Focus sets the application it resolves to.
const focusPID = 1

// Sim feeds simulated input to a Hook. Do not call Run on the Hook while simulating.
// Sim is not safe for concurrent use.
type Sim struct {
	target simulate.Target
	clock  time.Time
	app    ivent.App
	x, y   int32
}

// New puts h in simulation mode and returns a Sim that drives it.
// Time starts at an arbitrary instant and only moves with Advance; Hold bindings use real timers.
func New(h *ivent.Hook) *Sim {
	s := &Sim{clock: time.Unix(1_000_000_000, 0)}
	s.target = simulate.Attach(h, func() time.Time { return s.clock },
		func(int) (string, string) { return s.app.Name, s.app.ID })
	return s
}

// Press presses keys in order and reports whether any press was blocked by a Suppress binding.
// Side-independent modifiers such as key.Ctrl press their left key.
func (s *Sim) Press(codes ...key.Code) (blocked bool) {
	return s.send(hook.KeyDown, codes)
}

// Release releases keys in order and reports whether any release was blocked.
func (s *Sim) Release(codes ...key.Code) (blocked bool) {
	return s.send(hook.KeyUp, codes)
}

// Tap presses the keys in order and releases them in reverse order.
func (s *Sim) Tap(codes ...key.Code) {
	s.Press(codes...)
	for i := len(codes) - 1; i >= 0; i-- {
		s.Release(codes[i])
	}
}

// Type taps each combination of a spec such as "Ctrl+Shift+A" or "Ctrl+K Ctrl+C".
func (s *Sim) Type(spec string) error {
	steps, err := key.ParseSequence(spec)
	if err != nil {
		return err
	}
	for _, step := range steps {
		s.Tap(step.Codes()...)
	}
	return nil
}

// MoveMouse moves the mouse to x, y.
func (s *Sim) MoveMouse(x, y int32) {
	s.x, s.y = x, y
	s.target.Handle(hook.Event{Kind: hook.MouseMove, X: x, Y: y, PID: focusPID})
}

// Scroll scrolls the mouse wheel by delta.
func (s *Sim) Scroll(delta int32) {
	s.target.Handle(hook.Event{Kind: hook.MouseWheel, X: s.x, Y: s.y, Delta: delta, PID: focusPID})
}

// Advance moves the simulated clock forward, for DoubleTap intervals and sequence timeouts.
func (s *Sim) Advance(d time.Duration) {
	s.clock = s.clock.Add(d)
}

// Focus makes app the application that receives the following events, for OnlyApps and ExceptApps.
func (s *Sim) Focus(app ivent.App) {
	s.app = app
}

// Flush runs the callbacks queued so far on the calling goroutine, including Async ones.
func (s *Sim) Flush() {
	s.target.Flush()
}

func (s *Sim) send(kind hook.Kind, codes []key.Code) (blocked bool) {
	for _, c := range codes {
		ev := hook.Event{Kind: kind, Code: key.Physical(c), X: s.x, Y: s.y, PID: focusPID}
		if s.target.Handle(ev) {
			blocked = true
		}
	}
	return blocked
}
