package ivent

import (
	"context"
	"errors"

	"github.com/ironpark/ivent/key"
)

// ErrRecording is returned by Record when another recording is in progress.
var ErrRecording = errors.New("ivent: already recording")

// Record waits for the user to press a key combination and returns it, for example to let users
// choose a shortcut in a settings screen. Bindings do not fire while recording, and the recorded key
// press is blocked when possible (see CanSuppress) so the focused application does not receive it.
//
// The combination is complete when a key other than a modifier is pressed (such as Ctrl+Shift+A),
// or when modifiers alone are pressed and released (such as Cmd). Plain left and right clicks are
// ignored so the user can still click around. Modifiers in the result are side-independent.
// The hook must be running.
func (h *Hook) Record(ctx context.Context) (key.Combo, error) {
	r := &recorder{result: make(chan key.Combo, 1)}
	h.mu.Lock()
	if h.recorder != nil {
		h.mu.Unlock()
		return key.Combo{}, ErrRecording
	}
	h.recorder = r
	h.mu.Unlock()
	defer func() {
		h.mu.Lock()
		if h.recorder == r {
			h.recorder = nil
		}
		h.mu.Unlock()
	}()

	select {
	case c := <-r.result:
		return c, nil
	case <-ctx.Done():
		return key.Combo{}, ctx.Err()
	}
}

type recorder struct {
	modifiers key.Table // modifiers held at the latest modifier press
	done      bool
	result    chan key.Combo
}

// handle processes a key event while recording and reports whether to block it.
// It is called with Hook.mu held.
func (r *recorder) handle(ev Event) (suppress bool) {
	switch {
	case ev.Kind == KeyDown && !ev.Repeat && key.IsModifier(ev.Key):
		r.modifiers = ev.state
	case ev.Kind == KeyDown && !ev.Repeat:
		combo := ev.state.Combo().SideIndependent()
		if (ev.Key == key.MouseLeft || ev.Key == key.MouseRight) && combo == key.NewCombo(ev.Key) {
			return false
		}
		r.finish(combo)
		return true
	case ev.Kind == KeyUp && ev.state == (key.Table{}) && r.modifiers != (key.Table{}):
		r.finish(r.modifiers.Combo().SideIndependent())
	}
	return false
}

func (r *recorder) finish(c key.Combo) {
	r.done = true
	r.result <- c
}
