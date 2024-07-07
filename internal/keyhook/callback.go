package keyhook

// #include "stdint.h"
import "C"

//export keyEventGoCallback
func keyEventGoCallback(pid C.int, keyCode C.uint8_t, down C.uint8_t) {
	State.SetKeyState(int(keyCode), down == 1)
}
