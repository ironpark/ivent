package ivent

import (
	"context"
	"github.com/ironpark/ivent/internal/keyhook"
	"github.com/ironpark/ivent/key"
	"strings"
	"sync"
)

var (
	combinations []*Comb
	mu           = sync.RWMutex{}
)

type Opt int

const (
	AllowOtherInputs = Opt(0x01)
	AllowRepeat      = Opt(0x02)
)

type Comb struct {
	Combination key.Table
	Callback    func()
	Option      Opt
}

func init() {
	// Register the update callback to be called whenever the key state is updated.
	keyhook.State.AddUpdateCallback(update)
}

// NewCombFromStr creates a new Comb from a string of key codes.
// Example input: "A+S+D"
// It splits the string by "+" and creates a key combination.
func NewCombFromStr(codes string, callback func(), options ...Opt) *Comb {
	codeNames := strings.Split(codes, "+")
	comb := key.MakeTable(key.Codes(codeNames...)...)
	var option Opt
	for _, opt := range options {
		option |= opt
	}
	return &Comb{
		Combination: comb,
		Callback:    callback,
		Option:      option,
	}
}

// NewComb creates a new Comb from a slice of key codes.
// It creates a key combination and assigns the provided callback and options.
func NewComb(codes []key.Code, callback func(), options ...Opt) *Comb {
	comb := key.MakeTable(codes...)
	var option Opt
	for _, opt := range options {
		option |= opt
	}
	return &Comb{
		Combination: comb,
		Callback:    callback,
		Option:      option,
	}
}

// update is the callback function that gets called whenever the key state is updated.
// It checks the current state against registered combinations and triggers the appropriate callbacks.
func update(state key.Table, pressed bool) {
	if !pressed {
		return
	}
	mu.RLock()
	defer mu.RUnlock()
	for _, comb := range combinations {
		if comb.Combination.Eq(state) {
			comb.Callback()
			continue
		}
		if comb.Option&AllowOtherInputs == 0 && comb.Combination.IsSubsetOf(state) {
			comb.Callback()
			continue
		}
	}
}

// Start begins listening for input events and runs until the context is done.
func Start(ctx context.Context) {
	go keyhook.Start()
	<-ctx.Done()
	keyhook.Stop()
}

// Register registers new key combinations to be monitored.
// It adds the combinations to the list of monitored combinations.
func Register(comb ...*Comb) {
	mu.Lock()
	defer mu.Unlock()
	for _, c := range comb {
		combinations = append(combinations, c)
	}
}

// Reset clears the list of monitored combinations.
func Reset() {
	mu.Lock()
	defer mu.Unlock()
	combinations = nil
}

// Remove removes a specific combination from the list of monitored combinations.
func Remove(comb *Comb) {
	mu.Lock()
	defer mu.Unlock()
	for i, c := range combinations {
		if c == comb {
			combinations = append(combinations[:i], combinations[i+1:]...)
			return
		}
	}
}
