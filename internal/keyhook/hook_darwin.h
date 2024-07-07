
#ifndef __KEY_HOOK_H__
#define __KEY_HOOK_H__
#include <stdlib.h>
#include <stdint.h>

//pid C.int, keyCode C.uint8_t, down C.uint8_t
void keyEventGoCallback(int pid, uint8_t keyCode, uint8_t down);
void stop();
void start();
#endif