package hook

/*
#cgo CFLAGS: -x objective-c -Wimplicit-function-declaration
#cgo LDFLAGS: -framework Cocoa -framework IOKit -framework Carbon
#include "hook_common.h"
#include "hook_darwin.h"
*/
import "C"

import (
	"fmt"

	"github.com/ironpark/ivent/key"
)

func platformStart(cfg Config) error {
	if C.start(C.bool(cfg.Keyboard), C.bool(cfg.Mouse)) != 0 {
		return fmt.Errorf("%w: grant Input Monitoring permission to this process", ErrHookFailed)
	}
	return nil
}

func charForKey(code key.Code) (rune, bool) {
	c := C.charForKey(C.int(code))
	return rune(c), c >= 0
}

func platformStop() {
	C.stop()
}

// KeyState reports whether the OS considers the key to be held down. ok is false if the OS cannot tell.
func KeyState(code key.Code) (pressed, ok bool) {
	for i, button := range key.MouseButtons {
		if code == button {
			return bool(C.buttonPressed(C.int(i))), true
		}
	}
	return bool(C.keyPressed(C.int(code))), true
}

// Send posts a synthetic key event.
func Send(code key.Code, down bool) error {
	if C.sendKey(C.int(code), C.bool(down)) != 0 {
		return fmt.Errorf("ivent: failed to send key %s", code.Name())
	}
	return nil
}

// ProcessInfo returns the display name and an identifier (bundle ID or path) of a process.
// A pid of 0 resolves to the frontmost application.
func ProcessInfo(pid int) (resolvedPID int, name, id string, ok bool) {
	if pid == 0 {
		pid = int(C.frontmostPID())
	}
	var nameBuf, idBuf [1024]C.char
	if !C.processInfo(C.int(pid), &nameBuf[0], C.int(len(nameBuf)), &idBuf[0], C.int(len(idBuf))) {
		return pid, "", "", false
	}
	return pid, C.GoString(&nameBuf[0]), C.GoString(&idBuf[0]), true
}

// Permissions reports whether the process may listen to input (monitor) and block or send input (control).
// If prompt is true, the OS asks the user for missing permissions.
func Permissions(prompt bool) (monitor, control bool) {
	var m, c C.bool
	C.permissions(C.bool(prompt), &m, &c)
	return bool(m), bool(c)
}
