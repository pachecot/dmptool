package report

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xuri/excelize/v2"
)

const (
	csvExt  = ".csv"
	xlsxExt = ".xlsx"
)

func NewCmdReport() *cobra.Command {
	cc := &cobra.Command{
		Use:   "report <subcommand> <dump file>",
		Short: "generate structured reports from a dump file",
		Long: `Generate structured reports from a Continuum dump file.

Available subcommands:
  io      - IO points (InfinityInput and InfinityOutput) with electrical and engineering scale info
  points  - Point summary (String, Numeric, Input, Output, DateTime) with value and setpoint
  alarms  - Points that have alarm links, expanded one row per alarm
  logs    - Points that have logging configured
`,
		Args: cobra.MinimumNArgs(2),
	}
	cc.AddCommand(
		NewCmdIO(),
		NewCmdPoints(),
		NewCmdAlarms(),
		NewCmdLogs(),
	)
	return cc
}

func colWidths(headers []string, table [][]string) []int {
	ws := make([]int, len(headers))
	for i, h := range headers {
		ws[i] = len(h)
	}
	for _, row := range table {
		for i, cell := range row {
			if i < len(ws) && len(cell) > ws[i] {
				ws[i] = len(cell)
			}
		}
	}
	return ws
}

func writeText(w io.Writer, headers []string, table [][]string) {
	ws := colWidths(headers, table)

	formats := make([]string, len(headers))
	for i, width := range ws {
		if i > 0 {
			formats[i] = fmt.Sprintf("  %%-%ds", width)
		} else {
			formats[i] = fmt.Sprintf("%%-%ds", width)
		}
	}

	for i, h := range headers {
		fmt.Fprintf(w, formats[i], h)
	}
	fmt.Fprintln(w)

	for i, width := range ws {
		fmt.Fprintf(w, formats[i], strings.Repeat("-", width))
	}
	fmt.Fprintln(w)

	for _, row := range table {
		for i, cell := range row {
			fmt.Fprintf(w, formats[i], cell)
		}
		fmt.Fprintln(w)
	}
}

func writeCSV(fileName string, headers []string, table [][]string) {
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
	if err := csvW.Write(headers); err != nil {
		log.Fatalln("error writing record to csv:", err)
	}
	for _, row := range table {
		if err := csvW.Write(row); err != nil {
			log.Fatalln("error writing record to csv:", err)
		}
	}
	csvW.Flush()
}

func writeXLSX(fName string, headers []string, table [][]string) {
	ws := colWidths(headers, table)

	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	for i, width := range ws {
		f.SetColWidth("Sheet1", fmt.Sprintf("%c", 'A'+i), fmt.Sprintf("%c", 'A'+i), float64(width))
	}

	f.SetSheetRow("Sheet1", "A1", &headers)
	for i, row := range table {
		f.SetSheetRow("Sheet1", fmt.Sprintf("%c%d", 'A', i+2), &row)
	}

	if err := f.SaveAs(fName); err != nil {
		fmt.Println(err)
	}
}

func outputTable(outFile string, headers []string, table [][]string) {
	switch strings.ToLower(path.Ext(outFile)) {
	case xlsxExt:
		writeXLSX(outFile, headers, table)
	case csvExt:
		writeCSV(outFile, headers, table)
	default:
		w := os.Stdout
		if outFile != "" {
			f, err := os.Create(outFile)
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
		writeText(w, headers, table)
	}
}
