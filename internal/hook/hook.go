// Package hook installs the OS input hook and forwards raw events to a Handler.
package hook

import (
	"context"
	"errors"
	"runtime"
	"sync/atomic"

	"github.com/ironpark/ivent/key"
)

var (
	// ErrAlreadyRunning is returned by Run when the hook is already running.
	ErrAlreadyRunning = errors.New("ivent: hook is already running")
	// ErrHookFailed is returned by Run when the OS input hook cannot be installed.
	ErrHookFailed = errors.New("ivent: failed to install input hook")
	// ErrUnsupported is returned for features that are not available on the current platform.
	ErrUnsupported = errors.New("ivent: not supported on this platform")
)

// Kind is the type of a raw input event.
type Kind uint8

const (
	KeyDown Kind = iota + 1 // key or mouse button pressed (also sent for auto-repeat)
	KeyUp                   // key or mouse button released
	MouseMove
	MouseWheel
)

// Event flags.
const (
	FlagInjected = 1 << iota // synthesized by software rather than a physical device
	FlagOwn                  // sent by this process with Send
)

// Event is a raw input event reported by the OS hook.
type Event struct {
	Kind  Kind
	Code  key.Code // key code; mouse buttons are reported as key.MouseLeft etc.
	X, Y  int32    // mouse position
	Delta int32    // wheel delta
	PID   int      // process that receives the event, 0 if unknown
	Flags uint8
}

// Handler receives raw events on the OS hook thread. It must return quickly.
type Handler interface {
	// Handle processes an event and reports whether it should be blocked.
	Handle(Event) (suppress bool)
	// Reset is called when the OS may have dropped events, so the key state is unreliable.
	Reset()
}

// Config selects what the hook listens to.
type Config struct {
	Keyboard bool
	Mouse    bool
	// Exclusive grabs keyboards on Linux so that events can be blocked. Ignored elsewhere.
	Exclusive bool
	// OnReady, if set, is called on the goroutine of Run once the hook is installed.
	OnReady func(canSuppress bool)
}

var (
	running     atomic.Bool
	suppressing atomic.Bool
	handler     atomic.Pointer[Handler]
	// readyCh is signalled by the hook thread once the hook is installed and platformStop can be used.
	readyCh = make(chan struct{}, 1)
)

// Suppressing reports whether the running hook can block events: always on Windows, on macOS when the
// process has Accessibility permission, and on Linux when keyboards are grabbed (Config.Exclusive).
func Suppressing() bool {
	return running.Load() && suppressing.Load()
}

// KeyForChar returns the key that types r (without modifiers) on the current keyboard layout.
// It returns false if no such key exists or the platform cannot tell (Linux).
func KeyForChar(r rune) (key.Code, bool) {
	for code := key.Code(0); code < 256; code++ {
		if !key.Known(code) || key.IsModifier(code) || key.IsMouseButton(code) || key.IsKeypad(code) {
			continue
		}
		if c, ok := charForKey(code); ok && c == r {
			return code, true
		}
	}
	return 0, false
}

// Run installs the OS input hook and delivers events to h until ctx is done.
func Run(ctx context.Context, cfg Config, h Handler) error {
	if !running.CompareAndSwap(false, true) {
		return ErrAlreadyRunning
	}
	defer running.Store(false)
	handler.Store(&h)
	defer handler.Store(nil)

	done := make(chan error, 1)
	go func() {
		// The hook and its run loop / message loop must live on a single OS thread.
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		done <- platformStart(cfg)
	}()

	select {
	case <-readyCh:
	case err := <-done:
		return err
	}
	if cfg.OnReady != nil {
		cfg.OnReady(suppressing.Load())
	}
	select {
	case <-ctx.Done():
		platformStop()
		return <-done
	case err := <-done:
		return err
	}
}

func dispatch(ev Event) bool {
	if h := handler.Load(); h != nil {
		return (*h).Handle(ev)
	}
	return false
}

func reset() {
	if h := handler.Load(); h != nil {
		(*h).Reset()
	}
}

func ready(canSuppress bool) {
	suppressing.Store(canSuppress)
	readyCh <- struct{}{}
}
