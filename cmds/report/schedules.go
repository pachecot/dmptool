package report

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/pachecot/dmptool/dmp"
	"github.com/spf13/cobra"
)

var scheduleHeaders = []string{
	"Name", "Device", "Kind", "Day/Date", "Time", "Value", "ValueType", "Priority",
}

var dayNames = [7]string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}

type schedulesHandler struct {
	dmp.EmptyHandler
	devices []string
	names   []string
	results []*dmp.Object
}

func (h *schedulesHandler) Object(obj *dmp.Object) {
	if obj.Type != "Schedule" {
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

type schedulesCommand struct {
	FileName string
	OutFile  string
	Devices  []string
	Names    []string
}

func (cmd *schedulesCommand) Execute() {
	r, err := os.Open(cmd.FileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file %s, %v", cmd.FileName, err)
		return
	}
	table := schedulesReport(r, cmd.Devices, cmd.Names)
	outputTable(cmd.OutFile, "Alarm Enrollments Report", scheduleHeaders, table)
}

func schedulesReport(r io.Reader, devices, names []string) [][]string {
	h := &schedulesHandler{
		devices: devices,
		names:   names,
	}
	dmp.Parse(r, h)

	var rows []scheduleRow
	for _, obj := range h.results {
		sched := dmp.ParseSchedule(obj)
		for dayIdx, day := range sched.Weekly {
			dayName := dayNames[dayIdx]
			for _, tv := range day.TimeValues {
				rows = append(rows, newWeeklyRow(sched, dayIdx, dayName, tv))
			}
		}
		for _, ex := range sched.Exceptions {
			dayDate := exceptionPeriodLabel(ex.Period)
			if len(ex.Schedule) == 0 {
				rows = append(rows, newExceptionRow(sched, dayDate, "", "", "", ex.Priority))
				continue
			}
			for _, tv := range ex.Schedule {
				rows = append(rows, newExceptionRow(sched, dayDate, tv.Time, tv.Value, tv.ValueType, ex.Priority))
			}
		}
	}

	sortScheduleRows(rows)

	table := make([][]string, len(rows))
	for i, r := range rows {
		table[i] = r.cols
	}
	return table
}

func exceptionPeriodLabel(p dmp.ExceptionPeriod) string {
	if p.IsRange {
		return fmt.Sprintf("%s to %s", p.StartDate, p.EndDate)
	}
	return p.Date
}

// scheduleRow holds a table row with sort keys.
type scheduleRow struct {
	device  string
	name    string
	kindOrd int // 0=Weekly, 1=Exception
	dayOrd  int // day-of-week index for weekly; 7 for exceptions
	dayDate string
	timeOrd int // minutes since midnight
	cols    []string
}

// scheduleTimeMinutes converts "H:MM:SS AM/PM" to minutes since midnight for sorting.
func scheduleTimeMinutes(t string) int {
	if t == "" {
		return -1
	}
	var h, m, s int
	var ampm string
	fmt.Sscanf(t, "%d:%d:%d %s", &h, &m, &s, &ampm)
	ampm = strings.ToUpper(ampm)
	if ampm == "PM" && h != 12 {
		h += 12
	} else if ampm == "AM" && h == 12 {
		h = 0
	}
	return h*60 + m
}

func newWeeklyRow(sched dmp.Schedule, dayIdx int, dayName string, tv dmp.TimeValue) scheduleRow {
	return scheduleRow{
		device:  sched.DeviceId,
		name:    sched.Name,
		kindOrd: 0,
		dayOrd:  dayIdx,
		dayDate: dayName,
		timeOrd: scheduleTimeMinutes(tv.Time),
		cols:    []string{sched.Name, sched.DeviceId, "Weekly", dayName, tv.Time, tv.Value, tv.ValueType, ""},
	}
}

func newExceptionRow(sched dmp.Schedule, dayDate string, timeStr, val, valType, priority string) scheduleRow {
	return scheduleRow{
		device:  sched.DeviceId,
		name:    sched.Name,
		kindOrd: 1,
		dayOrd:  7,
		dayDate: dayDate,
		timeOrd: scheduleTimeMinutes(timeStr),
		cols:    []string{sched.Name, sched.DeviceId, "Exception", dayDate, timeStr, val, valType, priority},
	}
}

func sortScheduleRows(rows []scheduleRow) {
	sort.SliceStable(rows, func(i, j int) bool {
		a, b := rows[i], rows[j]
		if a.device != b.device {
			return a.device < b.device
		}
		if a.name != b.name {
			return a.name < b.name
		}
		if a.kindOrd != b.kindOrd {
			return a.kindOrd < b.kindOrd
		}
		if a.dayOrd != b.dayOrd {
			return a.dayOrd < b.dayOrd
		}
		if a.dayDate != b.dayDate {
			return a.dayDate < b.dayDate
		}
		return a.timeOrd < b.timeOrd
	})
}

func NewCmdSchedules() *cobra.Command {
	cmd := &schedulesCommand{}
	cc := &cobra.Command{
		Use:     "schedules <dump file>",
		Short:   "report schedule objects with weekly and exception entries",
		Aliases: []string{"sched", "schedule"},
		Args:    cobra.MinimumNArgs(1),
		Run: func(c *cobra.Command, args []string) {
			cmd.FileName = args[0]
			cmd.Execute()
		},
	}
	cc.Flags().StringVarP(&cmd.OutFile, "output", "o", "", "output file (csv, xlsx, or text)")
	cc.Flags().StringSliceVarP(&cmd.Devices, "devices", "d", []string{}, "filter by device path")
	cc.Flags().StringSliceVarP(&cmd.Names, "names", "n", []string{}, "filter by schedule name")
	return cc
}
