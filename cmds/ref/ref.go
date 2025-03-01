package ref

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"slices"
	"strings"
	"unicode"

	"github.com/tpacheco/dmptool/dmp"
	"github.com/xuri/excelize/v2"
	"golang.org/x/exp/maps"
)

const (
	csvExt  = ".csv"
	xlsxExt = ".xlsx"
)

type refHandler struct {
	dmp.EmptyHandler
	refs         map[string][]*dmp.Object
	withGraphics bool
	withCode     bool
	withAlarms   bool
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
	if start >= 0 && s[start] == ' ' {
		return nil
	}
	end := i + 1
	if end < len(s) && s[end] == ' ' {
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
			lines := strings.Split(byteCode, "\n")
			for _, line := range lines {
				line = removeComment(line)
				refs := parseRefs(line)
				for _, r := range refs {
					h.refs[r] = append(h.refs[r], do)
				}
			}
			return
		}

	default:
		if !h.withAlarms {
			return
		}
		if links, ok := do.Properties["AlarmLinks"]; ok {
			alarms := dmp.ParseAlarmLinks(links)
			for _, alarm := range alarms {
				if alarm == nil {
					continue
				}
				r := alarm.Path
				h.refs[r] = append(h.refs[r], do)
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
	Alarms   bool
	Code     bool
}

func (cmd *Command) Execute() {

	h := &refHandler{
		refs:         make(map[string][]*dmp.Object),
		withGraphics: cmd.Graphics,
		withCode:     cmd.Code,
		withAlarms:   cmd.Alarms,
	}

	dmpPath := dmp.ParseFile(cmd.FileName, h)

	refs := maps.Keys(h.refs)
	slices.Sort(refs)

	if !cmd.All {
		refs = slices.DeleteFunc(refs, func(s string) bool {
			return strings.HasPrefix(s, dmpPath)
		})
	}

	if len(refs) == 0 {
		fmt.Println("No references found.")
		return
	}

	if cmd.Bare {
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

		for _, v := range refs {
			fmt.Fprintln(w, v)
		}
		return
	}

	table := getTable(cmd, refs, h)

	switch strings.ToLower(path.Ext(cmd.OutFile)) {
	case xlsxExt:
		writeXlsx(cmd.OutFile, table)
	case csvExt:
		writeCSV(cmd.OutFile, table)
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
		writeFile(w, dmpPath, table)
	}
}

func writeFile(w io.Writer, dmpPath string, table [][]string) {
	fmt.Fprintf(w, "Device external references\n\n  Source device: %s\n\n", dmpPath)
	printTable(w, table, false)
}

func getTable(cmd *Command, refs []string, h *refHandler) [][]string {
	f := withoutSource
	if cmd.Sources {
		f = withSource
		if cmd.ShowType {
			f = withSourceAndType
		}
	}
	return f(refs, h)
}

func withoutSource(refs []string, h *refHandler) [][]string {

	table := make([][]string, 0, len(h.refs))
	table = append(table, []string{"External Reference", "Count"})

	for _, v := range refs {
		table = append(table, []string{v, fmt.Sprintf("%5d", len(h.refs[v]))})
	}
	return table
}

func withSourceAndType(refs []string, h *refHandler) [][]string {

	table := make([][]string, 0, len(h.refs))
	table = append(table, []string{"External Reference", "Source", "Type Name"})

	for _, v := range refs {
		for _, do := range h.refs[v] {
			table = append(table, []string{v, do.Path, do.Type})
		}
	}
	return table
}

func withSource(refs []string, h *refHandler) [][]string {

	table := make([][]string, 0, len(h.refs))
	table = append(table, []string{"External Reference", "Source"})
	for _, v := range refs {
		for _, do := range h.refs[v] {
			table = append(table, []string{v, do.Path})
		}
	}
	return table
}

func printTable(w io.Writer, table [][]string, elide bool) {

	col_sep := " "
	col_pad := " "

	widths := lens(table)

	{
		// print header
		pad := ""
		for i, col := range table[0] {
			fmt.Fprintf(w, "%s%s", pad, col)
			pad = strings.Repeat(col_pad, widths[i]-len(col)) + col_sep
		}
		fmt.Fprintln(w)
	}

	{
		// print header line
		pad := ""
		for _, col := range widths {
			fmt.Fprintf(w, "%s%s", pad, strings.Repeat("-", col))
			pad = col_sep
		}
		fmt.Fprintln(w)
	}

	ref := ""
	for _, row := range table[1:] {
		pad := ""
		for i, col := range row {
			if elide && i == 0 && ref == col {
				col = "-"
			}
			fmt.Fprintf(w, "%s%s", pad, col)
			pad = strings.Repeat(col_pad, widths[i]-len(col)) + col_sep
		}
		fmt.Fprintln(w)
		ref = row[0]
	}
}

func lens(table [][]string) []int {
	widths := make([]int, len(table[1]))
	for _, row := range table {
		for i, col := range row {
			n := len(col)
			if widths[i] < n {
				widths[i] = n
			}
		}
	}
	return widths
}

func writeCSV(fileName string, table [][]string) {

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

	header := table[0]
	if err := csvW.Write(header); err != nil {
		log.Fatalln("error writing record to csv:", err)
	}

	for _, row := range table[1:] {
		if err := csvW.Write(row); err != nil {
			log.Fatalln("error writing record to csv:", err)
		}
	}
	csvW.Flush()
}

func writeXlsx(fName string, table [][]string) {

	ws := lens(table)

	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	for i, col := range ws {
		f.SetColWidth("Sheet1", fmt.Sprintf("%c", 'A'+i), fmt.Sprintf("%c", 'A'+i), float64(col))
	}

	for i, row := range table {
		f.SetSheetRow("Sheet1", fmt.Sprintf("%c%d", 'A', i+1), &row)
	}

	if err := f.SaveAs(fName); err != nil {
		fmt.Println(err)
	}
}
