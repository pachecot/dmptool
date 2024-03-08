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

// fixDoubleEOL removes the extra empty lines that are
// inserted into the dmp file bytecode sections.
func fixDoubleEOL(lines []string) []string {
	next := []string{}
	count := 0
	for _, line := range lines {
		if line == "" {
			count++
			if count%2 != 0 {
				continue
			}
		} else {
			count = 0
		}
		next = append(next, line)
	}
	return next
}

type parser interface {
	parse(string, int) parser
}

type blockParser struct {
	prev  *objectParser
	name  string
	end   string
	lines []string
}

func (p *blockParser) parse(line string, n int) parser {
	txt := strings.Trim(line, " ")
	if txt == p.end {
		p.prev.properties[p.name] = strings.Join(p.lines, "\r\n")
		return p.prev
	}
	p.lines = append(p.lines, line)
	return p
}

type codeParser struct {
	prev     *objectParser
	path     string
	kind     string
	modified time.Time
	lines    []string
	h        Handler
}

func (p *codeParser) parse(line string, n int) parser {
	txt := strings.Trim(line, " ")
	if txt == "EndByteCode" {
		lines := fixDoubleEOL(p.lines)
		p.prev.properties["ByteCode"] = strings.Join(lines, "\r\n")
		p.h.Code(&Code{
			Path:     p.path,
			Type:     p.kind,
			Modified: p.modified,
			Lines:    lines,
		})
		return p.prev
	}
	p.lines = append(p.lines, line)
	return p
}

type dictionaryParser struct {
	prev  parser
	h     Handler
	lines []string
}

func (p *dictionaryParser) parse(line string, n int) parser {
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
	prev       parser
	name       string
	path       string
	modified   time.Time
	last       string
	properties map[string]string
	h          Handler
}

func (p *objectParser) parse(line string, n int) parser {
	k, v, _ := split(line)

	if p.properties == nil {
		p.properties = make(map[string]string)
	}

	switch k {

	case "EndObject":
		p.h.Object(&Object{
			Path:       p.path,
			Modified:   p.modified,
			Properties: p.properties,
		})
		p.h.End("Object", p.name)
		return p.prev

	case "LastChange":
		p.properties[k] = v
		if t, err := ParseTime(v); err == nil {
			p.modified = t
		}

	case "{": // start of a CDT
		return &blockParser{
			prev:  p,
			name:  p.last,
			lines: []string{line},
			end:   "EndOfCDT",
		}

	case "Array", "Members", "AlarmLinks":
		return &blockParser{
			prev: p,
			name: k,
			end:  "End" + k,
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
	prev parser
	name string
	path string
	h    Handler
}

func (p *controllerParser) parse(line string, n int) parser {
	k, v, _ := split(line)
	switch k {
	case "Object":
		p.h.Begin(k, v)
		return &objectParser{
			prev: p,
			name: v,
			path: filepath.Join(p.path, v),
			h:    p.h,
		}
	case "InfinetCtlr":
		p.h.Begin(k, v)
		return &infControllerParser{
			prev: p,
			name: v,
			path: filepath.Join(p.path, v),
			h:    p.h,
		}
	case "EndController":
		p.h.Begin("Controller", p.name)
		return p.prev
	}
	return p
}

type dmpParser struct {
	path    string
	name    string
	devPath string
	h       Handler
}

func (p *dmpParser) parse(line string, n int) parser {
	k, v, _ := split(line)
	switch k {

	case "Path":
		p.path = v
		p.name = v

	case "Dictionary":
		p.h.Begin(k, v)
		p.devPath = filepath.Join(p.path, v)
		return &dictionaryParser{
			prev: p,
			h:    p.h,
		}

	case "InfinetCtlr":
		p.h.Begin(k, v)
		return &infControllerParser{
			prev: p,
			name: v,
			path: filepath.Join(p.path, v),
			h:    p.h,
		}

	case "BeginController":
		p.h.Begin(k, v)
		return &controllerParser{
			prev: p,
			name: v,
			path: filepath.Join(p.path, v),
			h:    p.h,
		}

	case "Object":
		p.h.Begin(k, v)
		return &objectParser{
			prev: p,
			name: v,
			path: filepath.Join(p.path, v),
			h:    p.h,
		}
	}

	return p
}

type infControllerParser struct {
	prev parser
	name string
	path string
	h    Handler
}

func (p *infControllerParser) parse(line string, n int) parser {
	k, v, _ := split(line)
	switch k {
	case "Object":
		p.h.Begin(k, v)
		return &objectParser{
			prev: p,
			name: v,
			path: filepath.Join(p.path, v),
			h:    p.h,
		}
	case "EndInfinetCtlr":
		p.h.End("InfinetCtlr", p.name)
		return p.prev
	}
	return p
}

func ParseFile(file string, h Handler) string {
	scanner := &scanner{
		FileName: file,
	}

	err := scanner.open()
	if err != nil {
		fmt.Println("error opening file:", file)
		return ""
	}
	defer scanner.close()

	p := &dmpParser{h: h}

	err = scanner.scan(p)
	if err != nil {
		fmt.Println("Error:", err)
	}

	return p.devPath
}
