package list

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path"
	"slices"
	"sort"
	"strings"

	"github.com/tpacheco/dmptool/dmp"
	"github.com/xuri/excelize/v2"
)

type listHandler struct {
	dmp.EmptyHandler
	fields   []string
	types    []string
	names    []string
	devices  []string
	results  []*dmp.Object
	whereExp expression
}

const (
	csvExt  = ".csv"
	xlsxExt = ".xlsx"
)

func (h *listHandler) Object(do *dmp.Object) {

	if len(h.types) > 0 && !slices.Contains(h.types, do.Type) {
		return
	}

	if len(h.names) > 0 && !slices.ContainsFunc(h.names, func(f string) bool {
		return strings.Contains(do.Name, f)
	}) {
		return
	}

	if len(h.devices) > 0 && !slices.ContainsFunc(h.devices, func(f string) bool {
		return strings.Contains(do.Path, f)
	}) {
		return
	}

	if h.whereExp != nil && !h.whereExp.match(do) {
		return
	}

	h.results = append(h.results, do)
}

type Command struct {
	FileName string
	OutFile  string
	Fields   []string
	Types    []string
	Filter   string
	Names    []string
	Devices  []string
	Ordering []string
}

func (cmd *Command) Execute() {

	h := &listHandler{
		fields:  cmd.Fields,
		names:   cmd.Names,
		devices: cmd.Devices,
		types:   cmd.Types,
	}

	if cmd.Filter != "" {
		h.whereExp = parseWhere(cmd.Filter)
	}

	dmp.ParseFile(cmd.FileName, h)

	// if no fields are given then display the fields
	// or if first field is *, -, or ? then display the
	// fields
	switch len(cmd.Fields) {
	case 0:
		processFields(h)
		return
	case 1:
		switch cmd.Fields[0] {
		case "?", "-", "*":
			processFields(h)
			return
		}
	}

	table := buildTable(cmd, h)

	if len(cmd.Ordering) > 0 {
		err := reorder(cmd.Ordering, cmd.Fields, table)
		if err != nil {
			fmt.Printf("could not reorder results: %s", err)
			return
		}
	}

	switch strings.ToLower(path.Ext(cmd.OutFile)) {
	case xlsxExt:
		writeXlsx(cmd.OutFile, cmd, table)
	case csvExt:
		writeCSV(cmd.OutFile, cmd, table)
	default:
		w := os.Stdout
		if cmd.OutFile != "" {
			f, err := os.Create(cmd.OutFile)
			if err != nil {
				fmt.Println("could not create file")
				return
			}
			defer func() {
				f.Sync()
				f.Close()
			}()
			w = f
		}
		writeFile(w, cmd, table)
	}
}

func listFields(h *listHandler) []string {
	fields := make(map[string]struct{})
	for _, obj := range h.results {
		for k := range obj.Properties {
			fields[k] = struct{}{}
		}
	}
	names := make([]string, 0, len(fields))
	for k := range fields {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

func processFields(h *listHandler) {
	for _, name := range listFields(h) {
		fmt.Println(name)
	}
}

func buildTable(cmd *Command, h *listHandler) [][]string {

	cols := len(cmd.Fields)
	table := make([][]string, 0, len(h.results))

	for _, obj := range h.results {
		row := make([]string, cols)
		table = append(table, row)
		for i, n := range cmd.Fields {
			if p, ok := obj.Properties[n]; ok {
				row[i] = p
			}
		}
	}
	return table
}

func writeFile(w *os.File, cmd *Command, table [][]string) {

	cols := len(cmd.Fields)

	ws := widths(cmd, table)

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

func widths(cmd *Command, table [][]string) []int {
	cols := len(cmd.Fields)
	ws := make([]int, cols)
	for i, n := range cmd.Fields {
		if ws[i] < len(n) {
			ws[i] = len(n)
		}
	}
	for _, row := range table {
		for i, n := range row {
			if ws[i] < len(n) {
				ws[i] = len(n)
			}
		}
	}
	return ws
}

func writeCSV(fileName string, cmd *Command, table [][]string) {

	w, err := os.Create(fileName)
	if err != nil {
		fmt.Println("could not create file")
		return
	}
	defer func() {
		w.Sync()
		w.Close()
	}()

	csvW := csv.NewWriter(w)
	if err := csvW.Write(cmd.Fields); err != nil {
		log.Fatalln("error writing record to csv:", err)
	}

	for _, row := range table {
		if err := csvW.Write(row); err != nil {
			log.Fatalln("error writing record to csv:", err)
		}
	}
	csvW.Flush()
}

func writeXlsx(fName string, cmd *Command, table [][]string) {

	ws := widths(cmd, table)

	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	for i, col := range ws {
		f.SetColWidth("Sheet1", fmt.Sprintf("%c", 'A'+i), fmt.Sprintf("%c", 'A'+i), float64(col))
	}

	f.SetSheetRow("Sheet1", "A1", &cmd.Fields)
	for i, row := range table {
		f.SetSheetRow("Sheet1", fmt.Sprintf("%c%d", 'A', i+2), &row)
	}

	if err := f.SaveAs(fName); err != nil {
		fmt.Println(err)
	}
}
