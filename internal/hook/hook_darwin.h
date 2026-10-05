#ifndef HOOK_DARWIN_H
#define HOOK_DARWIN_H

#include <stdbool.h>

bool buttonPressed(int button);
int frontmostPID(void);
bool processInfo(int pid, char *name, int nameLen, char *ident, int identLen);
void permissions(bool prompt, bool *monitor, bool *control);

#endif // HOOK_DARWIN_H
