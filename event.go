package ivent

import (
	"sync"
	"time"

	"github.com/ironpark/ivent/internal/hook"
	"github.com/ironpark/ivent/key"
)

// EventKind is the type of an input event.
type EventKind uint8

const (
	// KeyDown is a key or mouse button press. It is also reported for auto-repeat (see Event.Repeat).
	KeyDown EventKind = iota + 1
	// KeyUp is a key or mouse button release.
	KeyUp
	// MouseMove is a mouse movement.
	MouseMove
	// MouseWheel is a vertical scroll.
	MouseWheel
)

func (k EventKind) String() string {
	switch k {
	case KeyDown:
		return "KeyDown"
	case KeyUp:
		return "KeyUp"
	case MouseMove:
		return "MouseMove"
	case MouseWheel:
		return "MouseWheel"
	}
	return "Unknown"
}

// Event is an input event. Mouse buttons are reported as KeyDown / KeyUp with key.MouseLeft etc.
type Event struct {
	Kind EventKind
	// Key is the key or mouse button of KeyDown and KeyUp events.
	Key key.Code
	// X and Y are the mouse position (not available on Linux).
	X, Y int32
	// Delta is the scroll amount of MouseWheel events.
	Delta int32
	// Repeat is true for a KeyDown of a key that is already held (auto-repeat).
	Repeat bool
	// Injected is true for events synthesized by software (only delivered with WithInjected).
	Injected bool
	// PID is the process that receives the event, or 0 if unknown.
	PID int
	// Time is when the event was received.
	Time time.Time

	state  key.Table     // keys held after the event
	lookup func(int) App // resolves PID; nil uses the OS
}

// Pressed returns the keys and mouse buttons held down after the event.
func (e Event) Pressed() []key.Code {
	return e.state.PressedKeys()
}

// Combo returns the keys held down after the event as a combination of physical keys,
// for example to display it with Combo.Display or to compare it with Combo.Match.
func (e Event) Combo() key.Combo {
	return e.state.Combo()
}

// IsPressed reports whether a key is held down after the event.
// Side-independent modifiers such as key.Ctrl count either side.
func (e Event) IsPressed(code key.Code) bool {
	return key.NewCombo(code).MatchSubset(e.state)
}

// App describes the application that receives an event.
type App struct {
	PID int
	// Name is the application or executable name, such as "Safari" or "chrome".
	Name string
	// ID is the bundle identifier on macOS (such as "com.apple.Safari") or the executable path on Windows.
	ID string
}

// App returns the application that receives the event. It is empty if it cannot be determined (always on Linux).
func (e Event) App() App {
	if e.lookup != nil {
		return e.lookup(e.PID)
	}
	return lookupProcess(e.PID)
}

// maxCachedApps bounds the app cache; it is cleared when full, since pids are rarely reused.
const maxCachedApps = 256

var (
	appCacheMu sync.Mutex
	appCache   = map[int]App{}
)

func lookupProcess(pid int) App {
	if pid != 0 {
		appCacheMu.Lock()
		app, ok := appCache[pid]
		appCacheMu.Unlock()
		if ok {
			return app
		}
	}
	resolved, name, id, ok := hook.ProcessInfo(pid)
	app := App{PID: resolved, Name: name, ID: id}
	// pid 0 means "frontmost", which changes, so only real pids are cached.
	if ok && pid != 0 {
		appCacheMu.Lock()
		if len(appCache) >= maxCachedApps {
			clear(appCache)
		}
		appCache[pid] = app
		appCacheMu.Unlock()
	}
	return app
}
