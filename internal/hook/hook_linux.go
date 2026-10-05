package hook

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/ironpark/ivent/key"
)

// linuxHook reads events from the evdev devices in /dev/input, which needs read access to them
// (root or the "input" group). With Config.Exclusive, it grabs keyboards so that only this process
// receives their events, and forwards the events it does not block through a uinput device.
type linuxHook struct {
	cfg    Config
	out    *uinput // forwarding device, nil unless keyboards are grabbed
	notify *os.File
	stop   chan struct{}
	wg     sync.WaitGroup

	mu      sync.Mutex
	devices map[string]*device // by path, nil once stopped
}

type device struct {
	inputDevice
	f       *os.File
	out     io.Writer   // receives the events of the device once grabbed, nil if it is not to be grabbed
	grabbed atomic.Bool // events are forwarded to out
}

var current atomic.Pointer[linuxHook]

func platformStart(cfg Config) error {
	l := &linuxHook{cfg: cfg, stop: make(chan struct{}), devices: map[string]*device{}}
	if cfg.Exclusive {
		// Without the forwarding device nothing is grabbed, and Suppressing reports false.
		if out, err := newUinput(); err == nil {
			l.out = out
			settle()
			go l.forwardLEDs()
		}
	}
	current.Store(l)
	if l.scan() == 0 {
		l.shutdown()
		return fmt.Errorf("%w: cannot open /dev/input devices (run as root or join the input group)", ErrHookFailed)
	}
	// Without inotify, devices plugged in later are not seen.
	if notify, err := watchDir("/dev/input"); err == nil {
		l.notify = notify
		go l.watch()
	}
	ready(l.out != nil)
	<-l.stop
	l.shutdown()
	return nil
}

func platformStop() {
	if l := current.Load(); l != nil {
		close(l.stop)
	}
}

// shutdown closes the devices, which ends their reads and releases the grabs. When the forwarding
// device is destroyed, the kernel and display servers release the keys still held through it.
func (l *linuxHook) shutdown() {
	current.Store(nil)
	l.mu.Lock()
	devices := l.devices
	l.devices = nil
	l.mu.Unlock()
	if l.notify != nil {
		l.notify.Close()
	}
	for _, d := range devices {
		d.f.Close()
	}
	l.wg.Wait()
	if l.out != nil {
		l.out.Close()
	}
}

// scan opens the keyboards and mice that are not open yet and returns how many devices are open.
func (l *linuxHook) scan() int {
	infos, _ := listDevices()
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.devices == nil {
		return 0
	}
	for _, info := range infos {
		wanted := l.cfg.Keyboard && info.keyboard || l.cfg.Mouse && info.mouse
		if _, open := l.devices[info.path]; open || !wanted || l.own(info) {
			continue
		}
		l.open(info)
	}
	return len(l.devices)
}

// own reports whether d is a virtual device of this process, whose events are not read.
func (l *linuxHook) own(d inputDevice) bool {
	if l.out != nil && d.sysfs == l.out.sysfs {
		return true
	}
	s := senderSysfs.Load()
	return s != nil && d.sysfs == *s
}

// open opens a device and starts reading it. It is called with l.mu held.
func (l *linuxHook) open(info inputDevice) {
	d := &device{inputDevice: info}
	flag := os.O_RDONLY
	if l.out != nil && info.grabbable() {
		d.out = l.out
		flag = os.O_RDWR // lets forwardLEDs update the keyboard's LEDs
	}
	f, err := os.OpenFile(info.path, flag, 0)
	if err != nil && flag != os.O_RDONLY {
		f, err = os.Open(info.path)
	}
	if err != nil {
		return
	}
	d.f = f
	l.devices[info.path] = d
	l.wg.Add(1)
	go func() {
		defer l.wg.Done()
		readDevice(d)
		f.Close()
		l.mu.Lock()
		if l.devices[info.path] == d {
			delete(l.devices, info.path)
		}
		l.mu.Unlock()
	}()
}

// watch rescans the devices when device files are created or their permissions change.
func (l *linuxHook) watch() {
	buf := make([]byte, 4096)
	for {
		if _, err := l.notify.Read(buf); err != nil {
			return
		}
		l.scan()
	}
}

func watchDir(dir string) (*os.File, error) {
	fd, err := syscall.InotifyInit1(syscall.IN_CLOEXEC | syscall.IN_NONBLOCK)
	if err != nil {
		return nil, err
	}
	if _, err := syscall.InotifyAddWatch(fd, dir, syscall.IN_CREATE|syscall.IN_ATTRIB); err != nil {
		syscall.Close(fd)
		return nil, err
	}
	// A non-blocking descriptor uses the runtime poller, so Close ends a pending Read.
	return os.NewFile(uintptr(fd), "inotify"), nil
}

// forwardLEDs passes LED changes, such as Caps Lock, from the display server to the grabbed keyboards,
// which only listen to the process that grabbed them. It ends when the forwarding device is closed.
func (l *linuxHook) forwardLEDs() {
	buf := make([]byte, inputEventSize)
	for {
		if _, err := io.ReadFull(l.out, buf); err != nil {
			return
		}
		if typ, _, _ := decodeEvent(buf); typ != evLed {
			continue
		}
		frame := appendSyn(bytes.Clone(buf))
		l.mu.Lock()
		for _, d := range l.devices {
			if d.grabbed.Load() {
				d.f.Write(frame)
			}
		}
		l.mu.Unlock()
	}
}

// readDevice delivers the events of d until it fails. Once d is grabbed, the events that
// are not blocked are forwarded to d.out, one SYN_REPORT frame at a time.
func readDevice(d *device) {
	var flags uint8
	if d.virtual() {
		flags = FlagInjected
	}
	// A key held while grabbing would never be released for the display server, so the grab waits.
	tryGrab := func() {
		if d.out != nil && !d.grabbed.Load() && !anyKeyHeld(d.f) && ioctlInt(d.f, eviocgrab, 1) == nil {
			d.grabbed.Store(true)
		}
	}
	tryGrab()
	buf := make([]byte, 64*inputEventSize) // evdev returns whole events, as many as fit
	var frame []byte
	dropping, released := false, false
	for {
		n, err := d.f.Read(buf)
		if err != nil {
			return
		}
		for p := buf[:n-n%inputEventSize]; len(p) > 0; p = p[inputEventSize:] {
			ev := p[:inputEventSize]
			typ, code, value := decodeEvent(ev)
			switch {
			case typ == evSyn && code == synDropped:
				// The kernel buffer overflowed: events up to the next SYN_REPORT are incomplete.
				dropping = true
				frame = frame[:0]
				reset()
			case typ == evSyn && code == synReport:
				if !dropping && len(frame) > 0 && d.grabbed.Load() {
					d.out.Write(append(frame, ev...))
				}
				if released {
					tryGrab()
				}
				dropping, released = false, false
				frame = frame[:0]
			case dropping:
			default:
				released = released || typ == evKey && value == 0
				suppress := handleInput(typ, code, value, flags)
				if !suppress && typ != evMsc && d.grabbed.Load() {
					frame = append(frame, ev...)
				}
			}
		}
	}
}

func handleInput(typ, code uint16, value int32, flags uint8) (suppress bool) {
	switch {
	case typ == evKey && code >= btnLeft && code <= btnExtra:
		return dispatch(Event{Kind: keyKind(value != 0), Code: key.MouseButtons[code-btnLeft], Flags: flags})
	case typ == evKey && code < 256:
		// value is 0 for release, 1 for press and 2 for auto-repeat.
		return dispatch(Event{Kind: keyKind(value != 0), Code: key.Code(code), Flags: flags})
	case typ == evRel && code == relWheel:
		return dispatch(Event{Kind: MouseWheel, Delta: value, Flags: flags})
	}
	return false
}

func anyKeyHeld(f *os.File) bool {
	bits, err := keyState(f)
	return err != nil || bits != [keyBytes]byte{}
}

// evdevCode converts a key code to its Linux event code.
func evdevCode(code key.Code) uint16 {
	if i, ok := mouseButton(code); ok {
		return btnLeft + uint16(i)
	}
	return uint16(code)
}

// KeyState reports whether any open device has the key held down, according to the kernel.
// ok is false if the hook is not running.
func KeyState(code key.Code) (pressed, ok bool) {
	l := current.Load()
	if l == nil {
		return false, false
	}
	c := evdevCode(code)
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, d := range l.devices {
		bits, err := keyState(d.f)
		if err != nil {
			continue
		}
		ok = true
		if bits[c/8]&(1<<(c%8)) != 0 {
			return true, true
		}
	}
	return false, ok
}

// charForKey is not supported on Linux: evdev reports physical keys only.
func charForKey(code key.Code) (rune, bool) {
	return 0, false
}

// Send sends a key through the virtual uinput device of the process.
func Send(code key.Code, down bool) error {
	v, err := sendDevice()
	if err != nil {
		return err
	}
	return v.sendKey(evdevCode(code), down)
}

// ProcessInfo is not supported on Linux, where events carry no process.
func ProcessInfo(pid int) (resolvedPID int, name, id string, ok bool) {
	return pid, "", "", false
}

// Permissions reports whether the process may listen to input (monitor) and block or send input (control).
// On Linux, monitor means at least one keyboard or mouse can be opened, and control means /dev/uinput
// can be written.
func Permissions(prompt bool) (monitor, control bool) {
	infos, _ := listDevices()
	for _, info := range infos {
		if !info.keyboard && !info.mouse {
			continue
		}
		if f, err := os.Open(info.path); err == nil {
			f.Close()
			monitor = true
			break
		}
	}
	if f, err := os.OpenFile("/dev/uinput", os.O_WRONLY, 0); err == nil {
		f.Close()
		control = true
	}
	return monitor, control
}
