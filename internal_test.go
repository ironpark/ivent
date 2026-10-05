package ivent

import (
	"errors"
	"sync"
	"testing"

	"github.com/ironpark/ivent/internal/hook"
	"github.com/ironpark/ivent/key"
)

// newTestHook returns a Hook whose OS key state always reports keys as held.
func newTestHook(opts ...Option) *Hook {
	h := New(append([]Option{WithWarnings(func(error) {})}, opts...)...)
	h.keyState = func(key.Code) (bool, bool) { return true, true }
	h.canSuppress = func() bool { return true }
	return h
}

func (h *Hook) flushForTest() int {
	n := 0
	for {
		select {
		case c := <-h.dispatch:
			c.fn(c.ev)
			n++
		default:
			return n
		}
	}
}

func TestInjected(t *testing.T) {
	for _, tt := range []struct {
		opts  []Option
		flags uint8
		want  int
	}{
		{nil, hook.FlagInjected, 0},
		{[]Option{WithInjected()}, hook.FlagInjected, 1},
		{[]Option{WithInjected()}, hook.FlagInjected | hook.FlagOwn, 0}, // never react to our own Send
	} {
		h := newTestHook(tt.opts...)
		h.BindCombo(key.NewCombo(key.A), func(Event) {})
		h.handle(hook.Event{Kind: hook.KeyDown, Code: key.A, Flags: tt.flags})
		if got := h.flushForTest(); got != tt.want {
			t.Errorf("flags %d with %d options: %d calls, want %d", tt.flags, len(tt.opts), got, tt.want)
		}
	}
}

func TestResync(t *testing.T) {
	h := newTestHook()
	h.handle(hook.Event{Kind: hook.KeyDown, Code: key.LeftCtrl})
	// The OS reports Ctrl as released: its key up was missed.
	h.keyState = func(c key.Code) (bool, bool) { return c != key.LeftCtrl, true }
	h.BindCombo(key.NewCombo(key.C), func(Event) {})
	h.handle(hook.Event{Kind: hook.KeyDown, Code: key.C})
	if h.flushForTest() != 1 {
		t.Fatal("a stuck key must be released when the OS no longer reports it")
	}
}

func TestCannotSuppressWarning(t *testing.T) {
	var mu sync.Mutex
	var warnings []error
	done := make(chan struct{}, 1)
	h := New(WithWarnings(func(err error) {
		mu.Lock()
		warnings = append(warnings, err)
		mu.Unlock()
		done <- struct{}{}
	}))
	h.canSuppress = func() bool { return false }
	h.BindCombo(key.NewCombo(key.J), func(Event) {}, Suppress())
	h.ready(false)
	<-done
	mu.Lock()
	defer mu.Unlock()
	if len(warnings) != 1 || !errors.Is(warnings[0], ErrCannotSuppress) {
		t.Fatalf("warnings = %v", warnings)
	}
	if h.handle(hook.Event{Kind: hook.KeyDown, Code: key.J}) {
		t.Fatal("events must not be blocked when the hook cannot block")
	}
}

func TestDroppedWarning(t *testing.T) {
	warned := make(chan error, 10)
	h := New(WithBuffer(1), WithWarnings(func(err error) { warned <- err }))
	h.keyState = func(key.Code) (bool, bool) { return true, true }
	h.BindCombo(key.NewCombo(key.A), func(Event) {}, Repeat())
	for i := 0; i < 3; i++ {
		h.handle(hook.Event{Kind: hook.KeyDown, Code: key.A})
	}
	if err := <-warned; !errors.Is(err, ErrCallbacksDropped) || h.Dropped() != 2 {
		t.Fatalf("warning = %v, dropped = %d", err, h.Dropped())
	}
}

func TestIsPowerOf10(t *testing.T) {
	for n, want := range map[uint64]bool{1: true, 2: false, 10: true, 11: false, 100: true, 1000: true, 1010: false} {
		if isPowerOf10(n) != want {
			t.Errorf("isPowerOf10(%d) != %v", n, want)
		}
	}
}
