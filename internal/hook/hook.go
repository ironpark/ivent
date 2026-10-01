package hook

import (
	"context"
	"errors"
	"runtime"
	"sync/atomic"
)

var (
	// ErrAlreadyRunning is returned by Start when the hook is already running.
	ErrAlreadyRunning = errors.New("ivent: hook is already running")
	// ErrHookFailed is returned by Start when the OS input hook cannot be installed.
	ErrHookFailed = errors.New("ivent: failed to install input hook")
)

var running atomic.Bool

// Start installs the OS input hook and processes events until ctx is done.
// Key state callbacks are invoked on the calling goroutine.
func Start(ctx context.Context) error {
	if !running.CompareAndSwap(false, true) {
		return ErrAlreadyRunning
	}
	defer running.Store(false)

	done := make(chan error, 1)
	go func() {
		// The hook and its run loop / message loop must live on a single OS thread.
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		done <- platformStart()
	}()

	// platformStop is only valid once the hook thread has signalled ready.
	select {
	case <-readyCh:
	case err := <-done:
		return err
	}

	for {
		select {
		case event := <-eventCh:
			State.SetKeyState(event.keyCode(), event.down())
		case <-ctx.Done():
			platformStop()
			return <-done
		}
	}
}
