package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/spf13/cobra"
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

type parser interface {
	Parse(string) parser
}

type codeParser struct {
	prev     parser
	count    int
	path     string
	modified time.Time
	lines    []string
}

func (s *codeParser) Parse(line string) parser {
	txt := strings.Trim(line, " ")
	if txt == "EndByteCode" {
		file := s.path + ".pe"
		err := os.MkdirAll(path.Dir(file), os.ModeDir)
		if err != nil {
			fmt.Println("error creating directory: ", err)
			return s.prev
		}
		os.WriteFile(file, []byte(strings.Join(s.lines, "\n")), os.ModePerm)
		if !s.modified.IsZero() {
			os.Chtimes(file, s.modified, s.modified)
		}
		return s.prev
	}
	if line == "" {
		s.count++
		if s.count%2 == 0 {
			return s
		}
	} else {
		s.count = 0
	}
	s.lines = append(s.lines, line)
	return s
}

type dictionaryParser struct {
	prev parser
}

func (s *dictionaryParser) Parse(line string) parser {
	k, _, _ := split(line)
	switch k {
	case "Dictionary":
		return &dictionaryParser{
			prev: s,
		}
	case "EndDictionary":
		return s.prev
	}
	return s
}

type objectParser struct {
	prev     parser
	deviceId string
	path     string
	modified time.Time
}

func (s *objectParser) Parse(line string) parser {
	k, v, _ := split(line)
	switch k {
	case "EndObject":
		return s.prev
	case "DeviceId":
		s.deviceId = v
	case "LastChange":
		if t, err := parseDmpTime(v); err == nil {
			s.modified = t
		}
	case "ByteCode":
		return &codeParser{
			path:     s.path,
			modified: s.modified,
			prev:     s,
		}
	}
	return s
}

type controllerParser struct {
	prev parser
	path string
}

func (s *controllerParser) Parse(line string) parser {
	k, v, _ := split(line)
	switch k {
	case "Object":
		return &objectParser{
			prev: s,
			path: path.Join(s.path, v),
		}
	case "InfinetCtlr":
		return &infControllerParser{
			prev: s,
			path: path.Join(s.path, v),
		}
	case "EndController":
		return s.prev
	}
	return s
}

type dmpParser struct {
	outDir string
}

func (s *dmpParser) Parse(line string) parser {
	k, v, _ := split(line)
	switch k {
	case "Dictionary":
		return &dictionaryParser{
			prev: s,
		}
	case "InfinetCtlr":
		return &infControllerParser{
			prev: s,
			path: path.Join(s.outDir, v),
		}
	case "BeginController":
		return &controllerParser{
			prev: s,
			path: path.Join(s.outDir, v),
		}
	case "Object":
		return &objectParser{
			prev: s,
			path: path.Join(s.outDir, v),
		}
	}
	return s
}

type infControllerParser struct {
	prev parser
	path string
}

func (s *infControllerParser) Parse(line string) parser {
	k, v, _ := split(line)
	switch k {
	case "Object":
		return &objectParser{
			prev: s,
			path: path.Join(s.path, v),
		}
	case "EndInfinetCtlr":
		return s.prev
	}
	return s
}

type peCommand struct {
	fileName string
	outDir   string
	file     *os.File
	scanner  *bufio.Scanner
}

func (cmd *peCommand) Open() error {
	var err error
	cmd.file, err = os.Open(cmd.fileName)
	if err != nil {
		return err
	}
	cmd.scanner = bufio.NewScanner(cmd.file)
	return nil
}

func (cmd *peCommand) Close() error {
	return cmd.file.Close()
}

func (cmd *peCommand) Scan(pf parser) error {
	for cmd.scanner.Scan() {
		if err := cmd.scanner.Err(); err != nil {
			return err
		}
		pf = pf.Parse(cmd.scanner.Text())
	}
	return nil
}

func (cmd *peCommand) Execute() {
	err := cmd.Open()
	if err != nil {
		fmt.Println("error opening file:", cmd.fileName)
		return
	}
	defer cmd.Close()

	p := &dmpParser{
		outDir: cmd.outDir,
	}
	err = cmd.Scan(p)
	if err != nil {
		fmt.Println("Error:", err)
	}
}

func main() {

	cmd := &cobra.Command{
		Use:   "dmptool <command>",
		Short: "continuum dump file tool",
	}

	peCmd := &cobra.Command{
		Use:   "pe <dump file> <output directory>",
		Short: "extract all PE program files into individual files a directory structure",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			c := &peCommand{
				fileName: args[0],
				outDir:   args[1],
			}
			c.Execute()
		},
	}

	cmd.AddCommand(peCmd)
	cmd.Execute()
}
