#include <windows.h>
#include <wctype.h>
#include "hook_common.h"
#include "hook_windows.h"

// Marks events sent by sendKey so they can be recognized when they come back through the hook.
#define IVENT_EXTRA_INFO ((ULONG_PTR)0x69766E74)

static HHOOK keyboardHook = NULL;
static HHOOK mouseHook = NULL;
static SRWLOCK lock = SRWLOCK_INIT;
static DWORD hookThreadId = 0;

// charForKey returns the character a key types without modifiers on the foreground window's layout, or -1.
int charForKey(int code) {
    HKL layout = GetKeyboardLayout(GetWindowThreadProcessId(GetForegroundWindow(), NULL));
    UINT ch = MapVirtualKeyExW((UINT)code, MAPVK_VK_TO_CHAR, layout);
    // The high bit marks a dead key.
    if (ch == 0 || (ch & 0x80000000)) {
        return -1;
    }
    return (int)towlower((wint_t)ch);
}

int foregroundPID(void) {
    DWORD pid = 0;
    HWND window = GetForegroundWindow();
    if (window) {
        GetWindowThreadProcessId(window, &pid);
    }
    return (int)pid;
}

static int eventFlags(DWORD flags, DWORD injectedFlag, ULONG_PTR extraInfo) {
    if (extraInfo == IVENT_EXTRA_INFO) {
        return IVENT_FLAG_INJECTED | IVENT_FLAG_OWN;
    }
    return (flags & injectedFlag) ? IVENT_FLAG_INJECTED : 0;
}

LRESULT CALLBACK KeyboardProc(int nCode, WPARAM wParam, LPARAM lParam) {
    if (nCode == HC_ACTION) {
        KBDLLHOOKSTRUCT* kbd = (KBDLLHOOKSTRUCT*)lParam;
        // AltGr sends a fake Left Ctrl (scan code 0x21D) before Right Alt. Ignore it so that
        // AltGr is not mistaken for Ctrl+Alt.
        bool fakeCtrl = kbd->vkCode == VK_LCONTROL && (kbd->scanCode & 0x200);
        if (!fakeCtrl && kbd->vkCode < 256) {
            bool down = wParam == WM_KEYDOWN || wParam == WM_SYSKEYDOWN;
            int flags = eventFlags(kbd->flags, LLKHF_INJECTED, kbd->dwExtraInfo);
            // Keyboard events carry no position; report the cursor like the other platforms do.
            POINT pt = {0};
            GetCursorPos(&pt);
            if (goEvent(down ? IVENT_KEY_DOWN : IVENT_KEY_UP, (int)kbd->vkCode, pt.x, pt.y, 0, foregroundPID(), flags)) {
                return 1;
            }
        }
    }
    return CallNextHookEx(keyboardHook, nCode, wParam, lParam);
}

LRESULT CALLBACK MouseProc(int nCode, WPARAM wParam, LPARAM lParam) {
    if (nCode == HC_ACTION) {
        MSLLHOOKSTRUCT* mouse = (MSLLHOOKSTRUCT*)lParam;
        int kind = 0, button = 0, delta = 0;
        switch (wParam) {
            case WM_LBUTTONDOWN: kind = IVENT_KEY_DOWN; button = 0; break;
            case WM_LBUTTONUP:   kind = IVENT_KEY_UP;   button = 0; break;
            case WM_RBUTTONDOWN: kind = IVENT_KEY_DOWN; button = 1; break;
            case WM_RBUTTONUP:   kind = IVENT_KEY_UP;   button = 1; break;
            case WM_MBUTTONDOWN: kind = IVENT_KEY_DOWN; button = 2; break;
            case WM_MBUTTONUP:   kind = IVENT_KEY_UP;   button = 2; break;
            case WM_XBUTTONDOWN:
            case WM_XBUTTONUP:
                kind = wParam == WM_XBUTTONDOWN ? IVENT_KEY_DOWN : IVENT_KEY_UP;
                button = HIWORD(mouse->mouseData) == XBUTTON1 ? 3 : 4;
                break;
            case WM_MOUSEMOVE:
                kind = IVENT_MOUSE_MOVE;
                break;
            case WM_MOUSEWHEEL:
                kind = IVENT_MOUSE_WHEEL;
                delta = GET_WHEEL_DELTA_WPARAM(mouse->mouseData) / WHEEL_DELTA;
                break;
        }
        if (kind) {
            int flags = eventFlags(mouse->flags, LLMHF_INJECTED, mouse->dwExtraInfo);
            // The target process only matters for buttons; skip the lookup for frequent moves.
            int pid = 0;
            if (kind == IVENT_KEY_DOWN || kind == IVENT_KEY_UP) {
                flags |= IVENT_FLAG_BUTTON;
                pid = foregroundPID();
            }
            if (goEvent(kind, button, mouse->pt.x, mouse->pt.y, delta, pid, flags)) {
                return 1;
            }
        }
    }
    return CallNextHookEx(mouseHook, nCode, wParam, lParam);
}

static void unhookAll() {
    if (keyboardHook) {
        UnhookWindowsHookEx(keyboardHook);
        keyboardHook = NULL;
    }
    if (mouseHook) {
        UnhookWindowsHookEx(mouseHook);
        mouseHook = NULL;
    }
}

// start returns 0 on success or the GetLastError code of the failed hook installation.
int start(bool keyboard, bool mouse) {
    MSG msg;
    // Make sure this thread has a message queue before stop can post WM_QUIT to it.
    PeekMessage(&msg, NULL, WM_USER, WM_USER, PM_NOREMOVE);

    if (keyboard) {
        keyboardHook = SetWindowsHookEx(WH_KEYBOARD_LL, KeyboardProc, NULL, 0);
        if (!keyboardHook) goto fail;
    }
    if (mouse) {
        mouseHook = SetWindowsHookEx(WH_MOUSE_LL, MouseProc, NULL, 0);
        if (!mouseHook) goto fail;
    }

    AcquireSRWLockExclusive(&lock);
    hookThreadId = GetCurrentThreadId();
    ReleaseSRWLockExclusive(&lock);

    // Low-level hooks can always block events.
    goReady(true);
    while (GetMessage(&msg, NULL, 0, 0) > 0) {
        TranslateMessage(&msg);
        DispatchMessage(&msg);
    }

    AcquireSRWLockExclusive(&lock);
    hookThreadId = 0;
    ReleaseSRWLockExclusive(&lock);
    // Hooks are removed by the thread that installed them.
    unhookAll();
    return 0;

fail: {
        DWORD err = GetLastError();
        unhookAll();
        return err ? (int)err : -1;
    }
}

// stop can be called from any thread once start has signalled ready.
// WM_QUIT stays queued until the message loop reads it.
void stop(void) {
    AcquireSRWLockExclusive(&lock);
    if (hookThreadId != 0) {
        PostThreadMessage(hookThreadId, WM_QUIT, 0, 0);
    }
    ReleaseSRWLockExclusive(&lock);
}

bool keyPressed(int code) {
    return (GetAsyncKeyState(code) & 0x8000) != 0;
}

static bool isExtendedKey(int vk) {
    switch (vk) {
        case VK_RCONTROL: case VK_RMENU: case VK_LWIN: case VK_RWIN: case VK_APPS:
        case VK_INSERT: case VK_DELETE: case VK_HOME: case VK_END: case VK_PRIOR: case VK_NEXT:
        case VK_LEFT: case VK_UP: case VK_RIGHT: case VK_DOWN:
        case VK_NUMLOCK: case VK_DIVIDE: case VK_SNAPSHOT:
            return true;
        default:
            return false;
    }
}

// sendKey returns 0 on success or the GetLastError code.
int sendKey(int code, bool down) {
    INPUT input = {0};
    input.type = INPUT_KEYBOARD;
    input.ki.wVk = (WORD)code;
    input.ki.wScan = (WORD)MapVirtualKey(code, MAPVK_VK_TO_VSC);
    input.ki.dwFlags = (down ? 0 : KEYEVENTF_KEYUP) | (isExtendedKey(code) ? KEYEVENTF_EXTENDEDKEY : 0);
    input.ki.dwExtraInfo = IVENT_EXTRA_INFO;
    if (SendInput(1, &input, sizeof(INPUT)) != 1) {
        DWORD err = GetLastError();
        return err ? (int)err : -1;
    }
    return 0;
}

// processPath writes the UTF-8 executable path of a process to out.
bool processPath(int pid, char *out, int outLen) {
    HANDLE process = OpenProcess(PROCESS_QUERY_LIMITED_INFORMATION, FALSE, (DWORD)pid);
    if (!process) {
        return false;
    }
    WCHAR path[MAX_PATH * 2];
    DWORD size = sizeof(path) / sizeof(path[0]);
    BOOL ok = QueryFullProcessImageNameW(process, 0, path, &size);
    CloseHandle(process);
    if (!ok) {
        return false;
    }
    return WideCharToMultiByte(CP_UTF8, 0, path, -1, out, outLen, NULL, NULL) > 0;
}

