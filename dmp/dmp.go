package dmp

import (
	"time"
)

const timeLayout = "1/2/2006 3:04:05 PM"

func ParseTime(value string) (time.Time, error) {
	return time.ParseInLocation(timeLayout, value, time.Local)
}
