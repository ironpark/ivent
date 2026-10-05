#ifndef HOOK_COMMON_H
#define HOOK_COMMON_H

#include <stdbool.h>

// Event kinds passed to goEvent. Must match hook.Kind.
enum {
    IVENT_KEY_DOWN = 1,
    IVENT_KEY_UP = 2,
    IVENT_MOUSE_MOVE = 3,
    IVENT_MOUSE_WHEEL = 4,
};

// Event flags. Must match hook.FlagInjected / hook.FlagOwn.
enum {
    IVENT_FLAG_INJECTED = 1,
    IVENT_FLAG_OWN = 2,
    // Key events of mouse buttons carry a button index (0 left, 1 right, 2 middle, 3-4 extra) as code.
    IVENT_FLAG_BUTTON = 4,
};

// Implemented in Go. goEvent returns non-zero to block the event.
int goEvent(int kind, int code, int x, int y, int delta, int pid, int flags);
void goReset(void);
// goReady is called by start once the hook is installed, with whether it can block events.
void goReady(bool suppress);

// Implemented per platform.
int start(bool keyboard, bool mouse);
// charForKey returns the character a key types without modifiers on the current layout, or -1.
int charForKey(int code);
void stop(void);
bool keyPressed(int code);
int sendKey(int code, bool down);

#endif // HOOK_COMMON_H
