package dmp

import "strings"

func split(line string) (string, string, bool) {
	ss := strings.SplitN(line, ":", 2)
	if len(ss) == 1 {
		return strings.Trim(ss[0], " "), "", false
	}
	return strings.Trim(ss[0], " "), strings.Trim(ss[1], " "), true
}

// fixDoubleEOL removes the extra empty lines that are
// in the dmp file bytecode sections.
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
