package hook

/*
#cgo CFLAGS: -I.
#cgo LDFLAGS: -L. -luser32
#include <stdlib.h>
#include "hook_windows.h"
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
	fmt.Println(C.start(C.LISTEN_MOUSEANDKEYBOARD))
}

func Start(ctx context.Context) {
	go start()
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-eventCh:
				State.SetKeyState(event.keyCode(), event.down())
			}
		}
	}(ctx)
	<-ctx.Done()
	C.stop()
}

func Stop() {
	C.stop()
}
