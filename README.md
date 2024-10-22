<p align="center">
    <img src="./assets/images/ivent.png" width="256">
</p>

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


func main() {
	// Create a context with a 10-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// Register key combinations with their respective callbacks
	ivent.Register(
		// Q+W+E combination
		ivent.NewComb([]key.Code{key.Q, key.W, key.E}, func() {
			fmt.Println("QWE")
		}),
		// A+S+D combination from string
		ivent.NewCombFromStr("A+S+D", func() {
			fmt.Println("ASD")
		}),
		// AllowOtherInputs: if true, allows other inputs after the key combination
		// For example, if you press this key combination, it also allows Z+X+C + <any keys>
		ivent.NewComb([]key.Code{key.Z, key.X, key.C}, func() {
			fmt.Println("ZXC")
		}, ivent.AllowOtherInputs),
	)

	// Start listening for input events with the given context
	ivent.Start(ctx)
}
```

## TODO
- [ ] More options for key combinations
- [x] Generate key-code table `key/codes.go` using `go:generate`
    - [x] Alias Support
    - [ ] Add windows,linux key-code config in `key/keycode.json` file
- [ ] Add alias for key codes (e.g. `RightBracket` -> `}`) 
- [ ] Add Support mouse events
- [ ] Implement key event triggering
- [ ] Add support for Windows and Linux