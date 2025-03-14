package list

import (
	"strings"
)

type expression interface {
	exp()
}

func (e kind) exp() {}

type kind int

const (
	k_unknown = iota
	// k_field
	k_text
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

type binOp struct {
	kind
	lv expression
	rv expression
}

type uniOp struct {
	kind
	rv expression
}

type errOp struct {
	kind
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
	case '_', '.', '%':
		return true
	default:
		return false
	}
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

func readWord(data []byte) (int, token) {
	p := 0
	for p < len(data) && isWord(data[p]) {
		p++
	}
	s := string(data[:p])
	lc := strings.ToLower(s)
	// test for key words
	if kw, ok := km[lc]; ok {
		return p, token{
			kind: kw,
		}
	}
	return p, token{
		kind: k_text,
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

func parse(tks []token) expression {
	if len(tks) == 1 {
		return tks[0]
	}
	var i int
	var last expression
	for i < len(tks) {
		switch tks[i].kind {
		case k_text:
			last = tks[i]
			i++
			continue
		case k_paren_op:
			return parse(tks[i+1:])
		case k_paren_cl:
			if last != nil {
				return last
			}
		case k_like, k_eq, k_ne, k_lt, k_le, k_gt, k_ge:
			if last == nil || len(tks) < i {
				return errOp{kind: k_op_error}
			}
			last = binOp{
				kind: tks[i].kind,
				lv:   last,
				rv:   tks[i+1],
			}
			i += 2
		case k_and, k_or:
			return binOp{
				kind: tks[i].kind,
				lv:   last,
				rv:   parse(tks[i+1:]),
			}
		case k_not:
			return uniOp{
				kind: tks[i].kind,
				rv:   parse(tks[i+1:]),
			}
		case k_op_error:
			return errOp{kind: k_op_error}
		default:
			return errOp{kind: k_op_error}
		}
	}
	return last
}

func parseWhere(f string) expression {
	tks := tokenize(f)
	return parse(tks)
}

func compareWith(op kind, p string, v string) bool {
	// fmt.Printf("%s %s %s\n", p, op, v)
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
