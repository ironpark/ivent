package hook

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"unsafe"

	"github.com/ironpark/ivent/key"
)

// Linux input event types and codes (linux/input-event-codes.h).
const (
	evKey = 0x01
	evRel = 0x02

	relWheel = 0x08

	btnLeft  = 0x110
	btnExtra = 0x114 // BTN_LEFT, RIGHT, MIDDLE, SIDE, EXTRA are consecutive
)

// inputEventSize is sizeof(struct input_event): a timeval followed by type, code and value.
var inputEventSize = int(unsafe.Sizeof(syscall.Timeval{})) + 8

var (
	devicesMu sync.Mutex
	devices   []*os.File
)

// platformStart reads events from the evdev devices in /dev/input.
// It needs read access to them, usually by being root or in the "input" group.
func platformStart(cfg Config) error {
	files, err := openDevices(cfg.Keyboard, cfg.Mouse)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrHookFailed, err)
	}
	devicesMu.Lock()
	devices = files
	devicesMu.Unlock()

	var wg sync.WaitGroup
	for _, f := range files {
		wg.Add(1)
		go func(f *os.File) {
			defer wg.Done()
			readDevice(f)
		}(f)
	}
	ready(false)
	wg.Wait()
	return nil
}

// platformStop closes the devices, which ends the blocked reads.
func platformStop() {
	devicesMu.Lock()
	defer devicesMu.Unlock()
	for _, f := range devices {
		f.Close()
	}
	devices = nil
}

func readDevice(f *os.File) {
	buf := make([]byte, inputEventSize)
	tv := inputEventSize - 8
	for {
		if _, err := io.ReadFull(f, buf); err != nil {
			return
		}
		typ := binary.NativeEndian.Uint16(buf[tv:])
		code := binary.NativeEndian.Uint16(buf[tv+2:])
		value := int32(binary.NativeEndian.Uint32(buf[tv+4:]))
		switch {
		case typ == evKey && code >= btnLeft && code <= btnExtra:
			dispatch(Event{Kind: keyKind(value), Code: key.MouseButtons[code-btnLeft]})
		case typ == evKey && code < 256:
			// value is 0 for release, 1 for press and 2 for auto-repeat.
			dispatch(Event{Kind: keyKind(value), Code: key.Code(code)})
		case typ == evRel && code == relWheel:
			dispatch(Event{Kind: MouseWheel, Delta: value})
		}
	}
}

func keyKind(value int32) Kind {
	if value == 0 {
		return KeyUp
	}
	return KeyDown
}

// openDevices opens the event devices of keyboards and mice.
func openDevices(keyboard, mouse bool) ([]*os.File, error) {
	paths, err := inputDevices(keyboard, mouse)
	if err != nil {
		return nil, err
	}
	var files []*os.File
	for _, path := range paths {
		if f, err := os.Open(path); err == nil {
			files = append(files, f)
		}
	}
	if len(files) == 0 {
		return nil, errors.New("cannot open /dev/input devices (run as root or join the input group)")
	}
	return files, nil
}

// inputDevices lists the event devices of keyboards and mice from /proc/bus/input/devices.
func inputDevices(keyboard, mouse bool) ([]string, error) {
	f, err := os.Open("/proc/bus/input/devices")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return parseDevices(f, keyboard, mouse)
}

func parseDevices(r io.Reader, keyboard, mouse bool) ([]string, error) {
	var paths []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		handlers, ok := strings.CutPrefix(scanner.Text(), "H: Handlers=")
		if !ok {
			continue
		}
		fields := strings.Fields(handlers)
		wanted := false
		var event string
		for _, h := range fields {
			switch {
			case h == "kbd" && keyboard, strings.HasPrefix(h, "mouse") && mouse:
				wanted = true
			case strings.HasPrefix(h, "event"):
				event = h
			}
		}
		if wanted && event != "" {
			paths = append(paths, filepath.Join("/dev/input", event))
		}
	}
	if len(paths) == 0 {
		return nil, errors.New("no keyboard or mouse found in /proc/bus/input/devices")
	}
	return paths, scanner.Err()
}

// charForKey is not supported on Linux: evdev reports physical keys only.
func charForKey(code key.Code) (rune, bool) {
	return 0, false
}

// KeyState reports whether the OS considers the key to be held down.
// Linux has no global query, so ok is always false.
func KeyState(code key.Code) (pressed, ok bool) {
	return false, false
}

// Send is not supported on Linux.
func Send(code key.Code, down bool) error {
	return fmt.Errorf("%w: sending keys", ErrUnsupported)
}

// ProcessInfo is not supported on Linux, where events carry no process.
func ProcessInfo(pid int) (resolvedPID int, name, id string, ok bool) {
	return pid, "", "", false
}

// Permissions reports whether the process may listen to input (monitor) and block or send input (control).
// On Linux, monitor means at least one input device can be opened. Blocking and sending are not supported.
func Permissions(prompt bool) (monitor, control bool) {
	files, err := openDevices(true, true)
	for _, f := range files {
		f.Close()
	}
	return err == nil, false
}
