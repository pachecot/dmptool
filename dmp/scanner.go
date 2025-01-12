package dmp

import (
	"bufio"
	"io"
)

// maxTokenSize is the for buffer used by the buffio Scanner.
// Found this to work. Not sure on max size, but default is too small.
const maxTokenSize = 400000

func scanWith(r io.Reader, p parser) error {

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, maxTokenSize), maxTokenSize)

	line := 0
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return err
		}
		// some lines are terminated with extra \r scan doesn't remove
		text := trimR(scanner.Text())
		p = p.parse(&token{
			value: text,
			line:  line,
		})
		line++
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}
