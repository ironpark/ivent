package hook

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync/atomic"
	"syscall"
	"unicode"
	"unsafe"

	"github.com/ironpark/ivent/key"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	procSetWindowsHookExW        = user32.NewProc("SetWindowsHookExW")
	procUnhookWindowsHookEx      = user32.NewProc("UnhookWindowsHookEx")
	procCallNextHookEx           = user32.NewProc("CallNextHookEx")
	procGetMessageW              = user32.NewProc("GetMessageW")
	procPeekMessageW             = user32.NewProc("PeekMessageW")
	procTranslateMessage         = user32.NewProc("TranslateMessage")
	procDispatchMessageW         = user32.NewProc("DispatchMessageW")
	procPostThreadMessageW       = user32.NewProc("PostThreadMessageW")
	procGetAsyncKeyState         = user32.NewProc("GetAsyncKeyState")
	procGetCursorPos             = user32.NewProc("GetCursorPos")
	procGetForegroundWindow      = user32.NewProc("GetForegroundWindow")
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	procGetKeyboardLayout        = user32.NewProc("GetKeyboardLayout")
	procMapVirtualKeyW           = user32.NewProc("MapVirtualKeyW")
	procMapVirtualKeyExW         = user32.NewProc("MapVirtualKeyExW")
	procSendInput                = user32.NewProc("SendInput")

	procGetCurrentThreadId         = kernel32.NewProc("GetCurrentThreadId")
	procQueryFullProcessImageNameW = kernel32.NewProc("QueryFullProcessImageNameW")
)

const (
	whKeyboardLL = 13
	whMouseLL    = 14
	hcAction     = 0

	wmQuit        = 0x0012
	wmUser        = 0x0400
	wmKeyDown     = 0x0100
	wmSysKeyDown  = 0x0104
	wmMouseMove   = 0x0200
	wmLButtonDown = 0x0201
	wmLButtonUp   = 0x0202
	wmRButtonDown = 0x0204
	wmRButtonUp   = 0x0205
	wmMButtonDown = 0x0207
	wmMButtonUp   = 0x0208
	wmMouseWheel  = 0x020A
	wmXButtonDown = 0x020B
	wmXButtonUp   = 0x020C

	llkhfInjected = 0x10
	llmhfInjected = 0x01
	xButton1      = 1
	wheelDelta    = 120
	vkLControl    = 0xA2

	inputKeyboard        = 1
	keyeventfExtendedKey = 0x1
	keyeventfKeyUp       = 0x2
	mapvkVKToVSC         = 0
	mapvkVKToChar        = 2

	processQueryLimitedInformation = 0x1000
)

type point struct{ X, Y int32 }

// kbdllHookStruct is KBDLLHOOKSTRUCT.
type kbdllHookStruct struct {
	VKCode, ScanCode, Flags, Time uint32
	ExtraInfo                     uintptr
}

// msllHookStruct is MSLLHOOKSTRUCT.
type msllHookStruct struct {
	Pt                     point
	MouseData, Flags, Time uint32
	ExtraInfo              uintptr
}

// msg is MSG.
type msg struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      point
	Private uint32
}

// keyboardInput is an INPUT holding a KEYBDINPUT. The nested struct gets the union's alignment.
type keyboardInput struct {
	Type uint32
	Ki   keybdInput
	_    [8]byte // the union is as large as MOUSEINPUT
}

// keybdInput is KEYBDINPUT.
type keybdInput struct {
	VK, Scan  uint16
	Flags     uint32
	Time      uint32
	ExtraInfo uintptr
}

var (
	// The callbacks are created once: Windows callbacks are never freed.
	keyboardProc = syscall.NewCallback(onKeyboard)
	mouseProc    = syscall.NewCallback(onMouse)

	hookThreadID atomic.Uintptr // thread running the message loop, 0 when stopped
)

// call calls a Windows function. Unlike LazyProc.Call, it does not allocate, which matters in the
// hook callbacks that run for every event.
func call(p *syscall.LazyProc, args ...uintptr) uintptr {
	r, _, _ := syscall.SyscallN(p.Addr(), args...)
	return r
}

func eventFlags(flags, injectedFlag uint32, extra uintptr) uint8 {
	if extra == ownEventTag {
		return FlagInjected | FlagOwn
	}
	if flags&injectedFlag != 0 {
		return FlagInjected
	}
	return 0
}

// isAltGrCtrl reports whether an event is the fake Left Ctrl (scan code 0x21D) that AltGr sends before
// Right Alt. It is ignored so that AltGr is not mistaken for Ctrl+Alt.
func isAltGrCtrl(kbd *kbdllHookStruct) bool {
	return kbd.VKCode == vkLControl && kbd.ScanCode&0x200 != 0
}

func onKeyboard(nCode, wParam uintptr, kbd *kbdllHookStruct) uintptr {
	if int32(nCode) == hcAction && kbd.VKCode < 256 && !isAltGrCtrl(kbd) {
		kind := KeyUp
		if wParam == wmKeyDown || wParam == wmSysKeyDown {
			kind = KeyDown
		}
		// Keyboard events carry no position; report the cursor like the other platforms do.
		var pt point
		call(procGetCursorPos, uintptr(unsafe.Pointer(&pt)))
		ev := Event{Kind: kind, Code: key.Code(kbd.VKCode), X: pt.X, Y: pt.Y, PID: foregroundPID(),
			Flags: eventFlags(kbd.Flags, llkhfInjected, kbd.ExtraInfo)}
		if dispatch(ev) {
			return 1
		}
	}
	// The hook handle argument is ignored.
	return call(procCallNextHookEx, 0, nCode, wParam, uintptr(unsafe.Pointer(kbd)))
}

func onMouse(nCode, wParam uintptr, m *msllHookStruct) uintptr {
	if int32(nCode) == hcAction {
		ev := Event{X: m.Pt.X, Y: m.Pt.Y, Flags: eventFlags(m.Flags, llmhfInjected, m.ExtraInfo)}
		button := -1
		switch wParam {
		case wmLButtonDown, wmLButtonUp:
			button, ev.Kind = 0, keyKind(wParam == wmLButtonDown)
		case wmRButtonDown, wmRButtonUp:
			button, ev.Kind = 1, keyKind(wParam == wmRButtonDown)
		case wmMButtonDown, wmMButtonUp:
			button, ev.Kind = 2, keyKind(wParam == wmMButtonDown)
		case wmXButtonDown, wmXButtonUp:
			button, ev.Kind = 4, keyKind(wParam == wmXButtonDown)
			if m.MouseData>>16 == xButton1 {
				button = 3
			}
		case wmMouseMove:
			ev.Kind = MouseMove
		case wmMouseWheel:
			ev.Kind = MouseWheel
			ev.Delta = int32(int16(m.MouseData>>16)) / wheelDelta
		}
		if button >= 0 {
			ev.Code = key.MouseButtons[button]
			// The target process only matters for buttons; skip the lookup for frequent moves.
			ev.PID = foregroundPID()
		}
		if ev.Kind != 0 && dispatch(ev) {
			return 1
		}
	}
	return call(procCallNextHookEx, 0, nCode, wParam, uintptr(unsafe.Pointer(m)))
}

// platformStart installs the hooks and runs the message loop that delivers their events.
// It runs on a locked OS thread, which the hooks and the loop must share.
func platformStart(cfg Config) error {
	var m msg
	// Make sure this thread has a message queue before platformStop can post WM_QUIT to it.
	call(procPeekMessageW, uintptr(unsafe.Pointer(&m)), 0, wmUser, wmUser, 0)

	var hooks []uintptr
	// Hooks are removed by the thread that installed them.
	defer func() {
		for _, h := range hooks {
			call(procUnhookWindowsHookEx, h)
		}
	}()
	install := func(id int, proc uintptr) error {
		h, _, err := procSetWindowsHookExW.Call(uintptr(id), proc, 0, 0)
		if h == 0 {
			return fmt.Errorf("%w: SetWindowsHookEx failed: %v", ErrHookFailed, err)
		}
		hooks = append(hooks, h)
		return nil
	}
	if cfg.Keyboard {
		if err := install(whKeyboardLL, keyboardProc); err != nil {
			return err
		}
	}
	if cfg.Mouse {
		if err := install(whMouseLL, mouseProc); err != nil {
			return err
		}
	}

	hookThreadID.Store(call(procGetCurrentThreadId))
	defer hookThreadID.Store(0)
	// Low-level hooks can always block events.
	ready(true)
	for int32(call(procGetMessageW, uintptr(unsafe.Pointer(&m)), 0, 0, 0)) > 0 {
		call(procTranslateMessage, uintptr(unsafe.Pointer(&m)))
		call(procDispatchMessageW, uintptr(unsafe.Pointer(&m)))
	}
	return nil
}

// platformStop can be called from any thread once platformStart has signalled ready.
// WM_QUIT stays queued until the message loop reads it.
func platformStop() {
	if tid := hookThreadID.Load(); tid != 0 {
		call(procPostThreadMessageW, tid, wmQuit, 0, 0)
	}
}

func foregroundPID() int {
	window := call(procGetForegroundWindow)
	if window == 0 {
		return 0
	}
	var pid uint32
	call(procGetWindowThreadProcessId, window, uintptr(unsafe.Pointer(&pid)))
	return int(pid)
}

// charForKey returns the character a key types without modifiers on the foreground window's layout.
func charForKey(code key.Code) (rune, bool) {
	window := call(procGetForegroundWindow)
	thread := call(procGetWindowThreadProcessId, window, 0)
	layout := call(procGetKeyboardLayout, thread)
	ch := call(procMapVirtualKeyExW, uintptr(code), mapvkVKToChar, layout)
	// The high bit marks a dead key.
	if ch == 0 || ch&0x80000000 != 0 {
		return 0, false
	}
	return unicode.ToLower(rune(uint32(ch))), true
}

// KeyState reports whether the OS considers the key to be held down. ok is false if the OS cannot tell.
// Mouse button key codes are virtual-key codes on Windows, so they need no special case.
func KeyState(code key.Code) (pressed, ok bool) {
	r := call(procGetAsyncKeyState, uintptr(code))
	return r&0x8000 != 0, true
}

// extendedKeys are the keys that need KEYEVENTF_EXTENDEDKEY.
var extendedKeys = map[key.Code]bool{
	key.RightCtrl: true, key.RightAlt: true, key.LeftSuper: true, key.RightSuper: true, key.Menu: true,
	key.Insert: true, key.Delete: true, key.Home: true, key.End: true, key.PageUp: true, key.PageDown: true,
	key.ArrowLeft: true, key.ArrowUp: true, key.ArrowRight: true, key.ArrowDown: true,
	key.NumLock: true, key.PadSlash: true, key.PrintScreen: true,
}

// Send sends a synthetic key event.
func Send(code key.Code, down bool) error {
	scan := call(procMapVirtualKeyW, uintptr(code), mapvkVKToVSC)
	in := keyboardInput{Type: inputKeyboard, Ki: keybdInput{VK: uint16(code), Scan: uint16(scan), ExtraInfo: ownEventTag}}
	if !down {
		in.Ki.Flags |= keyeventfKeyUp
	}
	if extendedKeys[code] {
		in.Ki.Flags |= keyeventfExtendedKey
	}
	n, _, err := procSendInput.Call(1, uintptr(unsafe.Pointer(&in)), unsafe.Sizeof(in))
	if n != 1 {
		return fmt.Errorf("ivent: failed to send key %s: %v", code.Name(), err)
	}
	return nil
}

// ProcessInfo returns the executable name and path of a process.
// A pid of 0 resolves to the process of the foreground window.
func ProcessInfo(pid int) (resolvedPID int, name, id string, ok bool) {
	if pid == 0 {
		pid = foregroundPID()
	}
	process, err := syscall.OpenProcess(processQueryLimitedInformation, false, uint32(pid))
	if err != nil {
		return pid, "", "", false
	}
	defer syscall.CloseHandle(process)
	var buf [syscall.MAX_PATH * 2]uint16
	size := uint32(len(buf))
	if r := call(procQueryFullProcessImageNameW, uintptr(process), 0, uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size))); r == 0 {
		return pid, "", "", false
	}
	path := syscall.UTF16ToString(buf[:size])
	base := filepath.Base(path)
	return pid, strings.TrimSuffix(base, filepath.Ext(base)), path, true
}

// Permissions reports whether the process may listen to input (monitor) and block or send input (control).
// Windows does not restrict low-level hooks, so both are always true.
func Permissions(prompt bool) (monitor, control bool) {
	return true, true
}
