package ref

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"unicode"

	"github.com/tpacheco/dmptool/dmp"
	"golang.org/x/exp/maps"
)

type refHandler struct {
	dmp.EmptyHandler
	refs         map[string][]*dmp.Object
	withGraphics bool
	withCode     bool
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

func (h *refHandler) Object(do *dmp.Object) {

	typ := do.Type

	switch typ {
	case "Graphics":
		if !h.withGraphics {
			return
		}
		if cdt, ok := do.Properties["PanelObjectList"]; ok {
			lines := strings.Split(cdt, "\n")
			for _, line := range lines {
				refs := parseRefs(line)
				for _, r := range refs {
					h.refs[r] = append(h.refs[r], do)
				}
			}
			return
		}

	case "InfinityFunction", "Program", "InfinityProgram":
		if !h.withCode {
			return
		}
		if byteCode, ok := do.Properties["ByteCode"]; ok {
			lines := strings.Split(byteCode, "")
			for _, line := range lines {
				line = removeComment(line)
				refs := parseRefs(line)
				for _, r := range refs {
					h.refs[r] = append(h.refs[r], do)
				}
			}
			return
		}
	}
}

type Command struct {
	FileName string
	OutFile  string
	Bare     bool
	All      bool
	Sources  bool
	ShowType bool
	Graphics bool
	Code     bool
}

func (cmd *Command) Execute() {

	h := &refHandler{
		refs:         make(map[string][]*dmp.Object),
		withGraphics: cmd.Graphics,
		withCode:     cmd.Code,
	}

	dmpPath := dmp.ParseFile(cmd.FileName, h)

	refs := maps.Keys(h.refs)
	slices.Sort(refs)

	if !cmd.All {
		refs = slices.DeleteFunc(refs, func(s string) bool {
			return strings.HasPrefix(s, dmpPath)
		})
	}

	w := os.Stdout

	if cmd.OutFile != "" {
		var err error
		w, err = os.Create(cmd.OutFile)
		if err != nil {
			fmt.Println("could not create file")
			return
		}
		defer func() {
			w.Sync()
			w.Close()
		}()
	}

	if cmd.Bare {
		for _, v := range refs {
			fmt.Fprintln(w, v)
		}
		return
	}

	if len(refs) == 0 {
		fmt.Fprintln(w, "No references found.")
		return
	}

	fmt.Fprintf(w, "Device external references\n\n  Source device: %s\n\n", dmpPath)

	if cmd.Sources {
		if cmd.ShowType {
			printWithSourceAndType(w, h)
			return
		}
		printWithSource(w, h)
		return
	}
	printWithoutSource(w, h)
}

func printWithoutSource(w io.Writer, h *refHandler) {
	refs := maps.Keys(h.refs)
	slices.Sort(refs)

	wRef := len("External Reference")
	for _, v := range refs {
		n := len(v)
		if wRef < n {
			wRef = n
		}
	}

	fmt.Fprintf(w, "%s%s %5s\n", "External Reference", strings.Repeat(" ", wRef-len("External Reference")), "Count")
	fmt.Fprintf(w, "%s%s %5s\n", strings.Repeat("-", wRef), "", "-----")
	for _, v := range refs {
		fmt.Fprintf(w, "%s%s %5d\n", v, strings.Repeat(" ", wRef-len(v)), len(h.refs[v]))
	}
}

func printWithSourceAndType(w io.Writer, h *refHandler) {

	refs := maps.Keys(h.refs)
	slices.Sort(refs)

	wRef := len("External Reference")
	for _, v := range refs {
		n := len(v)
		if wRef < n {
			wRef = n
		}
	}

	wSrc := 0
	for _, dos := range h.refs {
		for _, do := range dos {
			n := len(do.Path)
			if wSrc < n {
				wSrc = n
			}
		}
	}

	fmt.Fprintf(w, "%s%s %s%s %s\n", "External Reference", strings.Repeat(" ", wRef-len("External Reference")), "Source", strings.Repeat(" ", wSrc-len("Source")), "Type Name")
	fmt.Fprintf(w, "%s %s %s\n", strings.Repeat("-", wRef), strings.Repeat("-", wSrc), "----------------")
	for _, v := range refs {
		dos := h.refs[v]
		fmt.Fprintf(w, "%s%s %s%s %s\n", v, strings.Repeat(" ", wRef-len(v)), dos[0].Path, strings.Repeat(" ", wSrc-len(dos[0].Path)), dos[0].Type)
		for _, do := range dos[1:] {
			fmt.Fprintf(w, "-%s%s %s%s %s\n", strings.Repeat(" ", len(v)-1), strings.Repeat(" ", wRef-len(v)), do.Path, strings.Repeat(" ", wSrc-len(do.Path)), do.Type)
		}
	}

}

func printWithSource(w io.Writer, h *refHandler) {

	refs := maps.Keys(h.refs)
	slices.Sort(refs)

	wRef := len("External Reference")
	for _, v := range refs {
		n := len(v)
		if wRef < n {
			wRef = n
		}
	}

	wSrc := 0
	for _, dos := range h.refs {
		for _, do := range dos {
			n := len(do.Path)
			if wSrc < n {
				wSrc = n
			}
		}
	}

	fmt.Fprintf(w, "%s%s %s\n", "External Reference", strings.Repeat(" ", wRef-len("External Reference")), "Source")
	fmt.Fprintf(w, "%s %s\n", strings.Repeat("-", wRef), strings.Repeat("-", wSrc))
	for _, v := range refs {
		dos := h.refs[v]
		fmt.Fprintf(w, "%s%s %s\n", v, strings.Repeat(" ", wRef-len(v)), dos[0].Path)
		for _, do := range dos[1:] {
			fmt.Fprintf(w, "-%s%s %s\n", strings.Repeat(" ", len(v)-1), strings.Repeat(" ", wRef-len(v)), do.Path)
		}
	}
}
