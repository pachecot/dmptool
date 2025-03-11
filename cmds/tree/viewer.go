package tree

import (
	"fmt"
	"os"
	"strings"

	"github.com/tpacheco/dmptool/internal/terminal"
)

type command int

const (
	cmd_Noop command = iota // No Operation
	cmd_Quit                // Quit
	cmd_PgDn                // Page Down
	cmd_PgUp                // Page Up
	cmd_Dn                  // Down 1
	cmd_Up                  // Up 1
	cmd_Home                // Home
	cmd_End                 // End
)

type view struct {
	out    *os.File
	in     *os.File
	pos    int
	window struct {
		height int
		width  int
	}
	height int
	data   []*node
	name   string
}

func newView(name string, data []*node) *view {
	return &view{
		in:   os.Stdin,
		out:  os.Stdout,
		data: data,
		name: name,
	}
}

func (v *view) run() {

	t := terminal.New()
	defer func() {
		t.Close()
	}()
	sz := t.GetSize()
	if sz == nil {
		return
	}
	v.resize(*sz)
	listener := t.Open()

	for {
		select {
		case evt, ok := <-listener.KeyEvents:
			if !ok {
				return
			}
			switch getCmd(evt) {
			case cmd_Quit:
				return
			case cmd_Dn:
				if v.pos < len(v.data)-(v.height-v.data[v.pos].depth) {
					v.pos++
					v.redraw()
				}
			case cmd_PgDn:
				// rotate up until the last item is the first
				m := v.pos + (v.height - v.data[v.pos].depth)
				for v.pos < len(v.data)-(v.height-v.data[v.pos].depth) && v.pos < m {
					v.pos++
				}
				v.redraw()
			case cmd_PgUp:
				// rotate down until the first item is the last
				m := v.pos
				for v.pos > 0 && v.pos+(v.height-v.data[v.pos].depth) > m {
					v.pos--
				}
				v.redraw()
			case cmd_Up:
				if v.pos > 0 {
					v.pos--
					v.redraw()
				}
			case cmd_Home:
				if v.pos > 0 {
					v.pos = 0
					v.redraw()
				}
			case cmd_End:
				n := len(v.data)
				last := n - v.height
				for last < n-(v.height-v.data[last].depth) {
					last++
				}
				if v.pos != last {
					v.pos = last
					v.redraw()
				}
			}
		case sz, ok := <-listener.ResizeEvents:
			if !ok {
				return
			}
			if sz.Height != v.window.height || sz.Width != v.window.width {
				v.resize(sz)
			}
		}
	}
}

func (v *view) resize(sz terminal.Size) {
	v.window.width = sz.Width
	v.window.height = sz.Height
	v.height = sz.Height - 1
	v.out.WriteString(eraseDisplay)
	v.redraw()
}

func (v *view) redraw() {
	row := 1
	n := len(v.data)
	if v.height > len(v.data) {
		v.pos = 0
	} else {
		row = v.data[v.pos].depth
		n = v.height - row + 1
		if v.pos+n > len(v.data) {
			v.pos = len(v.data) - n
		}
	}
	v.header()
	v.out.WriteString(fmt.Sprintf(fCUP, row))
	for i := range n {
		v.out.WriteString(eraseLine)
		v.out.WriteString(v.data[v.pos+i].view())
		v.out.WriteString(LF)
	}
	v.menu()
}

func (v *view) menu() {
	v.out.WriteString(fmt.Sprintf(fCUP, v.window.height))
	v.out.WriteString(CR)
	v.out.WriteString(sgrReverse)
	v.out.WriteString(strings.Repeat(" ", v.window.width))
	v.out.WriteString(CR)
	v.out.WriteString(" Enter ESC or q to Exit || Navigation: Space/PgDn, PgUp, Up-Arrow, Down-Arrow ")
	n := v.pos + v.height
	if n > len(v.data) {
		n = len(v.data)
	}
	pct := fmt.Sprintf("%.1f%% ", (100.0*float32(n))/float32(len(v.data)))
	v.out.WriteString(fmt.Sprintf(fCUPos, v.window.width, v.window.width-len(pct)))
	v.out.WriteString(pct)
	v.out.WriteString(sgrReset)
}

func parents(n *node) []*node {
	if n == nil {
		return nil
	}
	return append(parents(n.parent), n)
}

func (v *view) header() {
	if v.pos == 0 {
		return
	}
	first := v.data[v.pos]
	v.out.WriteString(fmt.Sprintf(fCUP, 1))
	for _, p := range parents(first.parent) {
		v.out.WriteString(eraseLine)
		v.out.WriteString(fmt.Sprintf(sgrColor, 100))
		v.out.WriteString(strings.Repeat(" ", v.window.width))
		v.out.WriteString(CR)
		v.out.WriteString(p.view())
		v.out.WriteString(sgrReset)
		v.out.WriteString(LF)
	}
}

func getCmd(evt terminal.Key) command {
	if evt.Control.CtrlKey() {
		switch evt.Key {
		case terminal.VK_HOME:
			return cmd_Home
		case terminal.VK_END:
			return cmd_End
		}
		return 0
	}

	switch evt.Key {
	case terminal.VK_ESCAPE:
		return cmd_Quit
	case terminal.VK_UP:
		return cmd_Up
	case terminal.VK_DOWN:
		return cmd_Dn
	case terminal.VK_SPACE, terminal.VK_PGDN:
		return cmd_PgDn
	case terminal.VK_PGUP:
		return cmd_PgUp
	case terminal.VK_Q:
		return cmd_Quit
	}
	return 0
}

const (
	BEL = "\x07" //	Bell	Makes an audible noise.
	BS  = "\x08" //	Backspace	Moves the cursor left (but may "backwards wrap" if cursor is at start of line).
	HT  = "\x09" //	Tab	Moves the cursor right to next tab stop.
	LF  = "\x0A" //	Line Feed	Moves to next line, scrolls the display up if at bottom of the screen. Usually does not move horizontally, though programs should not rely on this.
	FF  = "\x0C" //	Form Feed	Move a printer to top of next page. Usually does not move horizontally, though programs should not rely on this. Effect on video terminals varies.
	CR  = "\x0D" //	Carriage Return	Moves the cursor to column zero.
	ESC = "\x1B" // Escape	Starts all the escape sequences

	CSI          = ESC + "["
	eraseLine    = CSI + "2K"
	eraseDisplay = CSI + "2J"
	scrollUP     = CSI + "S"
	scrollDN     = CSI + "T"
	fScrollUP    = CSI + "%dS"
	fScrollDN    = CSI + "%dT"
	fCUP         = CSI + "%dH"
	fCUPos       = CSI + "%d;%dH"
	sgrReverse   = CSI + "7m"
	sgrReset     = CSI + "0m"
	sgrColor     = CSI + "%dm"
	sgrFaint     = CSI + "2m"
)
