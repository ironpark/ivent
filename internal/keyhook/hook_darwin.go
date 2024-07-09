package keyhook

/*
#cgo CFLAGS: -x objective-c -Wimplicit-function-declaration
#cgo LDFLAGS: -framework Cocoa
#include <stdlib.h>
#include "hook_darwin.h"
*/
import "C"
import (
	"context"
	"fmt"
	"runtime"
)

func start() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	fmt.Println(C.start())
}

func Start(ctx context.Context) {
	go start()
	<-ctx.Done()
	C.stop()
}

func Stop() {
	C.stop()
}
