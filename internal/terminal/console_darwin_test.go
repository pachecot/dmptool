package terminal

import "testing"

func TestDecodeByte(t *testing.T) {
	cases := []struct {
		b    byte
		key  uint16
		ctrl KeyState
	}{
		{' ', VK_SPACE, 0},
		{'q', VK_Q, 0},
		{'Q', VK_Q, 0},
		{0x1A, uint16(VK_Z), LEFT_CTRL_PRESSED}, // Ctrl+Z
	}
	for _, c := range cases {
		k := decodeByte(c.b)
		if k.Key != c.key || k.Control != c.ctrl {
			t.Errorf("decodeByte(%#x) = %+v, want Key=%#x Control=%#x", c.b, k, c.key, c.ctrl)
		}
	}
}

func TestDecodeCSI(t *testing.T) {
	cases := []struct {
		seq  string
		key  uint16
		ctrl KeyState
	}{
		{"A", VK_UP, 0},
		{"B", VK_DOWN, 0},
		{"5~", VK_PGUP, 0},
		{"6~", VK_PGDN, 0},
		{"H", VK_HOME, 0},
		{"F", VK_END, 0},
		{"1;5H", VK_HOME, LEFT_CTRL_PRESSED}, // Ctrl+Home
		{"1;5F", VK_END, LEFT_CTRL_PRESSED},  // Ctrl+End
		{"7~", VK_HOME, 0},
		{"8~", VK_END, 0},
	}
	for _, c := range cases {
		k := decodeCSI([]byte(c.seq))
		if k.Key != c.key || k.Control != c.ctrl {
			t.Errorf("decodeCSI(%q) = %+v, want Key=%#x Control=%#x", c.seq, k, c.key, c.ctrl)
		}
	}
}

func TestXtermModifier(t *testing.T) {
	if got := xtermModifier(5); got&LEFT_CTRL_PRESSED == 0 {
		t.Errorf("xtermModifier(5) = %#x, want LEFT_CTRL_PRESSED set", got)
	}
	if got := xtermModifier(0); got != 0 {
		t.Errorf("xtermModifier(0) = %#x, want 0", got)
	}
}
