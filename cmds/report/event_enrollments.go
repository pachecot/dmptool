package report

import (
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/pachecot/dmptool/dmp"
	"github.com/spf13/cobra"
)

var eventEnrollmentHeaders = []string{
	"Name", "Alias", "Description", "EventType", "TimeDelay",
	"ObjPropRef", "StateValues", "COVIncrement", "LowLimit", "HighLimit", "Deadband",
	"EventNotificationId", "MsgTextOffNormal", "MsgTextNormal",
}

type eventEnrollmentsHandler struct {
	dmp.EmptyHandler
	devices []string
	names   []string
	results []*dmp.Object
}

func (h *eventEnrollmentsHandler) Object(obj *dmp.Object) {
	if obj.Type != "EventEnrollment" {
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

type eventEnrollmentsCommand struct {
	FileName string
	OutFile  string
	Devices  []string
	Names    []string
}

func (cmd *eventEnrollmentsCommand) Execute() {
	r, err := os.Open(cmd.FileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file %s, %v", cmd.FileName, err)
		return
	}
	table := eeoReport(r, cmd.Devices, cmd.Names)
	outputTable(cmd.OutFile, "Event Enrollment Report", eventEnrollmentHeaders, table)
}

func eeoReport(r io.Reader, devices, names []string) [][]string {
	h := &eventEnrollmentsHandler{
		devices: devices,
		names:   names,
	}
	dmp.Parse(r, h)
	sort.Slice(h.results, func(i, j int) bool {
		return h.results[i].Name < h.results[j].Name
	})

	table := make([][]string, 0, len(h.results))
	for _, obj := range h.results {
		ep := dmp.ParseEEEventParameters(obj.Properties["EventParameters"])
		table = append(table, []string{
			obj.Name,
			obj.Alias,
			obj.Properties["Description"],
			obj.Properties["EventType"],
			ep.TimeDelay,
			dmp.ParseObjPropRef(obj.Properties["ObjPropRef"]),
			ep.StateValues,
			ep.COVIncrement,
			ep.LowLimit,
			ep.HighLimit,
			ep.Deadband,
			obj.Properties["EventNotificationId"],
			obj.Properties["MsgTextOffNormal"],
			obj.Properties["MsgTextNormal"],
		})
	}

	return table
}

func NewCmdEventEnrollments() *cobra.Command {
	cmd := &eventEnrollmentsCommand{}
	cc := &cobra.Command{
		Use:     "event-enrollments <dump file>",
		Short:   "report event enrollments with event parameters",
		Aliases: []string{"ee"},
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
