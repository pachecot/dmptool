package list

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/tpacheco/dmptool/dmp"
)

type expression interface {
	exp()
	match(*dmp.Object) bool
}

func (e kind) exp() {}

func (t token) match(do *dmp.Object) bool {
	return strings.Contains(do.Name, t.text) || strings.Contains(do.Path, t.text)
}

// inOp represents a binary operation
type inOp struct {
	kind
	lv    expression
	items []expression
}

func (op inOp) String() string {
	b := make([]byte, 0)
	b = fmt.Appendf(b, "%s IN (", op.lv)
	for i := range op.items {
		if i > 0 {
			b = fmt.Append(b, ", ")
		}
		b = fmt.Appendf(b, "%s", op.items[i])
	}
	b = fmt.Append(b, ")")
	return string(b)
}

func (op inOp) match(do *dmp.Object) bool {
	lv, ok := op.lv.(token)
	if !ok {
		return false
	}
	v, ok := do.Properties[lv.text]
	if !ok {
		return false
	}
	for i := range op.items {
		rv, ok := op.items[i].(token)
		if !ok {
			continue
		}
		if v == rv.text {
			return true
		}
	}
	return false
}

// binOp represents a binary operation
type binOp struct {
	kind
	lv expression
	rv expression
}

func (op binOp) String() string {
	return fmt.Sprintf("%s %s %s", op.lv, op.kind, op.rv)
}

func (op binOp) match(do *dmp.Object) bool {

	switch op.kind {

	case k_and:
		return op.lv.match(do) && op.rv.match(do)

	case k_or:
		return op.lv.match(do) || op.rv.match(do)

	case k_like:
		lv := op.lv.(token)
		rv := op.rv.(token)
		return isLike(do.Properties[lv.text], rv.text)

	default:
		lv := op.lv.(token)
		rv := op.rv.(token)

		switch rv.kind {

		case k_number:
			if rn, err := strconv.Atoi(rv.text); err == nil {
				return compareWithInt(do.Properties, lv.text, op.kind, rn)
			}
			return compareWith(do.Properties, lv.text, op.kind, rv.text)

		default:
			return compareWith(do.Properties, lv.text, op.kind, rv.text)

		}
	}
}

// uniOp represents a unary operations
type uniOp struct {
	kind
	rv expression
}

func (op uniOp) String() string {
	return fmt.Sprintf("%s %s", op.kind, op.rv)
}

func (op uniOp) match(do *dmp.Object) bool {
	switch op.kind {
	case k_not:
		return !op.rv.match(do)
	default:
		return false
	}
}

type errOp struct {
	kind
	offset int
}

func (op errOp) match(do *dmp.Object) bool {
	return false
}

func compareWith(m map[string]string, key string, op kind, v string) bool {
	p, ok := m[key]
	if !ok {
		return false
	}
	switch op {
	case k_eq:
		return p == v
	case k_ne:
		return p != v
	case k_gt:
		return p > v
	case k_lt:
		return p < v
	case k_ge:
		return p >= v
	case k_le:
		return p <= v
	}
	return false
}

func compareWithInt(m map[string]string, key string, op kind, v int) bool {
	s, ok := m[key]
	if !ok {
		return false
	}
	p, err := strconv.Atoi(s)
	if err != nil {
		switch op {
		case k_eq:
			return s[0] == '0'
		case k_ne:
			return s[0] != '0'
		case k_gt:
			return s[0] > '0'
		case k_lt:
			return s[0] < '0'
		case k_ge:
			return s[0] >= '0'
		case k_le:
			return s[0] <= '0'
		}
		return false
	}

	switch op {
	case k_eq:
		return p == v
	case k_ne:
		return p != v
	case k_gt:
		return p > v
	case k_lt:
		return p < v
	case k_ge:
		return p >= v
	case k_le:
		return p <= v
	}
	return false
}

func isLike(s string, v string) bool {
	if v == "%" {
		return true
	}
	s = strings.ToLower(s)
	v = strings.ToLower(v)
	if !strings.Contains(v, "%") {
		return s == v
	}
	pt := strings.Split(v, "%")

	if len(pt[0]) > 0 {
		if !strings.HasPrefix(s, pt[0]) {
			return false
		}
		s = s[len(pt[0]):]
	}
	pt = pt[1:]

	for len(pt) > 0 {
		if len(pt[0]) == 0 {
			if len(pt) == 1 {
				return true
			}
			pt = pt[1:]
			continue
		}
		x := strings.Index(s, pt[0])
		if x < 0 {
			return false
		}
		s = s[x+len(pt[0]):]
		pt = pt[1:]
	}
	return len(s) == 0
}
