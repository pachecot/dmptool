package report

import (
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/pachecot/dmptool/dmp"
	"github.com/spf13/cobra"
)

var alarmEnrollmentHeaders = []string{
	"Name", "Alias", "Description", "EventType",
	"TimeDelay", "ACCExpression", "Limit", "Normal",
	"EventNotificationId", "MsgTextOffNormal", "MsgTextNormal",
}

type alarmEnrollmentsHandler struct {
	dmp.EmptyHandler
	devices []string
	names   []string
	results []*dmp.Object
}

func (h *alarmEnrollmentsHandler) Object(obj *dmp.Object) {
	if obj.Type != "AlarmEnrollment" {
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

type alarmEnrollmentsCommand struct {
	FileName string
	OutFile  string
	Devices  []string
	Names    []string
}

func (cmd *alarmEnrollmentsCommand) Execute() {
	r, err := os.Open(cmd.FileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file %s, %v", cmd.FileName, err)
		return
	}
	table := alarmEnrollmentsReport(r, cmd.Devices, cmd.Names)
	outputTable(cmd.OutFile, "Alarm Enrollments Report", alarmEnrollmentHeaders, table)
}

func alarmEnrollmentsReport(r io.Reader, devices, names []string) [][]string {
	h := &alarmEnrollmentsHandler{
		devices: devices,
		names:   names,
	}
	dmp.Parse(r, h)
	sort.Slice(h.results, func(i, j int) bool {
		return h.results[i].Name < h.results[j].Name
	})

	table := make([][]string, 0, len(h.results))
	for _, obj := range h.results {
		ep := dmp.ParseEventParameters(obj.Properties["EventParameters"])
		table = append(table, []string{
			obj.Name,
			obj.Alias,
			obj.Properties["Description"],
			obj.Properties["EventType"],
			ep.TimeDelay,
			ep.ACCExpression,
			ep.Limit,
			ep.Normal,
			obj.Properties["EventNotificationId"],
			obj.Properties["MsgTextOffNormal"],
			obj.Properties["MsgTextNormal"],
		})
	}

	return table
}

func NewCmdAlarmEnrollments() *cobra.Command {
	cmd := &alarmEnrollmentsCommand{}
	cc := &cobra.Command{
		Use:     "alarm-enrollments <dump file>",
		Short:   "report alarm enrollments with event parameters",
		Aliases: []string{"ae", "enrollments"},
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
