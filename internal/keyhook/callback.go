package keyhook

// #include "stdint.h"
import "C"

type keyEvent uint16

var eventCh = make(chan keyEvent, 10)

func (kev keyEvent) keyCode() uint8 {
	//keyEvent(keyCode)<<1 | keyEvent(down)
	return uint8(kev >> 1)
}

func (kev keyEvent) down() bool {
	//keyEvent(keyCode)<<1 | keyEvent(down)
	return kev&1 == 1
}

//export keyEventGoCallback
func keyEventGoCallback(pid C.int, keyCode C.uint8_t, down C.uint8_t) {
	// combine keyCode and down into a single uint16
	eventCh <- keyEvent(keyCode)<<1 | keyEvent(down)
}
