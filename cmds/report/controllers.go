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

var controllerTypes = []string{
	"InfinityInfinetController",
}

var controllerHeaders = []string{"DeviceId", "Name", "Alias", "Port", "CtrlrType", "SerialNum", "VersionNum", "Description"}

type controllerHandler struct {
	dmp.EmptyHandler
	types   []string
	devices []string
	names   []string
	results []*dmp.Object
}

func (h *controllerHandler) Object(obj *dmp.Object) {
	allowed := controllerTypes
	if len(h.types) > 0 {
		allowed = h.types
	}
	if !slices.Contains(allowed, obj.Type) {
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

type controllerCommand struct {
	FileName string
	OutFile  string
	Types    []string
	Devices  []string
	Names    []string
}

func (cmd *controllerCommand) Execute() {
	r, err := os.Open(cmd.FileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file %s, %v", cmd.FileName, err)
		return
	}
	table := controllerReport(r, cmd.Types, cmd.Devices, cmd.Names)
	outputTable(cmd.OutFile, "Controller report", controllerHeaders, table)

}

func controllerReport(r io.Reader, types, devices, names []string) [][]string {
	h := &controllerHandler{
		types:   types,
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
		row := make([]string, len(controllerHeaders))
		for i, field := range controllerHeaders {
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

func NewCmdControllers() *cobra.Command {
	cmd := &controllerCommand{}
	cc := &cobra.Command{
		Use:     "controller <dump file>",
		Short:   "report controller summary",
		Aliases: []string{"summary"},
		Args:    cobra.MinimumNArgs(1),
		Run: func(c *cobra.Command, args []string) {
			cmd.FileName = args[0]
			cmd.Execute()
		},
	}
	cc.Flags().StringVarP(&cmd.OutFile, "output", "o", "", "output file (csv, xlsx, or text)")
	cc.Flags().StringSliceVarP(&cmd.Types, "types", "t", []string{}, "filter by point type (e.g. InfinityNumeric)")
	cc.Flags().StringSliceVarP(&cmd.Devices, "devices", "d", []string{}, "filter by device path")
	cc.Flags().StringSliceVarP(&cmd.Names, "names", "n", []string{}, "filter by point name")
	return cc
}
