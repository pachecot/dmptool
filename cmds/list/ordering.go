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

	s1, s2 := a[o.col], b[o.col]
	if s1 == s2 {
		return 0
	}
	sign := 1
	if o.desc {
		sign = -1
	}

	if n1, e1 := strconv.Atoi(s1); e1 == nil {
		if n2, e2 := strconv.Atoi(s2); e2 == nil {
			if n1 < n2 {
				return -sign
			}
			return sign
		}
	}

	xs1 := partitionDigits(s1)
	xs2 := partitionDigits(s2)

	for i := range xs1 {
		if i >= len(xs2) {
			return -sign
		}
		if xs1[i] == xs2[i] {
			continue
		}
		if isDigit(xs1[i][0]) && isDigit(xs2[i][0]) {
			n1, _ := strconv.Atoi(xs1[i])
			n2, _ := strconv.Atoi(xs2[i])
			if n1 < n2 {
				return -sign
			}
			return sign
		}
		if xs1[i] < xs2[i] {
			return -sign
		}
		return sign
	}
	if len(xs2) > len(xs1) {
		return -sign
	}
	return 0
}

var (
	errOrderbyParseFailed       = errors.New("orderby failed parse")
	errOrderbyUnrecognizedOrder = errors.New("orderby direction unknown")
	errOrderbyUnknownField      = errors.New("orderby unknown field")
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
				return errOrderbyUnrecognizedOrder
			}
		default:
			return errOrderbyParseFailed
		}
		os[i].col = slices.Index(lc_fields, os[i].name)
		if os[i].col < 0 {
			return errOrderbyUnknownField
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

func partitionDigits(s string) []string {
	r := make([]string, 0)
	if len(s) == 0 {
		return []string{""}
	}
	inDigit := isDigit(s[0])
	left := 0
	for i := range s {
		if isDigit(s[i]) {
			if !inDigit && i > 0 {
				r = append(r, s[left:i])
				left = i
				inDigit = true
			}
		} else {
			if inDigit {
				r = append(r, s[left:i])
				left = i
				inDigit = false
			}
		}
	}
	r = append(r, s[left:])
	return r
}
