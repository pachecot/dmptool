package list

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/tpacheco/dmptool/dmp"
)

type listHandler struct {
	dmp.EmptyHandler
	fields  []string
	types   []string
	filters []string
	results []*dmp.Object
}

func (h *listHandler) Object(do *dmp.Object) {

	if len(h.filters) > 0 {
		if !slices.ContainsFunc(h.filters,
			func(f string) bool {
				return strings.Contains(do.Name, f) ||
					strings.Contains(do.Path, f)
			}) {
			return
		}
	}

	if len(h.types) > 0 {
		if !slices.Contains(h.types, do.Type) {
			return
		}
	}

	h.results = append(h.results, do)
}

type Command struct {
	FileName string
	OutFile  string
	Fields   []string
	Types    []string
	Filters  []string
	Record   bool
}

func (cmd *Command) Execute() {

	h := &listHandler{
		fields:  cmd.Fields,
		filters: cmd.Filters,
		types:   cmd.Types,
	}

	dmp.ParseFile(cmd.FileName, h)

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

	cols := len(cmd.Fields)
	ws := make([]int, cols)
	table := make([][]string, 0, len(h.results))

	for i, n := range cmd.Fields {
		if ws[i] < len(n) {
			ws[i] = len(n)
		}
	}

	for _, obj := range h.results {
		row := make([]string, cols)
		table = append(table, row)
		for i, n := range cmd.Fields {
			if n == "Name" {
				row[i] = obj.Name
				if ws[i] < len(obj.Name) {
					ws[i] = len(obj.Name)
				}
				continue
			}
			if p, ok := obj.Properties[n]; ok {
				row[i] = p
				if ws[i] < len(p) {
					ws[i] = len(p)
				}
			}
		}
	}

	formats := make([]string, cols)
	for i, w := range ws {
		if i > 0 {
			formats[i] = fmt.Sprintf("  %%-%ds", w)
		} else {
			formats[i] = fmt.Sprintf("%%-%ds", w)
		}
	}

	for i, t := range cmd.Fields {
		fmt.Fprintf(w, formats[i], t)
	}
	fmt.Fprintln(w)

	for i := range cmd.Fields {
		fmt.Fprintf(w, formats[i], strings.Repeat("-", ws[i]))
	}
	fmt.Fprintln(w)

	for _, row := range table {
		for i, t := range row {
			fmt.Fprintf(w, formats[i], t)
		}
		fmt.Fprintln(w)
	}
}
