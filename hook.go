package ivent

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ironpark/ivent/internal/hook"
	"github.com/ironpark/ivent/key"
)

// Errors returned by Run.
var (
	ErrAlreadyRunning = hook.ErrAlreadyRunning
	ErrHookFailed     = hook.ErrHookFailed
	ErrUnsupported    = hook.ErrUnsupported
)

// Warnings passed to the WithWarnings handler.
var (
	// ErrCannotSuppress means a binding uses Suppress, but the running hook cannot block events.
	ErrCannotSuppress = errors.New("ivent: cannot block events (needs Accessibility permission on macOS; not supported on Linux)")
	// ErrCallbacksDropped means callbacks were dropped because the buffer was full (see Async and WithBuffer).
	ErrCallbacksDropped = errors.New("ivent: callbacks dropped")
)

// Hook listens to global keyboard and mouse input and runs bindings.
// Only one Hook can run at a time in a process.
type Hook struct {
	cfg config

	mu         sync.Mutex
	state      key.Table // keys and mouse buttons held down
	suppressed key.Table // keys whose press was blocked; their repeat and release are blocked too
	bindings   []*binding
	paused     bool
	recorder   *recorder
	running    bool

	mouseX, mouseY atomic.Int32
	events         chan Event
	eventsUsed     atomic.Bool // Events was called, so events are worth building
	dispatch       chan call
	dropped        atomic.Uint64

	// Replaceable for simulation.
	now         func() time.Time
	keyState    func(key.Code) (pressed, ok bool)
	canSuppress func() bool
	lookupApp   func(int) App
}

// call is a binding callback queued for the dispatch goroutine.
type call struct {
	fn    func(Event)
	ev    Event
	async bool
}

func (c call) run() {
	if c.async {
		go c.fn(c.ev)
	} else {
		c.fn(c.ev)
	}
}

type config struct {
	mouse    bool
	injected bool
	buffer   int
	warn     func(error)
}

// Option configures a Hook.
type Option func(*config)

// WithInjected delivers events synthesized by software (other than this process's Send) as well.
// By default they are ignored.
func WithInjected() Option {
	return func(c *config) { c.injected = true }
}

// WithoutMouse stops listening to the mouse.
func WithoutMouse() Option {
	return func(c *config) { c.mouse = false }
}

// WithBuffer sets how many pending callbacks and events are buffered (default 256).
// When the buffer is full, callbacks and events are dropped, but the key state stays correct.
func WithBuffer(n int) Option {
	return func(c *config) { c.buffer = n }
}

// WithWarnings sets the handler for problems that do not stop the hook, such as ErrCannotSuppress
// and ErrCallbacksDropped. By default they are logged with log/slog. The handler runs on its own goroutine.
func WithWarnings(fn func(error)) Option {
	return func(c *config) { c.warn = fn }
}

// New creates a Hook. Call Run to start listening.
func New(opts ...Option) *Hook {
	cfg := config{
		mouse:  true,
		buffer: 256,
		warn:   func(err error) { slog.Warn(err.Error()) },
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &Hook{
		cfg:         cfg,
		events:      make(chan Event, cfg.buffer),
		dispatch:    make(chan call, cfg.buffer),
		now:         time.Now,
		keyState:    hook.KeyState,
		canSuppress: hook.Suppressing,
	}
}

// Run installs the OS input hook and blocks until ctx is done.
// Binding callbacks run one at a time, in order, on a goroutine started by Run (see Async).
// It returns ErrHookFailed if the hook cannot be installed (on macOS, usually a missing
// permission; see CheckPermission) and ErrAlreadyRunning if a Hook is already running.
func (h *Hook) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case c := <-h.dispatch:
				c.run()
			case <-ctx.Done():
				return
			}
		}
	}()
	cfg := hook.Config{Keyboard: true, Mouse: h.cfg.mouse, OnReady: h.ready}
	err := hook.Run(ctx, cfg, adapter{h})
	h.mu.Lock()
	h.running = false
	h.mu.Unlock()
	cancel()
	wg.Wait()
	return err
}

func (h *Hook) ready(canSuppress bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.running = true
	if canSuppress {
		return
	}
	for _, b := range h.bindings {
		if b.suppress {
			h.warn(fmt.Errorf("%w: %s", ErrCannotSuppress, b.info()))
		}
	}
}

// CanSuppress reports whether the running hook can block events for Suppress bindings:
// always on Windows, on macOS with Accessibility permission, never on Linux.
func (h *Hook) CanSuppress() bool {
	return h.canSuppress()
}

// Events returns a channel of every input event. Events are dropped if the channel is not read.
func (h *Hook) Events() <-chan Event {
	h.eventsUsed.Store(true)
	return h.events
}

// Pressed returns the keys and mouse buttons that are currently held down.
func (h *Hook) Pressed() key.Combo {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.state.Combo()
}

// ResetState marks every key as released.
func (h *Hook) ResetState() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.state = key.Table{}
	h.suppressed = key.Table{}
	for _, b := range h.bindings {
		b.reset()
	}
}

// Pause stops all bindings from firing until Resume. Events are still delivered.
func (h *Hook) Pause() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.paused = true
	for _, b := range h.bindings {
		b.reset()
	}
}

// Resume lets bindings fire again after Pause.
func (h *Hook) Resume() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.paused = false
}

// Paused reports whether bindings are paused.
func (h *Hook) Paused() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.paused
}

// Mouse returns the last known mouse position (not available on Linux).
func (h *Hook) Mouse() (x, y int32) {
	return h.mouseX.Load(), h.mouseY.Load()
}

// Dropped returns how many callbacks were dropped because the buffer was full.
func (h *Hook) Dropped() uint64 {
	return h.dropped.Load()
}

// adapter implements hook.Handler, keeping Handle out of the public API of Hook.
type adapter struct{ h *Hook }

func (a adapter) Handle(raw hook.Event) bool { return a.h.handle(raw) }
func (a adapter) Reset()                     { a.h.ResetState() }

// handle runs on the OS hook thread, so it only updates state, matches bindings and queues callbacks.
func (h *Hook) handle(raw hook.Event) bool {
	injected := raw.Flags&hook.FlagInjected != 0
	if injected && (!h.cfg.injected || raw.Flags&hook.FlagOwn != 0) {
		return false
	}
	ev := Event{
		Kind:     EventKind(raw.Kind),
		X:        raw.X,
		Y:        raw.Y,
		Delta:    raw.Delta,
		Injected: injected,
		PID:      raw.PID,
		Time:     h.now(),
		lookup:   h.lookupApp,
	}
	h.mouseX.Store(raw.X)
	h.mouseY.Store(raw.Y)
	if ev.Kind == MouseMove || ev.Kind == MouseWheel {
		// Mouse moves are frequent: only build the event if someone reads Events.
		if h.eventsUsed.Load() {
			h.mu.Lock()
			ev.state = h.state
			h.mu.Unlock()
			h.emit(ev)
		}
		return false
	}
	// Unknown keys are ignored, so they do not break exact matches.
	if !key.Known(raw.Code) {
		return false
	}
	ev.Key = raw.Code
	down := ev.Kind == KeyDown

	h.mu.Lock()
	// A repeat changes nothing, so there is nothing to resynchronize.
	if down && !h.state.IsKeyPressed(raw.Code) {
		h.resyncLocked()
	}
	prev := h.state
	changed := h.state.Set(raw.Code, down)
	ev.Repeat = down && !changed
	ev.state = h.state

	var fire []call
	suppress := false
	switch {
	case h.recorder != nil:
		suppress = h.recorder.handle(ev)
		if h.recorder.done {
			h.recorder = nil
		}
	case !h.paused:
		for _, b := range h.bindings {
			fired, s := b.handle(h, ev, prev, changed)
			if fired {
				fire = append(fire, b.call(ev))
			}
			suppress = suppress || s
		}
	}
	// Without the ability to block, nothing is tracked as blocked either.
	suppress = suppress && h.canSuppress()
	// Once a press is blocked, its repeats and release must be blocked as well.
	switch {
	case h.suppressed.IsKeyPressed(raw.Code):
		suppress = true
		if !down {
			h.suppressed.Set(raw.Code, false)
		}
	case suppress && down:
		h.suppressed.Set(raw.Code, true)
	}
	h.mu.Unlock()

	if h.eventsUsed.Load() {
		h.emit(ev)
	}
	for _, c := range fire {
		h.enqueue(c)
	}
	return suppress
}

// resyncLocked releases keys that the OS no longer reports as held. Key releases can be missed,
// for example while macOS secure input is active, which would otherwise leave keys stuck.
func (h *Hook) resyncLocked() {
	for _, code := range h.state.PressedKeys() {
		// Blocked keys never reach the OS key state on Windows, so they are kept.
		if h.suppressed.IsKeyPressed(code) {
			continue
		}
		if pressed, ok := h.keyState(code); ok && !pressed {
			h.state.Set(code, false)
		}
	}
}

func (h *Hook) emit(ev Event) {
	select {
	case h.events <- ev:
	default:
	}
}

func (h *Hook) enqueue(c call) {
	select {
	case h.dispatch <- c:
	default:
		// Warn on the 1st, 10th, 100th, ... drop to avoid flooding.
		if n := h.dropped.Add(1); isPowerOf10(n) {
			h.warn(fmt.Errorf("%w: %d so far; use Async for slow callbacks", ErrCallbacksDropped, n))
		}
	}
}

func isPowerOf10(n uint64) bool {
	for n >= 10 && n%10 == 0 {
		n /= 10
	}
	return n == 1
}

// warn reports a problem without blocking the caller, which may hold Hook.mu or be the hook thread.
func (h *Hook) warn(err error) {
	if h.cfg.warn != nil {
		go h.cfg.warn(err)
	}
}
