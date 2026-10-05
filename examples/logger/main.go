// Command logger prints every input event. Press Ctrl+C in the terminal to quit.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/ironpark/ivent"
	"github.com/ironpark/ivent/key"
)

func main() {
	h := ivent.New(ivent.WithInjected())
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		for e := range h.Events() {
			switch e.Kind {
			case ivent.MouseMove:
				continue // too noisy
			case ivent.MouseWheel:
				fmt.Printf("wheel %+d at %d,%d\n", e.Delta, e.X, e.Y)
			default:
				fmt.Printf("%-7s %-12s repeat=%-5v injected=%-5v held=%-20s app=%s\n",
					e.Kind, key.DisplayName(e.Key), e.Repeat, e.Injected, e.Combo().Format(key.StyleText), e.App().Name)
			}
		}
	}()
	if err := h.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
