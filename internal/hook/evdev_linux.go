package hook

import (
	"bufio"
	"encoding/binary"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

// Linux input event types and codes (linux/input-event-codes.h).
const (
	evSyn = 0x00
	evKey = 0x01
	evRel = 0x02
	evMsc = 0x04
	evLed = 0x11

	synReport  = 0
	synDropped = 3

	relX           = 0x00
	relY           = 0x01
	relHWheel      = 0x06
	relWheel       = 0x08
	relWheelHiRes  = 0x0b
	relHWheelHiRes = 0x0c

	keyA     = 30
	btnLeft  = 0x110
	btnExtra = 0x114 // BTN_LEFT, RIGHT, MIDDLE, SIDE, EXTRA are consecutive
	keyMax   = 0x2ff

	busUSB = 0x03
)

const (
	timevalSize = int(unsafe.Sizeof(syscall.Timeval{}))
	// inputEventSize is sizeof(struct input_event): a timeval followed by type, code and value.
	inputEventSize = timevalSize + 8
)

func decodeEvent(buf []byte) (typ, code uint16, value int32) {
	return binary.NativeEndian.Uint16(buf[timevalSize:]),
		binary.NativeEndian.Uint16(buf[timevalSize+2:]),
		int32(binary.NativeEndian.Uint32(buf[timevalSize+4:]))
}

// appendEvent appends an input_event with a zero time, which the kernel fills in.
func appendEvent(buf []byte, typ, code uint16, value int32) []byte {
	buf = append(buf, make([]byte, timevalSize)...)
	buf = binary.NativeEndian.AppendUint16(buf, typ)
	buf = binary.NativeEndian.AppendUint16(buf, code)
	return binary.NativeEndian.AppendUint32(buf, uint32(value))
}

// appendSyn ends a frame of events.
func appendSyn(buf []byte) []byte {
	return appendEvent(buf, evSyn, synReport, 0)
}

// ioctl request numbers, using the asm-generic encoding (amd64, arm64, 386, arm, riscv64, ...).
const (
	iocWrite = 1
	iocRead  = 2
)

func ioc(dir, typ, nr, size uintptr) uintptr {
	return dir<<30 | size<<16 | typ<<8 | nr
}

var (
	eviocgrab = ioc(iocWrite, 'E', 0x90, 4)
	eviocgkey = ioc(iocRead, 'E', 0x18, keyBytes)

	uiDevCreate  = ioc(0, 'U', 1, 0)
	uiDevSetup   = ioc(iocWrite, 'U', 3, unsafe.Sizeof(uinputSetup{}))
	uiSetEvBit   = ioc(iocWrite, 'U', 100, 4)
	uiSetKeyBit  = ioc(iocWrite, 'U', 101, 4)
	uiSetRelBit  = ioc(iocWrite, 'U', 102, 4)
	uiSetLedBit  = ioc(iocWrite, 'U', 105, 4)
	uiGetSysname = ioc(iocRead, 'U', 44, sysnameSize)
)

const (
	keyBytes    = keyMax/8 + 1 // size of a key state bitmap
	sysnameSize = 64
)

// ioctlInt runs an ioctl whose argument is an integer.
func ioctlInt(f *os.File, req uintptr, arg int) error {
	return control(f, func(fd uintptr) syscall.Errno {
		_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, req, uintptr(arg))
		return errno
	})
}

// ioctlPtr runs an ioctl whose argument is a pointer.
func ioctlPtr(f *os.File, req uintptr, arg unsafe.Pointer) error {
	return control(f, func(fd uintptr) syscall.Errno {
		_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, req, uintptr(arg))
		return errno
	})
}

// control runs fn with the descriptor of f. Unlike File.Fd, it keeps f non-blocking,
// so that closing f still ends a pending Read.
func control(f *os.File, fn func(fd uintptr) syscall.Errno) error {
	conn, err := f.SyscallConn()
	if err != nil {
		return err
	}
	var errno syscall.Errno
	if err := conn.Control(func(fd uintptr) { errno = fn(fd) }); err != nil {
		return err
	}
	if errno != 0 {
		return errno
	}
	return nil
}

// keyState reads the key state bitmap of an evdev device.
func keyState(f *os.File) (bits [keyBytes]byte, err error) {
	err = ioctlPtr(f, eviocgkey, unsafe.Pointer(&bits))
	return bits, err
}

// inputDevice is an entry of /proc/bus/input/devices.
type inputDevice struct {
	path     string // /dev/input/eventN
	sysfs    string // such as /devices/virtual/input/input23
	keyboard bool   // has the kbd handler
	mouse    bool   // has a mouse handler
	letters  bool   // reports KEY_A, so it is a typing keyboard rather than, say, a power button
	abs      bool   // reports absolute axes (touchpads, tablets, touch screens)
}

// virtual reports whether the device was created by software, for example with uinput.
func (d inputDevice) virtual() bool {
	return strings.HasPrefix(d.sysfs, "/devices/virtual/")
}

// grabbable reports whether the device is a physical keyboard that can be grabbed and forwarded
// through the virtual device without losing anything.
func (d inputDevice) grabbable() bool {
	return d.keyboard && d.letters && !d.abs && !d.virtual()
}

// listDevices reads the input devices from /proc/bus/input/devices.
func listDevices() ([]inputDevice, error) {
	f, err := os.Open("/proc/bus/input/devices")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return parseDevices(f)
}

func parseDevices(r io.Reader) ([]inputDevice, error) {
	var devices []inputDevice
	var d inputDevice
	flush := func() {
		if d.path != "" {
			devices = append(devices, d)
		}
		d = inputDevice{}
	}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			flush()
		} else if v, ok := strings.CutPrefix(line, "S: Sysfs="); ok {
			d.sysfs = v
		} else if v, ok := strings.CutPrefix(line, "H: Handlers="); ok {
			for _, h := range strings.Fields(v) {
				switch {
				case h == "kbd":
					d.keyboard = true
				case strings.HasPrefix(h, "mouse"):
					d.mouse = true
				case strings.HasPrefix(h, "event"):
					d.path = filepath.Join("/dev/input", h)
				}
			}
		} else if v, ok := strings.CutPrefix(line, "B: KEY="); ok {
			d.letters = hasBit(v, keyA)
		} else if v, ok := strings.CutPrefix(line, "B: ABS="); ok {
			d.abs = strings.Trim(v, "0 ") != ""
		}
	}
	flush()
	return devices, scanner.Err()
}

// hasBit reports whether bit is set in a capability bitmap of /proc/bus/input/devices:
// hexadecimal words of the kernel's long size, most significant first.
func hasBit(bitmap string, bit int) bool {
	words := strings.Fields(bitmap)
	i := len(words) - 1 - bit/strconv.IntSize
	if i < 0 {
		return false
	}
	w, err := strconv.ParseUint(words[i], 16, strconv.IntSize)
	return err == nil && w&(1<<(bit%strconv.IntSize)) != 0
}
