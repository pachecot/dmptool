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

func split(line string) (string, string) {
	ss := strings.SplitN(line, ":", 2)
	if len(ss) == 1 {
		return strings.Trim(ss[0], " "), ""
	}
	return strings.Trim(ss[0], " "), strings.Trim(ss[1], " ")
}

const dmpTimeLayout = "1/2/2006 3:04:05 PM"

func parseDmpTime(value string) (time.Time, error) {
	return time.ParseInLocation(dmpTimeLayout, value, time.Local)
}

type scanner interface {
	Scan(string) scanner
}

type codeScanner struct {
	prev     scanner
	count    int
	path     string
	modified time.Time
	lines    []string
}

func (s *codeScanner) Scan(line string) scanner {
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

type dictionaryScanner struct {
	prev scanner
}

func (s *dictionaryScanner) Scan(line string) scanner {
	k, _ := split(line)
	switch k {
	case "Dictionary":
		return &dictionaryScanner{
			prev: s,
		}
	case "EndDictionary":
		return s.prev
	}
	return s
}

type objectScanner struct {
	prev     scanner
	deviceId string
	path     string
	modified time.Time
}

func (s *objectScanner) Scan(line string) scanner {
	k, v := split(line)
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
		return &codeScanner{
			path:     s.path,
			modified: s.modified,
			prev:     s,
		}
	}
	return s
}

type controllerScanner struct {
	prev scanner
	path string
}

func (s *controllerScanner) Scan(line string) scanner {
	k, v := split(line)
	switch k {
	case "Object":
		return &objectScanner{
			prev: s,
			path: path.Join(s.path, v),
		}
	case "InfinetCtlr":
		return &infControllerScanner{
			prev: s,
			path: path.Join(s.path, v),
		}
	case "EndController":
		return s.prev
	}
	return s
}

type dmpScanner struct {
	outDir string
}

func (s *dmpScanner) Scan(line string) scanner {
	k, v := split(line)
	switch k {
	case "Dictionary":
		return &dictionaryScanner{
			prev: s,
		}
	case "InfinetCtlr":
		return &infControllerScanner{
			prev: s,
			path: path.Join(s.outDir, v),
		}
	case "BeginController":
		return &controllerScanner{
			prev: s,
			path: path.Join(s.outDir, v),
		}
	case "Object":
		return &objectScanner{
			prev: s,
			path: path.Join(s.outDir, v),
		}
	}
	return s
}

type infControllerScanner struct {
	prev scanner
	path string
}

func (s *infControllerScanner) Scan(line string) scanner {
	k, v := split(line)
	switch k {
	case "Object":
		return &objectScanner{
			prev: s,
			path: path.Join(s.path, v),
		}
	case "EndInfinetCtlr":
		return s.prev
	}
	return s
}

func cmdExtractPE(srcFile string, destDir string) {

	sf, err := os.Open(srcFile)
	if err != nil {
		fmt.Println("error opening file:", srcFile)
		return
	}
	defer sf.Close()

	ss := bufio.NewScanner(sf)

	var s scanner
	s = &dmpScanner{
		outDir: destDir,
	}

	for ss.Scan() {
		if err = ss.Err(); err != nil {
			fmt.Println("Error:", ss.Err())
			return
		}
		s = s.Scan(ss.Text())
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
			cmdExtractPE(args[0], args[1])
		},
	}

	cmd.AddCommand(peCmd)
	cmd.Execute()
}
