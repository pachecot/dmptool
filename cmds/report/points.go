package report

import (
	"slices"

	"github.com/pachecot/dmptool/dmp"
	"github.com/spf13/cobra"
)

var pointTypes = []string{
	"InfinityString",
	"InfinityNumeric",
	"InfinityInput",
	"InfinityOutput",
	"InfinityDateTime",
}

var pointHeaders = []string{"DeviceId", "Name", "Alias", "Type", "Value", "Setpoint"}

type pointsHandler struct {
	dmp.EmptyHandler
	types   []string
	devices []string
	names   []string
	results []*dmp.Object
}

func (h *pointsHandler) Object(obj *dmp.Object) {
	allowed := pointTypes
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

type pointsCommand struct {
	FileName string
	OutFile  string
	Types    []string
	Devices  []string
	Names    []string
}

func (cmd *pointsCommand) Execute() {
	h := &pointsHandler{
		types:   cmd.Types,
		devices: cmd.Devices,
		names:   cmd.Names,
	}
	dmp.ParseFile(cmd.FileName, h)

	table := make([][]string, 0, len(h.results))
	for _, obj := range h.results {
		row := make([]string, len(pointHeaders))
		for i, field := range pointHeaders {
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

	outputTable(cmd.OutFile, pointHeaders, table)
}

func NewCmdPoints() *cobra.Command {
	cmd := &pointsCommand{}
	cc := &cobra.Command{
		Use:     "points <dump file>",
		Short:   "report point summary (String, Numeric, Input, Output, DateTime) with value and setpoint",
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
