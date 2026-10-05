// Command hotkey shows the binding features of ivent. Press Ctrl+C in the terminal to quit.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/ironpark/ivent"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Ask for Input Monitoring (listening) and Accessibility (blocking) on macOS.
	need := ivent.Permission{Monitor: true, Control: true}
	if err := ivent.WaitForPermission(ctx, need); err != nil {
		log.Printf("permissions: %v (restart after granting Input Monitoring)", err)
	}

	h := ivent.New()
	say := func(msg string) func(ivent.Event) {
		return func(e ivent.Event) { fmt.Printf("%s (app: %s)\n", msg, e.App().Name) }
	}
	bind := func(spec, msg string, opts ...ivent.BindOption) {
		if _, err := h.Bind(spec, say(msg), opts...); err != nil {
			log.Fatal(err)
		}
	}

	bind("Ctrl+Shift+A", "Ctrl+Shift+A (either Ctrl and Shift)")
	bind("Ctrl+K Ctrl+C", "sequence Ctrl+K Ctrl+C")
	bind("Shift", "double Shift", ivent.DoubleTap(300*time.Millisecond))
	bind("Space", "held Space", ivent.Hold(time.Second))
	bind("RightCmd", "tapped Right Cmd / Win", ivent.OnRelease())
	bind("Ctrl+J", "blocked Ctrl+J", ivent.Suppress())
	bind("Ctrl+Z", "Ctrl+Z on the key labelled Z", ivent.ByCharacter())
	bind("Ctrl+MouseLeft", "Ctrl+click in a browser", ivent.OnlyApps("Safari", "Google Chrome", "chrome", "firefox"))

	// Ctrl+Shift+R records a new shortcut and binds it.
	if _, err := h.Bind("Ctrl+Shift+R", func(ivent.Event) {
		fmt.Println("press a shortcut to bind...")
		combo, err := h.Record(ctx)
		if err != nil {
			return
		}
		fmt.Println("bound", combo.Display())
		h.BindCombo(combo, say("recorded "+combo.Display()))
	}, ivent.Async(), ivent.OnRelease()); err != nil {
		log.Fatal(err)
	}

	go func() {
		time.Sleep(500 * time.Millisecond) // CanSuppress is known once the hook is running
		fmt.Println("listening... blocking available:", h.CanSuppress())
	}()
	if err := h.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
