package ivent

import (
	"context"
	"fmt"
	"github.com/ironpark/ivent/key"
	"testing"
	"time"
)

func TestStart(t *testing.T) {
	Register(
		NewComb([]key.Code{key.A, key.S, key.E}, func() { fmt.Println("ASD") }, AllowOtherInputs),
	)
	go Start(context.Background())
	time.Sleep(10 * time.Second)
}
