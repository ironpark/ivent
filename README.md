Ivent
========
Simple input event hook library for golang

## Warning
- This package is in a very early stage and may undergo breaking changes.
- While there are plans for future support on Windows and Linux, it currently only works on macOS.

## Usage
```bash
go get github.com/ironpark/ivent
```

```go
package main

import (
	"context"
	"fmt"
	"github.com/ironpark/ivent"
	"github.com/ironpark/ivent/key"
	"time"
)

func keyQWE() {
	fmt.Println("QWE")
}
func keyASD() {
	fmt.Println("ASD")
}
func keyZXC() {
	fmt.Println("ZXC")
}

func main() {
	// Create a context with a 10-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Register key combinations with their respective callbacks
	ivent.Register(
		ivent.NewComb([]key.Code{key.Q, key.W, key.E}, keyQWE), // Q+W+E combination
		ivent.NewCombFromStr("A+S+D", keyASD),                  // A+S+D combination from string
		// AllowOtherInputs: if true, allows other inputs after the key combination
		// For example, if you press this key combination, it also allows Z+X+C + <any keys>
		ivent.NewComb([]key.Code{key.Z, key.X, key.C}, keyZXC, ivent.AllowOtherInputs),
	)

	// Start listening for input events with the given context
	ivent.Start(ctx)
}
```