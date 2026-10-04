package ivent

import (
	"context"
	"github.com/ironpark/ivent/internal/hook"
	"github.com/ironpark/ivent/key"
	"sync"
)

var (
	combinations []*Comb
	mu           = sync.RWMutex{}
)

type Opt int

const (
	AllowOtherInputs = Opt(0x01)
	AllowRepeat      = Opt(0x02)
)

type Comb struct {
	Combination key.Combo
	Callback    func()
	Option      Opt
}

func init() {
	// Register the update callback to be called whenever the key state is updated.
	hook.State.AddUpdateCallback(update)
}

// Errors returned by Start.
var (
	ErrAlreadyRunning = hook.ErrAlreadyRunning
	ErrHookFailed     = hook.ErrHookFailed
)

// ParseComb creates a new Comb from a string of key names such as "Ctrl+Shift+A".
// It returns an error wrapping key.ErrUnknownKey if a key name cannot be resolved.
func ParseComb(codes string, callback func(), options ...Opt) (*Comb, error) {
	combo, err := key.ParseCombo(codes)
	if err != nil {
		return nil, err
	}
	return newComb(combo, callback, options), nil
}

// NewCombFromStr creates a new Comb from a string of key codes.
// Example input: "A+S+D"
// It splits the string by "+" and creates a key combination.
// A combination containing an unknown key name never matches; use ParseComb to detect it.
func NewCombFromStr(codes string, callback func(), options ...Opt) *Comb {
	combo, _ := key.ParseCombo(codes) // an empty Combo on error, which never matches
	return newComb(combo, callback, options)
}

// NewComb creates a new Comb from a slice of key codes.
// It creates a key combination and assigns the provided callback and options.
// Side-independent modifiers such as key.Ctrl match either the left or the right key.
func NewComb(codes []key.Code, callback func(), options ...Opt) *Comb {
	return newComb(key.NewCombo(codes...), callback, options)
}

func newComb(combo key.Combo, callback func(), options []Opt) *Comb {
	var option Opt
	for _, opt := range options {
		option |= opt
	}
	return &Comb{
		Combination: combo,
		Callback:    callback,
		Option:      option,
	}
}

// update is the callback function that gets called whenever the key state is updated.
// It checks the current state against registered combinations and triggers the appropriate callbacks.
// Callbacks run without holding the lock, so they may call Register, Remove or Reset.
func update(state key.Table, pressed bool, repeat bool) {
	if !pressed {
		return
	}
	mu.RLock()
	var matched []func()
	for _, comb := range combinations {
		if repeat && comb.Option&AllowRepeat == 0 {
			continue
		}
		// An exact match is also a subset match, so one check is enough.
		match := comb.Combination.Match
		if comb.Option&AllowOtherInputs != 0 {
			match = comb.Combination.MatchSubset
		}
		if match(state) {
			matched = append(matched, comb.Callback)
		}
	}
	mu.RUnlock()
	for _, callback := range matched {
		callback()
	}
}

// MouseXY returns the current x and y position of the mouse pointer.
func MouseXY() (X int32, Y int32) {
	return hook.MouseX(), hook.MouseY()
}

// MouseX returns the current x position of the mouse pointer.
func MouseX() int32 {
	return hook.MouseX()
}

// MouseY returns the current y position of the mouse pointer.
func MouseY() int32 {
	return hook.MouseY()
}

// Start begins listening for input events and blocks until the context is done.
// Combination callbacks are invoked on the goroutine that called Start.
// It returns ErrHookFailed if the OS hook cannot be installed (on macOS, usually missing
// Accessibility / Input Monitoring permission) and ErrAlreadyRunning if it is already running.
func Start(ctx context.Context) error {
	return hook.Start(ctx)
}

// ResetKeyState resets the key state to all keys being unpressed.
func ResetKeyState() {
	hook.State.Reset()
}

// Register registers new key combinations to be monitored.
// It adds the combinations to the list of monitored combinations.
func Register(comb ...*Comb) {
	mu.Lock()
	defer mu.Unlock()
	combinations = append(combinations, comb...)
}

// Reset clears the list of monitored combinations.
func Reset() {
	mu.Lock()
	defer mu.Unlock()
	combinations = nil
}

// Remove removes a specific combination from the list of monitored combinations.
func Remove(comb *Comb) {
	mu.Lock()
	defer mu.Unlock()
	for i, c := range combinations {
		if c == comb {
			combinations = append(combinations[:i], combinations[i+1:]...)
			return
		}
	}
}
