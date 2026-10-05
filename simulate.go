package ivent

import (
	"time"

	"github.com/ironpark/ivent/internal/hook"
	"github.com/ironpark/ivent/internal/simulate"
	"github.com/ironpark/ivent/key"
)

func init() {
	simulate.Attach = func(x any, now func() time.Time, apps func(int) (string, string)) simulate.Target {
		h := x.(*Hook)
		h.mu.Lock()
		defer h.mu.Unlock()
		h.now = now
		h.keyState = func(key.Code) (bool, bool) { return false, false }
		h.canSuppress = func() bool { return true }
		h.lookupApp = func(pid int) App {
			name, id := apps(pid)
			return App{PID: pid, Name: name, ID: id}
		}
		return simTarget{h}
	}
}

type simTarget struct{ h *Hook }

func (s simTarget) Handle(ev hook.Event) bool { return s.h.handle(ev) }

func (s simTarget) Flush() {
	for {
		select {
		case c := <-s.h.dispatch:
			// Async callbacks also run synchronously, so tests see their effects after Flush.
			c.fn(c.ev)
		default:
			return
		}
	}
}
