package hook

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/ironpark/ivent/key"
)

// These tests create real input devices. They need write access to /dev/uinput and read access to
// /dev/input, for example in a container started with --privileged -v /dev/input:/dev/input.

func requireUinput(t *testing.T) {
	t.Helper()
	if _, control := Permissions(false); !control {
		t.Skip("needs write access to /dev/uinput")
	}
}

// newKeyboard creates a uinput device, for example to play the part of a physical keyboard.
func newKeyboard(t *testing.T) *uinput {
	t.Helper()
	kb, err := newUinput()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { kb.Close() })
	return kb
}

// openNode opens the /dev/input node of a uinput device.
func openNode(t *testing.T, v *uinput, flag int) *os.File {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for {
		infos, _ := listDevices()
		for _, info := range infos {
			if info.sysfs == v.sysfs {
				if f, err := os.OpenFile(info.path, flag, 0); err == nil {
					t.Cleanup(func() { f.Close() })
					return f
				}
			}
		}
		if time.Now().After(deadline) {
			t.Fatalf("no device node for %s", v.sysfs)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// readEvents reads events from f until n match or a second passes, and returns their codes.
func readEvents(t *testing.T, f *os.File, n int, match func(typ, code uint16, value int32) bool) (codes []uint16) {
	t.Helper()
	f.SetReadDeadline(time.Now().Add(time.Second))
	defer f.SetReadDeadline(time.Time{})
	buf := make([]byte, inputEventSize)
	for len(codes) < n {
		if _, err := f.Read(buf); err != nil {
			return codes
		}
		if typ, code, value := decodeEvent(buf); match(typ, code, value) {
			codes = append(codes, code)
		}
	}
	return codes
}

// readPresses reads up to n key presses from f.
func readPresses(t *testing.T, f *os.File, n int) []uint16 {
	return readEvents(t, f, n, func(typ, _ uint16, value int32) bool { return typ == evKey && value == 1 })
}

type chanHandler struct {
	events chan Event
	block  key.Code
}

func (h chanHandler) Handle(e Event) bool {
	h.events <- e
	return e.Code == h.block
}
func (chanHandler) Reset() {}

func (h chanHandler) next(t *testing.T) Event {
	t.Helper()
	select {
	case e := <-h.events:
		return e
	case <-time.After(2 * time.Second):
		t.Fatal("no event")
		return Event{}
	}
}

func TestUinputRun(t *testing.T) {
	requireUinput(t)
	kb := newKeyboard(t)
	h := chanHandler{events: make(chan Event, 16)}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	ready := make(chan bool, 1)
	go func() {
		done <- Run(ctx, Config{Keyboard: true, Exclusive: true, OnReady: func(s bool) { ready <- s }}, h)
	}()
	select {
	case s := <-ready:
		if !s {
			t.Fatal("cannot suppress with Exclusive")
		}
	case err := <-done:
		t.Fatal(err)
	}

	kb.sendKey(uint16(key.A), true)
	if e := h.next(t); e.Code != key.A || e.Kind != KeyDown || e.Flags != FlagInjected {
		t.Fatalf("event = %+v, want injected A down", e)
	}
	if pressed, ok := KeyState(key.A); !pressed || !ok {
		t.Fatalf("KeyState(A) = %v, %v while held", pressed, ok)
	}
	kb.sendKey(uint16(key.A), false)
	h.next(t)
	if pressed, ok := KeyState(key.A); pressed || !ok {
		t.Fatalf("KeyState(A) = %v, %v after release", pressed, ok)
	}

	// A keyboard plugged in while running.
	late := newKeyboard(t)
	deadline := time.Now().Add(3 * time.Second)
	for got := false; !got; {
		if time.Now().After(deadline) {
			t.Fatal("hot-plugged keyboard is not read")
		}
		late.sendKey(uint16(key.B), true)
		late.sendKey(uint16(key.B), false)
		select {
		case e := <-h.events:
			if e.Code != key.B {
				t.Fatalf("event = %+v, want B", e)
			}
			got = true
		case <-time.After(100 * time.Millisecond):
		}
	}

	cancel()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

func TestUinputGrab(t *testing.T) {
	requireUinput(t)
	out := newKeyboard(t)
	kb := newKeyboard(t)
	forwarded := openNode(t, out, os.O_RDONLY)
	other := openNode(t, kb, os.O_RDONLY) // another program reading the keyboard
	h := chanHandler{events: make(chan Event, 16), block: key.B}
	useHandler(t, h)

	// The grab waits until no key is held.
	kb.sendKey(uint16(key.LeftShift), true)
	d := &device{f: openNode(t, kb, os.O_RDWR), out: out}
	l := &linuxHook{out: out, devices: map[string]*device{"kb": d}}
	go l.forwardLEDs()
	go readDevice(d)
	time.Sleep(50 * time.Millisecond)
	if d.grabbed.Load() {
		t.Fatal("grabbed while a key is held")
	}
	kb.sendKey(uint16(key.LeftShift), false)
	h.next(t)
	for deadline := time.Now().Add(time.Second); !d.grabbed.Load(); time.Sleep(10 * time.Millisecond) {
		if time.Now().After(deadline) {
			t.Fatal("not grabbed after the key was released")
		}
	}
	readPresses(t, other, 1) // drain the Shift press

	kb.sendKey(uint16(key.A), true)
	kb.sendKey(uint16(key.B), true) // blocked
	kb.sendKey(uint16(key.C), true)
	for range 3 {
		h.next(t)
	}
	if got := readPresses(t, forwarded, 3); len(got) != 2 || got[0] != uint16(key.A) || got[1] != uint16(key.C) {
		t.Fatalf("forwarded %v, want A and C", got)
	}
	if got := readPresses(t, other, 1); len(got) != 0 {
		t.Fatalf("grabbed keyboard leaked %v", got)
	}

	// LEDs set on the forwarding device reach the grabbed keyboard.
	openNode(t, out, os.O_WRONLY).Write(appendSyn(ev(evLed, 1, 1)))
	capsLED := func(typ, code uint16, value int32) bool { return typ == evLed && code == 1 && value == 1 }
	if got := readEvents(t, kb.File, 1, capsLED); len(got) != 1 {
		t.Fatal("Caps Lock LED not forwarded")
	}
}

func TestUinputSend(t *testing.T) {
	requireUinput(t)
	v, err := sendDevice()
	if err != nil {
		t.Fatal(err)
	}
	node := openNode(t, v, os.O_RDONLY)
	if err := Send(key.Z, true); err != nil {
		t.Fatal(err)
	}
	if err := Send(key.Z, false); err != nil {
		t.Fatal(err)
	}
	if got := readPresses(t, node, 1); len(got) != 1 || got[0] != uint16(key.Z) {
		t.Fatalf("sent %v, want Z", got)
	}
}
