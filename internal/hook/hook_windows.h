#ifndef HOOK_WINDOWS_H
#define HOOK_WINDOWS_H

#include <stdbool.h>

bool processPath(int pid, char *out, int outLen);
int foregroundPID(void);

#endif // HOOK_WINDOWS_H
