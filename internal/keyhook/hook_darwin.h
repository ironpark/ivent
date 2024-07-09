#ifndef HOOK_DARWIN_H
#define HOOK_DARWIN_H

#include <stdint.h>
#include <stdbool.h>

void keyEventGoCallback(int pid, uint8_t keyCode, uint8_t down);
void stop();
int start();
#endif // HOOK_DARWIN_H