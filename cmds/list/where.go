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

type kind int

const (
	k_unknown kind = iota
	k_text
	k_number
	k_pattern
	k_paren_op
	k_paren_cl
	k_eq
	k_ne
	k_lt
	k_le
	k_gt
	k_ge
	k_like
	k_and
	k_or
	k_not
	k_op_error
)

var (
	km = map[string]kind{
		"=":    k_eq,
		"==":   k_eq,
		"!=":   k_ne,
		"<":    k_lt,
		">":    k_gt,
		"<=":   k_le,
		">=":   k_ge,
		"like": k_like,
		"not":  k_not,
		"and":  k_and,
		"or":   k_or,
	}
)

func (k kind) String() string {

	switch k {
	case k_unknown:
		return "unknown"
	case k_text:
		return "text"
	case k_number:
		return "number"
	case k_pattern:
		return "pattern"
	case k_paren_op:
		return "("
	case k_paren_cl:
		return ")"
	case k_op_error:
		return "op_error"
	case k_eq:
		return "="
	case k_ne:
		return "!="
	case k_lt:
		return "<"
	case k_le:
		return "<="
	case k_gt:
		return ">"
	case k_ge:
		return ">="
	case k_like:
		return "like"
	case k_not:
		return "not"
	case k_and:
		return "and"
	case k_or:
		return "or"
	}
	return "unknown"
}

type token struct {
	kind
	p string
}

func (t token) String() string {
	if t.p == "" {
		return t.kind.String()
	}
	return fmt.Sprintf("%s(%s)", t.kind, t.p)
}

func (t token) match(do *dmp.Object) bool {
	return strings.Contains(do.Name, t.p) || strings.Contains(do.Path, t.p)
}

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
		return isLike(do.Properties[lv.p], rv.p)
	default:
		lv := op.lv.(token)
		rv := op.rv.(token)
		switch rv.kind {
		case k_number:
			left, ok := do.Properties[lv.p]
			if !ok {
				return false
			}
			if ln, err := strconv.Atoi(left); err == nil {
				if rn, err := strconv.Atoi(rv.p); err == nil {
					return compareWithInt(op.kind, ln, rn)
				}
			}
			return compareWith(op.kind, do.Properties[lv.p], rv.p)
		default:
			return compareWith(op.kind, do.Properties[lv.p], rv.p)
		}
	}
}

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
}

func (op errOp) match(do *dmp.Object) bool {
	return false
}

func isSpace(c byte) bool {
	switch c {
	case ' ':
		return true
	default:
		return false
	}
}

func isOperator(c byte) bool {
	switch c {
	case '=', '!', '<', '>':
		return true
	default:
		return false
	}
}

func isWord(c byte) bool {
	if ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z') {
		return true
	}
	switch c {
	case '_', '.':
		return true
	default:
		return false
	}
}

func isDigit(c byte) bool {
	return '0' <= c && c <= '9'
}

func skipSpace(data []byte) int {
	p := 0
	for p < len(data) && isSpace(data[p]) {
		p++
	}
	return p
}

func readOperator(data []byte) (int, token) {
	p := 0
	for p < len(data) && isOperator(data[p]) {
		p++
	}
	s := string(data[:p])
	lc := strings.ToLower(s)
	k, ok := km[lc]
	if !ok {
		return p, token{
			kind: k_op_error,
			p:    s,
		}
	}
	tk := token{
		kind: k,
		p:    s,
	}
	return p, tk
}

func readQuote(data []byte) (int, token) {
	q := data[0]
	p := 1
	for ; p < len(data) && data[p] != q; p++ {
	}
	s := string(data[1:p])
	if p == len(data) {
		return p, token{kind: k_op_error}
	}
	p++
	return p, token{
		kind: k_text,
		p:    s,
	}
}

func readWord(data []byte) (int, token) {
	nc := 0
	k := k_text
	p := 0
	for ; p < len(data); p++ {
		if isWord(data[p]) {
			continue
		}
		if isDigit(data[p]) {
			nc++
			continue
		}
		if data[p] == '%' {
			k = k_pattern
			continue
		}
		break
	}
	s := string(data[:p])

	// test for key words
	lc := strings.ToLower(s)
	if kw, ok := km[lc]; ok {
		return p, token{
			kind: kw,
		}
	}

	// if all chars are digits then set to number
	if nc == p {
		k = k_number
	}

	return p, token{
		kind: k,
		p:    s,
	}
}

func tokenize(s string) []token {
	tks := make([]token, 0)
	data := []byte(s)
	for i := 0; i < len(data); {
		switch data[i] {
		case '(':
			i++
			tks = append(tks, token{kind: k_paren_op})
		case ')':
			i++
			tks = append(tks, token{kind: k_paren_cl})
		case ' ':
			i += skipSpace(data[i:])
		case '!', '=', '<', '>':
			n, tk := readOperator(data[i:])
			i += n
			tks = append(tks, tk)
		case '"', '\'':
			n, tk := readQuote(data[i:])
			i += n
			tks = append(tks, tk)
		default:
			n, tk := readWord(data[i:])
			if n == 0 {
				i++
				continue
			}
			i += n
			tks = append(tks, tk)
		}
	}
	return tks
}

func parse(tks []token) (int, expression) {
	var last expression
	inParen := false
	for i := 0; i < len(tks); {
		switch k := tks[i].kind; k {

		case k_text:
			last = tks[i]
			i++
			continue

		case k_paren_op:
			inParen = true
			var n int
			i++
			n, last = parse(tks[i:])
			i += n

		case k_paren_cl:
			i++
			if !inParen {
				return i, last
			}
			if last == nil {
				return i, errOp{kind: k_op_error}
			}
			inParen = false

		case k_like:
			i++
			if last == nil || i >= len(tks) {
				return i, errOp{kind: k_op_error}
			}
			next := tks[i]
			i++
			if next.kind != k_text && next.kind != k_pattern {
				return i, errOp{kind: k_op_error}
			}
			last = binOp{
				kind: k,
				lv:   last,
				rv:   next,
			}

		case k_eq, k_ne, k_lt, k_le, k_gt, k_ge:
			i++
			if last == nil || i >= len(tks) {
				return i, errOp{kind: k_op_error}
			}
			next := tks[i]
			i++
			if next.kind != k_text && next.kind != k_number {
				return i, errOp{kind: k_op_error}
			}
			last = binOp{
				kind: k,
				lv:   last,
				rv:   next,
			}

		case k_and, k_or:
			i++
			n, rest := parse(tks[i:])
			i += n
			return i, binOp{
				kind: k,
				lv:   last,
				rv:   rest,
			}

		case k_not:
			i++
			n, rest := parse(tks[i:])
			i += n
			return i, uniOp{
				kind: k,
				rv:   rest,
			}

		default:
			return i, errOp{kind: k_op_error}
		}
	}
	return len(tks), last
}

func parseWhere(f string) expression {
	tks := tokenize(f)
	n, e := parse(tks)
	if _, ok := e.(errOp); ok {
		fmt.Println("error parsing where at ", n)
	}
	if n < len(tks) {
		fmt.Println("error parsing where incomplete ", n, len(tks))
	}
	return e
}

func compareWith(op kind, p, v string) bool {

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

func compareWithInt(op kind, p, v int) bool {
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
