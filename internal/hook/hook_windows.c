#include <windows.h>
#include <stdio.h>
#include "hook_windows.h"

static HHOOK keyboardHook = NULL;
static HHOOK mouseHook = NULL;
static SRWLOCK lock = SRWLOCK_INIT;
static DWORD hookThreadId = 0;

LRESULT CALLBACK KeyboardProc(int nCode, WPARAM wParam, LPARAM lParam) {
    if (nCode >= 0) {
        KBDLLHOOKSTRUCT* kbdStruct = (KBDLLHOOKSTRUCT*)lParam;
        DWORD processId;
        HWND foregroundWindow = GetForegroundWindow();
        GetWindowThreadProcessId(foregroundWindow, &processId);

        bool isKeyDown = (wParam == WM_KEYDOWN || wParam == WM_SYSKEYDOWN);
        keyEventGoCallback((int)processId, kbdStruct->vkCode, isKeyDown);
    }
    return CallNextHookEx(keyboardHook, nCode, wParam, lParam);
}

LRESULT CALLBACK MouseProc(int nCode, WPARAM wParam, LPARAM lParam) {
    if (nCode >= 0) {
        MSLLHOOKSTRUCT* mouseStruct = (MSLLHOOKSTRUCT*)lParam;
        DWORD processId;
        HWND foregroundWindow = GetForegroundWindow();
        GetWindowThreadProcessId(foregroundWindow, &processId);

        switch (wParam) {
            case WM_LBUTTONDOWN:
            case WM_LBUTTONUP:
            case WM_RBUTTONDOWN:
            case WM_RBUTTONUP:
            case WM_MBUTTONDOWN:
            case WM_MBUTTONUP: {
                int button = 0;
                if (wParam == WM_LBUTTONDOWN || wParam == WM_LBUTTONUP) button = 0;
                else if (wParam == WM_RBUTTONDOWN || wParam == WM_RBUTTONUP) button = 1;
                else if (wParam == WM_MBUTTONDOWN || wParam == WM_MBUTTONUP) button = 2;

                bool isDown = (wParam == WM_LBUTTONDOWN || wParam == WM_RBUTTONDOWN || wParam == WM_MBUTTONDOWN);
                mouseEventGoCallback((int)processId, mouseStruct->pt.x, mouseStruct->pt.y, button, isDown);
                break;
            }
            case WM_MOUSEMOVE:
                mouseMoveGoCallback((int)processId, mouseStruct->pt.x, mouseStruct->pt.y);
                break;
            case WM_MOUSEWHEEL:
                mouseWheelGoCallback((int)processId, mouseStruct->pt.x, mouseStruct->pt.y, GET_WHEEL_DELTA_WPARAM(wParam) / WHEEL_DELTA);
                break;
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
int start(ListenMode mode) {
    MSG msg;
    // Make sure this thread has a message queue before stop can post WM_QUIT to it.
    PeekMessage(&msg, NULL, WM_USER, WM_USER, PM_NOREMOVE);

    if (mode & LISTEN_KEYBOARD) {
        keyboardHook = SetWindowsHookEx(WH_KEYBOARD_LL, KeyboardProc, NULL, 0);
        if (!keyboardHook) goto fail;
    }
    if (mode & LISTEN_MOUSE) {
        mouseHook = SetWindowsHookEx(WH_MOUSE_LL, MouseProc, NULL, 0);
        if (!mouseHook) goto fail;
    }

    AcquireSRWLockExclusive(&lock);
    hookThreadId = GetCurrentThreadId();
    ReleaseSRWLockExclusive(&lock);

    hookReadyGoCallback();
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
void stop() {
    AcquireSRWLockExclusive(&lock);
    if (hookThreadId != 0) {
        PostThreadMessage(hookThreadId, WM_QUIT, 0, 0);
    }
    ReleaseSRWLockExclusive(&lock);
}
