package ref

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"unicode"

	"github.com/tpacheco/dmptool/dmp"
	"golang.org/x/exp/maps"
)

type refHandler struct {
	dmp.EmptyHandler
	refs map[string][]string
}

func isValid(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '\\' || r == '_' || r == '.'
}

func removeComment(s string) string {
	i := strings.Index(s, "'")
	if i < 0 {
		return s
	}
	return s[:i]
}

func parseRefs(s string) []string {
	i := strings.Index(s, "\\")
	if i < 0 {
		return nil
	}
	start := i - 1
	if s[start] == ' ' {
		return nil
	}
	end := i + 1
	if s[end] == ' ' {
		return nil
	}
	for start > 0 && isValid(rune(s[start-1])) {
		start--
	}
	for end < len(s) && isValid(rune(s[end])) {
		end++
	}
	return append([]string{s[start:end]}, parseRefs(s[end:])...)
}

func (h *refHandler) Code(c *dmp.Code) {
	for _, line := range c.Lines {
		line = removeComment(line)
		refs := parseRefs(line)
		for _, r := range refs {
			h.refs[r] = append(h.refs[r], c.Path)
		}
	}
}

type Command struct {
	FileName string
	OutFile  string
	Bare     bool
	All      bool
	Sources  bool
}

func (cmd *Command) Execute() {

	h := &refHandler{
		refs: make(map[string][]string),
	}

	dmpPath := dmp.ParseFile(cmd.FileName, h)

	refs := maps.Keys(h.refs)
	slices.Sort(refs)

	if !cmd.All {
		refs = slices.DeleteFunc(refs, func(s string) bool {
			return strings.HasPrefix(s, dmpPath)
		})
	}

	f := os.Stdout

	if cmd.OutFile != "" {
		var err error
		f, err = os.Create(cmd.OutFile)
		if err != nil {
			fmt.Println("could not create file")
			return
		}
		defer func() { f.Close() }()
	}

	if cmd.Bare {
		for _, v := range refs {
			fmt.Fprintln(f, v)
		}
		return
	}

	if len(refs) == 0 {
		fmt.Fprintln(f, "No references found.")
		return
	}

	fmt.Fprintf(f, "Device external references\n\n  Source device: %s\n\n", dmpPath)

	w := len("External Reference")
	for _, v := range refs {
		n := len(v)
		if w < n {
			w = n
		}
	}
	if !cmd.Sources {

		fmt.Fprintf(f, "%s%s %5s\n", "External Reference", strings.Repeat(" ", w-len("External Reference")), "Count")
		fmt.Fprintf(f, "%s%s %5s\n", strings.Repeat("-", w), "", "-----")
		for _, v := range refs {
			fmt.Fprintf(f, "%s%s %5d\n", v, strings.Repeat(" ", w-len(v)), len(h.refs[v]))
		}
		return
	}

	w1 := 0
	for _, xs := range h.refs {
		for _, x := range xs {
			n := len(x)
			if w1 < n {
				w1 = n
			}
		}
	}

	fmt.Fprintf(f, "%s%s %s\n", "External Reference", strings.Repeat(" ", w-len("External Reference")), "Source")
	fmt.Fprintf(f, "%s %s\n", strings.Repeat("-", w), strings.Repeat("-", w1))
	for _, v := range refs {
		xs := h.refs[v]
		fmt.Fprintf(f, "%s%s %s\n", v, strings.Repeat(" ", w-len(v)), xs[0])
		for _, x := range xs[1:] {
			fmt.Fprintf(f, "-%s%s %s\n", strings.Repeat(" ", len(v)-1), strings.Repeat(" ", w-len(v)), x)
		}
	}

}
