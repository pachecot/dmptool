package terminal

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows"
)

func (t *Terminal) getSize() *Size {
	bi := ConsoleScreenBufferInfo{}
	err := GetConsoleScreenBufferInfo(windows.Handle(t.out), &bi)
	if err != nil {
		fmt.Println("getSize() failed:", err)
		return nil
	}
	return &Size{
		Width:  int(bi.Size.X),
		Height: int(bi.Size.Y),
	}
}

func (t *Terminal) enableInput() error {
	_, err := SetConsoleModeFlag(windows.Handle(t.in), ENABLE_WINDOW_INPUT)
	return err
}

func (t *Terminal) open() Listener {

	inRec := make([]InputRecord, 8)
	var count uint32

	resizeEvents := make(chan Size)
	keyEvents := make(chan Key)
	mouseEvents := make(chan MouseEventRecord)
	done := make(chan struct{})

	go func() {
		for {

			event, _ := windows.WaitForSingleObject(windows.Handle(t.in), 1_000)
			switch event {
			case windows.WAIT_FAILED:
				return
			case windows.WAIT_ABANDONED, uint32(windows.WAIT_TIMEOUT):
				select {
				case <-done:
					return
				default:
					continue
				}
			}

			err := ReadConsoleInput(windows.Handle(t.in), inRec, &count)
			if err != nil {
				os.Stderr.WriteString(err.Error())
				return
			}

			for i := range count {
				switch inRec[i].EventType {

				case KeyEventType:
					evt := inRec[i].KeyEvent()
					if evt.KeyDown == 0 {
						continue
					}
					for range evt.RepeatCount {
						keyEvents <- Key{
							Rune:    rune(evt.Char),
							Key:     evt.VirtualKeyCode,
							Control: evt.ControlKeyState,
						}
					}

				case MouseEventType:
					evt := inRec[i].MouseEvent()
					if evt == nil {
						continue
					}
					mouseEvents <- *evt

				case ResizeEventType:
					evt := inRec[i].ResizeEvent()
					if evt == nil {
						continue
					}
					resizeEvents <- Size{
						Width:  int(evt.Width),
						Height: int(evt.Height),
					}
				}

				select {
				case <-done:
					return
				default:
				}
			}
		}
	}()
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
		t.isOpen = false
		t.closer = nil
	}

	return Listener{
		ResizeEvents: resizeEvents,
		KeyEvents:    keyEvents,
		MouseEvent:   mouseEvents,
	}
}
