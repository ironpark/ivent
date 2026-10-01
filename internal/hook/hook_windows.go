package hook

/*
#cgo CFLAGS: -I.
#cgo LDFLAGS: -luser32
#include <stdlib.h>
#include "hook_windows.h"
*/
import "C"
import "fmt"

func platformStart() error {
	if code := C.start(C.LISTEN_MOUSEANDKEYBOARD); code != 0 {
		return fmt.Errorf("%w: SetWindowsHookEx failed (error %d)", ErrHookFailed, int(code))
	}
	return nil
}

func platformStop() {
	C.stop()
}
