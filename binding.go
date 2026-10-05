package ivent

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/ironpark/ivent/internal/hook"
	"github.com/ironpark/ivent/key"
)

var (
	// ErrInvalidBinding is returned by Bind for an invalid combination of options.
	ErrInvalidBinding = errors.New("ivent: invalid binding")
	// ErrDuplicateBinding is returned by Bind when an identical binding already exists.
	ErrDuplicateBinding = errors.New("ivent: duplicate binding")
)

type triggerMode uint8

const (
	onPress triggerMode = iota
	onRelease
	onDoubleTap
	onHold
)

func (m triggerMode) String() string {
	return [...]string{"press", "release", "double tap", "hold"}[m]
}

// DefaultSequenceTimeout is the longest pause allowed between the steps of a sequence.
const DefaultSequenceTimeout = 2 * time.Second

// BindOption configures a binding.
type BindOption func(*binding)

// Contains also fires when other keys are held in addition to the combination.
// By default the held keys must match the combination exactly.
func Contains() BindOption {
	return func(b *binding) { b.contains = true }
}

// Repeat also fires on auto-repeat while the combination is held.
func Repeat() BindOption {
	return func(b *binding) { b.repeat = true }
}

// OnRelease fires when a key of the combination is released, if no other key was pressed
// while it was held. For example, binding "Cmd" with OnRelease fires when Cmd is tapped alone.
func OnRelease() BindOption {
	return func(b *binding) { b.mode = onRelease }
}

// DoubleTap fires when the combination is pressed twice within interval,
// with no other key pressed in between.
func DoubleTap(interval time.Duration) BindOption {
	return func(b *binding) {
		b.mode = onDoubleTap
		b.interval = interval
	}
}

// Hold fires once the combination has been held for d.
func Hold(d time.Duration) BindOption {
	return func(b *binding) {
		b.mode = onHold
		b.interval = d
	}
}

// Suppress blocks the key press that triggers the binding (and its repeats and release),
// so that other applications do not receive it. It cannot be combined with OnRelease or Hold,
// whose keys have already been delivered. Blocking needs Accessibility permission on macOS and
// WithExclusive on Linux; see Hook.CanSuppress.
func Suppress() BindOption {
	return func(b *binding) { b.suppress = true }
}

// OnlyApps restricts the binding to the given applications. Names are compared case-insensitively
// with App.Name, App.ID and the base name of App.ID. Not supported on Linux.
func OnlyApps(names ...string) BindOption {
	return func(b *binding) { b.onlyApps = append(b.onlyApps, names...) }
}

// ExceptApps disables the binding in the given applications. See OnlyApps for how names are matched.
func ExceptApps(names ...string) BindOption {
	return func(b *binding) { b.exceptApps = append(b.exceptApps, names...) }
}

// SequenceTimeout sets the longest pause allowed between the steps of a sequence
// (default DefaultSequenceTimeout).
func SequenceTimeout(d time.Duration) BindOption {
	return func(b *binding) { b.timeout = d }
}

// Async runs the callback on its own goroutine, so a slow callback does not delay other bindings.
// By default callbacks run one at a time, in order.
func Async() BindOption {
	return func(b *binding) { b.async = true }
}

// ByCharacter resolves single-character key names such as "z" or "/" to the key that types that
// character on the current keyboard layout, instead of the key at that position on a US layout.
// For example, "Ctrl+Z" then matches the key labelled Z on an AZERTY keyboard. The layout is read
// when the binding is created. Only valid with Bind; on Linux names keep their US positions.
func ByCharacter() BindOption {
	return func(b *binding) { b.byCharacter = true }
}

type binding struct {
	steps []key.Combo
	fn    func(Event)

	mode        triggerMode
	contains    bool
	repeat      bool
	suppress    bool
	async       bool
	byCharacter bool
	interval    time.Duration
	timeout     time.Duration
	onlyApps    []string
	exceptApps  []string

	// Matching state, guarded by Hook.mu.
	step      int       // sequence: number of steps matched so far
	stepTime  time.Time // sequence: time of the last matched step
	tapTime   time.Time // double tap: time of the first tap
	armed     bool      // release: the combination is held with no other key pressed since
	holdTimer *time.Timer
}

// Binding describes a registered binding.
type Binding struct {
	// Steps is the combination, or the combinations of a sequence.
	Steps []key.Combo
	// Trigger is "press", "release", "double tap" or "hold".
	Trigger    string
	OnlyApps   []string
	ExceptApps []string
}

// String returns the steps in the form accepted by Bind, such as "CTRL+K CTRL+C".
func (b Binding) String() string {
	parts := make([]string, len(b.Steps))
	for i, s := range b.Steps {
		parts[i] = s.String()
	}
	return strings.Join(parts, " ")
}

// Bind parses a combination such as "Ctrl+Shift+A", or a sequence such as "Ctrl+K Ctrl+C",
// and runs fn when it is pressed. It returns a function that removes the binding.
func (h *Hook) Bind(spec string, fn func(Event), opts ...BindOption) (unbind func(), err error) {
	b := newBinding(fn, opts)
	var resolve func(rune) (key.Code, bool)
	if b.byCharacter {
		resolve = hook.KeyForChar
	}
	if b.steps, err = key.ParseSequenceWith(spec, resolve); err != nil {
		return nil, err
	}
	return h.add(b)
}

// BindCombo runs fn when the combination is pressed. See Bind.
func (h *Hook) BindCombo(c key.Combo, fn func(Event), opts ...BindOption) (unbind func(), err error) {
	return h.BindSequence([]key.Combo{c}, fn, opts...)
}

// BindSequence runs fn when the combinations are pressed one after another, such as Ctrl+K then Ctrl+C.
// Keys of the next step (such as its modifiers) may be released and pressed again between steps.
// Trigger options such as Hold apply to the last step.
func (h *Hook) BindSequence(steps []key.Combo, fn func(Event), opts ...BindOption) (unbind func(), err error) {
	b := newBinding(fn, opts)
	if b.byCharacter {
		return nil, fmt.Errorf("%w: ByCharacter needs a key name string; use Bind", ErrInvalidBinding)
	}
	b.steps = slices.Clone(steps)
	return h.add(b)
}

func newBinding(fn func(Event), opts []BindOption) *binding {
	b := &binding{fn: fn, timeout: DefaultSequenceTimeout}
	for _, opt := range opts {
		opt(b)
	}
	return b
}

func (b *binding) validate() error {
	switch {
	case b.fn == nil:
		return fmt.Errorf("%w: nil callback", ErrInvalidBinding)
	case len(b.steps) == 0 || slices.ContainsFunc(b.steps, key.Combo.IsEmpty):
		return fmt.Errorf("%w: empty combination", ErrInvalidBinding)
	case b.suppress && (b.mode == onRelease || b.mode == onHold):
		return fmt.Errorf("%w: Suppress cannot be combined with %s triggers", ErrInvalidBinding, b.mode)
	case (b.mode == onDoubleTap || b.mode == onHold) && b.interval <= 0:
		return fmt.Errorf("%w: %s duration must be positive", ErrInvalidBinding, b.mode)
	case b.timeout <= 0:
		return fmt.Errorf("%w: sequence timeout must be positive", ErrInvalidBinding)
	case len(b.onlyApps) > 0 && len(b.exceptApps) > 0:
		return fmt.Errorf("%w: OnlyApps and ExceptApps cannot be combined", ErrInvalidBinding)
	}
	return nil
}

// sameTrigger reports whether two bindings fire on exactly the same input.
func (b *binding) sameTrigger(o *binding) bool {
	return slices.Equal(b.steps, o.steps) && b.mode == o.mode && b.contains == o.contains &&
		b.repeat == o.repeat && b.interval == o.interval &&
		slices.Equal(b.onlyApps, o.onlyApps) && slices.Equal(b.exceptApps, o.exceptApps)
}

func (h *Hook) add(b *binding) (unbind func(), err error) {
	if err := b.validate(); err != nil {
		return nil, err
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if slices.ContainsFunc(h.bindings, b.sameTrigger) {
		return nil, fmt.Errorf("%w: %s", ErrDuplicateBinding, b.info())
	}
	h.bindings = append(h.bindings, b)
	if b.suppress && h.running && !h.canSuppress() {
		h.warn(fmt.Errorf("%w: %s", ErrCannotSuppress, b.info()))
	}
	return func() { h.unbind(b) }, nil
}

func (b *binding) info() Binding {
	return Binding{
		Steps:      slices.Clone(b.steps),
		Trigger:    b.mode.String(),
		OnlyApps:   slices.Clone(b.onlyApps),
		ExceptApps: slices.Clone(b.exceptApps),
	}
}

// Bindings returns the registered bindings.
func (h *Hook) Bindings() []Binding {
	h.mu.Lock()
	defer h.mu.Unlock()
	infos := make([]Binding, len(h.bindings))
	for i, b := range h.bindings {
		infos[i] = b.info()
	}
	return infos
}

func (h *Hook) unbind(b *binding) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if i := slices.Index(h.bindings, b); i >= 0 {
		b.reset()
		h.bindings = slices.Delete(h.bindings, i, i+1)
	}
}

func (b *binding) reset() {
	b.step = 0
	b.stepTime = time.Time{}
	b.tapTime = time.Time{}
	b.armed = false
	b.cancelHold()
}

func (b *binding) cancelHold() {
	if b.holdTimer != nil {
		b.holdTimer.Stop()
		b.holdTimer = nil
	}
}

func (b *binding) matches(c key.Combo, state key.Table) bool {
	if b.contains {
		return c.MatchSubset(state)
	}
	return c.Match(state)
}

func (b *binding) appAllowed(ev Event) bool {
	names := b.onlyApps
	if len(names) == 0 {
		names = b.exceptApps
	}
	if len(names) == 0 {
		return true
	}
	app := ev.App()
	listed := slices.ContainsFunc(names, func(name string) bool {
		return strings.EqualFold(name, app.Name) || strings.EqualFold(name, app.ID) ||
			(app.ID != "" && strings.EqualFold(name, filepath.Base(app.ID)))
	})
	return listed == (len(b.onlyApps) > 0)
}

// handle updates the binding with a key event and reports whether the callback should run
// and whether the event should be blocked. prev is the key state before the event.
// Sequences advance through all but the last step here; the trigger mode applies to the last step.
// It is called with Hook.mu held.
func (b *binding) handle(h *Hook, ev Event, prev key.Table, changed bool) (fire, suppress bool) {
	last := len(b.steps) - 1
	if ev.Kind == KeyDown && !ev.Repeat && last > 0 {
		if b.step > 0 && ev.Time.Sub(b.stepTime) > b.timeout {
			b.reset()
		}
		// A key outside the last step restarts the sequence.
		if b.step == last && !b.matches(b.steps[last], ev.state) && !b.steps[last].Covers(ev.state) {
			b.reset()
		}
		if b.step < last {
			return false, b.advance(ev)
		}
	}
	if b.step < last {
		return false, false
	}
	fire, suppress = b.handleLast(h, ev, prev, changed)
	if fire {
		b.step = 0
	}
	return fire, suppress
}

// advance moves a sequence forward if ev completes its current step, and reports whether to block ev.
func (b *binding) advance(ev Event) (suppress bool) {
	if !b.matches(b.steps[b.step], ev.state) {
		// Pressing part of the current step (such as its modifier) keeps the progress.
		if b.step > 0 && b.steps[b.step].Covers(ev.state) {
			return false
		}
		b.step = 0
		if !b.matches(b.steps[0], ev.state) {
			return false
		}
	}
	if !b.appAllowed(ev) {
		b.step = 0
		return false
	}
	b.step++
	b.stepTime = ev.Time
	return b.suppress
}

// handleLast applies the trigger mode to the last (or only) step.
func (b *binding) handleLast(h *Hook, ev Event, prev key.Table, changed bool) (fire, suppress bool) {
	combo := b.steps[len(b.steps)-1]
	switch b.mode {
	case onPress:
		if ev.Kind == KeyDown && (!ev.Repeat || b.repeat) && b.matches(combo, ev.state) && b.appAllowed(ev) {
			return true, b.suppress
		}
	case onRelease:
		switch {
		case ev.Kind == KeyDown && !ev.Repeat:
			b.armed = b.matches(combo, ev.state)
		case ev.Kind == KeyUp && changed:
			armed := b.armed
			b.armed = false
			if armed && b.matches(combo, prev) && b.appAllowed(ev) {
				return true, false
			}
		}
	case onDoubleTap:
		if ev.Kind != KeyDown || ev.Repeat {
			break
		}
		if !b.matches(combo, ev.state) {
			b.tapTime = time.Time{}
			break
		}
		if !b.tapTime.IsZero() && ev.Time.Sub(b.tapTime) <= b.interval && b.appAllowed(ev) {
			b.tapTime = time.Time{}
			return true, b.suppress
		}
		b.tapTime = ev.Time
	case onHold:
		if ev.Repeat {
			break
		}
		if ev.Kind == KeyDown && b.matches(combo, ev.state) {
			// A hold already in progress keeps running (extra keys with Contains).
			if b.holdTimer == nil && b.appAllowed(ev) {
				b.startHold(h, ev)
			}
		} else if changed {
			b.cancelHold()
		}
	}
	return false, false
}

func (b *binding) startHold(h *Hook, ev Event) {
	var timer *time.Timer
	// Hook.mu is held, so timer is assigned before the callback can take the lock.
	timer = time.AfterFunc(b.interval, func() {
		h.mu.Lock()
		valid := b.holdTimer == timer
		if valid {
			b.holdTimer = nil
			b.step = 0
		}
		h.mu.Unlock()
		if valid {
			h.enqueue(b.call(ev))
		}
	})
	b.holdTimer = timer
}

func (b *binding) call(ev Event) call {
	return call{fn: b.fn, ev: ev, async: b.async}
}
