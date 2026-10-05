package ivent_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ironpark/ivent"
	"github.com/ironpark/ivent/iventtest"
	"github.com/ironpark/ivent/key"
)

// setup returns a Hook in simulation mode that discards warnings.
func setup(t *testing.T) (*ivent.Hook, *iventtest.Sim) {
	t.Helper()
	h := ivent.New(ivent.WithWarnings(func(error) {}))
	return h, iventtest.New(h)
}

// counter returns a callback and a function that flushes callbacks and returns the call count.
func counter(sim *iventtest.Sim) (func(ivent.Event), func() int) {
	n := 0
	return func(ivent.Event) { n++ }, func() int { sim.Flush(); return n }
}

func mustBind(t *testing.T, h *ivent.Hook, spec string, fn func(ivent.Event), opts ...ivent.BindOption) func() {
	t.Helper()
	unbind, err := h.Bind(spec, fn, opts...)
	if err != nil {
		t.Fatal(err)
	}
	return unbind
}

func TestExactMatch(t *testing.T) {
	h, sim := setup(t)
	fn, count := counter(sim)
	mustBind(t, h, "A+S+D", fn)

	sim.Press(key.A, key.S, key.D)
	sim.Press(key.F)
	if count() != 1 {
		t.Fatalf("expected 1 call, got %d", count())
	}
}

func TestContainsAndRepeat(t *testing.T) {
	h, sim := setup(t)
	contains, containsCount := counter(sim)
	repeat, repeatCount := counter(sim)
	mustBind(t, h, "A+S", contains, ivent.Contains())
	mustBind(t, h, "F", repeat, ivent.Repeat())

	sim.Press(key.D, key.A, key.S)
	sim.Release(key.D, key.A, key.S)
	sim.Press(key.F, key.F, key.F)
	if containsCount() != 1 || repeatCount() != 3 {
		t.Fatalf("Contains = %d (want 1), Repeat = %d (want 3)", containsCount(), repeatCount())
	}
}

func TestUnbind(t *testing.T) {
	h, sim := setup(t)
	fn, count := counter(sim)
	mustBind(t, h, "A", fn)()
	sim.Tap(key.A)
	if count() != 0 || len(h.Bindings()) != 0 {
		t.Fatal("unbound binding fired or is still listed")
	}
}

func TestSideIndependentModifier(t *testing.T) {
	h, sim := setup(t)
	any, anyCount := counter(sim)
	right, rightCount := counter(sim)
	mustBind(t, h, "Ctrl+C", any)
	mustBind(t, h, "RightCtrl+C", right)

	sim.Tap(key.RightCtrl, key.C)
	sim.Tap(key.LeftCtrl, key.C)
	if anyCount() != 2 || rightCount() != 1 {
		t.Fatalf("Ctrl+C = %d (want 2), RightCtrl+C = %d (want 1)", anyCount(), rightCount())
	}
}

func TestMouseButtonInCombo(t *testing.T) {
	h, sim := setup(t)
	var got ivent.Event
	mustBind(t, h, "Ctrl+MouseLeft", func(e ivent.Event) { got = e })
	sim.MoveMouse(10, 20)
	sim.Press(key.Ctrl, key.MouseLeft)
	sim.Flush()
	if got.Key != key.MouseLeft || got.X != 10 || got.Y != 20 || !got.IsPressed(key.Ctrl) {
		t.Fatalf("unexpected event %+v", got)
	}
}

func TestOnRelease(t *testing.T) {
	h, sim := setup(t)
	fn, count := counter(sim)
	mustBind(t, h, "Cmd", fn, ivent.OnRelease())

	sim.Press(key.LeftSuper)
	if count() != 0 {
		t.Fatal("OnRelease fired on press")
	}
	sim.Release(key.LeftSuper)
	sim.Tap(key.LeftSuper, key.C) // Cmd+C is not a tap of Cmd
	if count() != 1 {
		t.Fatalf("expected 1 call, got %d", count())
	}
}

func TestDoubleTap(t *testing.T) {
	h, sim := setup(t)
	fn, count := counter(sim)
	mustBind(t, h, "Shift", fn, ivent.DoubleTap(300*time.Millisecond))

	sim.Tap(key.LeftShift)
	sim.Advance(100 * time.Millisecond)
	sim.Tap(key.RightShift)
	if count() != 1 {
		t.Fatalf("double tap within interval must fire, got %d", count())
	}
	sim.Advance(time.Second)
	sim.Tap(key.LeftShift)
	sim.Advance(time.Second)
	sim.Tap(key.LeftShift)
	sim.Advance(time.Second)
	sim.Tap(key.LeftShift)
	sim.Tap(key.A)
	sim.Tap(key.LeftShift)
	if count() != 1 {
		t.Fatalf("slow taps or a key in between must not fire, got %d", count())
	}
}

func TestHold(t *testing.T) {
	h, sim := setup(t)
	fired := make(chan struct{}, 1)
	mustBind(t, h, "Space", func(ivent.Event) { fired <- struct{}{} }, ivent.Hold(30*time.Millisecond))
	wait := func() bool {
		deadline := time.After(200 * time.Millisecond)
		for {
			sim.Flush()
			select {
			case <-fired:
				return true
			case <-deadline:
				return false
			case <-time.After(5 * time.Millisecond):
			}
		}
	}

	sim.Press(key.SpaceBar)
	if !wait() {
		t.Fatal("hold must fire")
	}
	sim.Release(key.SpaceBar)
	sim.Tap(key.SpaceBar)
	if wait() {
		t.Fatal("hold released early must not fire")
	}
}

func TestSequence(t *testing.T) {
	h, sim := setup(t)
	fn, count := counter(sim)
	mustBind(t, h, "Ctrl+K Ctrl+C", fn)

	// Ctrl held for both steps.
	sim.Press(key.LeftCtrl)
	sim.Tap(key.K)
	sim.Tap(key.C)
	sim.Release(key.LeftCtrl)
	// Ctrl released and pressed again between steps.
	sim.Type("Ctrl+K Ctrl+C")
	if count() != 2 {
		t.Fatalf("expected 2 calls, got %d", count())
	}

	// A different key or a long pause resets the sequence.
	sim.Type("Ctrl+K X Ctrl+C")
	sim.Type("Ctrl+K")
	sim.Advance(ivent.DefaultSequenceTimeout + time.Millisecond)
	sim.Type("Ctrl+C")
	if count() != 2 {
		t.Fatalf("a wrong key or timeout must reset the sequence, got %d", count())
	}
}

func TestSequenceWithTrigger(t *testing.T) {
	h, sim := setup(t)
	fn, count := counter(sim)
	mustBind(t, h, "Ctrl+K Shift", fn, ivent.DoubleTap(300*time.Millisecond))

	sim.Type("Ctrl+K")
	sim.Tap(key.LeftShift)
	if count() != 0 {
		t.Fatal("the trigger of the last step must apply")
	}
	sim.Tap(key.LeftShift)
	if count() != 1 {
		t.Fatalf("expected 1 call, got %d", count())
	}
}

func TestSuppress(t *testing.T) {
	h, sim := setup(t)
	fn, count := counter(sim)
	mustBind(t, h, "Ctrl+J", fn, ivent.Suppress())

	if sim.Press(key.LeftCtrl) {
		t.Fatal("modifier must not be blocked")
	}
	if !sim.Press(key.J) || !sim.Press(key.J) || !sim.Release(key.J) {
		t.Fatal("trigger key, its repeat and its release must be blocked")
	}
	if sim.Press(key.K) {
		t.Fatal("unrelated key must not be blocked")
	}
	if count() != 1 {
		t.Fatalf("expected 1 call, got %d", count())
	}
}

func TestAppFilter(t *testing.T) {
	h, sim := setup(t)
	only, onlyCount := counter(sim)
	except, exceptCount := counter(sim)
	mustBind(t, h, "A", only, ivent.OnlyApps("safari"))
	mustBind(t, h, "A", except, ivent.ExceptApps("com.apple.Safari"))

	sim.Focus(ivent.App{Name: "Safari", ID: "com.apple.Safari"})
	sim.Tap(key.A)
	sim.Focus(ivent.App{Name: "Terminal", ID: "com.apple.Terminal"})
	sim.Tap(key.A)
	if onlyCount() != 1 || exceptCount() != 1 {
		t.Fatalf("OnlyApps = %d (want 1), ExceptApps = %d (want 1)", onlyCount(), exceptCount())
	}
}

func TestPause(t *testing.T) {
	h, sim := setup(t)
	fn, count := counter(sim)
	mustBind(t, h, "A", fn)
	h.Pause()
	sim.Tap(key.A)
	h.Resume()
	sim.Tap(key.A)
	if count() != 1 {
		t.Fatalf("expected 1 call, got %d", count())
	}
}

func TestAsync(t *testing.T) {
	h, sim := setup(t)
	fn, count := counter(sim)
	mustBind(t, h, "A", fn, ivent.Async())
	sim.Tap(key.A)
	if count() != 1 {
		t.Fatal("Flush must run Async callbacks")
	}
}

func TestBindValidation(t *testing.T) {
	h, _ := setup(t)
	nop := func(ivent.Event) {}
	invalid := []struct {
		spec string
		opts []ivent.BindOption
	}{
		{"A", []ivent.BindOption{ivent.Suppress(), ivent.OnRelease()}},
		{"A", []ivent.BindOption{ivent.Suppress(), ivent.Hold(time.Second)}},
		{"A", []ivent.BindOption{ivent.DoubleTap(0)}},
		{"A", []ivent.BindOption{ivent.OnlyApps("x"), ivent.ExceptApps("y")}},
		{"A B", []ivent.BindOption{ivent.SequenceTimeout(0)}},
	}
	for _, tt := range invalid {
		if _, err := h.Bind(tt.spec, nop, tt.opts...); !errors.Is(err, ivent.ErrInvalidBinding) {
			t.Errorf("Bind(%q) error = %v, want ErrInvalidBinding", tt.spec, err)
		}
	}
	if _, err := h.Bind("A", nil); !errors.Is(err, ivent.ErrInvalidBinding) {
		t.Errorf("nil callback: %v", err)
	}
	if _, err := h.Bind("Ctrl+Shitf", nop); !errors.Is(err, key.ErrUnknownKey) {
		t.Errorf("unknown key: %v", err)
	}
	if _, err := h.BindCombo(key.Combo{}, nop); !errors.Is(err, ivent.ErrInvalidBinding) {
		t.Errorf("empty combo: %v", err)
	}
	if _, err := h.BindCombo(key.NewCombo(key.A), nop, ivent.ByCharacter()); !errors.Is(err, ivent.ErrInvalidBinding) {
		t.Errorf("ByCharacter with BindCombo: %v", err)
	}

	mustBind(t, h, "Ctrl+A", nop)
	if _, err := h.Bind("ctrl + a", nop); !errors.Is(err, ivent.ErrDuplicateBinding) {
		t.Errorf("duplicate: %v", err)
	}
	if _, err := h.Bind("Ctrl+A", nop, ivent.OnlyApps("Safari")); err != nil {
		t.Errorf("same keys in another app must be allowed: %v", err)
	}
	if got := h.Bindings(); len(got) != 2 || got[0].String() != "CTRL+A" || got[0].Trigger != "press" {
		t.Errorf("Bindings() = %+v", got)
	}
}

func TestRecord(t *testing.T) {
	tests := []struct {
		name  string
		input func(*iventtest.Sim)
		want  key.Combo
	}{
		{"combo", func(s *iventtest.Sim) { s.Press(key.LeftCtrl, key.LeftShift, key.A) }, key.MustParseCombo("Ctrl+Shift+A")},
		{"modifier alone", func(s *iventtest.Sim) { s.Tap(key.RightSuper) }, key.NewCombo(key.Super)},
		{"plain click ignored", func(s *iventtest.Sim) { s.Tap(key.MouseLeft); s.Press(key.F5) }, key.NewCombo(key.F5)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, sim := setup(t)
			fn, count := counter(sim)
			mustBind(t, h, "Ctrl+Shift+A", fn)

			result := make(chan key.Combo, 1)
			go func() {
				c, _ := h.Record(context.Background())
				result <- c
			}()
			// Wait until Record has started.
			for _, err := h.Record(canceled()); !errors.Is(err, ivent.ErrRecording); _, err = h.Record(canceled()) {
				time.Sleep(time.Millisecond)
			}
			tt.input(sim)
			select {
			case got := <-result:
				if got != tt.want {
					t.Fatalf("Record = %s, want %s", got, tt.want)
				}
			case <-time.After(time.Second):
				t.Fatal("Record did not return")
			}
			if count() != 0 {
				t.Fatal("bindings must not fire while recording")
			}
		})
	}
}

func canceled() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

func TestEvents(t *testing.T) {
	h, sim := setup(t)
	events := h.Events()
	sim.MoveMouse(10, 20)
	sim.Press(key.LeftCtrl, key.A)
	mv, ctrl, a := <-events, <-events, <-events
	if mv.Kind != ivent.MouseMove || mv.X != 10 || ctrl.Key != key.LeftCtrl || a.Kind != ivent.KeyDown {
		t.Fatalf("unexpected events: %+v %+v %+v", mv, ctrl, a)
	}
	if a.Combo() != key.NewCombo(key.LeftCtrl, key.A) || a.Combo().SideIndependent().Format(key.StyleText) != "Ctrl+A" {
		t.Fatalf("Combo() = %s", a.Combo())
	}
	if h.Pressed() != a.Combo() {
		t.Fatalf("Pressed() = %s", h.Pressed())
	}
}
