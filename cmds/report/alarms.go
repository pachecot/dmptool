package report

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/pachecot/dmptool/dmp"
	"github.com/spf13/cobra"
)

var alarmHeaders = []string{
	"Alarm", "Enabled", "DeviceId", "Name", "Alias", "Type",
	"Point1", "Point2", "Point3", "Point4",
}

type alarmsHandler struct {
	dmp.EmptyHandler
	devices []string
	names   []string
	results []*dmp.Object
}

func (h *alarmsHandler) Object(obj *dmp.Object) {
	if obj.Properties["AlarmLinks"] == "" {
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

type alarmsCommand struct {
	FileName string
	OutFile  string
	Devices  []string
	Names    []string
}

func (cmd *alarmsCommand) Execute() {
	r, err := os.Open(cmd.FileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file %s, %v", cmd.FileName, err)
		return
	}
	table := alarmReport(r, cmd.Devices, cmd.Names)

	outputTable(cmd.OutFile, "Alarm Report", alarmHeaders, table)
}

func alarmReport(r io.Reader, devices, names []string) [][]string {
	h := &alarmsHandler{
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
		links := dmp.ParseAlarmLinks(obj.Properties["AlarmLinks"])
		slices.SortFunc(links, func(a, b *dmp.AlarmLink) int {
			return strings.Compare(a.Path, b.Path)
		})
		for _, link := range links {
			if link == nil {
				continue
			}
			table = append(table, []string{
				link.Path,
				strconv.FormatBool(link.Enabled),
				obj.DeviceId,
				obj.Name,
				obj.Alias,
				obj.Type,
				obj.Properties["Point1"],
				obj.Properties["Point2"],
				obj.Properties["Point3"],
				obj.Properties["Point4"],
			})
		}
	}

	return table
}

func NewCmdAlarms() *cobra.Command {
	cmd := &alarmsCommand{}
	cc := &cobra.Command{
		Use:     "alarms <dump file>",
		Short:   "report points with alarms, one row per alarm link",
		Aliases: []string{"alarm"},
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
