#ifndef HOOK_DARWIN_H
#define HOOK_DARWIN_H

#include <stdint.h>
#include <stdbool.h>
// Enum for listen modes
typedef enum {
    LISTEN_KEYBOARD = 1,
    LISTEN_MOUSE = 2,
    LISTEN_MOUSEANDKEYBOARD = 3
} ListenMode;

// Key Events
void keyEventGoCallback(int pid, uint8_t keyCode, uint8_t down);
// Mouse Events
void mouseEventGoCallback(int pid, int x, int y, int button, bool isDown);
void mouseMoveGoCallback(int pid, int x, int y);
void mouseWheelGoCallback(int pid, int x, int y, int deltaY);
// Start and Stop
void stop();
int start(ListenMode mode);
#endif // HOOK_DARWIN_H