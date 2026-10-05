// Package simulate connects package iventtest to the internals of package ivent.
package simulate

import (
	"time"

	"github.com/ironpark/ivent/internal/hook"
)

// Target is a Hook in simulation mode.
type Target interface {
	// Handle processes a synthetic event as if it came from the OS and reports whether it was blocked.
	Handle(hook.Event) bool
	// Flush runs the queued callbacks on the calling goroutine.
	Flush()
}

// Attach puts a Hook (an *ivent.Hook) in simulation mode: time comes from now, applications from apps,
// the OS key state is not consulted and blocking is always possible. It is set by package ivent.
var Attach func(h any, now func() time.Time, apps func(pid int) (name, id string)) Target
