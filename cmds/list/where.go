package list

import "strings"

func parseWhere(f string) (string, string, string) {
	i := strings.IndexAny(f, "!=<> ")
	if i < 0 {
		return f, "@", ""
	}
	k := strings.TrimSpace(f[:i])
	v := strings.Trim(f[i+1:], "!=<> ")
	if len(v) > 5 && strings.ToLower(v[:5]) == "like " {
		return k, "like", v[5:]
	}
	op := strings.TrimSpace(f[i : len(f)-len(v)])
	return k, op, v
}

func compareWith(op string, p string, v string) bool {
	switch op {
	case "=", "==":
		return p == v
	case "!=":
		return p != v
	case ">":
		return p > v
	case "<":
		return p < v
	case ">=":
		return p >= v
	case "<=":
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
