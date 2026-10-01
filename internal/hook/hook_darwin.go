package hook

/*
#cgo CFLAGS: -x objective-c -Wimplicit-function-declaration
#cgo LDFLAGS: -framework Cocoa
#include <stdlib.h>
#include "hook_darwin.h"
*/
import "C"
import "fmt"

func platformStart() error {
	if C.start(C.LISTEN_MOUSEANDKEYBOARD) != 0 {
		return fmt.Errorf("%w: grant Accessibility or Input Monitoring permission to this process", ErrHookFailed)
	}
	return nil
}

func platformStop() {
	C.stop()
}
