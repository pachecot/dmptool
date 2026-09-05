package terminal

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

func (t *Terminal) getSize() *Size {
	width, height, err := term.GetSize(int(t.out))
	if err != nil {
		fmt.Println("getSize() failed:", err)
		return nil
	}
	return &Size{
		Width:  width,
		Height: height,
	}
}

// enableInput switches the tty into raw mode (no echo, no line buffering,
// byte-at-a-time delivery) and arranges for it to be restored on Close,
// even if Open is never called.
func (t *Terminal) enableInput() error {
	state, err := term.MakeRaw(int(t.in))
	if err != nil {
		return err
	}
	t.rawState = state
	t.closer = t.restoreRaw
	return nil
}

func (t *Terminal) restoreRaw() {
	state, ok := t.rawState.(*term.State)
	if !ok || state == nil {
		return
	}
	term.Restore(int(t.in), state)
	t.rawState = nil
	t.closer = nil
}

func (t *Terminal) open() Listener {

	resizeEvents := make(chan Size)
	keyEvents := make(chan Key)
	mouseEvents := make(chan MouseEventRecord)
	done := make(chan struct{})

	// os.NewFile only registers an already-open fd with the runtime poller
	// (needed for SetReadDeadline, used below to make the read loop
	// cancellable) if the fd already has O_NONBLOCK set; set it explicitly
	// since t.in is inherited as a normal blocking tty fd.
	unix.SetNonblock(int(t.in), true)
	in := os.NewFile(t.in, "stdin")

	go keyReadLoop(in, keyEvents, done)
	go resizeLoop(t, resizeEvents, done)

	t.isOpen = true

	t.closer = func() {
		if !t.isOpen {
			return
		}
		go func() {
			done <- struct{}{}
		}()
		close(resizeEvents)
		close(keyEvents)
		// mouseEvents intentionally left unclosed: mirrors the same
		// asymmetry present in the windows implementation, and it is
		// never sent to since mouse reporting is not implemented here.
		t.isOpen = false
		t.restoreRaw()
	}

	return Listener{
		ResizeEvents: resizeEvents,
		KeyEvents:    keyEvents,
		MouseEvent:   mouseEvents,
	}
}

// keyReadLoop reads raw bytes from f, decodes them into Key events, and
// sends them on keyEvents until done is signaled or the read fails.
// It polls done roughly once a second via a read deadline, matching the
// cadence of the windows implementation's WaitForSingleObject loop.
func keyReadLoop(f *os.File, keyEvents chan<- Key, done <-chan struct{}) {
	buf := make([]byte, 1)
	for {
		select {
		case <-done:
			return
		default:
		}

		if err := f.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
			return
		}

		n, err := f.Read(buf)
		if err != nil {
			if os.IsTimeout(err) {
				continue
			}
			return
		}
		if n == 0 {
			continue
		}

		var key Key
		if buf[0] == 0x1B {
			k, ok := readEscapeSequence(f)
			if !ok {
				key = Key{Key: VK_ESCAPE}
			} else {
				key = k
			}
		} else {
			key = decodeByte(buf[0])
		}

		select {
		case keyEvents <- key:
		case <-done:
			return
		}
	}
}

// resizeLoop watches for SIGWINCH and reports the terminal's new size on
// resizeEvents until done is signaled.
func resizeLoop(t *Terminal, resizeEvents chan<- Size, done <-chan struct{}) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGWINCH)
	defer signal.Stop(sigChan)

	for {
		select {
		case <-sigChan:
			sz := t.getSize()
			if sz == nil {
				continue
			}
			select {
			case resizeEvents <- *sz:
			case <-done:
				return
			}
		case <-done:
			return
		}
	}
}
