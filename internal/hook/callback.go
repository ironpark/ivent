package hook

// #include <stdint.h>
// #include <stdbool.h>
import "C"
import (
	"sync/atomic"
)

type keyEvent uint16

var (
	mouseX int32
	mouseY int32
)

// eventCh is buffered so that the OS hook thread is never blocked by slow Go-side processing.
var eventCh = make(chan keyEvent, 1024)

func (kev keyEvent) keyCode() uint8 {
	return uint8(kev >> 1)
}

func (kev keyEvent) down() bool {
	return kev&1 == 1
}

func MouseX() int32 {
	return atomic.LoadInt32(&mouseX)
}

func MouseY() int32 {
	return atomic.LoadInt32(&mouseY)
}

//export keyEventGoCallback
func keyEventGoCallback(pid C.int, keyCode C.uint8_t, down C.uint8_t) {
	// combine keyCode and down into a single uint16
	ev := keyEvent(keyCode)<<1 | keyEvent(down)
	// Never block the OS hook thread: macOS disables a slow event tap and
	// Windows stalls system-wide input while a low-level hook is blocked.
	select {
	case eventCh <- ev:
	default:
	}
}

// readyCh is signalled by the OS hook thread once the hook is installed.
var readyCh = make(chan struct{}, 1)

//export hookReadyGoCallback
func hookReadyGoCallback() {
	readyCh <- struct{}{}
}

//export mouseEventGoCallback
func mouseEventGoCallback(pid C.int, x C.int, y C.int, button C.int, down C.bool) {
	atomic.StoreInt32(&mouseX, int32(x))
	atomic.StoreInt32(&mouseY, int32(y))
}

//export mouseMoveGoCallback
func mouseMoveGoCallback(pid C.int, x C.int, y C.int) {
	atomic.StoreInt32(&mouseX, int32(x))
	atomic.StoreInt32(&mouseY, int32(y))
}

//export mouseWheelGoCallback
func mouseWheelGoCallback(pid C.int, x C.int, y C.int, deltaY C.int) {
}
