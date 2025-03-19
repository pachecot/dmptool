package list

import (
	"fmt"
	"slices"
	"strings"
)

type kind int

const (
	k_unknown kind = iota
	k_parse_error
	k_op_error

	k_field
	k_string
	k_integer
	k_decimal
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

	k_minus
	k_plus

	k_like
	k_and
	k_or
	k_in
	k_is
	k_is_null
	k_is_not_null
	k_null
	k_where
	k_between
	k_select
	k_not
	k_order
	k_by
)

func (k kind) match(kinds ...kind) bool {
	return slices.Contains(kinds, k)
}

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
		"like":      k_like,
		"not":       k_not,
		"and":       k_and,
		"or":        k_or,
		"in":        k_in,
		"where":     k_where,
		"select":    k_select,
		"between":   k_between,
		"null":      k_null,
		"order":     k_order,
		"by":        k_by,
		"is":        k_is,
		"isnull":    k_is_null,
		"isnotnull": k_is_not_null,
	}
)

func (k kind) String() string {

	switch k {
	case k_unknown:
		return "unknown"
	case k_string:
		return "string"
	case k_field:
		return "field"
	case k_decimal:
		return "decimal"
	case k_integer:
		return "integer"
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
	case k_minus:
		return "-"
	case k_plus:
		return "+"
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
	case k_is_null:
		return "isnull"
	case k_is_not_null:
		return "isnotnull"
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

func scan(s string) []token {
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

		case ccComma:
			i++
			tks = append(tks, token{pos: i, kind: k_comma})

		case ccQuoteS:
			n, tk := readQuote(data[i:])
			tk.pos = i
			i += n
			tks = append(tks, tk)

		case ccDigit:
			n, tk := readNumber(data[i:])
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
		kind: k_string,
		text: s,
	}
}

func readDigits(data []byte) (int, token) {
	p := 0
	for ; p < len(data) && isDigit(data[p]); p++ {
	}
	return p, token{
		kind: k_integer,
		text: string(data[:p]),
	}
}

func readNumber(data []byte) (int, token) {
	k := k_integer
	p := 0
	if data[p] == '-' || data[p] == '+' {
		p++
	}
	n, _ := readDigits(data[p:])
	p += n
	if p == len(data) || isSpace(data[p]) {
		return p, token{kind: k, text: string(data[:p])}
	}
	// decimal float number
	if p < len(data) && data[p] == '.' {
		k = k_decimal
		p++
		n, _ := readDigits(data[p:])
		p += n
	}
	// exponential number
	if p < len(data) && (data[p] == 'e' || data[p] == 'E') {
		k = k_decimal
		p++
		if data[p] == '-' || data[p] == '+' {
			p++
		}
		n, _ := readDigits(data[p:])
		p += n
		if p < len(data) && data[p] == '.' {
			p++
			n, _ := readDigits(data[p:])
			p += n
		}
	}
	if p < len(data) {
		switch ccMap[data[p]] {
		case ccAlpha, ccLowerCase, ccUpperCase:
			return readWord(data)
		case ccSpace, ccOperator, ccComma, ccParenClose, ccSymbol:
			return p, token{
				kind: k,
				text: string(data[:p]),
			}
		default:
			return p, token{
				kind: k_parse_error,
				text: string(data[:p]),
			}
		}
	}
	return p, token{
		kind: k,
		text: string(data[:p]),
	}
}

func readWord(data []byte) (int, token) {
	nc := 0
	k := k_field
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
		k = k_integer
	}

	return p, token{
		kind: k,
		text: s,
	}
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
	ccQuoteS
	ccQuoteD
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

		'/': ccSymbol,
		'*': ccSymbol,
		'%': ccSymbol,
		'+': ccSymbol,
		'-': ccSymbol,
		'|': ccSymbol,

		'.': ccAlpha,
		'_': ccAlpha,
		'@': ccAlpha,
		'&': ccAlpha,

		'"':  ccQuoteD,
		'\'': ccQuoteS,

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
