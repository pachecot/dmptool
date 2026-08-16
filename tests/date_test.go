package tests_test

import (
	"fmt"
	"testing"
	"time"
)

func TestDates(t *testing.T) {
	dates := []string{
		"4/6/2018 6:26:49 AM EST",
		"4/5/2018 3:23:42 PM EST",
		"4/5/2018 2:01:37 PM EST",
		"4/5/2018 2:04:59 PM EST",
		"4/5/2018 12:20:03 PM EST",
		"4/5/2018 12:02:00 PM EST",
		"4/5/2018 2:04:04 PM EST",
		"4/5/2018 2:04:00 PM EST",
		"4/5/2018 12:01:22 PM EST",
		"4/5/2018 12:01:12 PM EST",
		"4/5/2018 12:19:39 PM EST",
		"4/5/2018 12:00:12 PM EST",
		"4/5/2018 2:03:49 PM EST",
		"4/5/2018 11:57:54 AM EST",
		"4/5/2018 12:04:39 PM EST",
		"4/6/2018 6:41:42 AM EST",
		"4/6/2018 6:41:55 AM EST",
		"4/5/2018 11:53:11 AM EST",
		"4/5/2018 11:52:19 AM EST",
		"4/6/2018 6:43:12 AM EST",
		"4/5/2018 11:50:05 AM EST",
		"4/6/2018 6:34:54 AM EST",
		"4/6/2018 6:43:32 AM EST",
		"4/5/2018 11:48:35 AM EST",
		"4/5/2018 11:47:18 AM EST",
		"4/6/2018 6:41:32 AM EST",
		"4/5/2018 3:14:49 PM EST",
	}

	for _, d := range dates {
		tm, err := time.Parse("1/2/2006 3:04:05 PM MST", d)
		if err != nil {
			t.Errorf("error parsing %s, %v", d, err)
		}
		fmt.Println(d, tm)
	}
}
