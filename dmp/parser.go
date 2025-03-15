package dmp

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	prop_alarm_links  = "AlarmLinks"
	prop_array        = "Array"
	prop_bytecode     = "ByteCode"
	prop_bytecode_end = "EndByteCode"
	prop_last_change  = "LastChange"
	prop_members      = "Members"
	prop_path         = "Path"

	tag_container       = "Container"
	tag_container_begin = "BeginContainer"
	tag_container_end   = "EndContainer"
	tag_device          = "Device"
	tag_device_end      = "EndDevice"

	tag_controller       = "Controller"
	tag_controller_begin = "BeginController"
	tag_controller_end   = "EndController"
	tag_dictionary       = "Dictionary"
	tag_dictionary_end   = "EndDictionary"
	tag_infinet_ctlr     = "InfinetCtlr"
	tag_infinet_ctlr_end = "EndInfinetCtlr"
	tag_object           = "Object"
	tag_object_end       = "EndObject"

	new_line = "\n"
)

type token struct {
	value string
	line  int
}

type parser interface {
	parse(*token) parser
}

type state struct {
	alias map[string]string
}

type dmpParser struct {
	name    string
	path    string
	devPath string
	h       Handler
	s       *state
}

func newParser(h Handler) *dmpParser {
	return &dmpParser{
		h: h,
		s: &state{
			alias: map[string]string{},
		},
	}
}

type infControllerParser struct {
	prev parser
	name string
	path string
	h    Handler
	s    *state
}

type deviceParser struct {
	prev parser
	name string
	path string
	h    Handler
	s    *state
}

type controllerParser struct {
	last parser
	name string
	path string
	h    Handler
	s    *state
}

type containerParser struct {
	last parser
	name string
	path string
	h    Handler
	s    *state
}

type dictionaryParser struct {
	prev   parser
	name   string
	path   string
	tables []*Table
	h      Handler
	s      *state
}

type tableParser struct {
	prev  *dictionaryParser
	table *Table
	s     *state
}

type objectParser struct {
	prev     parser
	lastProp string
	h        Handler
	obj      *Object
}

func newObject(name string, pth string) *Object {
	return &Object{
		Name:       name,
		Path:       pth,
		Properties: map[string]string{"Name": name},
	}
}

func newObjectParser(h Handler, p parser, s *state, name string, pth string) *objectParser {
	return &objectParser{
		h:    h,
		prev: p,
		obj:  newObject(name, pth),
		s:    s,
	}
}

type codeParser struct {
	prev  *objectParser
	lines []string
	s     *state
}

type blockParser struct {
	prev       *objectParser
	name       string
	endTag     string
	includeEnd bool
	lines      []string
	s          *state
}

func (p *blockParser) parse(tk *token) parser {
	txt := strings.Trim(tk.value, " ")
	if txt == p.endTag {
		if p.includeEnd {
			p.lines = append(p.lines, tk.value)
		}
		p.prev.obj.Properties[p.name] = strings.Join(p.lines, new_line)
		return p.prev
	}
	p.lines = append(p.lines, tk.value)
	return p
}

func (p *codeParser) parse(tk *token) parser {
	txt := strings.TrimSpace(tk.value)
	if txt == prop_bytecode_end {
		p.prev.obj.Properties[prop_bytecode] = strings.Join(p.lines, new_line)
		return p.prev
	}
	p.lines = append(p.lines, tk.value)
	return p
}

func (p *dictionaryParser) parse(tk *token) parser {
	if len(tk.value) == 0 {
		return p
	}
	values := strings.Split(tk.value, ":")
	for i := range values {
		values[i] = strings.TrimSpace(values[i])
	}
	switch values[0] {
	case tag_dictionary:
		p.h.Begin(tag_dictionary, values[1])
		return &dictionaryParser{
			prev: p,
			h:    p.h,
			name: values[1],
			path: filepath.Join(p.path, p.name),
			s:    p.s,
		}

	case tag_dictionary_end:
		p.h.Dictionary(&Dictionary{
			Path:   p.path,
			Name:   p.name,
			Tables: p.tables,
		})
		p.h.End(tag_dictionary, p.name)
		return p.prev

	case "'TYPE":
		values[0] = values[0][1:]
		return &tableParser{
			prev: p,
			table: &Table{
				Header: values,
			},
			s: p.s,
		}
	}
	return p
}

func (p *tableParser) parse(tk *token) parser {

	if len(tk.value) == 0 {
		p.prev.tables = append(p.prev.tables, p.table)
		return p.prev
	}

	cells := strings.Split(tk.value, ":")

	if len(cells) != len(p.table.Header) {
		// table changed so defer to parent parser
		p.prev.tables = append(p.prev.tables, p.table)
		return p.prev.parse(tk)
	}

	// trim whitespace for all columns
	for i := range cells {
		cells[i] = strings.TrimSpace(cells[i])
	}

	switch cells[0] {
	case
		tag_dictionary,
		tag_dictionary_end:
		// previous length checks should catch this and
		// should never get here
		p.prev.tables = append(p.prev.tables, p.table)
		return p.prev.parse(tk)
	}

	p.table.Rows = append(p.table.Rows, cells)
	return p
}

func (p *objectParser) parse(tk *token) parser {
	k, v, _ := split(tk.value)

	if p.obj.Properties == nil {
		p.obj.Properties = make(map[string]string)
	}

	switch k {

	case tag_object_end:
		p.h.Object(p.obj)
		p.h.End(tag_object, p.obj.Name)
		return p.prev

	case prop_last_change:
		p.obj.Properties[k] = v
		if t, err := ParseTime(v); err == nil {
			p.obj.Modified = t
		}

	case "Alias":
		p.obj.Properties[k] = v
		p.obj.Alias = v
		if p.obj.Name != v {
			// update path with alias
			pth := filepath.Join(filepath.Dir(p.obj.Path), v)
			p.s.alias[p.obj.Path] = pth
			p.obj.Path = pth
		}

	case "DeviceId":
		p.obj.Properties[k] = v
		p.obj.DeviceId = v

	case "Type":
		p.obj.Properties[k] = v
		p.obj.Type = v

	case "{": // start of a CDT
		return &blockParser{
			prev:   p,
			name:   p.lastProp,
			lines:  []string{tk.value},
			endTag: "EndOfCDT",
			s:      p.s,
		}

	case "PanelObjectList":
		return &blockParser{
			prev:       p,
			name:       k,
			endTag:     "}",
			includeEnd: true,
			s:          p.s,
		}

	case prop_array, prop_members, prop_alarm_links:
		return &blockParser{
			prev:   p,
			name:   k,
			endTag: "End" + k,
			s:      p.s,
		}

	case prop_bytecode:
		return &codeParser{
			prev: p,
			s:    p.s,
		}

	default:
		p.obj.Properties[k] = v
	}

	p.lastProp = k

	return p
}

func (p *controllerParser) parse(tk *token) parser {
	k, v, _ := split(tk.value)
	switch k {
	case tag_object:
		p.h.Begin(k, v)
		pth := filepath.Join(p.path, v)
		if np, ok := p.s.alias[pth]; ok {
			pth = np
		}
		return newObjectParser(p.h, p, p.s, v, pth)

	case tag_infinet_ctlr:
		p.h.Begin(k, v)
		pth := filepath.Join(p.path, v)
		if np, ok := p.s.alias[pth]; ok {
			pth = np
		}
		return &infControllerParser{
			prev: p,
			name: v,
			path: pth,
			h:    p.h,
			s:    p.s,
		}
	case tag_device:
		p.h.Begin(k, v)
		pth := filepath.Join(p.path, v)
		if np, ok := p.s.alias[pth]; ok {
			pth = np
		}
		return &deviceParser{
			prev: p,
			name: v,
			path: pth,
			h:    p.h,
			s:    p.s,
		}
	case tag_container_end:
		p.h.End(tag_container, p.name)
		return p.last
	case tag_controller_end:
		p.h.End(tag_controller, p.name)
		return p.last
	}
	return p
}

func (p *containerParser) parse(tk *token) parser {
	k, v, _ := split(tk.value)
	switch k {
	case tag_object:
		p.h.Begin(k, v)
		pth := filepath.Join(p.path, v)
		if np, ok := p.s.alias[pth]; ok {
			pth = np
		}
		return newObjectParser(p.h, p, p.s, v, pth)

	case tag_device:
		p.h.Begin(k, v)
		pth := filepath.Join(p.path, v)
		if np, ok := p.s.alias[pth]; ok {
			pth = np
		}
		return &deviceParser{
			prev: p,
			name: v,
			path: pth,
			h:    p.h,
			s:    p.s,
		}

	case tag_container_end:
		p.h.End(tag_container, p.name)
		return p.last

	}
	return p
}

func (p *dmpParser) parse(tk *token) parser {
	k, v, _ := split(tk.value)
	switch k {

	case prop_path:
		if p.path == "" {
			p.path = v
			p.name = v
		}
		p.h.Path(v)

	case tag_dictionary:
		p.h.Begin(k, v)
		p.devPath = filepath.Join(p.path, v)
		pth := filepath.Join(p.path, v)
		if np, ok := p.s.alias[pth]; ok {
			pth = np
		}
		return &dictionaryParser{
			prev: p,
			h:    p.h,
			name: v,
			path: pth,
			s:    p.s,
		}

	case tag_infinet_ctlr:
		p.h.Begin(k, v)
		pth := filepath.Join(p.path, v)
		if np, ok := p.s.alias[pth]; ok {
			pth = np
		}
		return &infControllerParser{
			prev: p,
			name: v,
			path: pth,
			h:    p.h,
			s:    p.s,
		}

	case tag_device:
		p.h.Begin(k, v)
		pth := filepath.Join(p.path, v)
		if np, ok := p.s.alias[pth]; ok {
			pth = np
		}
		return &deviceParser{
			prev: p,
			name: v,
			path: pth,
			h:    p.h,
			s:    p.s,
		}

	case tag_controller_begin:
		p.h.Begin(k, v)
		pth := filepath.Join(p.path, v)
		if np, ok := p.s.alias[pth]; ok {
			pth = np
		}
		return &controllerParser{
			last: p,
			name: v,
			path: pth,
			h:    p.h,
			s:    p.s,
		}

	case tag_container_begin:
		p.h.Begin(k, v)
		pth := filepath.Join(p.path, v)
		if np, ok := p.s.alias[pth]; ok {
			pth = np
		}
		return &containerParser{
			last: p,
			name: v,
			path: pth,
			h:    p.h,
			s:    p.s,
		}

	case tag_object:
		p.h.Begin(k, v)
		pth := filepath.Join(p.path, v)
		if np, ok := p.s.alias[pth]; ok {
			pth = np
		}
		return newObjectParser(p.h, p, p.s, v, pth)

	}
	return p
}

func (p *infControllerParser) parse(tk *token) parser {
	k, v, _ := split(tk.value)
	switch k {
	case tag_object:
		p.h.Begin(k, v)
		pth := filepath.Join(p.path, v)
		if np, ok := p.s.alias[pth]; ok {
			pth = np
		}
		return newObjectParser(p.h, p, p.s, v, pth)

	case tag_infinet_ctlr_end:
		p.h.End(tag_infinet_ctlr, p.name)
		return p.prev
	}
	return p
}

func (p *deviceParser) parse(tk *token) parser {
	k, v, _ := split(tk.value)
	switch k {
	case tag_object:
		p.h.Begin(k, v)
		pth := filepath.Join(p.path, v)
		if np, ok := p.s.alias[pth]; ok {
			pth = np
		}
		return newObjectParser(p.h, p, p.s, v, pth)

	case tag_device_end:
		p.h.End(tag_device, p.name)
		return p.prev
	}
	return p
}

// ParseFile if the main function to start parsing the file
// as the file is parsed events will call the Handler methods.
func ParseFile(file string, h Handler) string {

	r, err := os.Open(file)
	if err != nil {
		fmt.Println("error opening file:", file)
		return ""
	}
	defer r.Close()

	return Parse(r, h)
}

// ParseFile if the main function to start parsing the file
// as the file is parsed events will call the Handler methods.
func Parse(r io.Reader, h Handler) string {

	p := newParser(h)

	err := scanWith(r, p)
	if err != nil {
		fmt.Println("Error:", err)
	}
	return p.devPath
}

func ParseAlarmLinks(s string) []*AlarmLink {

	alarms := make([]*AlarmLink, 0)
	scanner := bufio.NewScanner(strings.NewReader(s))
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), ":")
		if len(fields) != 3 {
			continue
		}
		for i := range fields {
			fields[i] = strings.TrimSpace(fields[i])
		}
		id, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}
		for len(alarms) < id {
			alarms = append(alarms, nil)
		}
		alarms[id-1] = &AlarmLink{
			Id:      id,
			Path:    fields[0],
			Enabled: fields[2] == "Enabled",
		}
	}
	return alarms
}
