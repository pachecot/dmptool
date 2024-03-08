package dmp

import (
	"bufio"
	"os"
)

const maxTokenSize = 400000

type scanner struct {
	FileName string
	file     *os.File
	scanner  *bufio.Scanner
}

func (cmd *scanner) open() error {
	var err error
	cmd.file, err = os.Open(cmd.FileName)
	if err != nil {
		return err
	}
	cmd.scanner = bufio.NewScanner(cmd.file)
	cmd.scanner.Buffer(make([]byte, 0, maxTokenSize), maxTokenSize)
	return nil
}

func (cmd *scanner) close() error {
	return cmd.file.Close()
}

func (cmd *scanner) scan(p parser) error {
	line := 0
	for cmd.scanner.Scan() {
		if err := cmd.scanner.Err(); err != nil {
			return err
		}
		p = p.parse(cmd.scanner.Text(), line)
		line++
	}
	if err := cmd.scanner.Err(); err != nil {
		return err
	}
	return nil
}
