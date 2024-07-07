package keyhook

/*
#cgo CFLAGS: -x objective-c -Wimplicit-function-declaration
#cgo LDFLAGS: -framework Cocoa
#include <stdlib.h>
#include "hook_darwin.h"
*/
import "C"

func Start() {
	C.start()
}

func Stop() {
	C.stop()
}
