package hook

// #include "stdint.h"
import "C"
import (
	"sync/atomic"
)

type keyEvent uint16

var (
	mouseX int32
	mouseY int32
)

var eventCh = make(chan keyEvent, 10)

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
	eventCh <- keyEvent(keyCode)<<1 | keyEvent(down)
}

//export mouseEventGoCallback
func mouseEventGoCallback(pid C.int, x C.int, y C.int, button C.int, down C.int) {
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
