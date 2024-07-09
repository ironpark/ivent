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
		NewComb([]key.Code{key.A, key.S, key.D}, func() { fmt.Println("ASD") }, AllowOtherInputs),
	)
	go Start(context.Background())
	time.Sleep(24 * 30 * time.Hour)
}
