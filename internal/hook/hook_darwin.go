package hook

import (
	"fmt"
	"runtime"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"

	"github.com/ironpark/ivent/key"
)

// The system frameworks are called through purego, so building needs no cgo.

// CGEventType values (CoreGraphics/CGEventTypes.h).
const (
	cgLeftMouseDown          = 1
	cgLeftMouseUp            = 2
	cgRightMouseDown         = 3
	cgRightMouseUp           = 4
	cgMouseMoved             = 5
	cgLeftMouseDragged       = 6
	cgRightMouseDragged      = 7
	cgKeyDown                = 10
	cgKeyUp                  = 11
	cgFlagsChanged           = 12
	cgScrollWheel            = 22
	cgOtherMouseDown         = 25
	cgOtherMouseUp           = 26
	cgOtherMouseDragged      = 27
	cgTapDisabledByTimeout   = 0xFFFFFFFE
	cgTapDisabledByUserInput = 0xFFFFFFFF
)

// CGEventField values.
const (
	cgMouseButtonNumber     = 3
	cgKeyboardKeycode       = 9
	cgScrollDeltaAxis1      = 11
	cgTargetUnixProcessID   = 40
	cgEventSourceUserData   = 42
	cgEventSourceStateField = 45
)

const (
	// CGEventSourceStateID values.
	cgStatePrivate         = -1
	cgStateCombinedSession = 0
	cgStateHIDSystem       = 1

	cgHIDEventTap         = 0
	cgSessionEventTap     = 1
	cgHeadInsertEventTap  = 0
	cgTapOptionDefault    = 0
	cgTapOptionListenOnly = 1

	// Device-independent modifier flags.
	cgFlagShift     = 0x00020000
	cgFlagControl   = 0x00040000
	cgFlagAlternate = 0x00080000
	cgFlagCommand   = 0x00100000
	cgFlagFn        = 0x00800000

	// Device-dependent modifier flags (IOKit/hidsystem/IOLLEvent.h). The device-independent ones cannot
	// tell left and right apart, which leaves a key stuck when both sides are held and one is released.
	deviceLCtl   = 0x00000001
	deviceLShift = 0x00000002
	deviceRShift = 0x00000004
	deviceLCmd   = 0x00000008
	deviceRCmd   = 0x00000010
	deviceLAlt   = 0x00000020
	deviceRAlt   = 0x00000040
	deviceRCtl   = 0x00002000

	ioHIDRequestTypeListenEvent = 1
	ioHIDAccessTypeGranted      = 0

	ucKeyActionDown = 0
)

type cgPoint struct{ X, Y float64 }

var (
	cfRelease                     func(ref uintptr)
	cfRunLoopGetCurrent           func() uintptr
	cfRunLoopRun                  func()
	cfRunLoopStop                 func(rl uintptr)
	cfRunLoopWakeUp               func(rl uintptr)
	cfRunLoopAddSource            func(rl, source, mode uintptr)
	cfRunLoopRemoveSource         func(rl, source, mode uintptr)
	cfRunLoopPerformBlock         func(rl, mode uintptr, block objc.Block)
	cfMachPortCreateRunLoopSource func(allocator, port uintptr, order int) uintptr
	cfDataGetBytePtr              func(data uintptr) uintptr
	cfDictionaryCreate            func(allocator uintptr, keys, values *uintptr, n int, keyCallBacks, valueCallBacks uintptr) uintptr

	cgEventTapCreate            func(tap, place, options uint32, mask uint64, callback, refcon uintptr) uintptr
	cgEventTapEnable            func(tap uintptr, enable bool)
	cgEventSetIntegerValueField func(event uintptr, field uint32, value int64)
	cgEventSetFlags             func(event uintptr, flags uint64)
	cgEventGetLocation          func(event uintptr) cgPoint
	cgEventSourceCreate         func(state int32) uintptr
	cgEventSourceKeyState       func(state int32, code uint16) bool
	cgEventSourceButtonState    func(state int32, button uint32) bool
	cgEventCreateKeyboardEvent  func(source uintptr, code uint16, down bool) uintptr
	cgEventPost                 func(tap uint32, event uintptr)

	// Called for every event, so they are called directly rather than through reflection.
	addrEventGetIntegerValueField, addrEventGetFlags uintptr

	axIsProcessTrusted            func() bool
	axIsProcessTrustedWithOptions func(options uintptr) bool
	ioHIDCheckAccess              func(request uint32) uint32
	ioHIDRequestAccess            func(request uint32) bool

	tisCopyCurrentKeyboardLayoutInputSource func() uintptr
	tisGetInputSourceProperty               func(source, property uintptr) uintptr
	lmGetKbdType                            func() uint8
	ucKeyTranslate                          func(layout uintptr, code, action uint16, modifiers, keyboardType, options uint32,
		deadKeyState *uint32, maxLength uint, length *uint, chars *uint16) int32

	objcAutoreleasePoolPush func() uintptr
	objcAutoreleasePoolPop  func(pool uintptr)

	// Constants exported by the frameworks.
	kCFRunLoopCommonModes            uintptr
	kCFBooleanTrue, kCFBooleanFalse  uintptr
	kCFTypeDictionaryKeyCallBacks    uintptr // address of the struct
	kCFTypeDictionaryValueCallBacks  uintptr // address of the struct
	kAXTrustedCheckOptionPrompt      uintptr
	kTISPropertyUnicodeKeyLayoutData uintptr

	classNSWorkspace, classNSRunningApplication objc.ID
	selSharedWorkspace                          objc.SEL
	selFrontmostApplication                     objc.SEL
	selProcessIdentifier                        objc.SEL
	selRunningApplicationPID                    objc.SEL
	selLocalizedName                            objc.SEL
	selBundleIdentifier                         objc.SEL
	selExecutableURL                            objc.SEL
	selLastPathComponent                        objc.SEL
	selPath                                     objc.SEL
	selUTF8String                               objc.SEL

	tapCallback uintptr

	// sendSource is the private event source of Send. It does not track modifiers, see sentFlags.
	sendSource uintptr
)

// load binds the system frameworks on first use, so importing the package costs nothing and a
// missing symbol becomes an error instead of a crash.
var load = sync.OnceValue(func() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%w: %v", ErrUnsupported, r)
		}
	}()
	const frameworks = "/System/Library/Frameworks/"
	cf := dlopen(frameworks + "CoreFoundation.framework/CoreFoundation")
	appServices := dlopen(frameworks + "ApplicationServices.framework/ApplicationServices") // includes CoreGraphics
	iokit := dlopen(frameworks + "IOKit.framework/IOKit")
	carbon := dlopen(frameworks + "Carbon.framework/Carbon")
	dlopen(frameworks + "AppKit.framework/AppKit") // registers NSWorkspace and NSRunningApplication
	libobjc := dlopen("/usr/lib/libobjc.A.dylib")

	for _, f := range []struct {
		fptr any
		lib  uintptr
		name string
	}{
		{&cfRelease, cf, "CFRelease"},
		{&cfRunLoopGetCurrent, cf, "CFRunLoopGetCurrent"},
		{&cfRunLoopRun, cf, "CFRunLoopRun"},
		{&cfRunLoopStop, cf, "CFRunLoopStop"},
		{&cfRunLoopWakeUp, cf, "CFRunLoopWakeUp"},
		{&cfRunLoopAddSource, cf, "CFRunLoopAddSource"},
		{&cfRunLoopRemoveSource, cf, "CFRunLoopRemoveSource"},
		{&cfRunLoopPerformBlock, cf, "CFRunLoopPerformBlock"},
		{&cfMachPortCreateRunLoopSource, cf, "CFMachPortCreateRunLoopSource"},
		{&cfDataGetBytePtr, cf, "CFDataGetBytePtr"},
		{&cfDictionaryCreate, cf, "CFDictionaryCreate"},
		{&cgEventTapCreate, appServices, "CGEventTapCreate"},
		{&cgEventTapEnable, appServices, "CGEventTapEnable"},
		{&cgEventSetIntegerValueField, appServices, "CGEventSetIntegerValueField"},
		{&cgEventSetFlags, appServices, "CGEventSetFlags"},
		{&cgEventGetLocation, appServices, "CGEventGetLocation"},
		{&cgEventSourceCreate, appServices, "CGEventSourceCreate"},
		{&cgEventSourceKeyState, appServices, "CGEventSourceKeyState"},
		{&cgEventSourceButtonState, appServices, "CGEventSourceButtonState"},
		{&cgEventCreateKeyboardEvent, appServices, "CGEventCreateKeyboardEvent"},
		{&cgEventPost, appServices, "CGEventPost"},
		{&axIsProcessTrusted, appServices, "AXIsProcessTrusted"},
		{&axIsProcessTrustedWithOptions, appServices, "AXIsProcessTrustedWithOptions"},
		{&ioHIDCheckAccess, iokit, "IOHIDCheckAccess"},
		{&ioHIDRequestAccess, iokit, "IOHIDRequestAccess"},
		{&tisCopyCurrentKeyboardLayoutInputSource, carbon, "TISCopyCurrentKeyboardLayoutInputSource"},
		{&tisGetInputSourceProperty, carbon, "TISGetInputSourceProperty"},
		{&lmGetKbdType, carbon, "LMGetKbdType"},
		{&ucKeyTranslate, carbon, "UCKeyTranslate"},
		{&objcAutoreleasePoolPush, libobjc, "objc_autoreleasePoolPush"},
		{&objcAutoreleasePoolPop, libobjc, "objc_autoreleasePoolPop"},
	} {
		purego.RegisterLibFunc(f.fptr, f.lib, f.name)
	}
	addrEventGetIntegerValueField = dlsym(appServices, "CGEventGetIntegerValueField")
	addrEventGetFlags = dlsym(appServices, "CGEventGetFlags")

	kCFRunLoopCommonModes = loadPointer(cf, "kCFRunLoopCommonModes")
	kCFBooleanTrue = loadPointer(cf, "kCFBooleanTrue")
	kCFBooleanFalse = loadPointer(cf, "kCFBooleanFalse")
	kCFTypeDictionaryKeyCallBacks = dlsym(cf, "kCFTypeDictionaryKeyCallBacks")
	kCFTypeDictionaryValueCallBacks = dlsym(cf, "kCFTypeDictionaryValueCallBacks")
	kAXTrustedCheckOptionPrompt = loadPointer(appServices, "kAXTrustedCheckOptionPrompt")
	kTISPropertyUnicodeKeyLayoutData = loadPointer(carbon, "kTISPropertyUnicodeKeyLayoutData")

	classNSWorkspace = objc.ID(objc.GetClass("NSWorkspace"))
	classNSRunningApplication = objc.ID(objc.GetClass("NSRunningApplication"))
	selSharedWorkspace = objc.RegisterName("sharedWorkspace")
	selFrontmostApplication = objc.RegisterName("frontmostApplication")
	selProcessIdentifier = objc.RegisterName("processIdentifier")
	selRunningApplicationPID = objc.RegisterName("runningApplicationWithProcessIdentifier:")
	selLocalizedName = objc.RegisterName("localizedName")
	selBundleIdentifier = objc.RegisterName("bundleIdentifier")
	selExecutableURL = objc.RegisterName("executableURL")
	selLastPathComponent = objc.RegisterName("lastPathComponent")
	selPath = objc.RegisterName("path")
	selUTF8String = objc.RegisterName("UTF8String")

	tapCallback = purego.NewCallback(onEvent)
	sendSource = cgEventSourceCreate(cgStatePrivate)
	return nil
})

func dlopen(path string) uintptr {
	return must(purego.Dlopen(path, purego.RTLD_NOW|purego.RTLD_GLOBAL))
}

func dlsym(lib uintptr, name string) uintptr {
	return must(purego.Dlsym(lib, name))
}

// must panics on error. Panics in load are turned into its error.
func must(v uintptr, err error) uintptr {
	if err != nil {
		panic(err)
	}
	return v
}

// loadPointer returns the value of a global pointer variable, such as a CFStringRef constant.
func loadPointer(lib uintptr, name string) uintptr {
	addr := dlsym(lib, name)
	// Reinterpret the address as a pointer, which is valid for the lifetime of the process.
	return **(**uintptr)(unsafe.Pointer(&addr))
}

func eventField(event uintptr, field uint32) int64 {
	r, _, _ := purego.SyscallN(addrEventGetIntegerValueField, event, uintptr(field))
	return int64(r)
}

// modifierFlags returns the flags a modifier key sets while it is held, or 0 for other keys.
func modifierFlags(code key.Code) uint64 {
	switch code {
	case key.LeftShift:
		return cgFlagShift | deviceLShift
	case key.RightShift:
		return cgFlagShift | deviceRShift
	case key.LeftCtrl:
		return cgFlagControl | deviceLCtl
	case key.RightCtrl:
		return cgFlagControl | deviceRCtl
	case key.LeftAlt:
		return cgFlagAlternate | deviceLAlt
	case key.RightAlt:
		return cgFlagAlternate | deviceRAlt
	case key.LeftSuper:
		return cgFlagCommand | deviceLCmd
	case key.RightSuper:
		return cgFlagCommand | deviceRCmd
	case key.Fn:
		return cgFlagFn
	}
	return 0
}

// modifierPressed checks the side-specific flag of a modifier key.
func modifierPressed(code key.Code, flags uint64) bool {
	// Drop the device-independent bits, which are shared by both sides.
	return flags&modifierFlags(code)&^(cgFlagShift|cgFlagControl|cgFlagAlternate|cgFlagCommand) != 0
}

func eventFlags(event uintptr) uint8 {
	if eventField(event, cgEventSourceUserData) == ownEventTag {
		return FlagInjected | FlagOwn
	}
	// Events from physical devices come from the HID system state.
	if eventField(event, cgEventSourceStateField) != cgStateHIDSystem {
		return FlagInjected
	}
	return 0
}

var (
	eventTap uintptr // only used on the hook thread

	runLoopMu sync.Mutex
	runLoop   uintptr // run loop of the hook thread, 0 when stopped
)

// onEvent is the event tap callback. It returns the event to pass it on, or 0 to block it.
func onEvent(_ uintptr, typ uint32, event, _ uintptr) uintptr {
	if typ == cgTapDisabledByTimeout || typ == cgTapDisabledByUserInput {
		// https://stackoverflow.com/questions/2969110/cgeventtapcreate-breaks-down-mysteriously-with-key-down-events#2971217
		cgEventTapEnable(eventTap, true)
		// Events were missed while the tap was disabled.
		reset()
		return event
	}
	ev := Event{Flags: eventFlags(event)}
	switch typ {
	case cgKeyDown, cgKeyUp:
		ev.Kind = keyKind(typ == cgKeyDown)
		ev.Code = key.Code(eventField(event, cgKeyboardKeycode))
	case cgFlagsChanged:
		ev.Code = key.Code(eventField(event, cgKeyboardKeycode))
		flags, _, _ := purego.SyscallN(addrEventGetFlags, event)
		ev.Kind = keyKind(modifierPressed(ev.Code, uint64(flags)))
	case cgLeftMouseDown, cgRightMouseDown, cgOtherMouseDown, cgLeftMouseUp, cgRightMouseUp, cgOtherMouseUp:
		ev.Kind = keyKind(typ == cgLeftMouseDown || typ == cgRightMouseDown || typ == cgOtherMouseDown)
		button := eventField(event, cgMouseButtonNumber)
		if button < 0 || button >= int64(len(key.MouseButtons)) {
			return event
		}
		ev.Code = key.MouseButtons[button]
	case cgMouseMoved, cgLeftMouseDragged, cgRightMouseDragged, cgOtherMouseDragged:
		ev.Kind = MouseMove
	case cgScrollWheel:
		ev.Kind = MouseWheel
		ev.Delta = int32(eventField(event, cgScrollDeltaAxis1))
	default:
		return event
	}
	location := cgEventGetLocation(event)
	ev.X, ev.Y = int32(location.X), int32(location.Y)
	// The target process only matters for keys and buttons; skip the lookup for frequent moves.
	if ev.Kind != MouseMove {
		ev.PID = int(int32(eventField(event, cgTargetUnixProcessID)))
	}
	if dispatch(ev) {
		return 0
	}
	return event
}

// platformStart installs the event tap and runs the run loop that delivers its events.
// It runs on a locked OS thread, which the tap and the run loop must share.
func platformStart(cfg Config) error {
	if err := load(); err != nil {
		return fmt.Errorf("%w: %v", ErrHookFailed, err)
	}
	var mask uint64
	listen := func(types ...uint32) {
		for _, t := range types {
			mask |= 1 << t
		}
	}
	if cfg.Keyboard {
		listen(cgKeyDown, cgKeyUp, cgFlagsChanged)
	}
	if cfg.Mouse {
		listen(cgLeftMouseDown, cgLeftMouseUp, cgRightMouseDown, cgRightMouseUp, cgOtherMouseDown, cgOtherMouseUp,
			cgMouseMoved, cgLeftMouseDragged, cgRightMouseDragged, cgOtherMouseDragged, cgScrollWheel)
	}
	// A tap that can block events needs Accessibility permission; otherwise only listen.
	suppress := axIsProcessTrusted()
	options := uint32(cgTapOptionListenOnly)
	if suppress {
		options = cgTapOptionDefault
	}

	eventTap = cgEventTapCreate(cgSessionEventTap, cgHeadInsertEventTap, options, mask, tapCallback, 0)
	if eventTap == 0 {
		return fmt.Errorf("%w: grant Input Monitoring permission to this process", ErrHookFailed)
	}
	source := cfMachPortCreateRunLoopSource(0, eventTap, 0)
	rl := cfRunLoopGetCurrent()
	cfRunLoopAddSource(rl, source, kCFRunLoopCommonModes)
	cgEventTapEnable(eventTap, true)
	runLoopMu.Lock()
	runLoop = rl
	runLoopMu.Unlock()

	ready(suppress)
	cfRunLoopRun()

	// Tear down on the thread that owns the run loop.
	runLoopMu.Lock()
	runLoop = 0
	runLoopMu.Unlock()
	cgEventTapEnable(eventTap, false)
	cfRunLoopRemoveSource(rl, source, kCFRunLoopCommonModes)
	cfRelease(source)
	cfRelease(eventTap)
	eventTap = 0
	return nil
}

// platformStop can be called from any thread once platformStart has signalled ready.
// The stop is queued on the run loop, so it is not lost if the loop has not started running yet.
func platformStop() {
	runLoopMu.Lock()
	defer runLoopMu.Unlock()
	if runLoop == 0 {
		return
	}
	stop := objc.NewBlock(func(objc.Block) { cfRunLoopStop(cfRunLoopGetCurrent()) })
	cfRunLoopPerformBlock(runLoop, kCFRunLoopCommonModes, stop) // copies the block
	stop.Release()
	cfRunLoopWakeUp(runLoop)
}

// KeyState reports whether the OS considers the key to be held down. ok is false if the OS cannot tell.
func KeyState(code key.Code) (pressed, ok bool) {
	if load() != nil {
		return false, false
	}
	if i, ok := mouseButton(code); ok {
		return cgEventSourceButtonState(cgStateCombinedSession, uint32(i)), true
	}
	return cgEventSourceKeyState(cgStateCombinedSession, uint16(code)), true
}

var (
	sendMu sync.Mutex
	// sentFlags are the modifier flags held by keys sent with Send. A private event source does not track them.
	sentFlags uint64
)

// Send posts a synthetic key event.
func Send(code key.Code, down bool) error {
	if err := load(); err != nil {
		return err
	}
	event := cgEventCreateKeyboardEvent(sendSource, uint16(code), down)
	if event == 0 {
		return fmt.Errorf("ivent: failed to send key %s", code.Name())
	}
	defer cfRelease(event)
	sendMu.Lock()
	if down {
		sentFlags |= modifierFlags(code)
	} else {
		sentFlags &^= modifierFlags(code)
	}
	cgEventSetFlags(event, sentFlags)
	sendMu.Unlock()
	cgEventSetIntegerValueField(event, cgEventSourceUserData, ownEventTag)
	cgEventPost(cgHIDEventTap, event)
	return nil
}

// charForKey returns the character a key types without modifiers on the current layout.
func charForKey(code key.Code) (rune, bool) {
	if load() != nil {
		return 0, false
	}
	source := tisCopyCurrentKeyboardLayoutInputSource()
	if source == 0 {
		return 0, false
	}
	defer cfRelease(source)
	data := tisGetInputSourceProperty(source, kTISPropertyUnicodeKeyLayoutData)
	if data == 0 {
		return 0, false
	}
	var deadKeyState uint32
	var chars [4]uint16
	var length uint
	// Without options, a dead key types nothing and is skipped.
	status := ucKeyTranslate(cfDataGetBytePtr(data), uint16(code), ucKeyActionDown, 0, uint32(lmGetKbdType()), 0,
		&deadKeyState, uint(len(chars)), &length, &chars[0])
	if status != 0 || length != 1 {
		return 0, false
	}
	return rune(chars[0]), true
}

// ProcessInfo returns the display name and an identifier (bundle ID or path) of a process.
// A pid of 0 resolves to the frontmost application.
func ProcessInfo(pid int) (resolvedPID int, name, id string, ok bool) {
	if load() != nil {
		return pid, "", "", false
	}
	// The autorelease pool must be pushed and popped on the same thread.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	pool := objcAutoreleasePoolPush()
	defer objcAutoreleasePoolPop(pool)

	if pid == 0 {
		app := classNSWorkspace.Send(selSharedWorkspace).Send(selFrontmostApplication)
		pid = int(objc.Send[int32](app, selProcessIdentifier))
	}
	app := classNSRunningApplication.Send(selRunningApplicationPID, int32(pid))
	if app == 0 {
		return pid, "", "", false
	}
	// Messages to nil return nil, which becomes "".
	url := app.Send(selExecutableURL)
	name = objc.Send[string](app.Send(selLocalizedName), selUTF8String)
	if name == "" {
		name = objc.Send[string](url.Send(selLastPathComponent), selUTF8String)
	}
	id = objc.Send[string](app.Send(selBundleIdentifier), selUTF8String)
	if id == "" {
		id = objc.Send[string](url.Send(selPath), selUTF8String)
	}
	return pid, name, id, true
}

// Permissions reports whether the process may listen to input (monitor) and block or send input (control).
// If prompt is true, the OS asks the user for missing permissions.
func Permissions(prompt bool) (monitor, control bool) {
	if load() != nil {
		return false, false
	}
	keys, values := kAXTrustedCheckOptionPrompt, kCFBooleanFalse
	if prompt {
		values = kCFBooleanTrue
	}
	options := cfDictionaryCreate(0, &keys, &values, 1, kCFTypeDictionaryKeyCallBacks, kCFTypeDictionaryValueCallBacks)
	control = axIsProcessTrustedWithOptions(options)
	cfRelease(options)

	monitor = ioHIDCheckAccess(ioHIDRequestTypeListenEvent) == ioHIDAccessTypeGranted
	if !monitor && prompt {
		ioHIDRequestAccess(ioHIDRequestTypeListenEvent)
	}
	return monitor, control
}
