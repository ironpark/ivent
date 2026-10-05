package ivent_test

import (
	"context"
	"fmt"
	"time"

	"github.com/ironpark/ivent"
	"github.com/ironpark/ivent/iventtest"
	"github.com/ironpark/ivent/key"
)

func Example() {
	h := ivent.New()
	h.Bind("Ctrl+Shift+A", func(e ivent.Event) {
		fmt.Println("Ctrl+Shift+A in", e.App().Name)
	})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := h.Run(ctx); err != nil {
		fmt.Println(err) // e.g. ivent.ErrHookFailed: missing permission
	}
}

func ExampleHook_Bind_sequence() {
	h := ivent.New()
	unbind, err := h.Bind("Ctrl+K Ctrl+C", func(ivent.Event) {
		fmt.Println("comment")
	}, ivent.SequenceTimeout(time.Second))
	if err != nil {
		panic(err)
	}
	defer unbind()
}

func ExampleSuppress() {
	h := ivent.New()
	// Blocking needs Accessibility permission on macOS; see Hook.CanSuppress.
	h.Bind("Cmd+Q", func(ivent.Event) {
		fmt.Println("Cmd+Q blocked in Safari")
	}, ivent.Suppress(), ivent.OnlyApps("Safari"))
}

func ExampleDoubleTap() {
	h := ivent.New()
	h.Bind("Shift", func(ivent.Event) {
		fmt.Println("double Shift")
	}, ivent.DoubleTap(300*time.Millisecond))
}

func ExampleByCharacter() {
	h := ivent.New()
	// Matches the key labelled Z on any layout (on AZERTY, the key at the US "W" position).
	h.Bind("Ctrl+Z", func(ivent.Event) { fmt.Println("undo") }, ivent.ByCharacter())
}

func ExampleHook_Record() {
	h := ivent.New()
	go h.Run(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	fmt.Println("Press a shortcut...")
	combo, err := h.Record(ctx)
	if err != nil {
		return
	}
	fmt.Println("Recorded", combo.Display())
	h.BindCombo(combo, func(ivent.Event) { fmt.Println("shortcut") })
}

func ExampleSend() {
	if err := ivent.SendString("Cmd+C"); err != nil {
		fmt.Println(err)
	}
}

func Example_testing() {
	h := ivent.New()
	h.Bind("Ctrl+K Ctrl+C", func(ivent.Event) { fmt.Println("comment") })

	sim := iventtest.New(h)
	sim.Type("Ctrl+K Ctrl+C")
	sim.Flush()
	// Output: comment
}

func ExampleCombo_Display() {
	c := key.MustParseCombo("Shift+Ctrl+A")
	fmt.Println(c.Format(key.StyleText))
	fmt.Println(c.Format(key.StyleSymbols))
	// Output:
	// Ctrl+Shift+A
	// ⌃⇧A
}
