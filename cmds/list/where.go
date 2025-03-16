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

	k_op_error

	k_text
	k_number
	k_pattern

	k_comma
	k_paren_left
	k_paren_right
	k_eq
	k_ne
	k_lt
	k_le
	k_gt
	k_ge

	k_like
	k_and
	k_or
	k_in
	k_is
	k_null
	k_where
	k_between
	k_select
	k_not
	k_order
	k_by
)

var (
	// opMap is a lookup for operator symbols
	opMap = map[string]kind{
		"=":  k_eq,
		"==": k_eq,
		"!=": k_ne,
		"<>": k_ne,
		"<":  k_lt,
		">":  k_gt,
		"<=": k_le,
		">=": k_ge,
	}

	// kwMap is a lookup for key words
	kwMap = map[string]kind{
		"like":    k_like,
		"not":     k_not,
		"and":     k_and,
		"or":      k_or,
		"in":      k_in,
		"where":   k_where,
		"select":  k_select,
		"between": k_between,
		"null":    k_null,
		"order":   k_order,
		"by":      k_by,
		"is":      k_by,
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
	case k_paren_left:
		return "("
	case k_paren_right:
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
	case k_comma:
		return ","
	case k_between:
		return "between"
	case k_null:
		return "null"
	case k_in:
		return "in"
	case k_is:
		return "is"
	case k_select:
		return "select"
	case k_where:
		return "where"
	case k_order:
		return "order"
	case k_by:
		return "by"
	default:
		panic(fmt.Sprintf("unexpected list.kind: %#v", k))
	}
}

type token struct {
	kind
	pos  int
	text string
}

func (t token) String() string {
	if t.text == "" {
		return t.kind.String()
	}
	return fmt.Sprintf("%s(%s)", t.kind, t.text)
}

func (t token) match(do *dmp.Object) bool {
	return strings.Contains(do.Name, t.text) || strings.Contains(do.Path, t.text)
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

func isSpace(c byte) bool {
	return ccMap[c] == ccSpace
}

func isOperator(c byte) bool {
	return ccMap[c] == ccOperator
}

func isWord(c byte) bool {
	switch ccMap[c] {
	case ccLowerCase, ccUpperCase, ccAlpha:
		return true
	default:
		return false
	}
}

func isDigit(c byte) bool {
	return ccMap[c] == ccDigit
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
	k, ok := opMap[lc]
	if !ok {
		return p, token{
			kind: k_op_error,
			text: s,
		}
	}
	tk := token{
		kind: k,
		text: s,
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
		return p, token{kind: k_unknown}
	}
	p++
	return p, token{
		kind: k_text,
		text: s,
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
	if kw, ok := kwMap[lc]; ok {
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
		text: s,
	}
}

func tokenize(s string) []token {
	tks := make([]token, 0)
	data := []byte(s)
	for i := 0; i < len(data); {

		switch ccMap[data[i]] {

		case ccParenOpen:
			i++
			tks = append(tks, token{pos: i, kind: k_paren_left})

		case ccParenClose:
			i++
			tks = append(tks, token{pos: i, kind: k_paren_right})

		case ccSpace:
			i += skipSpace(data[i:])

		case ccOperator:
			n, tk := readOperator(data[i:])
			tk.pos = i
			i += n
			tks = append(tks, tk)

		case ccQuot:
			n, tk := readQuote(data[i:])
			tk.pos = i
			i += n
			tks = append(tks, tk)

		default:
			n, tk := readWord(data[i:])
			tk.pos = i
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
		pos := tks[i].pos
		switch k := tks[i].kind; k {

		case k_text:
			last = tks[i]
			i++
			continue

		case k_paren_left:
			inParen = true
			var n int
			i++
			n, last = parse(tks[i:])
			i += n

		case k_paren_right:
			i++
			if !inParen {
				return i, last
			}
			if last == nil {
				return i, errOp{offset: pos}
			}
			inParen = false

		case k_like:
			i++
			if last == nil || i >= len(tks) {
				return i, errOp{offset: pos}
			}
			next := tks[i]
			i++
			switch next.kind {

			case k_text, k_pattern:
				last = binOp{
					kind: k,
					lv:   last,
					rv:   next,
				}

			default:
				return i, errOp{offset: pos}
			}

		case k_eq, k_ne, k_lt, k_le, k_gt, k_ge:
			i++
			if last == nil || i >= len(tks) {
				return i, errOp{offset: pos}
			}
			next := tks[i]
			i++
			switch next.kind {

			case k_text, k_number, k_null:
				last = binOp{
					kind: k,
					lv:   last,
					rv:   next,
				}

			default:
				return i, errOp{offset: pos}

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
			return i, errOp{offset: pos}
		}
	}
	return len(tks), last
}

func parseWhere(f string) expression {
	tks := tokenize(f)
	_, exp := parse(tks)
	if op, ok := exp.(errOp); ok {
		fmt.Printf("error parsing where at : %d\n", op.offset)
	}
	return exp
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

type ccType int

const (
	ccNA    ccType = 0
	ccSpace ccType = 1 << iota
	ccDigit
	ccAlpha
	ccLowerCase
	ccUpperCase
	ccOperator
	ccQuot
	ccParenOpen
	ccParenClose
	ccComma
	ccSymbol
)

var (
	ccMap = [256]ccType{

		' ':  ccSpace,
		'\n': ccSpace,
		'\r': ccSpace,
		'\t': ccSpace,

		'!': ccOperator,
		'<': ccOperator,
		'=': ccOperator,
		'>': ccOperator,

		'(': ccParenOpen,
		')': ccParenClose,

		',': ccComma,

		'%': ccSymbol,
		'.': ccAlpha,
		'_': ccAlpha,

		'"':  ccQuot,
		'\'': ccQuot,

		'0': ccDigit,
		'1': ccDigit,
		'2': ccDigit,
		'3': ccDigit,
		'4': ccDigit,
		'5': ccDigit,
		'6': ccDigit,
		'7': ccDigit,
		'8': ccDigit,
		'9': ccDigit,

		'A': ccUpperCase,
		'B': ccUpperCase,
		'C': ccUpperCase,
		'D': ccUpperCase,
		'E': ccUpperCase,
		'F': ccUpperCase,
		'G': ccUpperCase,
		'H': ccUpperCase,
		'I': ccUpperCase,
		'J': ccUpperCase,
		'K': ccUpperCase,
		'L': ccUpperCase,
		'M': ccUpperCase,
		'N': ccUpperCase,
		'O': ccUpperCase,
		'P': ccUpperCase,
		'Q': ccUpperCase,
		'R': ccUpperCase,
		'S': ccUpperCase,
		'T': ccUpperCase,
		'U': ccUpperCase,
		'V': ccUpperCase,
		'W': ccUpperCase,
		'X': ccUpperCase,
		'Y': ccUpperCase,
		'Z': ccUpperCase,

		'a': ccLowerCase,
		'b': ccLowerCase,
		'c': ccLowerCase,
		'd': ccLowerCase,
		'e': ccLowerCase,
		'f': ccLowerCase,
		'g': ccLowerCase,
		'h': ccLowerCase,
		'i': ccLowerCase,
		'j': ccLowerCase,
		'k': ccLowerCase,
		'l': ccLowerCase,
		'm': ccLowerCase,
		'n': ccLowerCase,
		'o': ccLowerCase,
		'p': ccLowerCase,
		'q': ccLowerCase,
		'r': ccLowerCase,
		's': ccLowerCase,
		't': ccLowerCase,
		'u': ccLowerCase,
		'v': ccLowerCase,
		'w': ccLowerCase,
		'x': ccLowerCase,
		'y': ccLowerCase,
		'z': ccLowerCase,
	}
)
