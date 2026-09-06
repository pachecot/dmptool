package report

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/pachecot/dmptool/dmp"
	"github.com/spf13/cobra"
)

var logHeaders = []string{"DeviceId", "Name", "LogType", "Size", "LogIntervalSecs"}

type logsHandler struct {
	dmp.EmptyHandler
	devices []string
	names   []string
	results []*dmp.Object
}

func (h *logsHandler) Object(obj *dmp.Object) {
	if obj.Properties["LogType"] == "" {
		return
	}
	if !obj.MatchesDevices(h.devices) {
		return
	}
	if !obj.MatchesNames(h.names) {
		return
	}
	h.results = append(h.results, obj)
}

type logsCommand struct {
	FileName string
	OutFile  string
	Devices  []string
	Names    []string
}

func (cmd *logsCommand) Execute() {

	r, err := os.Open(cmd.FileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file %s, %v", cmd.FileName, err)
		return
	}
	table := logsReport(r, cmd.Devices, cmd.Names)
	outputTable(cmd.OutFile, "Log report", logHeaders, table)
}

func logsReport(r io.Reader, devices, names []string) [][]string {
	h := &logsHandler{
		devices: devices,
		names:   names,
	}
	dmp.Parse(r, h)

	slices.SortFunc(h.results, func(a, b *dmp.Object) int {
		if n := strings.Compare(a.DeviceId, b.DeviceId); n != 0 {
			return n
		}
		return strings.Compare(a.Name, b.Name)
	})

	table := make([][]string, 0, len(h.results))
	for _, obj := range h.results {
		row := make([]string, len(logHeaders))
		for i, field := range logHeaders {
			switch field {
			case "DeviceId":
				row[i] = obj.DeviceId
			case "Name":
				row[i] = obj.Name
			default:
				row[i] = obj.Properties[field]
			}
		}
		table = append(table, row)
	}
	return table
}

func NewCmdLogs() *cobra.Command {
	cmd := &logsCommand{}
	cc := &cobra.Command{
		Use:     "logs <dump file>",
		Short:   "report points with logging configured",
		Aliases: []string{"log"},
		Args:    cobra.MinimumNArgs(1),
		Run: func(c *cobra.Command, args []string) {
			cmd.FileName = args[0]
			cmd.Execute()
		},
	}
	cc.Flags().StringVarP(&cmd.OutFile, "output", "o", "", "output file (csv, xlsx, or text)")
	cc.Flags().StringSliceVarP(&cmd.Devices, "devices", "d", []string{}, "filter by device path")
	cc.Flags().StringSliceVarP(&cmd.Names, "names", "n", []string{}, "filter by point name")
	return cc
}
