package key

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Code int

// Invalid stands for a key that does not exist: an unresolved name, or a key that the current
// platform does not have (such as Fn on Windows). It is never reported, so a combination containing it never matches.
const Invalid = Code(0xFF)

// ErrUnknownKey is returned when a key name cannot be resolved to a key code.
var ErrUnknownKey = errors.New("unknown key name")

func (k Code) Name() string {
	return Name(k)
}

// Known reports whether the key code has a name in the key table of the current platform.
func Known(code Code) bool {
	_, ok := codeToName[code]
	return ok
}

// MouseButtons lists the mouse buttons in the order the OS numbers them (left, right, middle, extra 1, extra 2).
var MouseButtons = [...]Code{MouseLeft, MouseRight, MouseMiddle, MouseX1, MouseX2}

// IsMouseButton reports whether code is a mouse button.
func IsMouseButton(code Code) bool {
	for _, button := range MouseButtons {
		if code == button {
			return true
		}
	}
	return false
}

// IsKeypad reports whether code is a numeric keypad key.
func IsKeypad(code Code) bool {
	return strings.HasPrefix(codeToDisplay[code], "Pad")
}

func Name(code Code) string {
	name, ok := codeToName[code]
	if ok {
		return name
	}
	return fmt.Sprintf("Unk%d", code)
}

// Names converts a slice of key codes to a slice of key names.
// If the key code is not found, it will be named as "Unk" followed by the code.
func Names(codes ...Code) []string {
	names := make([]string, len(codes))
	for i, code := range codes {
		names[i] = Name(code)
	}
	return names
}

// Parse converts a key name to a key code.
// Names are case-insensitive and surrounding whitespace is ignored.
// "Unk<N>" resolves to the raw key code N (0-255).
func Parse(name string) (Code, error) {
	upper := strings.ToUpper(strings.TrimSpace(name))
	if code, ok := nameToCode[upper]; ok {
		return code, nil
	}
	if after, found := strings.CutPrefix(upper, "UNK"); found {
		code, err := strconv.Atoi(after)
		if err == nil && code >= 0 && code < tableSize {
			return Code(code), nil
		}
	}
	return Invalid, fmt.Errorf("%w: %q", ErrUnknownKey, name)
}
