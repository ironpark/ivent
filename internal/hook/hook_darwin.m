#import <ApplicationServices/ApplicationServices.h>
#import <Cocoa/Cocoa.h>
#import "hook_help_darwin.m"
#include <stdio.h>
#include <pthread.h>
#include "hook_darwin.h"

static CFMachPortRef eventTap = NULL;
static CFRunLoopSourceRef runLoopSource = NULL;
static CFRunLoopRef runLoop = NULL;
static pthread_mutex_t mutex = PTHREAD_MUTEX_INITIALIZER;


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

inline static Boolean isModifierPressed(CGKeyCode keycode,CGEventFlags flags) {
    switch (keycode) {
        case kVK_Shift:        return (flags & IVENT_DEVICE_LSHIFT) != 0;
        case kVK_RightShift:   return (flags & IVENT_DEVICE_RSHIFT) != 0;
        case kVK_Control:      return (flags & IVENT_DEVICE_LCTL) != 0;
        case kVK_RightControl: return (flags & IVENT_DEVICE_RCTL) != 0;
        case kVK_Option:       return (flags & IVENT_DEVICE_LALT) != 0;
        case kVK_RightOption:  return (flags & IVENT_DEVICE_RALT) != 0;
        case kVK_Command:      return (flags & IVENT_DEVICE_LCMD) != 0;
        case kVK_RightCommand: return (flags & IVENT_DEVICE_RCMD) != 0;
        case kVK_Function:     return (flags & kCGEventFlagMaskSecondaryFn) != 0;
        default:
        return false;
    }
}

static CGEventRef eventCallback(CGEventTapProxy proxy, CGEventType type, CGEventRef event, void *refcon) {
    if (type == kCGEventTapDisabledByTimeout || type == kCGEventTapDisabledByUserInput) {
        // https://stackoverflow.com/questions/2969110/cgeventtapcreate-breaks-down-mysteriously-with-key-down-events#2971217
        CGEventTapEnable(eventTap, true);
        return event;
    }
     pid_t pid = (pid_t)CGEventGetIntegerValueField(event, kCGEventTargetUnixProcessID);
     switch (type) {
           case kCGEventKeyDown:
           case kCGEventKeyUp: {
               CGKeyCode keycode = (CGKeyCode)CGEventGetIntegerValueField(event, kCGKeyboardEventKeycode);
               keyEventGoCallback(pid, keycode, type == kCGEventKeyDown);
               break;
           }
           case kCGEventFlagsChanged: {
               CGKeyCode keycode = (CGKeyCode)CGEventGetIntegerValueField(event, kCGKeyboardEventKeycode);
               bool keyDown = isModifierPressed(keycode, CGEventGetFlags(event));
               keyEventGoCallback(pid, keycode, keyDown);
               break;
           }
           case kCGEventLeftMouseDown:
           case kCGEventLeftMouseUp:
           case kCGEventRightMouseDown:
           case kCGEventRightMouseUp:
           case kCGEventOtherMouseDown:
           case kCGEventOtherMouseUp: {
               CGPoint location = CGEventGetLocation(event);
               int button = (int)CGEventGetIntegerValueField(event, kCGMouseEventButtonNumber);
               bool isDown = (type == kCGEventLeftMouseDown || type == kCGEventRightMouseDown || type == kCGEventOtherMouseDown);
               mouseEventGoCallback(pid, (int)location.x, (int)location.y, button, isDown);
               break;
           }
           case kCGEventMouseMoved:
           case kCGEventLeftMouseDragged:
           case kCGEventRightMouseDragged:
           case kCGEventOtherMouseDragged: {
               CGPoint location = CGEventGetLocation(event);
               mouseMoveGoCallback(pid, (int)location.x, (int)location.y);
               break;
           }
           case kCGEventScrollWheel: {
               CGPoint location = CGEventGetLocation(event);
               int64_t deltaY = CGEventGetIntegerValueField(event, kCGScrollWheelEventDeltaAxis1);
               mouseWheelGoCallback(pid, (int)location.x, (int)location.y, (int)deltaY);
               break;
           }
           default:
              break;
       }
    return event;
}


int start(ListenMode mode) {
    CGEventMask eventMask = 0;

    if (mode & LISTEN_KEYBOARD) {
        eventMask |= (1 << kCGEventKeyDown) | (1 << kCGEventKeyUp) | (1 << kCGEventFlagsChanged);
    }
    if (mode & LISTEN_MOUSE) {
        eventMask |= (1 << kCGEventLeftMouseDown) | (1 << kCGEventLeftMouseUp) |
                     (1 << kCGEventRightMouseDown) | (1 << kCGEventRightMouseUp) |
                     (1 << kCGEventOtherMouseDown) | (1 << kCGEventOtherMouseUp) |
                     (1 << kCGEventMouseMoved) | (1 << kCGEventLeftMouseDragged) |
                     (1 << kCGEventRightMouseDragged) | (1 << kCGEventOtherMouseDragged) |
                     (1 << kCGEventScrollWheel);
    }

    pthread_mutex_lock(&mutex);
    eventTap = CGEventTapCreate(kCGSessionEventTap, kCGHeadInsertEventTap, kCGEventTapOptionListenOnly, eventMask, eventCallback, NULL);
    if (!eventTap) {
        pthread_mutex_unlock(&mutex);
        return -1;
    }
    runLoopSource = CFMachPortCreateRunLoopSource(kCFAllocatorDefault, eventTap, 0);
    runLoop = CFRunLoopGetCurrent();
    CFRunLoopAddSource(runLoop, runLoopSource, kCFRunLoopCommonModes);
    CGEventTapEnable(eventTap, true);
    pthread_mutex_unlock(&mutex);

    hookReadyGoCallback();
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
void stop() {
    pthread_mutex_lock(&mutex);
    if (runLoop) {
        CFRunLoopPerformBlock(runLoop, kCFRunLoopCommonModes, ^{
            CFRunLoopStop(CFRunLoopGetCurrent());
        });
        CFRunLoopWakeUp(runLoop);
    }
    pthread_mutex_unlock(&mutex);
}
