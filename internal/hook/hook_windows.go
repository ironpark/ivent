package hook

/*
#cgo LDFLAGS: -luser32
#include "hook_common.h"
#include "hook_windows.h"
*/
import "C"

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ironpark/ivent/key"
)

func platformStart(cfg Config) error {
	if code := C.start(C.bool(cfg.Keyboard), C.bool(cfg.Mouse)); code != 0 {
		return fmt.Errorf("%w: SetWindowsHookEx failed (error %d)", ErrHookFailed, int(code))
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
// Mouse button key codes are virtual-key codes on Windows, so they need no special case.
func KeyState(code key.Code) (pressed, ok bool) {
	return bool(C.keyPressed(C.int(code))), true
}

// Send sends a synthetic key event.
func Send(code key.Code, down bool) error {
	if errCode := C.sendKey(C.int(code), C.bool(down)); errCode != 0 {
		return fmt.Errorf("ivent: failed to send key %s (error %d)", code.Name(), int(errCode))
	}
	return nil
}

// ProcessInfo returns the executable name and path of a process.
// A pid of 0 resolves to the process of the foreground window.
func ProcessInfo(pid int) (resolvedPID int, name, id string, ok bool) {
	if pid == 0 {
		pid = int(C.foregroundPID())
	}
	var buf [1024]C.char
	if !C.processPath(C.int(pid), &buf[0], C.int(len(buf))) {
		return pid, "", "", false
	}
	path := C.GoString(&buf[0])
	base := filepath.Base(path)
	return pid, strings.TrimSuffix(base, filepath.Ext(base)), path, true
}

// Permissions reports whether the process may listen to input (monitor) and block or send input (control).
// Windows does not restrict low-level hooks, so both are always true.
func Permissions(prompt bool) (monitor, control bool) {
	return true, true
}
