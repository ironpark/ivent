package ivent

import (
	"fmt"

	"github.com/ironpark/ivent/internal/hook"
	"github.com/ironpark/ivent/key"
)

// Send presses the keys of the combination in order and releases them in reverse order.
// Side-independent modifiers such as key.Ctrl are sent as their left key.
// Events sent by this process never trigger its own bindings. On Linux, keys are sent through a
// virtual uinput device, which needs write access to /dev/uinput.
func Send(c key.Combo) error {
	codes := c.Codes()
	for i, code := range codes {
		if err := Press(code); err != nil {
			releaseAll(codes[:i])
			return err
		}
	}
	return releaseAll(codes)
}

func releaseAll(codes []key.Code) error {
	var first error
	for i := len(codes) - 1; i >= 0; i-- {
		if err := Release(codes[i]); err != nil && first == nil {
			first = err
		}
	}
	return first
}

// SendString parses a combination such as "Ctrl+C" and sends it.
func SendString(s string) error {
	c, err := key.ParseCombo(s)
	if err != nil {
		return err
	}
	return Send(c)
}

// Press sends a key press. Mouse buttons are not supported.
func Press(code key.Code) error {
	return sendKey(code, true)
}

// Release sends a key release. Mouse buttons are not supported.
func Release(code key.Code) error {
	return sendKey(code, false)
}

func sendKey(code key.Code, down bool) error {
	code = key.Physical(code)
	if !key.Known(code) || key.IsMouseButton(code) {
		return fmt.Errorf("%w: cannot send %s", ErrUnsupported, code.Name())
	}
	return hook.Send(code, down)
}
