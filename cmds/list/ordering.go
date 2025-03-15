package list

import (
	"errors"
	"slices"
	"strconv"
	"strings"
)

type ordering struct {
	col  int
	name string
	desc bool
}

func (o ordering) compare(a, b []string) int {
	sign := 1
	if o.desc {
		sign = -1
	}
	s1, s2 := a[o.col], b[o.col]

	if n1, e1 := strconv.Atoi(s1); e1 == nil {
		if n2, e2 := strconv.Atoi(s2); e2 == nil {
			switch {
			case n1 == n2:
				return 0
			case n1 < n2:
				return -1 * sign
			case n1 > n2:
				return 1 * sign
			}
			return 0
		}
	}
	switch {
	case s1 == s2:
		return 0
	case s1 < s2:
		return -1 * sign
	case s1 > s2:
		return 1 * sign
	}
	return 0
}

var (
	errBadOrder = errors.New("bad order")
)

func reorder(orders []string, fields []string, table [][]string) error {

	lc_fields := make([]string, len(fields))
	for i := range fields {
		lc_fields[i] = strings.ToLower(fields[i])
	}
	os := make([]ordering, len(orders))
	for i := range orders {
		o := strings.ToLower(orders[i])
		do := strings.Split(strings.TrimSpace(o), " ")

		switch len(do) {
		case 1:
			os[i].name = do[0]
		case 2:
			os[i].name = do[1]
			switch do[0] {
			case "acs":
				os[i].desc = false
			case "desc":
				os[i].desc = true
			default:
				return errBadOrder
			}
		default:
			return errBadOrder
		}
		os[i].col = slices.Index(lc_fields, os[i].name)
		if os[i].col < 0 {
			return errBadOrder
		}
	}
	slices.SortFunc(table, func(a, b []string) int {
		for i := range os {
			r := os[i].compare(a, b)
			if r != 0 {
				return r
			}
		}
		return 0
	})
	return nil
}
