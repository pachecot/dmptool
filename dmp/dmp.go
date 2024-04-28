// Package dmp is an event stream parser for continuum dmpfiles.
package dmp

import (
	"time"
)

const timeLayout = "1/2/2006 3:04:05 PM"

// ParseTime parses a dmpfile timestamp string and returns the time value it represents.
func ParseTime(value string) (time.Time, error) {
	return time.ParseInLocation(timeLayout, value, time.Local)
}
