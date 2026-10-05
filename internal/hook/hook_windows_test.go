package hook

import (
	"testing"
	"unsafe"
)

// TestStructLayout checks the Go copies of Windows structs against their C sizes.
func TestStructLayout(t *testing.T) {
	is64 := unsafe.Sizeof(uintptr(0)) == 8
	pick := func(size64, size32 uintptr) uintptr {
		if is64 {
			return size64
		}
		return size32
	}
	var in keyboardInput
	for _, c := range []struct {
		name      string
		got, want uintptr
	}{
		{"INPUT", unsafe.Sizeof(in), pick(40, 28)},
		{"INPUT.ki offset", unsafe.Offsetof(in.Ki), pick(8, 4)},
		{"INPUT.ki.dwExtraInfo offset", unsafe.Offsetof(in.Ki) + unsafe.Offsetof(in.Ki.ExtraInfo), pick(24, 16)},
		{"KBDLLHOOKSTRUCT", unsafe.Sizeof(kbdllHookStruct{}), pick(24, 20)},
		{"KBDLLHOOKSTRUCT.dwExtraInfo offset", unsafe.Offsetof(kbdllHookStruct{}.ExtraInfo), 16},
		{"MSLLHOOKSTRUCT", unsafe.Sizeof(msllHookStruct{}), pick(32, 24)},
		{"MSLLHOOKSTRUCT.mouseData offset", unsafe.Offsetof(msllHookStruct{}.MouseData), 8},
		{"MSLLHOOKSTRUCT.dwExtraInfo offset", unsafe.Offsetof(msllHookStruct{}.ExtraInfo), pick(24, 20)},
		// Including lPrivate, which older SDKs only declare on _MAC; the extra space is unused.
		{"MSG", unsafe.Sizeof(msg{}), pick(48, 32)},
	} {
		if c.got != c.want {
			t.Errorf("sizeof %s = %d, want %d", c.name, c.got, c.want)
		}
	}
}
