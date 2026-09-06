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

var ioTypes = []string{"InfinityInput", "InfinityOutput"}

var ioHeaders = []string{
	"DeviceId", "Name", "Alias", "Channel", "Type",
	"ElecType", "ElecScaleTop", "ElecScaleBot", "EngScaleTop", "EngScaleBot", "Conversion",
}

type ioHandler struct {
	dmp.EmptyHandler
	devices []string
	names   []string
	results []*dmp.Object
}

func (h *ioHandler) Object(obj *dmp.Object) {
	if !slices.Contains(ioTypes, obj.Type) {
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

type ioCommand struct {
	FileName string
	OutFile  string
	Devices  []string
	Names    []string
}

func (cmd *ioCommand) Execute() {
	r, err := os.Open(cmd.FileName)
	if err != nil {
		fmt.Printf("error opening file")
		return
	}
	table := ioReport(r, cmd.Devices, cmd.Names)
	outputTable(cmd.OutFile, "IO Report", ioHeaders, table)
}

func ioReport(r io.Reader, devices, names []string) [][]string {
	h := &ioHandler{
		devices: devices,
		names:   names,
	}
	dmp.Parse(r, h)

	slices.SortFunc(h.results, func(a, b *dmp.Object) int {
		if n := strings.Compare(a.DeviceId, b.DeviceId); n != 0 {
			return n
		}
		if n := strings.Compare(a.Type, b.Type); n != 0 {
			return n
		}
		if n := strings.Compare(a.Properties["Channel"], b.Properties["Channel"]); n != 0 {
			return n
		}
		return strings.Compare(a.Name, b.Name)
	})

	table := make([][]string, 0, len(h.results))
	for _, obj := range h.results {
		row := make([]string, len(ioHeaders))
		for i, field := range ioHeaders {
			switch field {
			case "DeviceId":
				row[i] = obj.DeviceId
			case "Type":
				row[i] = obj.Type
			case "Name":
				row[i] = obj.Name
			case "Alias":
				row[i] = obj.Alias
			default:
				row[i] = obj.Properties[field]
			}
		}
		table = append(table, row)
	}
	return table
}

func NewCmdIO() *cobra.Command {
	cmd := &ioCommand{}
	cc := &cobra.Command{
		Use:     "io <dump file>",
		Short:   "report IO points with electrical and engineering scale configuration",
		Aliases: []string{"inputs", "outputs"},
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
