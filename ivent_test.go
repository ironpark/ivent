package ivent

import (
	"errors"
	"testing"

	"github.com/ironpark/ivent/internal/hook"
	"github.com/ironpark/ivent/key"
)

func press(codes ...key.Code) {
	for _, c := range codes {
		hook.State.SetKeyState(uint8(c), true)
	}
}

func release(codes ...key.Code) {
	for _, c := range codes {
		hook.State.SetKeyState(uint8(c), false)
	}
}

func setup(t *testing.T) {
	t.Helper()
	Reset()
	ResetKeyState()
	t.Cleanup(func() {
		Reset()
		ResetKeyState()
	})
}

func TestExactMatch(t *testing.T) {
	setup(t)
	count := 0
	Register(NewComb([]key.Code{key.A, key.S, key.D}, func() { count++ }))

	press(key.A, key.S, key.D)
	if count != 1 {
		t.Fatalf("expected 1 call, got %d", count)
	}
	press(key.F)
	if count != 1 {
		t.Fatalf("exact combination must not fire with extra keys, got %d", count)
	}
}

func TestAllowOtherInputs(t *testing.T) {
	setup(t)
	count := 0
	Register(NewComb([]key.Code{key.A, key.S}, func() { count++ }, AllowOtherInputs))

	press(key.D, key.A, key.S)
	if count != 1 {
		t.Fatalf("expected 1 call, got %d", count)
	}
}

func TestAllowRepeat(t *testing.T) {
	setup(t)
	plain, repeat := 0, 0
	Register(
		NewComb([]key.Code{key.A}, func() { plain++ }),
		NewComb([]key.Code{key.A}, func() { repeat++ }, AllowRepeat),
	)

	press(key.A, key.A, key.A)
	if plain != 1 {
		t.Fatalf("expected 1 call without AllowRepeat, got %d", plain)
	}
	if repeat != 3 {
		t.Fatalf("expected 3 calls with AllowRepeat, got %d", repeat)
	}
}

func TestCallbackCanModifyRegistrations(t *testing.T) {
	setup(t)
	var comb *Comb
	comb = NewComb([]key.Code{key.A}, func() {
		Remove(comb)
		Register(NewComb([]key.Code{key.B}, func() {}))
		ResetKeyState()
	})
	Register(comb)

	// Deadlocks if callbacks are invoked while holding a lock.
	press(key.A)
	release(key.A)
	press(key.A)
}

func TestParseComb(t *testing.T) {
	setup(t)
	if _, err := ParseComb("Ctrl+Shitf", func() {}); !errors.Is(err, key.ErrUnknownKey) {
		t.Fatalf("expected ErrUnknownKey, got %v", err)
	}

	count := 0
	comb, err := ParseComb("ctrl + shift + a", func() { count++ })
	if err != nil {
		t.Fatal(err)
	}
	Register(comb)
	press(key.LeftCtrl, key.LeftShift, key.A)
	if count != 1 {
		t.Fatalf("expected 1 call, got %d", count)
	}
}

func TestUnknownNameNeverMatches(t *testing.T) {
	setup(t)
	count := 0
	// Previously "Shitf" resolved to code 0, which is key.A on macOS.
	Register(NewCombFromStr("Ctrl+Shitf", func() { count++ }))
	press(key.LeftCtrl, key.A)
	if count != 0 {
		t.Fatalf("combination with unknown key must not fire, got %d", count)
	}
}

func TestSideIndependentModifier(t *testing.T) {
	setup(t)
	any, right := 0, 0
	Register(
		NewComb([]key.Code{key.Ctrl, key.C}, func() { any++ }),
		NewCombFromStr("RightCtrl+C", func() { right++ }),
	)

	press(key.RightCtrl, key.C)
	release(key.RightCtrl, key.C)
	press(key.LeftCtrl, key.C)
	release(key.LeftCtrl, key.C)
	if any != 2 {
		t.Fatalf("Ctrl+C must match both sides, got %d", any)
	}
	if right != 1 {
		t.Fatalf("RightCtrl+C must match only the right side, got %d", right)
	}

	// Both sides held still count as Ctrl.
	press(key.LeftCtrl, key.RightCtrl, key.C)
	if any != 3 {
		t.Fatalf("Ctrl+C must match with both sides held, got %d", any)
	}
}

func TestSideIndependentModifierExact(t *testing.T) {
	setup(t)
	count := 0
	Register(NewComb([]key.Code{key.Ctrl, key.C}, func() { count++ }))

	press(key.LeftCtrl, key.LeftShift, key.C)
	if count != 0 {
		t.Fatalf("Ctrl+C must not match Ctrl+Shift+C, got %d", count)
	}
}
