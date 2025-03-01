package list

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/tpacheco/dmptool/dmp"
	"github.com/xuri/excelize/v2"
)

type listHandler struct {
	dmp.EmptyHandler
	fields  []string
	types   []string
	filters []string
	names   []string
	devices []string
	results []*dmp.Object
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

	if len(h.filters) > 0 && !slices.ContainsFunc(h.filters, func(f string) bool {
		if i := strings.Index(f, "="); i > 0 {
			k := strings.TrimSpace(f[:i])
			v := strings.TrimSpace(f[i+1:])
			p, ok := do.Properties[k]
			if !ok {
				return false
			}
			return p == v || strings.Contains(p, v)
		}
		return strings.Contains(do.Name, f) || strings.Contains(do.Path, f)
	}) {
		return
	}

	h.results = append(h.results, do)
}

type Command struct {
	FileName string
	OutFile  string
	Fields   []string
	Types    []string
	Filters  []string
	Names    []string
	Devices  []string
	Record   bool
}

func (cmd *Command) Execute() {

	h := &listHandler{
		fields:  cmd.Fields,
		names:   cmd.Names,
		devices: cmd.Devices,
		filters: cmd.Filters,
		types:   cmd.Types,
	}

	dmp.ParseFile(cmd.FileName, h)

	table := buildTable(cmd, h)

	switch strings.ToLower(path.Ext(cmd.OutFile)) {
	case xlsxExt:
		writeXlsx(cmd.OutFile, cmd, table)
	case csvExt:
		writeCSV(cmd.OutFile, cmd, table)
	default:
		writeFile(os.Stdout, cmd, table)
	}
}

func buildTable(cmd *Command, h *listHandler) [][]string {

	cols := len(cmd.Fields)
	table := make([][]string, 0, len(h.results))

	for _, obj := range h.results {
		row := make([]string, cols)
		table = append(table, row)
		for i, n := range cmd.Fields {
			if n == "Name" {
				row[i] = obj.Name
				continue
			}
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
