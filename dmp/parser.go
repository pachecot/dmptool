package dmp

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

func split(line string) (string, string, bool) {
	ss := strings.SplitN(line, ":", 2)
	if len(ss) == 1 {
		return strings.Trim(ss[0], " "), "", false
	}
	return strings.Trim(ss[0], " "), strings.Trim(ss[1], " "), true
}

const dmpTimeLayout = "1/2/2006 3:04:05 PM"

func parseDmpTime(value string) (time.Time, error) {
	return time.ParseInLocation(dmpTimeLayout, value, time.Local)
}

type Parser interface {
	Parse(string, int) Parser
}

type Handler interface {
	Code(p *Code)
	Object(p *objectParser)
	StartElement(p Parser)
	EndElement(p Parser)
}

type EmptyHandler struct{}

func (h *EmptyHandler) Code(p *Code)           {}
func (h *EmptyHandler) Object(p *objectParser) {}
func (h *EmptyHandler) StartElement(p Parser)  {}
func (h *EmptyHandler) EndElement(p Parser)    {}

type ForwardingHandler struct{ h Handler }

func (h *ForwardingHandler) Code(p *Code)           { h.h.Code(p) }
func (h *ForwardingHandler) Object(p *objectParser) { h.h.Object(p) }
func (h *ForwardingHandler) StartElement(p Parser)  { h.h.StartElement(p) }
func (h *ForwardingHandler) EndElement(p Parser)    { h.h.EndElement(p) }

type arrayParser struct {
	prev  *objectParser
	lines []string
	h     Handler
}

func (p *arrayParser) Parse(line string, n int) Parser {
	txt := strings.Trim(line, " ")
	if txt == "EndArray" {
		p.prev.properties["Array"] = strings.Join(p.lines, "\r\n")
		return p.prev
	}
	p.lines = append(p.lines, line)
	return p
}

type membersParser struct {
	prev  *objectParser
	lines []string
	h     Handler
}

func (p *membersParser) Parse(line string, n int) Parser {
	txt := strings.Trim(line, " ")
	if txt == "EndMembers" {
		p.prev.properties["Members"] = strings.Join(p.lines, "\r\n")
		return p.prev
	}
	p.lines = append(p.lines, line)
	return p
}

type alarmLinksParser struct {
	prev  *objectParser
	lines []string
	h     Handler
}

func (p *alarmLinksParser) Parse(line string, n int) Parser {
	txt := strings.Trim(line, " ")
	if txt == "EndAlarmLinks" {
		p.prev.properties["AlarmLinks"] = strings.Join(p.lines, "\r\n")
		return p.prev
	}
	p.lines = append(p.lines, line)
	return p
}

type Code struct {
	Path     string
	Type     string
	Modified time.Time
	Lines    []string
}

type codeParser struct {
	prev     *objectParser
	path     string
	kind     string
	modified time.Time
	lines    []string
	h        Handler
}

func (p *codeParser) Parse(line string, n int) Parser {
	txt := strings.Trim(line, " ")
	if txt == "EndByteCode" {
		p.prev.properties["ByteCode"] = strings.Join(p.lines, "\r\n")
		p.h.Code(&Code{
			Path:     p.path,
			Type:     p.kind,
			Modified: p.modified,
			Lines:    p.lines,
		})
		return p.prev
	}
	p.lines = append(p.lines, line)
	return p
}

type dictionaryParser struct {
	prev  Parser
	h     Handler
	lines []string
}

func (p *dictionaryParser) Parse(line string, n int) Parser {
	k, _, _ := split(line)
	switch k {
	case "Dictionary":
		return &dictionaryParser{
			prev: p,
			h:    p.h,
		}
	case "EndDictionary":
		return p.prev
	}

	p.lines = append(p.lines, line)
	return p
}

type objectParser struct {
	prev       Parser
	path       string
	modified   time.Time
	last       string
	cdt        []string
	properties map[string]string
	h          Handler
}

func (p *objectParser) Parse(line string, n int) Parser {
	k, v, _ := split(line)

	if p.properties == nil {
		p.properties = make(map[string]string)
	}

	if p.cdt != nil {
		// CDT mode, just read lines until end is encountered
		if k == "EndOfCDT" {
			p.properties[p.last] = strings.Join(p.cdt, "\r\n")
			p.cdt = nil
		} else {
			p.cdt = append(p.cdt, line)
		}
		return p
	}

	switch k {

	case "EndObject":
		p.h.EndElement(p)
		return p.prev

	case "LastChange":
		p.properties[k] = v
		if t, err := parseDmpTime(v); err == nil {
			p.modified = t
		}

	case "{": // start of a CDT
		p.cdt = []string{line}
		return p

	case "Array":
		return &arrayParser{
			prev: p,
			h:    p.h,
		}

	case "Members":
		return &membersParser{
			prev: p,
			h:    p.h,
		}

	case "AlarmLinks":
		return &alarmLinksParser{
			prev: p,
			h:    p.h,
		}

	case "ByteCode":
		return &codeParser{
			path:     p.path,
			modified: p.modified,
			kind:     p.properties["Type"],
			prev:     p,
			h:        p.h,
		}

	default:
		p.properties[k] = v
	}

	p.last = k

	return p
}

type controllerParser struct {
	prev Parser
	path string
	h    Handler
}

func (p *controllerParser) Parse(line string, n int) Parser {
	k, v, _ := split(line)
	switch k {
	case "Object":
		return &objectParser{
			prev: p,
			path: filepath.Join(p.path, v),
			h:    p.h,
		}
	case "InfinetCtlr":
		return &infControllerParser{
			prev: p,
			path: filepath.Join(p.path, v),
			h:    p.h,
		}
	case "EndController":
		return p.prev
	}
	return p
}

type dmpParser struct {
	path    string
	devPath string
	h       Handler
}

func (p *dmpParser) Parse(line string, n int) Parser {
	k, v, _ := split(line)
	switch k {
	case "Path":
		p.path = v
		return p
	case "Dictionary":
		p.devPath = filepath.Join(p.path, v)
		return &dictionaryParser{
			prev: p,
			h:    p.h,
		}
	case "InfinetCtlr":
		return &infControllerParser{
			prev: p,
			path: filepath.Join(p.path, v),
			h:    p.h,
		}
	case "BeginController":
		return &controllerParser{
			prev: p,
			path: filepath.Join(p.path, v),
			h:    p.h,
		}
	case "Object":
		return &objectParser{
			prev: p,
			path: filepath.Join(p.path, v),
			h:    p.h,
		}
	}
	return p
}

type infControllerParser struct {
	prev Parser
	path string
	h    Handler
}

func (s *infControllerParser) Parse(line string, n int) Parser {
	k, v, _ := split(line)
	switch k {
	case "Object":
		return &objectParser{
			prev: s,
			path: filepath.Join(s.path, v),
			h:    s.h,
		}
	case "EndInfinetCtlr":
		return s.prev
	}
	return s
}

func NewParser(h Handler) Parser {
	return &dmpParser{h: h}
}

func ParseFile(file string, h Handler) string {
	scanner := &Scanner{
		FileName: file,
	}

	err := scanner.Open()
	if err != nil {
		fmt.Println("error opening file:", file)
		return ""
	}
	defer scanner.Close()

	p := &dmpParser{h: h}

	err = scanner.Scan(p)
	if err != nil {
		fmt.Println("Error:", err)
	}

	return p.devPath
}
