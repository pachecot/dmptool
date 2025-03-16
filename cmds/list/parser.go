package list

import "fmt"

func parseList(tks []token) (int, []expression) {
	if len(tks) == 0 {
		return 0, nil
	}
	if tks[0].kind != k_paren_left {
		return 0, nil
	}
	list := make([]expression, 0)
	n := 1
	for n < len(tks) {
		switch tks[n].kind {
		case k_paren_right:
			n++
			return n, list
		case k_comma:
			n++
			continue
		default:
			m, item := parse(tks[n:])
			n += m
			list = append(list, item)
		}
	}
	return n, nil
}

func parse(tks []token) (int, expression) {
	var last expression
	inParen := false
	for i := 0; i < len(tks); {
		pos := tks[i].pos
		switch k := tks[i].kind; k {

		case k_comma:
			return i, last

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
			if !inParen {
				return i, last
			}
			i++
			if last == nil {
				return i, errOp{offset: pos}
			}
			inParen = false

		case k_in:
			i++
			if last == nil || i >= len(tks) {
				return i, errOp{offset: pos}
			}
			n, items := parseList(tks[i:])
			i += n
			last = inOp{
				lv:    last,
				items: items,
			}

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
	tks := scan(f)
	_, exp := parse(tks)
	if op, ok := exp.(errOp); ok {
		fmt.Printf("error parsing where at : %d\n", op.offset)
	}
	return exp
}
