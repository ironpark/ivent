package hook

import (
	"bytes"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// uinputSetup is struct uinput_setup.
type uinputSetup struct {
	ID           [4]uint16 // bustype, vendor, product, version
	Name         [80]byte
	FFEffectsMax uint32
}

// uinput is a virtual keyboard and mouse created through /dev/uinput. Closing it destroys the device.
type uinput struct {
	*os.File
	sysfs string // sysfs path, to recognize the device in /proc/bus/input/devices
}

// virtualKeys lists the key codes the virtual device reports. Joystick, gamepad and tablet buttons are
// left out, so that the device is not taken for one of those.
var virtualKeys = [][2]uint16{{1, 0xff}, {btnLeft, 0x117}, {0x160, 0x21f}}

func newUinput() (_ *uinput, err error) {
	f, err := os.OpenFile("/dev/uinput", os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			f.Close()
		}
	}()
	for _, ev := range []int{evSyn, evKey, evRel, evLed} {
		if err := ioctlInt(f, uiSetEvBit, ev); err != nil {
			return nil, err
		}
	}
	for _, r := range virtualKeys {
		for c := r[0]; c <= r[1]; c++ {
			if err := ioctlInt(f, uiSetKeyBit, int(c)); err != nil {
				return nil, err
			}
		}
	}
	for _, rel := range []int{relX, relY, relHWheel, relWheel, relWheelHiRes, relHWheelHiRes} {
		if err := ioctlInt(f, uiSetRelBit, rel); err != nil {
			return nil, err
		}
	}
	for led := range 5 { // Num, Caps and Scroll Lock, Compose, Kana
		if err := ioctlInt(f, uiSetLedBit, led); err != nil {
			return nil, err
		}
	}
	setup := uinputSetup{ID: [4]uint16{busUSB, 0x6976, 0x656e, 1}}
	copy(setup.Name[:], "ivent virtual input")
	if err := ioctlPtr(f, uiDevSetup, unsafe.Pointer(&setup)); err != nil {
		return nil, err
	}
	if err := ioctlInt(f, uiDevCreate, 0); err != nil {
		return nil, err
	}
	var name [sysnameSize]byte
	if err := ioctlPtr(f, uiGetSysname, unsafe.Pointer(&name)); err != nil {
		return nil, err
	}
	sysname, _, _ := bytes.Cut(name[:], []byte{0})
	return &uinput{File: f, sysfs: "/devices/virtual/input/" + string(sysname)}, nil
}

// settle waits for the display server to open a new device. Events sent before that are lost.
func settle() {
	time.Sleep(200 * time.Millisecond)
}

func (v *uinput) sendKey(code uint16, down bool) error {
	value := int32(0)
	if down {
		value = 1
	}
	_, err := v.Write(appendSyn(appendEvent(nil, evKey, code, value)))
	return err
}

var (
	senderMu    sync.Mutex
	sender      *uinput
	senderSysfs atomic.Pointer[string] // read by the hook without waiting for a device being created
)

// sendDevice returns the virtual device used by Send, creating it on first use.
// It needs write access to /dev/uinput, usually by being root or through a udev rule.
func sendDevice() (*uinput, error) {
	senderMu.Lock()
	defer senderMu.Unlock()
	if sender == nil {
		v, err := newUinput()
		if err != nil {
			return nil, fmt.Errorf("ivent: cannot create a virtual input device (needs write access to /dev/uinput): %w", err)
		}
		senderSysfs.Store(&v.sysfs)
		settle()
		sender = v
	}
	return sender, nil
}
