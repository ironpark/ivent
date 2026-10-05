#import <ApplicationServices/ApplicationServices.h>
#import <Cocoa/Cocoa.h>
#import <IOKit/hidsystem/IOHIDLib.h>
#import <Carbon/Carbon.h> // kVK_* key codes
#include <pthread.h>
#include <string.h>
#include "hook_common.h"
#include "hook_darwin.h"

static CFMachPortRef eventTap = NULL;
static CFRunLoopSourceRef runLoopSource = NULL;
static CFRunLoopRef runLoop = NULL;
static pthread_mutex_t mutex = PTHREAD_MUTEX_INITIALIZER;

// Marks events posted by sendKey so they can be recognized when they come back through the tap.
#define IVENT_USER_DATA 0x6976656E74

// Device-dependent modifier flags (IOKit/hidsystem/IOLLEvent.h).
// The device-independent masks (kCGEventFlagMaskShift, ...) cannot tell left and right apart,
// which leaves a key stuck when both sides are held and one is released.
#define IVENT_DEVICE_LCTL   0x00000001
#define IVENT_DEVICE_LSHIFT 0x00000002
#define IVENT_DEVICE_RSHIFT 0x00000004
#define IVENT_DEVICE_LCMD   0x00000008
#define IVENT_DEVICE_RCMD   0x00000010
#define IVENT_DEVICE_LALT   0x00000020
#define IVENT_DEVICE_RALT   0x00000040
#define IVENT_DEVICE_RCTL   0x00002000

// modifierFlags returns the flags a modifier key sets while it is held, or 0 for other keys.
static CGEventFlags modifierFlags(CGKeyCode keycode) {
    switch (keycode) {
        case kVK_Shift:        return kCGEventFlagMaskShift | IVENT_DEVICE_LSHIFT;
        case kVK_RightShift:   return kCGEventFlagMaskShift | IVENT_DEVICE_RSHIFT;
        case kVK_Control:      return kCGEventFlagMaskControl | IVENT_DEVICE_LCTL;
        case kVK_RightControl: return kCGEventFlagMaskControl | IVENT_DEVICE_RCTL;
        case kVK_Option:       return kCGEventFlagMaskAlternate | IVENT_DEVICE_LALT;
        case kVK_RightOption:  return kCGEventFlagMaskAlternate | IVENT_DEVICE_RALT;
        case kVK_Command:      return kCGEventFlagMaskCommand | IVENT_DEVICE_LCMD;
        case kVK_RightCommand: return kCGEventFlagMaskCommand | IVENT_DEVICE_RCMD;
        case kVK_Function:     return kCGEventFlagMaskSecondaryFn;
        default:               return 0;
    }
}

// isModifierPressed checks the side-specific flag of a modifier key.
inline static bool isModifierPressed(CGKeyCode keycode, CGEventFlags flags) {
    // Drop the device-independent bits, which are shared by both sides.
    CGEventFlags mask = modifierFlags(keycode) &
        ~(CGEventFlags)(kCGEventFlagMaskShift | kCGEventFlagMaskControl | kCGEventFlagMaskAlternate | kCGEventFlagMaskCommand);
    return mask != 0 && (flags & mask) != 0;
}

static int eventFlags(CGEventRef event) {
    if (CGEventGetIntegerValueField(event, kCGEventSourceUserData) == IVENT_USER_DATA) {
        return IVENT_FLAG_INJECTED | IVENT_FLAG_OWN;
    }
    // Events from physical devices come from the HID system state.
    if (CGEventGetIntegerValueField(event, kCGEventSourceStateID) != kCGEventSourceStateHIDSystemState) {
        return IVENT_FLAG_INJECTED;
    }
    return 0;
}

static CGEventRef eventCallback(CGEventTapProxy proxy, CGEventType type, CGEventRef event, void *refcon) {
    if (type == kCGEventTapDisabledByTimeout || type == kCGEventTapDisabledByUserInput) {
        // https://stackoverflow.com/questions/2969110/cgeventtapcreate-breaks-down-mysteriously-with-key-down-events#2971217
        CGEventTapEnable(eventTap, true);
        // Events were missed while the tap was disabled.
        goReset();
        return event;
    }
    int kind = 0, code = 0, delta = 0, flags = eventFlags(event);
    switch (type) {
        case kCGEventKeyDown:
        case kCGEventKeyUp:
            kind = type == kCGEventKeyDown ? IVENT_KEY_DOWN : IVENT_KEY_UP;
            code = (int)CGEventGetIntegerValueField(event, kCGKeyboardEventKeycode);
            break;
        case kCGEventFlagsChanged:
            code = (int)CGEventGetIntegerValueField(event, kCGKeyboardEventKeycode);
            kind = isModifierPressed(code, CGEventGetFlags(event)) ? IVENT_KEY_DOWN : IVENT_KEY_UP;
            break;
        case kCGEventLeftMouseDown:
        case kCGEventRightMouseDown:
        case kCGEventOtherMouseDown:
        case kCGEventLeftMouseUp:
        case kCGEventRightMouseUp:
        case kCGEventOtherMouseUp:
            kind = (type == kCGEventLeftMouseDown || type == kCGEventRightMouseDown || type == kCGEventOtherMouseDown)
                ? IVENT_KEY_DOWN : IVENT_KEY_UP;
            code = (int)CGEventGetIntegerValueField(event, kCGMouseEventButtonNumber);
            flags |= IVENT_FLAG_BUTTON;
            break;
        case kCGEventMouseMoved:
        case kCGEventLeftMouseDragged:
        case kCGEventRightMouseDragged:
        case kCGEventOtherMouseDragged:
            kind = IVENT_MOUSE_MOVE;
            break;
        case kCGEventScrollWheel:
            kind = IVENT_MOUSE_WHEEL;
            delta = (int)CGEventGetIntegerValueField(event, kCGScrollWheelEventDeltaAxis1);
            break;
        default:
            return event;
    }
    CGPoint location = CGEventGetLocation(event);
    pid_t pid = (pid_t)CGEventGetIntegerValueField(event, kCGEventTargetUnixProcessID);
    if (goEvent(kind, code, (int)location.x, (int)location.y, delta, pid, flags)) {
        return NULL;
    }
    return event;
}

int start(bool keyboard, bool mouse) {
    CGEventMask eventMask = 0;
    if (keyboard) {
        eventMask |= CGEventMaskBit(kCGEventKeyDown) | CGEventMaskBit(kCGEventKeyUp) | CGEventMaskBit(kCGEventFlagsChanged);
    }
    if (mouse) {
        eventMask |= CGEventMaskBit(kCGEventLeftMouseDown) | CGEventMaskBit(kCGEventLeftMouseUp) |
                     CGEventMaskBit(kCGEventRightMouseDown) | CGEventMaskBit(kCGEventRightMouseUp) |
                     CGEventMaskBit(kCGEventOtherMouseDown) | CGEventMaskBit(kCGEventOtherMouseUp) |
                     CGEventMaskBit(kCGEventMouseMoved) | CGEventMaskBit(kCGEventLeftMouseDragged) |
                     CGEventMaskBit(kCGEventRightMouseDragged) | CGEventMaskBit(kCGEventOtherMouseDragged) |
                     CGEventMaskBit(kCGEventScrollWheel);
    }
    // A tap that can block events needs Accessibility permission; otherwise only listen.
    bool suppress = AXIsProcessTrusted();
    CGEventTapOptions options = suppress ? kCGEventTapOptionDefault : kCGEventTapOptionListenOnly;

    pthread_mutex_lock(&mutex);
    eventTap = CGEventTapCreate(kCGSessionEventTap, kCGHeadInsertEventTap, options, eventMask, eventCallback, NULL);
    if (!eventTap) {
        pthread_mutex_unlock(&mutex);
        return -1;
    }
    runLoopSource = CFMachPortCreateRunLoopSource(kCFAllocatorDefault, eventTap, 0);
    runLoop = CFRunLoopGetCurrent();
    CFRunLoopAddSource(runLoop, runLoopSource, kCFRunLoopCommonModes);
    CGEventTapEnable(eventTap, true);
    pthread_mutex_unlock(&mutex);

    goReady(suppress);
    CFRunLoopRun();

    // Tear down on the thread that owns the run loop.
    pthread_mutex_lock(&mutex);
    CGEventTapEnable(eventTap, false);
    CFRunLoopRemoveSource(runLoop, runLoopSource, kCFRunLoopCommonModes);
    CFRelease(runLoopSource);
    CFRelease(eventTap);
    runLoopSource = NULL;
    eventTap = NULL;
    runLoop = NULL;
    pthread_mutex_unlock(&mutex);
    return 0;
}

// stop can be called from any thread once start has signalled ready.
// The stop is queued on the run loop, so it is not lost if the loop has not started running yet.
void stop(void) {
    pthread_mutex_lock(&mutex);
    if (runLoop) {
        CFRunLoopPerformBlock(runLoop, kCFRunLoopCommonModes, ^{
            CFRunLoopStop(CFRunLoopGetCurrent());
        });
        CFRunLoopWakeUp(runLoop);
    }
    pthread_mutex_unlock(&mutex);
}

bool keyPressed(int code) {
    return CGEventSourceKeyState(kCGEventSourceStateCombinedSessionState, (CGKeyCode)code);
}

bool buttonPressed(int button) {
    return CGEventSourceButtonState(kCGEventSourceStateCombinedSessionState, (CGMouseButton)button);
}

// Modifier flags held by keys sent with sendKey. A private event source does not track them.
static CGEventFlags sentFlags = 0;

int sendKey(int code, bool down) {
    CGEventSourceRef source = CGEventSourceCreate(kCGEventSourceStatePrivate);
    if (!source) {
        return -1;
    }
    CGEventRef event = CGEventCreateKeyboardEvent(source, (CGKeyCode)code, down);
    if (!event) {
        CFRelease(source);
        return -1;
    }
    pthread_mutex_lock(&mutex);
    CGEventFlags flags = modifierFlags((CGKeyCode)code);
    if (down) {
        sentFlags |= flags;
    } else {
        sentFlags &= ~flags;
    }
    CGEventSetFlags(event, sentFlags);
    pthread_mutex_unlock(&mutex);
    CGEventSetIntegerValueField(event, kCGEventSourceUserData, IVENT_USER_DATA);
    CGEventPost(kCGHIDEventTap, event);
    CFRelease(event);
    CFRelease(source);
    return 0;
}

// charForKey returns the character a key types without modifiers on the current layout, or -1.
int charForKey(int code) {
    TISInputSourceRef source = TISCopyCurrentKeyboardLayoutInputSource();
    if (!source) {
        return -1;
    }
    int result = -1;
    CFDataRef data = (CFDataRef)TISGetInputSourceProperty(source, kTISPropertyUnicodeKeyLayoutData);
    if (data) {
        const UCKeyboardLayout *layout = (const UCKeyboardLayout *)CFDataGetBytePtr(data);
        UInt32 deadKeyState = 0;
        UniChar chars[4];
        UniCharCount length = 0;
        OSStatus status = UCKeyTranslate(layout, (UInt16)code, kUCKeyActionDown, 0, LMGetKbdType(),
                                         kUCKeyTranslateNoDeadKeysBit, &deadKeyState, 4, &length, chars);
        if (status == noErr && length == 1) {
            result = chars[0];
        }
    }
    CFRelease(source);
    return result;
}

int frontmostPID(void) {
    @autoreleasepool {
        return [[NSWorkspace sharedWorkspace] frontmostApplication].processIdentifier;
    }
}

bool processInfo(int pid, char *name, int nameLen, char *ident, int identLen) {
    @autoreleasepool {
        NSRunningApplication *app = [NSRunningApplication runningApplicationWithProcessIdentifier:pid];
        if (!app) {
            return false;
        }
        NSString *n = app.localizedName ?: app.executableURL.lastPathComponent ?: @"";
        NSString *i = app.bundleIdentifier ?: app.executableURL.path ?: @"";
        strlcpy(name, n.UTF8String, nameLen);
        strlcpy(ident, i.UTF8String, identLen);
        return true;
    }
}

void permissions(bool prompt, bool *monitor, bool *control) {
    @autoreleasepool {
        NSDictionary *options = @{(__bridge id)kAXTrustedCheckOptionPrompt: @(prompt)};
        *control = AXIsProcessTrustedWithOptions((__bridge CFDictionaryRef)options);
    }
    if (@available(macOS 10.15, *)) {
        *monitor = IOHIDCheckAccess(kIOHIDRequestTypeListenEvent) == kIOHIDAccessTypeGranted;
        if (!*monitor && prompt) {
            IOHIDRequestAccess(kIOHIDRequestTypeListenEvent);
        }
    } else {
        *monitor = true;
    }
}
