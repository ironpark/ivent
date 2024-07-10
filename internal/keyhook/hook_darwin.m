#import <ApplicationServices/ApplicationServices.h>
#import <Cocoa/Cocoa.h>
#include <stdio.h>
#include <pthread.h>
#include "hook_darwin.h"

// Go 콜백 함수를 참조하기 위한 함수 포인터 타입
// void keyEventGoCallback(pid_t pid, CGKeyCode keycode, bool down);
static Boolean restart_tap = false;

static CFMachPortRef eventTap = NULL;
static CFRunLoopSourceRef runLoopSource = NULL;
static pthread_mutex_t mutex = PTHREAD_MUTEX_INITIALIZER;


static CGEventRef keyEventCGEventCallback(CGEventTapProxy proxy, CGEventType type, CGEventRef event, void *refcon) {
    if (type == kCGEventTapDisabledByTimeout || type == kCGEventTapDisabledByUserInput) {
        // https://stackoverflow.com/questions/2969110/cgeventtapcreate-breaks-down-mysteriously-with-key-down-events#2971217
        CGEventTapEnable(eventTap, true);
        return event;
    }
    if (type == kCGEventKeyDown || type == kCGEventKeyUp || type == kCGEventFlagsChanged) {
        // 이벤트의 프로세스 ID를 확인
        pid_t pid = (pid_t)CGEventGetIntegerValueField(event, kCGEventTargetUnixProcessID);
        // 키 코드 가져오기
        CGKeyCode keycode = (CGKeyCode)CGEventGetIntegerValueField(event, kCGKeyboardEventKeycode);
        // 키 다운/업 이벤트 출력
        bool keyDown = (type == kCGEventKeyDown || (type == kCGEventFlagsChanged && (CGEventGetFlags(event) & (kCGEventFlagMaskCommand | kCGEventFlagMaskShift | kCGEventFlagMaskAlternate | kCGEventFlagMaskControl))));
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