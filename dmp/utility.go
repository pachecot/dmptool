package dmp

import "strings"

func split(line string) (string, string, bool) {
	ss := strings.SplitN(line, ":", 2)
	if len(ss) == 1 {
		return strings.Trim(ss[0], " "), "", false
	}
	return strings.Trim(ss[0], " "), strings.Trim(ss[1], " "), true
}

func trimR(s string) string {
	i := len(s)
	for i > 0 && s[i-1] == '\r' {
		i--
	}
	return s[:i]
}
