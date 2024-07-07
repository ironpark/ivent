package keyhook

import (
	"github.com/ironpark/ivent/key"
	"sync"
)

var State = NewKeyState()

type UpdateCallback func(state key.Table, pressed bool)

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
	ks.states = [4]uint64{
		0, 0, 0, 0,
	}
}

// AddUpdateCallback adds a new callback to be called when the key states are updated.
func (ks *KeyState) AddUpdateCallback(callback UpdateCallback) {
	ks.mu.Lock()
	defer ks.mu.Unlock()
	ks.updateCallback = append(ks.updateCallback, callback)
}

// RemoveUpdateCallback removes a callback from the list of update callbacks.
func (ks *KeyState) RemoveUpdateCallback(callback UpdateCallback) {
	ks.mu.Lock()
	defer ks.mu.Unlock()
	for i, cb := range ks.updateCallback {
		if &cb == &callback {
			ks.updateCallback = append(ks.updateCallback[:i], ks.updateCallback[i+1:]...)
			break
		}
	}
}

// SetKeyState sets the state of a specific key.
func (ks *KeyState) SetKeyState(keycode int, down bool) bool {
	byteIndex := keycode / 64
	bitIndex := uint(keycode % 64)
	mask := uint64(1 << bitIndex)

	ks.mu.Lock()
	defer ks.mu.Unlock()

	oldState := ks.states[byteIndex]
	newState := oldState

	if down {
		newState |= mask
	} else {
		newState &^= mask
	}

	ks.states[byteIndex] = newState
	updated := oldState != newState
	if updated && len(ks.updateCallback) > 0 {
		for _, callback := range ks.updateCallback {
			callback(ks.states, down)
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
