#import <ApplicationServices/ApplicationServices.h>
#import <Cocoa/Cocoa.h>
#import "hook_help_darwin.m"
#include <stdio.h>
#include <pthread.h>
#include "hook_darwin.h"

static CFMachPortRef eventTap = NULL;
static CFRunLoopSourceRef runLoopSource = NULL;
static pthread_mutex_t mutex = PTHREAD_MUTEX_INITIALIZER;


inline static Boolean isModifierPressed(CGKeyCode keycode,CGEventFlags flags) {
    switch (keycode) {
        case kVK_RightShift:
        case kVK_Shift:
        return (flags & kCGEventFlagMaskShift) != 0;
        case kVK_RightControl:
        case kVK_Control:
        return (flags & kCGEventFlagMaskControl) != 0;
        case kVK_RightOption:
        case kVK_Option:
        return (flags & kCGEventFlagMaskAlternate) != 0;
        case kVK_RightCommand:
        case kVK_Command:
        return (flags & kCGEventFlagMaskCommand) != 0;
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
    if (eventTap == NULL) {
        pthread_mutex_lock(&mutex);
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

        eventTap = CGEventTapCreate(kCGSessionEventTap, kCGHeadInsertEventTap, kCGEventTapOptionListenOnly, eventMask, eventCallback, NULL);
        if (!eventTap) {
            pthread_mutex_unlock(&mutex);
            return -1;
        }
        pthread_mutex_unlock(&mutex);

        runLoopSource = CFMachPortCreateRunLoopSource(kCFAllocatorDefault, eventTap, 0);
        CFRunLoopAddSource(CFRunLoopGetCurrent(), runLoopSource, kCFRunLoopCommonModes);
        CGEventTapEnable(eventTap, true);
        CFRunLoopRun();
    }
    return 0;
}

void stop() {
    pthread_mutex_lock(&mutex);
    if (eventTap) {
        CGEventTapEnable(eventTap, false);
        CFRunLoopRemoveSource(CFRunLoopGetCurrent(), runLoopSource, kCFRunLoopCommonModes);
        CFRelease(runLoopSource);
        CFRelease(eventTap);
        runLoopSource = NULL;
        eventTap = NULL;
    }
    pthread_mutex_unlock(&mutex);
}