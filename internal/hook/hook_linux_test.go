package hook

import (
	"bytes"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/ironpark/ivent/key"
)

// recorder records events and blocks presses of block.
type recorder struct {
	events []Event
	block  key.Code
	resets int
}

func (r *recorder) Handle(e Event) bool {
	r.events = append(r.events, e)
	return e.Code == r.block && r.block != 0
}
func (r *recorder) Reset() { r.resets++ }

func useHandler(t *testing.T, h Handler) {
	handler.Store(&h)
	t.Cleanup(func() { handler.Store(nil) })
}

// pipeDevice returns a device that reads the given events.
func pipeDevice(t *testing.T, info inputDevice, events ...[]byte) *device {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { r.Close() })
	for _, ev := range events {
		w.Write(ev)
	}
	w.Close()
	return &device{inputDevice: info, f: r}
}

func ev(typ, code uint16, value int32) []byte {
	return appendEvent(nil, typ, code, value)
}

var syn = ev(evSyn, synReport, 0)

func TestReadDevice(t *testing.T) {
	rec := &recorder{}
	useHandler(t, rec)
	readDevice(pipeDevice(t, inputDevice{},
		ev(evKey, uint16(key.A), 1),
		ev(evKey, uint16(key.A), 2), // auto-repeat
		ev(evKey, uint16(key.A), 0),
		ev(evKey, btnLeft+1, 1), // right button
		ev(evRel, relWheel, -1),
		syn,
	))

	want := []Event{
		{Kind: KeyDown, Code: key.A},
		{Kind: KeyDown, Code: key.A},
		{Kind: KeyUp, Code: key.A},
		{Kind: KeyDown, Code: key.MouseRight},
		{Kind: MouseWheel, Delta: -1},
	}
	if !reflect.DeepEqual(rec.events, want) {
		t.Fatalf("events = %+v, want %+v", rec.events, want)
	}
}

func TestReadDeviceVirtual(t *testing.T) {
	rec := &recorder{}
	useHandler(t, rec)
	readDevice(pipeDevice(t, inputDevice{sysfs: "/devices/virtual/input/input9"}, ev(evKey, uint16(key.A), 1)))
	if len(rec.events) != 1 || rec.events[0].Flags != FlagInjected {
		t.Fatalf("events = %+v, want one injected event", rec.events)
	}
}

func TestReadDeviceForward(t *testing.T) {
	rec := &recorder{block: key.B}
	useHandler(t, rec)
	d := pipeDevice(t, inputDevice{},
		ev(evMsc, 4, 30), ev(evKey, uint16(key.A), 1), syn, // forwarded without MSC_SCAN
		ev(evKey, uint16(key.B), 1), syn, // blocked: empty frame
		ev(evKey, uint16(key.C), 1), ev(evSyn, synDropped, 0), ev(evKey, uint16(key.D), 1), syn, // dropped
		ev(evKey, uint16(key.E), 1), syn,
	)
	var out bytes.Buffer
	d.out = &out
	d.grabbed.Store(true)
	readDevice(d)

	// C is read before SYN_DROPPED, but its frame is discarded with the dropped events.
	want := bytes.Join([][]byte{ev(evKey, uint16(key.A), 1), syn, ev(evKey, uint16(key.E), 1), syn}, nil)
	if !bytes.Equal(out.Bytes(), want) {
		t.Fatalf("forwarded %x, want %x", out.Bytes(), want)
	}
	if rec.resets != 1 {
		t.Fatalf("resets = %d, want 1", rec.resets)
	}
}

func TestParseDevices(t *testing.T) {
	const devices = `I: Bus=0019 Vendor=0000 Product=0001 Version=0000
N: Name="Power Button"
S: Sysfs=/devices/LNXSYSTM:00/LNXPWRBN:00/input/input0
H: Handlers=kbd event0
B: EV=3
B: KEY=10000000000000 0

I: Bus=0011 Vendor=0001 Product=0001 Version=ab41
N: Name="AT Translated Set 2 keyboard"
S: Sysfs=/devices/platform/i8042/serio0/input/input3
H: Handlers=sysrq kbd leds event3
B: EV=120013
B: KEY=402000000 3803078f800d001 feffffdfffefffff fffffffffffffffe

I: Bus=0011 Vendor=0002 Product=0013 Version=0006
N: Name="SynPS/2 Synaptics TouchPad"
S: Sysfs=/devices/platform/i8042/serio1/input/input4
H: Handlers=mouse0 event4
B: KEY=e520 10000 0 0 0 0
B: ABS=660800011000003

I: Bus=0003 Vendor=6976 Product=656e Version=0001
N: Name="ivent virtual input"
S: Sysfs=/devices/virtual/input/input23
H: Handlers=sysrq kbd mouse1 leds event5
B: KEY=ffffffffffffffff fffffffffffffffe

I: Bus=0000 Vendor=0000 Product=0000 Version=0000
N: Name="Lid Switch"
S: Sysfs=/devices/LNXSYSTM:00/input/input6
H: Handlers=event6
`
	got, err := parseDevices(strings.NewReader(devices))
	if err != nil {
		t.Fatal(err)
	}
	want := []inputDevice{
		{path: "/dev/input/event0", sysfs: "/devices/LNXSYSTM:00/LNXPWRBN:00/input/input0", keyboard: true},
		{path: "/dev/input/event3", sysfs: "/devices/platform/i8042/serio0/input/input3", keyboard: true, letters: true},
		{path: "/dev/input/event4", sysfs: "/devices/platform/i8042/serio1/input/input4", mouse: true, abs: true},
		{path: "/dev/input/event5", sysfs: "/devices/virtual/input/input23", keyboard: true, mouse: true, letters: true},
		{path: "/dev/input/event6", sysfs: "/devices/LNXSYSTM:00/input/input6"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("devices =\n%+v\nwant\n%+v", got, want)
	}
	var grabbable []string
	for _, d := range got {
		if d.grabbable() {
			grabbable = append(grabbable, d.path)
		}
	}
	if !reflect.DeepEqual(grabbable, []string{"/dev/input/event3"}) {
		t.Fatalf("grabbable = %v", grabbable)
	}
}
