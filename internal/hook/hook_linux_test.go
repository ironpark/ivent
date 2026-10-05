package hook

import (
	"encoding/binary"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/ironpark/ivent/key"
)

type recorder struct{ events []Event }

func (r *recorder) Handle(e Event) bool { r.events = append(r.events, e); return false }
func (r *recorder) Reset()              {}

func inputEvent(typ, code uint16, value int32) []byte {
	buf := make([]byte, inputEventSize)
	tv := inputEventSize - 8
	binary.NativeEndian.PutUint16(buf[tv:], typ)
	binary.NativeEndian.PutUint16(buf[tv+2:], code)
	binary.NativeEndian.PutUint32(buf[tv+4:], uint32(value))
	return buf
}

func TestReadDevice(t *testing.T) {
	rec := &recorder{}
	var h Handler = rec
	handler.Store(&h)
	defer handler.Store(nil)

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	for _, ev := range [][]byte{
		inputEvent(evKey, uint16(key.A), 1),
		inputEvent(evKey, uint16(key.A), 2), // auto-repeat
		inputEvent(evKey, uint16(key.A), 0),
		inputEvent(evKey, btnLeft+1, 1), // right button
		inputEvent(evRel, relWheel, -1),
		inputEvent(0, 0, 0), // EV_SYN, ignored
	} {
		w.Write(ev)
	}
	w.Close()
	readDevice(r)

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

func TestParseDevices(t *testing.T) {
	const devices = `I: Bus=0019 Vendor=0000 Product=0001 Version=0000
N: Name="Power Button"
H: Handlers=kbd event0

I: Bus=0011 Vendor=0001 Product=0001 Version=ab41
N: Name="AT Translated Set 2 keyboard"
H: Handlers=sysrq kbd leds event3

I: Bus=0011 Vendor=0002 Product=0013 Version=0006
N: Name="ImPS/2 Logitech Wheel Mouse"
H: Handlers=mouse0 event4

I: Bus=0000 Vendor=0000 Product=0000 Version=0000
N: Name="Lid Switch"
H: Handlers=event5
`
	got, err := parseDevices(strings.NewReader(devices), true, false)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"/dev/input/event0", "/dev/input/event3"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("keyboards = %v, want %v", got, want)
	}
	got, _ = parseDevices(strings.NewReader(devices), true, true)
	if len(got) != 3 {
		t.Fatalf("keyboards and mice = %v", got)
	}
}
