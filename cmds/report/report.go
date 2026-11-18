package report

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xuri/excelize/v2"
)

const (
	csvExt  = ".csv"
	xlsxExt = ".xlsx"
)

type reportCmd struct {
	FileName string
	outfile  string
}

func NewCmdReport() *cobra.Command {

	cmd := &reportCmd{}
	cc := &cobra.Command{
		Use:   "report [subcommand] <dump file>",
		Short: "generate structured reports from a dump file",
		Long: `Generate structured reports from a Continuum dump file.

Available subcommands:
  io                - IO points (InfinityInput and InfinityOutput) with electrical and engineering scale info
  points            - Point summary (String, Numeric, Input, Output, DateTime) with value and setpoint
  controllers       - Controller summary
  alarms            - Points that have alarm links, expanded one row per alarm
  logs              - Points that have logging configured
  alarm-enrollments - AlarmEnrollment objects with parsed event parameters
  event-enrollments - EventEnrollment objects with parsed event parameters and object property reference
  schedules         - Schedule objects with weekly and exception time-value entries
`,
		Args: cobra.MinimumNArgs(1),
		Run: func(c *cobra.Command, args []string) {
			cmd.FileName = args[0]
			cmd.Execute()
		},
	}
	cc.AddCommand(
		NewCmdIO(),
		NewCmdPoints(),
		NewCmdControllers(),
		NewCmdAlarms(),
		NewCmdLogs(),
		NewCmdAlarmEnrollments(),
		NewCmdEventEnrollments(),
		NewCmdSchedules(),
	)
	cc.Flags().StringVarP(&cmd.outfile, "output", "o", "", "output file name (xlsx)")

	return cc
}

func (cmd *reportCmd) Execute() {
	r, err := os.Open(cmd.FileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open file %s, error: %v", cmd.FileName, err)
		return
	}

	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	r.Seek(0, 0)
	table := ioReport(r, nil, nil)
	if len(table) > 0 {
		writeXLSXSheet(f, "IO Points", ioHeaders, table)
	}

	r.Seek(0, 0)
	table = controllerReport(r, nil, nil, nil)
	if len(table) > 0 {
		writeXLSXSheet(f, "Controllers", controllerHeaders, table)
	}

	r.Seek(0, 0)
	table = pointsReport(r, nil, nil, nil)
	if len(table) > 0 {
		writeXLSXSheet(f, "Points", pointHeaders, table)
	}

	r.Seek(0, 0)
	table = alarmReport(r, nil, nil)
	if len(table) > 0 {
		writeXLSXSheet(f, "Alarms", alarmHeaders, table)
	}

	r.Seek(0, 0)
	table = alarmEnrollmentsReport(r, nil, nil)
	if len(table) > 0 {
		writeXLSXSheet(f, "Alarm Enrollemnts", alarmEnrollmentHeaders, table)
	}

	r.Seek(0, 0)
	table = eeoReport(r, nil, nil)
	if len(table) > 0 {
		writeXLSXSheet(f, "Event Enrollemnts", eventEnrollmentHeaders, table)
	}

	r.Seek(0, 0)
	table = schedulesReport(r, nil, nil)
	if len(table) > 0 {
		writeXLSXSheet(f, "Schedules", scheduleHeaders, table)
	}

	r.Seek(0, 0)
	table = eeoReport(r, nil, nil)
	if len(table) > 0 {
		writeXLSXSheet(f, "Event Enrollemnts", ioHeaders, table)
	}

	if cmd.outfile == "" {
		cmd.outfile = strings.TrimSuffix(cmd.FileName, filepath.Ext(cmd.FileName)) + xlsxExt
	}

	if err := f.SaveAs(cmd.outfile); err != nil {
		fmt.Println(err)
	}
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

func toXLSXRow(row []string) []interface{} {
	cells := make([]interface{}, len(row))
	for i, s := range row {
		if s != "" && s[0] >= '0' && s[0] <= '9' {
			if n, err := strconv.ParseInt(s, 10, 64); err == nil {
				cells[i] = n
			} else if n, err := strconv.ParseFloat(s, 64); err == nil {
				cells[i] = n
			} else {
				cells[i] = s
			}
		} else {
			cells[i] = s
		}
	}
	return cells
}

func writeXLSX(fName string, sName string, headers []string, table [][]string) {

	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	writeXLSXSheet(f, sName, headers, table)

	if err := f.SaveAs(fName); err != nil {
		fmt.Println(err)
	}
}

func writeXLSXSheet(f *excelize.File, sName string, headers []string, table [][]string) {

	ws := colWidths(headers, table)

	if id, _ := f.GetSheetIndex(sName); id == -1 {
		f.NewSheet(sName)
	}

	for i, width := range ws {
		f.SetColWidth(sName, fmt.Sprintf("%c", 'A'+i), fmt.Sprintf("%c", 'A'+i), float64(width))
	}

	f.SetSheetRow(sName, "A1", &headers)
	for i, row := range table {
		xlsxRow := toXLSXRow(row)
		f.SetSheetRow(sName, fmt.Sprintf("%c%d", 'A', i+2), &xlsxRow)
	}
}

func outputTable(outFile string, name string, headers []string, table [][]string) {
	switch strings.ToLower(path.Ext(outFile)) {
	case xlsxExt:
		writeXLSX(outFile, name, headers, table)
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
