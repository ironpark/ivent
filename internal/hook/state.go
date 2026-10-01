package hook

import (
	"github.com/ironpark/ivent/key"
	"sync"
)

var State = NewKeyState()

// UpdateCallback is called with a snapshot of the key states.
// repeat is true when a key that is already down is reported down again (auto-repeat).
type UpdateCallback func(state key.Table, pressed bool, repeat bool)

// KeyState is a struct that stores the current key states.
type KeyState struct {
	states         key.Table
	mu             sync.RWMutex
	updateCallback []UpdateCallback
}

// NewKeyState creates and returns a new KeyState.
func NewKeyState() *KeyState {
	return &KeyState{}
}

// Reset resets the key states to the initial state.
func (ks *KeyState) Reset() {
	ks.mu.Lock()
	defer ks.mu.Unlock()
	ks.states = key.Table{}
}

// AddUpdateCallback adds a new callback to be called when the key states are updated.
func (ks *KeyState) AddUpdateCallback(callback UpdateCallback) {
	ks.mu.Lock()
	defer ks.mu.Unlock()
	ks.updateCallback = append(ks.updateCallback, callback)
}

// SetKeyState sets the state of a specific key.
// Callbacks are invoked after the lock is released, so they may safely call back into KeyState.
func (ks *KeyState) SetKeyState(keycode uint8, down bool) bool {
	// ignore unknown keys
	if !key.Known(key.Code(keycode)) {
		return false
	}
	ks.mu.Lock()
	updated := ks.states.Set(key.Code(keycode), down)
	snapshot := ks.states
	callbacks := ks.updateCallback
	ks.mu.Unlock()

	// A down event that changes nothing is an auto-repeat.
	if updated || down {
		for _, callback := range callbacks {
			callback(snapshot, down, !updated)
		}
	}
	return updated
}

// IsKeyPressed checks if a specific key is pressed.
func (ks *KeyState) IsKeyPressed(keycode key.Code) bool {
	ks.mu.RLock()
	defer ks.mu.RUnlock()
	return ks.states.IsKeyPressed(keycode)
}

// CheckKeyCombination checks if all specified key codes are pressed.
func (ks *KeyState) CheckKeyCombination(keyCodes ...key.Code) bool {
	ks.mu.RLock()
	defer ks.mu.RUnlock()
	return ks.states.CheckKeyCombination(keyCodes...)
}
