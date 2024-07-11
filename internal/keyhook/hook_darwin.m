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
        return false;
    }
}

static CGEventRef keyEventCGEventCallback(CGEventTapProxy proxy, CGEventType type, CGEventRef event, void *refcon) {
    if (type == kCGEventTapDisabledByTimeout || type == kCGEventTapDisabledByUserInput) {
        // https://stackoverflow.com/questions/2969110/cgeventtapcreate-breaks-down-mysteriously-with-key-down-events#2971217
        CGEventTapEnable(eventTap, true);
        return event;
    }
    if (type == kCGEventKeyDown || type == kCGEventKeyUp ) {
        // Check the process ID of the event
        pid_t pid = (pid_t)CGEventGetIntegerValueField(event, kCGEventTargetUnixProcessID);
        // Get keycode
        CGKeyCode keycode = (CGKeyCode)CGEventGetIntegerValueField(event, kCGKeyboardEventKeycode);
        // Call Go callback function
        keyEventGoCallback(pid, keycode,  type == kCGEventKeyDown);
        return event;
    } else if(type == kCGEventFlagsChanged){
        // Check the process ID of the event
        pid_t pid = (pid_t)CGEventGetIntegerValueField(event, kCGEventTargetUnixProcessID);
        // Get keycode
        CGKeyCode keycode = (CGKeyCode)CGEventGetIntegerValueField(event, kCGKeyboardEventKeycode);
        // Call Go callback function
        bool keyDown = isModifierPressed(keycode,(CGEventGetFlags(event)));
        keyEventGoCallback(pid, keycode, keyDown);
        return event;
    }
    return event;
}

int start() {
    if (eventTap == NULL) {
        pthread_mutex_lock(&mutex);
        CGEventMask eventMask = (1 << kCGEventKeyDown) | (1 << kCGEventKeyUp) | (1 << kCGEventFlagsChanged);
        eventTap = CGEventTapCreate(kCGSessionEventTap, kCGHeadInsertEventTap, kCGEventTapOptionListenOnly, eventMask, keyEventCGEventCallback, NULL);
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