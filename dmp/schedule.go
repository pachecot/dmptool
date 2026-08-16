package dmp

import "strings"

type TimeValue struct {
	Time      string
	ValueType string
	Value     string
}

type DaySchedule struct {
	TimeValues []TimeValue
}

type ExceptionPeriod struct {
	IsRange   bool
	StartDate string
	EndDate   string
	Date      string
}

type ExceptionEntry struct {
	Period   ExceptionPeriod
	Schedule []TimeValue
	Priority string
}

type Schedule struct {
	Name         string
	Alias        string
	DeviceId     string
	Description  string
	EffStartDate string
	EffEndDate   string
	DefaultType  string
	DefaultValue string
	Weekly       [7]DaySchedule
	Exceptions   []ExceptionEntry
}

// cdtValue holds either a leaf string or a block of child items.
type cdtValue struct {
	str   string
	items []cdtItem
}

// cdtItem is a key-value pair where the value may be a leaf or nested block.
type cdtItem struct {
	key   string
	value cdtValue
}

// parseCDTItems parses CDT items from lines starting at *pos until a closing
// "}" or end of slice. Returns the items found.
func parseCDTItems(lines []string, pos *int) []cdtItem {
	var items []cdtItem
	for *pos < len(lines) {
		line := strings.TrimSpace(lines[*pos])
		*pos++
		switch line {
		case "", "EndOfCDT":
			continue
		case "}":
			return items
		case "{":
			children := parseCDTItems(lines, pos)
			items = append(items, cdtItem{value: cdtValue{items: children}})
			continue
		}
		k, v, ok := split(line)
		if !ok {
			continue
		}
		// peek: if the next non-empty line is "{", the value is a child block
		if *pos < len(lines) && strings.TrimSpace(lines[*pos]) == "{" {
			*pos++ // consume "{"
			children := parseCDTItems(lines, pos)
			items = append(items, cdtItem{key: k, value: cdtValue{items: children}})
		} else {
			items = append(items, cdtItem{key: k, value: cdtValue{str: v}})
		}
	}
	return items
}

// parseCDT parses a CDT block string (starting with "{") into cdtItems.
func parseCDT(s string) []cdtItem {
	lines := strings.Split(s, "\n")
	pos := 0
	// skip leading whitespace-only lines and consume the opening "{"
	for pos < len(lines) {
		if strings.TrimSpace(lines[pos]) == "{" {
			pos++
			break
		}
		pos++
	}
	return parseCDTItems(lines, &pos)
}

func extractTimeValue(items []cdtItem) TimeValue {
	var tv TimeValue
	for _, item := range items {
		switch item.key {
		case "time":
			tv.Time = item.value.str
		case "value":
			for _, vi := range item.value.items {
				switch vi.key {
				case "Type":
					tv.ValueType = vi.value.str
				case "Value":
					tv.Value = vi.value.str
				}
			}
		}
	}
	return tv
}

func extractTimeValues(items []cdtItem) []TimeValue {
	var tvs []TimeValue
	for _, item := range items {
		// anonymous block entries in a TimeValueList
		if item.key == "" {
			tv := extractTimeValue(item.value.items)
			if tv.Time != "" {
				tvs = append(tvs, tv)
			}
		}
	}
	return tvs
}

// ParseWeeklySchedule parses the WeeklySchedule CDT property into 7 DaySchedule
// entries (index 0=Monday … 6=Sunday).
func ParseWeeklySchedule(s string) [7]DaySchedule {
	var days [7]DaySchedule
	// top-level items are 7 anonymous blocks (one per day)
	dayItems := parseCDT(s)
	for i, dayItem := range dayItems {
		if i >= 7 {
			break
		}
		for _, inner := range dayItem.value.items {
			if inner.key == "TimeValueList" {
				days[i].TimeValues = extractTimeValues(inner.value.items)
				break
			}
		}
	}
	return days
}

func parseExceptionPeriod(items []cdtItem) ExceptionPeriod {
	// structure: Type: 260, Value: { Type: 261|11, Value: {startdate/enddate} | date }
	var ep ExceptionPeriod
	for _, item := range items {
		if item.key == "Value" && item.value.items != nil {
			var innerType, innerStr string
			var innerBlock []cdtItem
			for _, inner := range item.value.items {
				switch inner.key {
				case "Type":
					innerType = inner.value.str
				case "Value":
					if inner.value.items != nil {
						innerBlock = inner.value.items
					} else {
						innerStr = inner.value.str
					}
				}
			}
			switch innerType {
			case "261":
				ep.IsRange = true
				for _, di := range innerBlock {
					switch di.key {
					case "startdate":
						ep.StartDate = di.value.str
					case "enddate":
						ep.EndDate = di.value.str
					}
				}
			case "11":
				ep.Date = innerStr
			}
		}
	}
	return ep
}

func parseExceptionEntry(items []cdtItem) ExceptionEntry {
	var entry ExceptionEntry
	for _, item := range items {
		switch item.key {
		case "period":
			entry.Period = parseExceptionPeriod(item.value.items)
		case "exception":
			entry.Schedule = extractTimeValues(item.value.items)
		case "priority":
			entry.Priority = item.value.str
		}
	}
	return entry
}

// ParseExceptionSched parses the ExceptionSched CDT property into a slice of
// ExceptionEntry values.
func ParseExceptionSched(s string) []ExceptionEntry {
	var entries []ExceptionEntry
	for _, item := range parseCDT(s) {
		if item.key == "" && item.value.items != nil {
			entries = append(entries, parseExceptionEntry(item.value.items))
		}
	}
	return entries
}

// ParseEffectivePeriod parses the EffectivePeriod CDT property.
func ParseEffectivePeriod(s string) (string, string) {
	var start, end string
	for _, item := range parseCDT(s) {
		switch item.key {
		case "startdate":
			start = item.value.str
		case "enddate":
			end = item.value.str
		}
	}
	return start, end
}

// ParseScheduleDefault parses the ScheduleDefault CDT property.
func ParseScheduleDefault(s string) (string, string) {
	var typ, val string
	for _, item := range parseCDT(s) {
		switch item.key {
		case "Type":
			typ = item.value.str
		case "Value":
			val = item.value.str
		}
	}
	return typ, val
}

// ParseSchedule builds a Schedule from a Schedule-type Object.
func ParseSchedule(obj *Object) Schedule {
	sched := Schedule{
		Name:        obj.Name,
		Alias:       obj.Alias,
		DeviceId:    obj.DeviceId,
		Description: obj.Properties["Description"],
	}
	sched.EffStartDate, sched.EffEndDate = ParseEffectivePeriod(obj.Properties["EffectivePeriod"])
	sched.DefaultType, sched.DefaultValue = ParseScheduleDefault(obj.Properties["ScheduleDefault"])
	sched.Weekly = ParseWeeklySchedule(obj.Properties["WeeklySchedule"])
	sched.Exceptions = ParseExceptionSched(obj.Properties["ExceptionSched"])
	return sched
}
