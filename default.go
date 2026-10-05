package ivent

import (
	"context"
	"runtime"
	"time"

	"github.com/ironpark/ivent/internal/hook"
	"github.com/ironpark/ivent/key"
)

// Default is the Hook used by the package-level functions.
var Default = New()

// Run runs the Default hook. See Hook.Run.
func Run(ctx context.Context) error {
	return Default.Run(ctx)
}

// Bind binds fn on the Default hook. See Hook.Bind.
func Bind(spec string, fn func(Event), opts ...BindOption) (unbind func(), err error) {
	return Default.Bind(spec, fn, opts...)
}

// Record records a combination on the Default hook. See Hook.Record.
func Record(ctx context.Context) (key.Combo, error) {
	return Default.Record(ctx)
}

// Permission reports what the process is allowed to do.
type Permission struct {
	// Monitor allows listening to input (Input Monitoring on macOS).
	Monitor bool
	// Control allows blocking input with Suppress and sending input (Accessibility on macOS,
	// write access to /dev/uinput on Linux).
	Control bool
}

// CheckPermission reports the input permissions of the process. If prompt is true, macOS asks the user
// to grant missing permissions.
func CheckPermission(prompt bool) Permission {
	monitor, control := hook.Permissions(prompt)
	return Permission{Monitor: monitor, Control: control}
}

// WaitForPermission asks for the permissions in need (on macOS) and waits until they are granted
// or ctx is done. On macOS, Input Monitoring usually takes effect only after the process restarts,
// while Accessibility takes effect immediately. On Linux it fails at once with ErrUnsupported if the
// permissions are missing, since they cannot be granted while the process runs.
func WaitForPermission(ctx context.Context, need Permission) error {
	granted := func(p Permission) bool {
		return (!need.Monitor || p.Monitor) && (!need.Control || p.Control)
	}
	if granted(CheckPermission(true)) {
		return nil
	}
	if runtime.GOOS == "linux" {
		return ErrUnsupported
	}
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if granted(CheckPermission(false)) {
				return nil
			}
		}
	}
}
