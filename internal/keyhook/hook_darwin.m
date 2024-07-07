#include <ApplicationServices/ApplicationServices.h>
#include <stdio.h>
#import "hook_darwin.h"
#include <pthread.h>

// Go 콜백 함수를 참조하기 위한 함수 포인터 타입
// void keyEventGoCallback(pid_t pid, CGKeyCode keycode, bool down);

static CFMachPortRef eventTap = NULL;
static CFRunLoopSourceRef runLoopSource = NULL;
static pthread_mutex_t mutex = PTHREAD_MUTEX_INITIALIZER;

static CGEventRef keyEventCGEventCallback(CGEventTapProxy proxy, CGEventType type, CGEventRef event, void *refcon) {
    if (type == kCGEventKeyDown || type == kCGEventKeyUp || type == kCGEventFlagsChanged) {
        // 이벤트의 프로세스 ID를 확인
        pid_t pid = (pid_t)CGEventGetIntegerValueField(event, kCGEventTargetUnixProcessID);
        // 키 코드 가져오기
        CGKeyCode keycode = (CGKeyCode)CGEventGetIntegerValueField(event, kCGKeyboardEventKeycode);
        // 키 다운/업 이벤트 출력
        bool keyDown = (type == kCGEventKeyDown || (type == kCGEventFlagsChanged && (CGEventGetFlags(event) & (kCGEventFlagMaskCommand | kCGEventFlagMaskShift | kCGEventFlagMaskAlternate | kCGEventFlagMaskControl))));
        keyEventGoCallback(pid, keycode, keyDown);
    }
    return event;
}

void start() {
    pthread_mutex_lock(&mutex);
    if (eventTap == NULL) {
        CGEventMask eventMask = (1 << kCGEventKeyDown) | (1 << kCGEventKeyUp) | (1 << kCGEventFlagsChanged);
        eventTap = CGEventTapCreate(kCGSessionEventTap, kCGHeadInsertEventTap, kCGEventTapOptionListenOnly, eventMask, keyEventCGEventCallback, NULL);
        if (!eventTap) {
            fprintf(stderr, "Failed to create event tap\n");
            pthread_mutex_unlock(&mutex);
            exit(1);
        }
        runLoopSource = CFMachPortCreateRunLoopSource(kCFAllocatorDefault, eventTap, 0);
        CFRunLoopAddSource(CFRunLoopGetCurrent(), runLoopSource, kCFRunLoopCommonModes);
        CGEventTapEnable(eventTap, true);
    }
    pthread_mutex_unlock(&mutex);
    CFRunLoopRun();
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
        CFRunLoopStop(CFRunLoopGetCurrent());
    }
    pthread_mutex_unlock(&mutex);
}