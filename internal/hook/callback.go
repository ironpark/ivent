//go:build darwin || windows

package hook

// #include "hook_common.h"
import "C"

import "github.com/ironpark/ivent/key"

//export goEvent
func goEvent(kind, code, x, y, delta, pid, flags C.int) C.int {
	ev := Event{Kind: Kind(kind), Code: key.Code(code), X: int32(x), Y: int32(y), Delta: int32(delta), PID: int(pid)}
	ev.Flags = uint8(flags &^ C.IVENT_FLAG_BUTTON)
	if flags&C.IVENT_FLAG_BUTTON != 0 {
		if code < 0 || int(code) >= len(key.MouseButtons) {
			return 0
		}
		ev.Code = key.MouseButtons[code]
	}
	if dispatch(ev) {
		return 1
	}
	return 0
}

//export goReset
func goReset() {
	reset()
}

//export goReady
func goReady(suppress C.bool) {
	ready(bool(suppress))
}
