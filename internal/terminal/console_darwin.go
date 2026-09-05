package terminal

// byte and escape-sequence decoding for macOS/unix ttys.
//
// There is no OS-level "console input record" API on macOS, so key events
// are reconstructed from the raw byte stream a terminal emulator sends for
// each keypress (ANSI/VT100 CSI and SS3 escape sequences). The VK_* codes
// from keys.go (originally Windows virtual-key codes) are reused here as a
// platform-neutral key identifier so callers don't need to know which OS
// they're on.

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Coord and MouseEventRecord mirror the shape of the windows console API
// types of the same name so that Listener.MouseEvent (declared in the
// shared terminal.go) has a concrete type on darwin too. Mouse reporting
// itself is not implemented here, so nothing ever populates or sends a
// MouseEventRecord.
type Coord struct {
	X int16
	Y int16
}

type MouseEventRecord struct {
	MousePosition   Coord
	ButtonState     ButtonState
	ControlKeyState KeyState
	EventFlags      MouseFlags
}

// escSeqTimeout bounds how long we wait for the rest of an escape sequence
// after seeing a lone 0x1B byte, so a bare Esc keypress can still be
// reported promptly. A slow link (e.g. SSH) could in principle split a
// sequence across this deadline and have it misreported as bare Esc.
const escSeqTimeout = 50 * time.Millisecond

// decodeByte maps a single non-ESC input byte to a Key.
func decodeByte(b byte) Key {
	switch b {
	case 0x08, 0x7F:
		return Key{Rune: rune(b), Key: VK_BACK}
	case 0x09:
		return Key{Rune: rune(b), Key: VK_TAB}
	case 0x0A, 0x0D:
		return Key{Rune: rune(b), Key: VK_RETURN}
	case 0x20:
		return Key{Rune: rune(b), Key: VK_SPACE}
	}

	if b >= 0x01 && b <= 0x1A {
		// Ctrl+letter: Ctrl+A is 0x01 .. Ctrl+Z is 0x1A.
		return Key{Rune: rune(b), Key: uint16(VK_A) + uint16(b-1), Control: LEFT_CTRL_PRESSED}
	}

	switch {
	case b >= '0' && b <= '9':
		return Key{Rune: rune(b), Key: uint16(VK_0) + uint16(b-'0')}
	case b >= 'a' && b <= 'z':
		return Key{Rune: rune(b), Key: uint16(VK_A) + uint16(b-'a')}
	case b >= 'A' && b <= 'Z':
		return Key{Rune: rune(b), Key: uint16(VK_A) + uint16(b-'A')}
	}

	return Key{Rune: rune(b)}
}

// readEscapeSequence consumes bytes following a lone 0x1B, returning the
// decoded Key and true if a full sequence was recognized, or false if
// nothing followed before escSeqTimeout (a bare Esc keypress).
func readEscapeSequence(f *os.File) (Key, bool) {
	if err := f.SetReadDeadline(time.Now().Add(escSeqTimeout)); err != nil {
		return Key{}, false
	}

	buf := make([]byte, 1)
	n, err := f.Read(buf)
	if n == 0 || err != nil {
		return Key{}, false
	}

	switch buf[0] {
	case '[':
		return readCSI(f)
	case 'O':
		n, err := f.Read(buf)
		if n == 0 || err != nil {
			return Key{}, false
		}
		return decodeSS3(buf[0]), true
	default:
		// Alt+key convention: Esc followed directly by a printable byte.
		k := decodeByte(buf[0])
		k.Control |= LEFT_ALT_PRESSED
		return k, true
	}
}

// readCSI reads the remainder of a CSI sequence ("\x1b[...") one byte at a
// time until a final byte (a letter, or '~') is seen, then decodes it.
func readCSI(f *os.File) (Key, bool) {
	var seq []byte
	buf := make([]byte, 1)
	for len(seq) < 16 {
		n, err := f.Read(buf)
		if n == 0 || err != nil {
			return Key{}, false
		}
		seq = append(seq, buf[0])
		b := buf[0]
		if (b >= 'A' && b <= 'Z') || b == '~' {
			return decodeCSI(seq), true
		}
	}
	return Key{}, false
}

// decodeCSI decodes a CSI sequence body (everything after "\x1b[",
// including the final letter or '~') into a Key.
func decodeCSI(seq []byte) Key {
	if len(seq) == 0 {
		return Key{}
	}
	final := seq[len(seq)-1]
	params := string(seq[:len(seq)-1])

	// Split "<n1>;<n2>" style parameters.
	var p []int
	for _, s := range strings.Split(params, ";") {
		v, err := strconv.Atoi(s)
		if err != nil {
			v = 0
		}
		p = append(p, v)
	}

	var mod KeyState
	if len(p) >= 2 {
		mod = xtermModifier(p[1])
	}

	if final == '~' {
		if len(p) == 0 {
			return Key{}
		}
		switch p[0] {
		case 1, 7:
			return Key{Key: VK_HOME, Control: mod}
		case 2:
			return Key{Key: VK_INSERT, Control: mod}
		case 3:
			return Key{Key: VK_DELETE, Control: mod}
		case 4, 8:
			return Key{Key: VK_END, Control: mod}
		case 5:
			return Key{Key: VK_PGUP, Control: mod}
		case 6:
			return Key{Key: VK_PGDN, Control: mod}
		}
		return Key{}
	}

	switch final {
	case 'A':
		return Key{Key: VK_UP, Control: mod}
	case 'B':
		return Key{Key: VK_DOWN, Control: mod}
	case 'C':
		return Key{Key: VK_RIGHT, Control: mod}
	case 'D':
		return Key{Key: VK_LEFT, Control: mod}
	case 'H':
		return Key{Key: VK_HOME, Control: mod}
	case 'F':
		return Key{Key: VK_END, Control: mod}
	}
	return Key{}
}

// decodeSS3 decodes an SS3 sequence ("\x1bO<c>"), used by some terminals
// (e.g. iTerm2 in application-cursor-keys mode) for Home/End/F1-F4.
func decodeSS3(c byte) Key {
	switch c {
	case 'H':
		return Key{Key: VK_HOME}
	case 'F':
		return Key{Key: VK_END}
	case 'P':
		return Key{Key: VK_F1}
	case 'Q':
		return Key{Key: VK_F2}
	case 'R':
		return Key{Key: VK_F3}
	case 'S':
		return Key{Key: VK_F4}
	}
	return Key{}
}

// xtermModifier decodes an xterm modifier parameter (as used in sequences
// like "\x1b[1;5H") into KeyState bits. The wire value is bias-1 encoded:
// bit0=Shift, bit1=Alt, bit2=Ctrl.
func xtermModifier(param int) KeyState {
	if param <= 0 {
		return 0
	}
	m := param - 1
	var s KeyState
	if m&0x1 != 0 {
		s |= SHIFT_PRESSED
	}
	if m&0x2 != 0 {
		s |= LEFT_ALT_PRESSED
	}
	if m&0x4 != 0 {
		s |= LEFT_CTRL_PRESSED
	}
	return s
}
