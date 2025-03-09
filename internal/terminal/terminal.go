package terminal

import (
	"fmt"
	"os"
)

type Terminal struct {
	in     uintptr
	out    uintptr
	isOpen bool
	closer func()
}

type Listener struct {
	ResizeEvents <-chan Size
	KeyEvents    <-chan Key
	MouseEvent   <-chan MouseEventRecord
}

type Size struct {
	Width  int
	Height int
}

type Key struct {
	Rune    rune
	Key     uint16
	Control KeyState
}

func New() *Terminal {
	t := &Terminal{
		in:  os.Stdin.Fd(),
		out: os.Stdout.Fd(),
	}
	err := t.enableInput()
	if err != nil {
		fmt.Println("enableInput failed:", err)
		return nil
	}
	return t
}

func (t *Terminal) Close() {
	if t.closer == nil {
		return
	}
	t.closer()
}

func (t *Terminal) GetSize() *Size {
	return t.getSize()
}

func (t *Terminal) Open() Listener {
	return t.open()
}
